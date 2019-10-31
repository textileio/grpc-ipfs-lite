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
	client                                 pb.IpfsLiteClient
	stringToAdd                            string = "hola"
	refFile                                *pb.Node
	refNode0, refNode1, refNode2, refNode3 *cbor.Node
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

	stream, err := client.AddFile(ctx)
	if err != nil {
		t.Fatalf("failed to AddFile: %v", err)
	}

	stream.Send(&pb.AddFileRequest{Payload: &pb.AddFileRequest_AddParams{AddParams: &pb.AddParams{}}})
	stream.Send(&pb.AddFileRequest{Payload: &pb.AddFileRequest_Chunk{Chunk: []byte(stringToAdd)}})
	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("failed to CloseAndRecv AddFile: %v", err)
	}
	refFile = resp.GetNode()
}

// func TestGetFile(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	resp, err := client.GetFile(ctx, &pb.GetFileRequest{Cid: refFile.Block.GetCid()})
// 	if err != nil {
// 		t.Fatalf("failed to GetFile: %v", err)
// 	}

// 	val := string(resp.GetData())
// 	if val != stringToAdd {
// 		t.Fatalf("wanted %s but got: %s", stringToAdd, val)
// 	}
// }

func TestHasBlock(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.HasBlock(ctx, &pb.HasBlockRequest{Cid: refFile.Block.GetCid()})
	if err != nil {
		t.Fatalf("failed to HasBlock: %v", err)
	}
	if !resp.GetHasBlock() {
		t.Fatal("should have found block but didn't")
	}
}

func TestAddNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	node0data := map[string]interface{}{
		"name": "node0",
	}
	node0, err := createNode(node0data)
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	block0 := pb.Block{
		Cid:     node0.Cid().String(),
		RawData: node0.RawData(),
	}

	_, err = client.AddNode(ctx, &pb.AddNodeRequest{Block: &block0})
	if err != nil {
		t.Fatalf("failed to add node: %v", err)
	}
	refNode0 = node0
}

func TestAddNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	node1Data := map[string]interface{}{
		"name": "node1",
		"link": refNode0.Cid(),
	}
	node1, err := createNode(node1Data)
	if err != nil {
		t.Fatalf("failed to create node0: %v", err)
	}
	block1 := pb.Block{
		Cid:     node1.Cid().String(),
		RawData: node1.RawData(),
	}

	node2Data := map[string]interface{}{
		"name": "node2",
		"link": node1.Cid(),
	}
	node2, err := createNode(node2Data)
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}
	block2 := pb.Block{
		Cid:     node2.Cid().String(),
		RawData: node2.RawData(),
	}

	node3Data := map[string]interface{}{
		"name": "node3",
		"link": node2.Cid(),
		"foo": map[string]interface{}{
			"bar": map[string]interface{}{
				"baz": "hi",
			},
			"what": "boom",
		},
	}
	node3, err := createNode(node3Data)
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}
	block3 := pb.Block{
		Cid:     node3.Cid().String(),
		RawData: node3.RawData(),
	}

	_, err = client.AddNodes(ctx, &pb.AddNodesRequest{Blocks: []*pb.Block{&block1, &block2, &block3}})
	if err != nil {
		t.Fatalf("failed to add node: %v", err)
	}
	refNode1 = node1
	refNode2 = node2
	refNode3 = node3
}

func TestGetNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetNode(ctx, &pb.GetNodeRequest{Cid: refNode0.Cid().String()})
	if err != nil {
		t.Fatalf("failed to GetNode: %v", err)
	}
	got := resp.GetNode().Block.GetCid()
	excpected := refNode0.Cid().String()
	if got != excpected {
		t.Fatalf("excpected cid %s but got: %s", excpected, got)
	}
}

func TestGetNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cids := []string{refNode0.Cid().String(), refNode1.Cid().String(), refNode2.Cid().String(), refNode3.Cid().String()}

	stream, err := client.GetNodes(ctx, &pb.GetNodesRequest{Cids: cids})
	if err != nil {
		t.Fatalf("failed to GetNodes: %v", err)
	}
	results := []*pb.Node{}
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
		results = append(results, resp.GetNode())
	}
	expected := len(cids)
	got := len(results)
	if got != expected {
		t.Fatalf("excpected %d results but got: %d", expected, got)
	}
}

func TestResolveLink(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.ResolveLink(ctx, &pb.ResolveLinkRequest{NodeCid: refNode3.Cid().String(), Path: []string{"link", "name"}})
	if err != nil {
		t.Fatalf("failed to ResolveLink: %v", err)
	}
	if len(resp.GetRemainingPath()) != 1 || resp.GetRemainingPath()[0] != "name" {
		t.Fatal("unexpected remaining path")
	}
	if resp.GetLink().GetCid() != refNode2.Cid().String() {
		t.Fatal("unexpected link cid")
	}
}

func TestResolve(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Resolve(ctx, &pb.ResolveRequest{NodeCid: refNode3.Cid().String(), Path: []string{"link", "name"}})
	if err != nil {
		t.Fatalf("failed to Resolve: %v", err)
	}
	if len(resp.GetRemainingPath()) != 1 || resp.GetRemainingPath()[0] != "name" {
		t.Fatal("unexpected remaining path")
	}
	// ToDo: something with resp.Object
}

func TestTree(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Tree(ctx, &pb.TreeRequest{NodeCid: refNode3.Cid().String(), Path: "", Depth: -1})
	if err != nil {
		t.Fatalf("failed to Tree: %v", err)
	}

	if len(resp.GetPaths()) != 6 {
		t.Fatal("unexpected number of tree paths")
	}
}

func TestRemoveNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.RemoveNode(ctx, &pb.RemoveNodeRequest{Cid: refNode3.Cid().String()})
	if err != nil {
		t.Fatalf("failed to RemoveNode: %v", err)
	}
}

func TestRemoveNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cids := []string{refNode0.Cid().String(), refNode0.Cid().String(), refNode1.Cid().String(), refNode2.Cid().String()}

	_, err := client.RemoveNodes(ctx, &pb.RemoveNodesRequest{Cids: cids})
	if err != nil {
		t.Fatalf("failed to RemoveNodes: %v", err)
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

func createNode(data map[string]interface{}) (*cbor.Node, error) {
	codec := uint64(multihash.SHA2_256)
	return cbor.WrapObject(data, codec, multihash.DefaultLengths[codec])
}
