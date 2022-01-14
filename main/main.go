package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mychat "github.com/igorcrevar/libp2pchatpubsubdiscovery/mychat"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

var logger = log.New(os.Stdout,
	"TRACE: ",
	log.Ldate|log.Ltime|log.Lshortfile)

func readData(chatRoom *mychat.ChatRoom) {
	for msg := range chatRoom.Messages {
		fmt.Printf("New message: %v, %v\n", msg.SenderID, msg.SenderNick)
		fmt.Printf(msg.Message)
	}
}

func getMessageAndRecieverFromInput(input string) (string, string) {
	receiver := ""
	if pos := strings.Index(input, "--- "); pos == 0 {
		if pos2 := strings.Index(input[4:], " --- "); pos2 != -1 {
			receiver = input[4 : pos2+4]
			input = input[pos2+9:]
		}
	}
	return receiver, input
}

func writeData(chatRoom *mychat.ChatRoom, config mychat.Config) {
	toStrings := func(ls []peer.ID) []string {
		rs := make([]string, len(ls))
		for i, x := range ls {
			rs[i] = x.Pretty()
		}
		return rs
	}

	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s/%s $$: \n", config.RoomName, config.NickName)
		inputData, err := stdReader.ReadString('\n')
		if err != nil {
			logger.Print("Error reading from stdin: ", err)
			continue
		}

		receiver, msg := getMessageAndRecieverFromInput(inputData)

		if receiver == "listpeers" {
			fmt.Println(strings.Join(toStrings(chatRoom.ListPeers()), ", "))
		} else if msg != "" {
			err = chatRoom.Publish(msg, receiver)
			if err != nil {
				logger.Print("Error while publishing message: ", err)
			}
		}
	}
}

func main() {
	ctx := context.Background()
	config := mychat.ParseFlags()

	if config.Help {
		fmt.Println("This program demonstrates a simple p2p chat application using libp2p")
		fmt.Println()
		fmt.Println("Usage: Run './chat in two different terminals. Let them connect to the bootstrap nodes, announce themselves and connect to the peers")
		flag.PrintDefaults()
		return
	}

	hostP2P, err := mychat.CreateHost(*config)
	if err != nil {
		panic(err)
	}
	logger.Print("Host created. We are: ", hostP2P.ID())
	logger.Print("Nickname: ", config.NickName)
	logger.Print("We want to join room: ", config.RoomName)
	logger.Print(hostP2P.Addrs())

	hostP2P.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(network network.Network, conn network.Conn) {
			logger.Print("peer connected: ", conn.RemotePeer().Pretty())
		},
		DisconnectedF: func(network network.Network, conn network.Conn) {
			logger.Print("peer disconnected: ", conn.RemotePeer().Pretty())
		},
	})

	discovery, err := mychat.NewMyDicovery(ctx, hostP2P, *config, logger)
	if err != nil {
		panic(err)
	}

	go discovery.AdvertiseAndFindPeers(*config)

	chatRoom, err := mychat.JoinChatRoom(ctx, hostP2P, *config, logger)
	if err != nil {
		panic(err)
	}

	go readData(chatRoom)
	go writeData(chatRoom, *config)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	logger.Print("Received signal, shutting down...")

	// shut the node down
	if err := hostP2P.Close(); err != nil {
		panic(err)
	}
}
