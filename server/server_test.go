package server

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/ipfs/go-log"
	"github.com/ipfs/go-merkledag"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/multiformats/go-multiaddr"
	multihash "github.com/multiformats/go-multihash"
	pb "github.com/textileio/grpc-ipfs-lite/ipfs-lite"
	"google.golang.org/grpc"
)

var (
	client                                                     pb.IpfsLiteClient
	stringToAdd                                                string = "hola"
	refFile, refLargeFile                                      *pb.Node
	refNode0, refNode1, refNode2, refNode3                     *cbor.Node
	refProtoNode0, refProtoNode1, refProtoNode2, refProtoNode3 *merkledag.ProtoNode
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

func TestGetFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.GetFile(ctx, &pb.GetFileRequest{Cid: refFile.Block.GetCid()})
	if err != nil {
		t.Fatalf("failed to GetFile: %v", err)
	}

	buffer := bytes.NewBuffer([]byte{})
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to receive file chunk: %v", err)
		}
		buffer.Write(resp.GetChunk())
	}

	val := string(buffer.Bytes())
	if val != stringToAdd {
		t.Fatalf("wanted %s but got: %s", stringToAdd, val)
	}
}

func TestAddLargeFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.AddFile(ctx)
	if err != nil {
		t.Fatalf("failed to AddLargeFile: %v", err)
	}

	file, err := os.Open("test-data/IMG_2293.jpeg")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	const BufferSize = 1024

	buffer := make([]byte, BufferSize)

	stream.Send(&pb.AddFileRequest{Payload: &pb.AddFileRequest_AddParams{AddParams: &pb.AddParams{}}})

	for {
		bytesread, err := file.Read(buffer)

		if err != nil {
			if err != io.EOF {
				t.Fatalf("failed while reading file: %v", err)
			}
			break
		}
		stream.Send(&pb.AddFileRequest{Payload: &pb.AddFileRequest_Chunk{Chunk: buffer[:bytesread]}})
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("failed to CloseAndRecv AddLargeFile: %v", err)
	}
	refLargeFile = resp.GetNode()
}

func TestGetLargeFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.GetFile(ctx, &pb.GetFileRequest{Cid: refLargeFile.Block.GetCid()})
	if err != nil {
		t.Fatalf("failed to GetFile: %v", err)
	}

	out, err := os.Create("/tmp/out.jpeg")
	defer out.Close()
	if err != nil {
		t.Fatalf("failed to create out file: %v", err)
	}

	buffer := bytes.NewBuffer([]byte{})
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to receive file chunk: %v", err)
		}
		_, err = out.Write(resp.GetChunk())
		if err != nil {
			t.Fatalf("failed to write chunk to file: %v", err)
		}

		_, err = buffer.Write(resp.GetChunk())
		if err != nil {
			t.Fatalf("failed to write chunk to buffer: %v", err)
		}
	}

	out.Sync()

	got := int32(len(buffer.Bytes()))
	expected := refLargeFile.GetStat().GetCumulativeSize() - refLargeFile.GetStat().GetLinksSize()
	if got != expected {
		t.Fatalf("wanted %d but got: %d", expected, got)
	}
}

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
		t.Fatalf("failed to add nodes: %v", err)
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

func TestAddProtoNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	node := merkledag.NodeWithData([]byte(stringToAdd))

	// Don't need to provide Cid for ProtoNode
	block := pb.Block{
		RawData: node.RawData(),
	}

	_, err := client.AddNode(ctx, &pb.AddNodeRequest{Block: &block})
	if err != nil {
		t.Fatalf("failed to add node: %v", err)
	}
	refProtoNode0 = node
}

func TestAddProtoNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	node1 := merkledag.NodeWithData([]byte(stringToAdd))
	node1.AddNodeLink("link", refProtoNode0)
	block1 := pb.Block{
		RawData: node1.RawData(),
	}
	node2 := merkledag.NodeWithData([]byte(stringToAdd))
	node2.AddNodeLink("link", node1)
	block2 := pb.Block{
		RawData: node2.RawData(),
	}
	node3 := merkledag.NodeWithData([]byte(stringToAdd))
	node3.AddNodeLink("link", node2)
	block3 := pb.Block{
		RawData: node3.RawData(),
	}
	_, err := client.AddNodes(ctx, &pb.AddNodesRequest{Blocks: []*pb.Block{&block1, &block2, &block3}})
	if err != nil {
		t.Fatalf("failed to add nodes: %v", err)
	}
	refProtoNode1 = node1
	refProtoNode2 = node2
	refProtoNode3 = node3
}

func TestGetProtoNode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetNode(ctx, &pb.GetNodeRequest{Cid: refProtoNode0.Cid().String()})
	if err != nil {
		t.Fatalf("failed to GetProtoNode: %v", err)
	}
	got := resp.GetNode().Block.GetCid()
	excpected := refProtoNode0.Cid().String()
	if got != excpected {
		t.Fatalf("excpected cid %s but got: %s", excpected, got)
	}
}

func TestGetProtoNodes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cids := []string{refProtoNode0.Cid().String(), refProtoNode1.Cid().String(), refProtoNode2.Cid().String(), refProtoNode3.Cid().String()}

	stream, err := client.GetNodes(ctx, &pb.GetNodesRequest{Cids: cids})
	if err != nil {
		t.Fatalf("failed to GetProtoNodes: %v", err)
	}
	results := []*pb.Node{}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to GetProtoNodes: %v", err)
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

	ds, err := ipfslite.BadgerDatastore("/tmp/test")
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
