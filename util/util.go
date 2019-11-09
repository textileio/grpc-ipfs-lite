package util

import (
	"context"
	"fmt"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-log"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/textileio/grpc-ipfs-lite/peermanager"
)

type manager struct {
	datastore datastore.Batching
	host      host.Host
	dht       *dht.IpfsDHT
	peer      *ipfslite.Peer
}

func (m manager) Peer() *ipfslite.Peer {
	return m.peer
}

func (m manager) Stop() error {
	var lastError error
	if err := m.datastore.Close(); err != nil {
		lastError = err
	}
	if err := m.host.Close(); err != nil {
		lastError = err
	}
	if err := m.dht.Close(); err != nil {
		lastError = err
	}
	return lastError
}

// NewPeerManager creates a new server.PeerManager
func NewPeerManager(ctx context.Context, datastorePath string, port int, debug bool) (peermanager.PeerManager, error) {
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

	var peerManager = manager{}

	var err error
	peerManager.datastore, err = ipfslite.BadgerDatastore(datastorePath)
	if err != nil {
		return nil, err
	}

	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, err
	}

	multiAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)
	listen, _ := multiaddr.NewMultiaddr(multiAddr)

	peerManager.host, peerManager.dht, err = ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{listen},
		ipfslite.Libp2pOptionsExtra...,
	)

	if err != nil {
		return nil, err
	}

	peerManager.peer, err = ipfslite.New(ctx, peerManager.datastore, peerManager.host, peerManager.dht, nil)
	if err != nil {
		return nil, err
	}

	peerManager.peer.Bootstrap(ipfslite.DefaultBootstrapPeers())

	return peerManager, nil
}
