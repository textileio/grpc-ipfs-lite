package mobile

import (
	"context"
	"fmt"

	"github.com/textileio/grpc-ipfs-lite/server"
	"github.com/textileio/grpc-ipfs-lite/util"
)

var (
	cancel context.CancelFunc
)

const ipfsPort = 4006
const grpcPort = 10001

// Start starts the mobile ipfs-lite peer and gRPC server
func Start(datastorePath string, debug bool) (int, error) {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	peerManager, err := util.NewPeerManager(ctx, datastorePath, ipfsPort, debug)
	if err != nil {
		return 0, err
	}

	host := fmt.Sprintf("localhost:%d", grpcPort)

	// TODO: run this in a goroutine, but need to get any error back out, but this blocks, so how do you know it is running with no error?
	go server.StartServer(peerManager, host)

	return grpcPort, nil
}

// Stop stops the embedded grpc and ipfs servers
func Stop() error {
	cancel()
	return server.StopServer()
}
