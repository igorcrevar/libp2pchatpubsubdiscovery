package mychat

import (
	"context"
	"encoding/json"
	"log"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *ChatMessage

	ctx                context.Context
	gossipPubsub       *pubsub.PubSub
	pubsubTopic        *pubsub.Topic
	pubsubSubscription *pubsub.Subscription

	roomName string
	self     peer.ID
	nickName string
	logger   *log.Logger
}

// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
type ChatMessage struct {
	Message          string
	SenderID         string
	SenderNick       string
	ReceiverNickName string
}

// JoinChatRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func JoinChatRoom(ctx context.Context, hostP2P host.Host, config Config, logger *log.Logger) (*ChatRoom, error) {
	// create a new PubSub service using the GossipSub router
	gossipPubsub, err := pubsub.NewGossipSub(ctx, hostP2P)
	if err != nil {
		return nil, err
	}

	// join the pubsub topic
	topic, err := gossipPubsub.Join(topicName(config.RoomName))
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &ChatRoom{
		ctx:                ctx,
		gossipPubsub:       gossipPubsub,
		pubsubTopic:        topic,
		pubsubSubscription: sub,
		self:               hostP2P.ID(),
		nickName:           config.NickName,
		roomName:           config.RoomName,
		Messages:           make(chan *ChatMessage, ChatRoomBufSize),
	}

	// start reading messages from the subscription in a loop
	go cr.readLoop()
	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *ChatRoom) Publish(message string, to string) error {
	m := ChatMessage{
		Message:          message,
		SenderID:         cr.self.Pretty(),
		SenderNick:       cr.nickName,
		ReceiverNickName: to,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return cr.pubsubTopic.Publish(cr.ctx, msgBytes)
}

func (cr *ChatRoom) ListPeers() []peer.ID {
	return cr.gossipPubsub.ListPeers(topicName(cr.roomName))
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (cr *ChatRoom) readLoop() {
	for {
		msg, err := cr.pubsubSubscription.Next(cr.ctx)
		if err != nil {
			close(cr.Messages)
			cr.logger.Print("chat room readLoop error", err)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == cr.self {
			continue
		}
		cm := new(ChatMessage)
		err = json.Unmarshal(msg.Data, cm)

		if err == nil {
			// send valid messages onto the Messages channel
			if cm.ReceiverNickName == "" || cm.ReceiverNickName == cr.nickName {
				// message is ment to be read only by user with nickname specified in ReceiverNickName
				cr.Messages <- cm
			}
		} else {
			cr.logger.Print("chat room readLoop unmarshal error", err)
		}
	}
}

func topicName(roomName string) string {
	return "chat-room:" + roomName
}
