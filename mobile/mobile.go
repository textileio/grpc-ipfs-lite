package mobile

import (
	"context"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-log"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/grpc-ipfs-lite/server"
)

// Start starts the mobile ipfs-lite peer and gRPC server
func Start(datastorePath string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetLogLevel("*", "warn")

	ds, err := ipfslite.BadgerDatastore(datastorePath)
	if err != nil {
		return err
	}
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return err
	}

	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{listen},
	)

	if err != nil {
		return err
	}

	lite, err := ipfslite.New(ctx, ds, h, dht, nil)
	if err != nil {
		return err
	}

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	// TODO: run this in a goroutine, but need to get any error back out, but this blocks, so how do you know it is running with no error?
	go server.StartServer(lite, "localhost:10000")

	return nil
}

// Stop stops the embedded grpc and ipfs servers
func Stop() {
	server.StopServer()
}
