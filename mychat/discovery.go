package mychat

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2pdiscovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type MyDiscovery struct {
	kademliaDHT *dht.IpfsDHT
	ctx         context.Context
	hostP2P     host.Host
	logger      *log.Logger
}

func NewMyDicovery(ctx context.Context, hostP2P host.Host, config Config, logger *log.Logger) (*MyDiscovery, error) {
	// var options []dht.Option
	// if asServer {
	// 	options = append(options, dht.Mode(dht.ModeServer))
	// }
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, hostP2P, dht.BootstrapPeersFunc(func() []peer.AddrInfo {
		boostrapPeerAddrs := make([]peer.AddrInfo, 0, len(config.BootstrapPeers))
		for _, x := range config.BootstrapPeers {
			peerInfo, err := peer.AddrInfoFromP2pAddr(x)
			if err == nil {
				boostrapPeerAddrs = append(boostrapPeerAddrs, *peerInfo)
			}
		}
		return boostrapPeerAddrs
	}))
	if err != nil {
		return nil, err
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	logger.Print("bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

	return &MyDiscovery{kademliaDHT: kademliaDHT, ctx: ctx, hostP2P: hostP2P, logger: logger}, nil
}

func (discovery MyDiscovery) AdvertiseAndFindPeers(config Config) {
	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	routingDiscovery := libp2pdiscovery.NewRoutingDiscovery(discovery.kademliaDHT)
	libp2pdiscovery.Advertise(discovery.ctx, routingDiscovery, config.RendezvousString)
	discovery.logger.Println("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	hostP2P := discovery.hostP2P
	for {
		peersChan, err := routingDiscovery.FindPeers(discovery.ctx, config.RendezvousString)
		if err != nil {
			discovery.logger.Print("find peers: ", err)
		}
		for peer := range peersChan {
			if peer.ID == hostP2P.ID() {
				continue
			}
			status := hostP2P.Network().Connectedness(peer.ID)
			if status == network.CanConnect || status == network.NotConnected {
				_, err = hostP2P.Network().DialPeer(discovery.ctx, peer.ID)
				if err != nil {
					hostP2P.Network().Peerstore().RemovePeer(peer.ID) // TODO: remove peer?
					discovery.logger.Print("error dialing finded peer: ", peer.ID, " ", err)
				} else {
					discovery.logger.Print("connected to peer: ", peer.ID)
				}
			}
		}
	}
}

// func (discovery MyDiscovery) ConnectToBoostrapNodes(config Config) (int, int) {
// 	// Let's connect to the bootstrap nodes first. They will tell us about the
// 	// other nodes in the network.
// 	successCount := 0
// 	var wg sync.WaitGroup
// 	for _, x := range config.BootstrapPeers {
// 		peerAddr := x
// 		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
// 		if err != nil {
// 			discovery.logger.Print("error while connecting to boostrap node, fail to convert peer address: ", peerAddr.String(), " ", err)
// 			continue
// 		}
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			if err := discovery.hostP2P.Connect(discovery.ctx, *peerInfo); err != nil {
// 				discovery.logger.Print("error while connecting to boostrap node: ", peerAddr.String(), " ", err)
// 			} else {
// 				successCount++
// 				discovery.logger.Print("connection established with bootstrap node: ", peerAddr.String())
// 			}
// 		}()
// 	}
// 	wg.Wait()
// 	return successCount, len(config.BootstrapPeers)
// }
