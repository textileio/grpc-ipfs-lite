package peermanager

import ipfslite "github.com/hsanjuan/ipfs-lite"

// PeerManager provides a Peer instance and stops it when requested
type PeerManager interface {
	Peer() *ipfslite.Peer
	Stop() error
}
