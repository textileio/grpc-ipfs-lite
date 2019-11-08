package util

import (
	"context"
	"fmt"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-log"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
)

// NewPeer creates a new ipfslite.Peer
func NewPeer(ctx context.Context, datastorePath string, port int, debug bool) (*ipfslite.Peer, error) {
	logLevel := "WARNING"
	if debug {
		logLevel = "DEBUG"
	}
	if err := log.SetLogLevel("*", logLevel); err != nil {
		return nil, err
	}

	// Bootstrappers are using 1024 keys. See:
	// https://github.com/ipfs/infra/issues/378
	crypto.MinRsaKeyBits = 1024

	ds, err := ipfslite.BadgerDatastore(datastorePath)
	if err != nil {
		return nil, err
	}

	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, err
	}

	multiAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	listen, _ := multiaddr.NewMultiaddr(multiAddr)

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{listen},
		ipfslite.Libp2pOptionsExtra...,
	)

	if err != nil {
		return nil, err
	}

	lite, err := ipfslite.New(ctx, ds, h, dht, nil)
	if err != nil {
		return nil, err
	}

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())
	return lite, nil
}
