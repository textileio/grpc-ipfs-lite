package mobile

import (
	"context"

	"github.com/textileio/grpc-ipfs-lite/server"
	"github.com/textileio/grpc-ipfs-lite/util"
)

var (
	cancel context.CancelFunc
)

// Start starts the mobile ipfs-lite peer and gRPC server
func Start(datastorePath string, debug bool) error {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	lite, err := util.NewPeer(ctx, datastorePath, 4005, debug)
	if err != nil {
		return err
	}

	// TODO: run this in a goroutine, but need to get any error back out, but this blocks, so how do you know it is running with no error?
	go server.StartServer(lite, "localhost:10000")

	return nil
}

// Stop stops the embedded grpc and ipfs servers
func Stop() {
	server.StopServer()
	cancel()
}
