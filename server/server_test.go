package server

import (
	"context"
	"io"
	"testing"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/ipfs/go-log"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/multiformats/go-multiaddr"
	multihash "github.com/multiformats/go-multihash"
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

// func TestAddNode(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	// client.AddNode(ctx,)
// }

func TestGetNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetNode(ctx, &pb.GetNodeRequest{Cid: node.Block.GetCid()})
	if err != nil {
		t.Fatalf("failed to GetNode: %v", err)
	}
	got := resp.GetNode().Block.GetCid()
	excpected := node.Block.GetCid()
	if got != excpected {
		t.Fatalf("excpected cid %s but got: %s", excpected, got)
	}
}

func TestGetNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.GetNodes(ctx, &pb.GetNodesRequest{Cids: []string{node.Block.GetCid()}})
	if err != nil {
		t.Fatalf("failed to GetNodes: %v", err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to GetNodes: %v", err)
		}
		if resp.GetError() != "" {
			t.Fatalf("received error %s", resp.GetError())
		}
		got := resp.GetNode().Block.GetCid()
		excpected := node.Block.GetCid()
		if got != excpected {
			t.Fatalf("excpected cid %s but got: %s", excpected, got)
		}
	}
}

func TestResolveLink(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.ResolveLink(ctx, &pb.ResolveLinkRequest{NodeCid: node.Block.GetCid(), Path: []string{}})
	if err != nil {
		t.Fatalf("failed to ResolveLink: %v", err)
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

func TestCreateNode(t *testing.T) {
	m := map[string]string{
		"akey": "avalue",
	}

	codec := uint64(multihash.SHA2_256)

	node, err := cbor.WrapObject(m, codec, multihash.DefaultLengths[codec])
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	n := map[string]interface{}{
		"foo":  "bar",
		"link": node.Cid(),
	}

	node2, err := cbor.WrapObject(n, codec, multihash.DefaultLengths[codec])
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}

	t.Logf("final node: %v", node2)
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
