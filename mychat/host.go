package mychat

import (
	"crypto/rand"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/mr-tron/base58"
)

const RSABits = 2048

func CreateHost(config Config) (host.Host, error) {
	var privateKey crypto.PrivKey
	var err error
	if config.PrivateKey == "" {
		// generate new private key
		privateKey, _, err = crypto.GenerateKeyPairWithReader(crypto.RSA, RSABits, rand.Reader)
		if err != nil {
			return nil, err
		}
	} else {
		// unmarshall private key from config
		pkBytes, err := base58.Decode(config.PrivateKey)
		if err != nil {
			return nil, err
		}
		privateKey, err = crypto.UnmarshalPrivateKey(pkBytes)
		if err != nil {
			return nil, err
		}
	}

	if config.OutputPrivateKey {
		bytes, err := crypto.MarshalPrivateKey(privateKey)
		if err == nil {
			// TODO: log error?
			fmt.Println("private key:", base58.Encode(bytes))
		}
	}

	// libp2p.New constructs a new libp2p Host. Other options can be added here.
	hostP2P, err := libp2p.New(
		libp2p.Identity(privateKey),
		libp2p.ListenAddrs(config.ListenMultiAdresses()...))
	if err != nil {
		return nil, err
	}
	return hostP2P, nil
}
