package server

import (
	"context"
	"testing"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-log"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
	pb "textile.io/grpc-ipfs-lite/ipfs-lite"
)

var (
	client      pb.IpfsLiteClient
	stringToAdd string = "hola"
	node        pb.Node
)

func TestSetup(t *testing.T) {
	peer, err := newPeer()
	if err != nil {
		t.Fatalf("failed to create peer: %v", err)
	}

	go StartServer(peer, "localhost:10000")

	conn, err := grpc.Dial("localhost:10000", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to grpc dial: %v", err)
	}
	client = pb.NewIpfsLiteClient(conn)
}

func TestAddFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := pb.AddFileRequest{
		Data: []byte(stringToAdd),
	}
	resp, err := client.AddFile(ctx, &req)
	if err != nil {
		t.Fatalf("failed to AddFile: %v", err)
	}
	node = *resp.GetNode()
}

func TestGetFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetFile(ctx, &pb.GetFileRequest{Cid: node.Block.GetCid()})
	if err != nil {
		t.Fatalf("failed to GetFile: %v", err)
	}

	val := string(resp.GetData())
	if val != stringToAdd {
		t.Fatalf("wanted %s but got: %s", stringToAdd, val)
	}
}

func TestHashOnRead(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.HashOnRead(ctx, &pb.HashOnReadRequest{HashOnRead: true})
	if err != nil {
		t.Fatalf("failed to HashOnRead: %v", err)
	}
}

func newPeer() (*ipfslite.Peer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetLogLevel("*", "warn")

	ds, err := ipfslite.BadgerDatastore("test")
	if err != nil {
		return nil, err
	}
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		return nil, err
	}

	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{listen},
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
