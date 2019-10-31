package server

import (
	"context"
	"fmt"
	"io"
	"net"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"google.golang.org/grpc"
	pb "textile.io/grpc-ipfs-lite/ipfs-lite"
)

type ipfsLiteServer struct {
	pb.UnimplementedIpfsLiteServer

	peer *ipfslite.Peer
}

const getFileChunkSize = 1024

// StartServer starts the gRPC server
func StartServer(peer *ipfslite.Peer, host string) error {
	lis, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	server := &ipfsLiteServer{
		peer: peer,
	}
	pb.RegisterIpfsLiteServer(grpcServer, server)
	return grpcServer.Serve(lis)
}

type addFileResult struct {
	Node format.Node
	Err  error
}

func addFile(ctx context.Context, peer *ipfslite.Peer, addParams *pb.AddParams, reader io.Reader, ch chan addFileResult) {
	defer close(ch)
	params := &ipfslite.AddParams{
		Layout:    addParams.GetLayout(),
		Chunker:   addParams.GetChunker(),
		RawLeaves: addParams.GetRawLeaves(),
		Hidden:    addParams.GetHidden(),
		Shard:     addParams.GetShared(),
		NoCopy:    addParams.GetNoCopy(),
		HashFun:   addParams.GetHashFun(),
	}
	node, err := peer.AddFile(ctx, reader, params)
	if err != nil {
		ch <- addFileResult{Err: err}
		return
	}
	ch <- addFileResult{Node: node}
}

func (s *ipfsLiteServer) AddFile(srv pb.IpfsLite_AddFileServer) error {
	req, err := srv.Recv()
	if err != nil {
		return err
	}
	var addParams *pb.AddParams
	switch payload := req.GetPayload().(type) {
	case *pb.AddFileRequest_AddParams:
		addParams = payload.AddParams
	default:
		return fmt.Errorf("expexted AddParams for AddFileRequest.Payload but got %T", payload)
	}

	reader, writer := io.Pipe()

	addFileChannel := make(chan addFileResult)
	go addFile(srv.Context(), s.peer, addParams, reader, addFileChannel)

	for {
		req, err := srv.Recv()
		if err == io.EOF {
			writer.Close()
			break
		} else if err != nil {
			writer.CloseWithError(err)
			break
		}
		switch payload := req.GetPayload().(type) {
		case *pb.AddFileRequest_Chunk:
			_, writeErr := writer.Write(payload.Chunk)
			if writeErr != nil {
				return writeErr
			}
		default:
			return fmt.Errorf("expected Chunk for AddFileRequest.Payload but got %T", payload)
		}
	}

	addFileResult := <-addFileChannel
	if addFileResult.Err != nil {
		return addFileResult.Err
	}

	respNode, err := nodeToPbNode(addFileResult.Node)
	if err != nil {
		return err
	}

	return srv.SendAndClose(&pb.AddFileResponse{Node: respNode})
}

func (s *ipfsLiteServer) GetFile(req *pb.GetFileRequest, srv pb.IpfsLite_GetFileServer) error {
	cid, err := cid.Decode(req.GetCid())
	if err != nil {
		return err
	}

	reader, err := s.peer.GetFile(srv.Context(), cid)
	if err != nil {
		return err
	}

	for {
		buffer := make([]byte, getFileChunkSize)
		size, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		srv.Send(&pb.GetFileResponse{Chunk: buffer[:size]})
		if err == io.EOF {
			return nil
		}
	}
}

func (s *ipfsLiteServer) HasBlock(ctx context.Context, req *pb.HasBlockRequest) (*pb.HasBlockResponse, error) {
	cid, err := cid.Decode(req.GetCid())
	if err != nil {
		return nil, err
	}
	hasBlock, err := s.peer.HasBlock(cid)
	if err != nil {
		return nil, err
	}
	return &pb.HasBlockResponse{HasBlock: hasBlock}, nil
}

func (s *ipfsLiteServer) AddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.AddNodeResponse, error) {
	cid, err := cid.Decode(req.Block.GetCid())
	if err != nil {
		return nil, err
	}

	// TODO: there is also just blocks.NewBlock()
	block, err := blocks.NewBlockWithCid(req.Block.GetRawData(), cid)
	if err != nil {
		return nil, err
	}

	node, err := format.Decode(block)
	if err != nil {
		return nil, err
	}

	err = s.peer.Add(ctx, node)
	if err != nil {
		return nil, err
	}

	return &pb.AddNodeResponse{}, nil
}

func (s *ipfsLiteServer) AddNodes(ctx context.Context, req *pb.AddNodesRequest) (*pb.AddNodesResponse, error) {
	nodes := make([]format.Node, len(req.GetBlocks()))

	for i, pbBlock := range req.GetBlocks() {
		cid, err := cid.Decode(pbBlock.GetCid())
		if err != nil {
			return nil, err
		}

		// TODO: there is also just blocks.NewBlock()
		block, err := blocks.NewBlockWithCid(pbBlock.GetRawData(), cid)
		if err != nil {
			return nil, err
		}

		node, err := format.Decode(block)
		if err != nil {
			return nil, err
		}

		nodes[i] = node
	}

	err := s.peer.AddMany(ctx, nodes)
	if err != nil {
		return nil, err
	}

	return &pb.AddNodesResponse{}, nil
}

func (s *ipfsLiteServer) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	cid, err := cid.Decode(req.GetCid())
	if err != nil {
		return nil, err
	}
	// TODO: use session() NodeGetter or Peer NodeGetter methods directly?
	node, err := s.peer.Get(ctx, cid)
	if err != nil {
		return nil, err
	}

	respNode, err := nodeToPbNode(node)
	if err != nil {
		return nil, err
	}

	return &pb.GetNodeResponse{Node: respNode}, nil
}

func (s *ipfsLiteServer) GetNodes(req *pb.GetNodesRequest, srv pb.IpfsLite_GetNodesServer) error {
	ctx := context.TODO()
	cids := make([]cid.Cid, len(req.GetCids()))
	for i, cidString := range req.GetCids() {
		cid, err := cid.Decode(cidString)
		if err != nil {
			return err
		}
		cids[i] = cid
	}
	// TODO: use session() NodeGetter or Peer NodeGetter methods directly?
	ch := s.peer.GetMany(ctx, cids)
	for {
		result, ok := <-ch
		if ok == false {
			break
		} else {
			resp := pb.GetNodesResponse{}
			if result.Err != nil {
				resp.Option = &pb.GetNodesResponse_Error{Error: result.Err.Error()}
			} else {
				node, err := nodeToPbNode(result.Node)
				if err != nil {
					resp.Option = &pb.GetNodesResponse_Error{Error: err.Error()}
				} else {
					resp.Option = &pb.GetNodesResponse_Node{Node: node}
				}
			}
			srv.Send(&resp)
		}
	}
	return nil
}

func (s *ipfsLiteServer) RemoveNode(ctx context.Context, req *pb.RemoveNodeRequest) (*pb.RemoveNodeResponse, error) {
	cid, err := cid.Decode(req.GetCid())
	if err != nil {
		return nil, err
	}
	err = s.peer.Remove(ctx, cid)
	if err != nil {
		return nil, err
	}
	return &pb.RemoveNodeResponse{}, nil
}

func (s *ipfsLiteServer) RemoveNodes(ctx context.Context, req *pb.RemoveNodesRequest) (*pb.RemoveNodesResponse, error) {
	cids := make([]cid.Cid, len(req.GetCids()))
	for i, reqCid := range req.GetCids() {
		cid, err := cid.Decode(reqCid)
		if err != nil {
			return nil, err
		}
		cids[i] = cid
	}
	err := s.peer.RemoveMany(ctx, cids)
	if err != nil {
		return nil, err
	}
	return &pb.RemoveNodesResponse{}, nil
}

func (s *ipfsLiteServer) ResolveLink(ctx context.Context, req *pb.ResolveLinkRequest) (*pb.ResolveLinkResponse, error) {
	cid, err := cid.Decode(req.GetNodeCid())
	if err != nil {
		return nil, err
	}

	node, err := s.peer.Get(ctx, cid)
	if err != nil {
		return nil, err
	}

	link, remainingPath, err := node.ResolveLink(req.GetPath())
	if err != nil {
		return nil, err
	}

	respLink := pb.Link{
		Name: link.Name,
		Size: int64(link.Size),
		Cid:  link.Cid.String(),
	}

	resp := pb.ResolveLinkResponse{
		Link:          &respLink,
		RemainingPath: remainingPath,
	}

	return &resp, nil
}

func (s *ipfsLiteServer) Resolve(ctx context.Context, req *pb.ResolveRequest) (*pb.ResolveResponse, error) {
	cid, err := cid.Decode(req.GetNodeCid())
	if err != nil {
		return nil, err
	}

	node, err := s.peer.Get(ctx, cid)
	if err != nil {
		return nil, err
	}

	// ToDo: Figure out how to use this _ interface{} data
	_, remainingPath, err := node.Resolve(req.GetPath())
	if err != nil {
		return nil, err
	}
	return &pb.ResolveResponse{RemainingPath: remainingPath}, nil
}

func (s *ipfsLiteServer) Tree(ctx context.Context, req *pb.TreeRequest) (*pb.TreeResponse, error) {
	cid, err := cid.Decode(req.GetNodeCid())
	if err != nil {
		return nil, err
	}

	node, err := s.peer.Get(ctx, cid)
	if err != nil {
		return nil, err
	}

	paths := node.Tree(req.GetPath(), int(req.GetDepth()))

	return &pb.TreeResponse{Paths: paths}, nil
}

func nodeToPbNode(node format.Node) (*pb.Node, error) {
	respBlock := pb.Block{
		RawData: node.RawData(),
		Cid:     node.Cid().String(),
	}

	respLinks := make([]*pb.Link, len(node.Links()))
	for i, link := range node.Links() {
		respLink := pb.Link{
			Name: link.Name,
			Size: int64(link.Size),
			Cid:  link.Cid.String(),
		}
		respLinks[i] = &respLink
	}

	stat, err := node.Stat()
	if err != nil {
		return nil, err
	}

	respStat := pb.NodeStat{
		Hash:           stat.Hash,
		NumLinks:       int32(stat.NumLinks),
		BlockSize:      int32(stat.BlockSize),
		LinksSize:      int32(stat.LinksSize),
		DataSize:       int32(stat.DataSize),
		CumulativeSize: int32(stat.CumulativeSize),
	}

	size, err := node.Size()
	if err != nil {
		return nil, err
	}

	return &pb.Node{
		Block: &respBlock,
		Links: respLinks,
		Stat:  &respStat,
		Size:  int64(size),
	}, nil
}
