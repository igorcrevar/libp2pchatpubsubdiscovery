package mychat

import (
	"flag"
	"fmt"
	"strings"
	"time"

	maddr "github.com/multiformats/go-multiaddr"
)

// A new type we need for writing a custom flag parser
type addrList []maddr.Multiaddr

type Config struct {
	RendezvousString  string
	BootstrapPeers    addrList
	ListenAddresses   addrList
	NickName          string
	RoomName          string
	FindPeersTimeSecs int
	Port              int
	PrivateKey        string
	OutputPrivateKey  bool
	Help              bool
}

func ParseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.RendezvousString, "rendezvous", "meet me here",
		"Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "listen", "Adds a multiaddress to the listen list")
	flag.IntVar(&config.Port, "port", 35005, "port for communication if listen addresses not specified")
	flag.IntVar(&config.FindPeersTimeSecs, "findpeertime", 10, "tick time for find peers thread")
	flag.StringVar(&config.NickName, "nick", "", "nickname to use in chat. will be generated if empty")
	flag.StringVar(&config.RoomName, "room", "crew-chat-room", "name of chat room to join")
	flag.StringVar(&config.PrivateKey, "pk", "", "serialized private key used by this host")
	flag.BoolVar(&config.OutputPrivateKey, "opk", false, "print serialized private key in console")
	flag.BoolVar(&config.Help, "h", false, "displays help")
	flag.Parse()
	if config.ListenAddresses == nil {
		config.ListenAddresses, _ = StringsToAddrs([]string{fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.Port)})
	}
	if config.NickName == "" {
		config.NickName = GenerateRandString(14)
	}

	return config
}

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

func StringsToAddrs(addrStrings []string) (maddrs []maddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}

func (config Config) ListenMultiAdresses() []maddr.Multiaddr {
	return []maddr.Multiaddr(config.ListenAddresses)
}

func (config Config) BoostrapPeersMultiAdresses() []maddr.Multiaddr {
	return []maddr.Multiaddr(config.BootstrapPeers)
}

func (config Config) FindPeersTimeSecsDuration() time.Duration {
	return time.Duration(config.FindPeersTimeSecs * int(time.Second))
}
