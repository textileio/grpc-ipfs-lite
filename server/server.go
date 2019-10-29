package server

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "textile.io/grpc-ipfs-lite/ipfs-lite"
)

type ipfsLiteServer struct {
	pb.UnimplementedIpfsLiteServer

	// should this be a file var or here?
	peer *ipfslite.Peer
}

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

func (s *ipfsLiteServer) AddFile(ctx context.Context, req *pb.AddFileRequest) (*pb.AddFileResponse, error) {
	reader := bytes.NewReader(req.GetData())
	addParams := ipfslite.AddParams{
		Layout:    req.AddParams.GetLayout(),
		Chunker:   req.AddParams.GetChunker(),
		RawLeaves: req.AddParams.GetRawLeaves(),
		Hidden:    req.AddParams.GetHidden(),
		Shard:     req.AddParams.GetShared(),
		NoCopy:    req.AddParams.GetNoCopy(),
		HashFun:   req.AddParams.GetHashFun(),
	}
	node, err := s.peer.AddFile(ctx, reader, &addParams)
	if err != nil {
		return nil, err
	}

	respNode, err := nodeToPbNode(node)
	if err != nil {
		return nil, err
	}

	return &pb.AddFileResponse{Node: respNode}, nil
}

func (s *ipfsLiteServer) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	cid, err := cid.Decode(req.GetCid())
	if err != nil {
		return nil, err
	}

	reader, err := s.peer.GetFile(ctx, cid)
	if err != nil {
		return nil, err
	}

	// ToDo: don't read all the data into memory at once
	buffer, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &pb.GetFileResponse{Data: buffer}, nil
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
	return nil, status.Errorf(codes.Unimplemented, "method RemoveNode not implemented")
}

func (s *ipfsLiteServer) RemoveNodes(ctx context.Context, req *pb.RemoveNodesRequest) (*pb.RemoveNodesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveNodes not implemented")
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
	return nil, status.Errorf(codes.Unimplemented, "method Resolve not implemented")
}

func (s *ipfsLiteServer) Tree(ctx context.Context, req *pb.TreeRequest) (*pb.TreeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Tree not implemented")
}

func (s *ipfsLiteServer) DeleteBlock(ctx context.Context, req *pb.DeleteBlockRequest) (*pb.DeleteBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteBlock not implemented")
}

func (s *ipfsLiteServer) HasBlock(ctx context.Context, req *pb.HasBlockRequest) (*pb.HasBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HasBlock not implemented")
}

func (s *ipfsLiteServer) GetBlock(ctx context.Context, req *pb.GetBlockRequest) (*pb.GetBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}

func (s *ipfsLiteServer) GetBlockSize(ctx context.Context, req *pb.GetBlockSizeRequest) (*pb.GetBlockSizeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockSize not implemented")
}

func (s *ipfsLiteServer) PutBlock(ctx context.Context, req *pb.PutBlockRequest) (*pb.PutBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutBlock not implemented")
}

func (s *ipfsLiteServer) PutBlocks(ctx context.Context, req *pb.PutBlocksRequest) (*pb.PutBlocksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutBlocks not implemented")
}

func (s *ipfsLiteServer) AllKeys(req *pb.AllKeysRequest, srv pb.IpfsLite_AllKeysServer) error {
	return status.Errorf(codes.Unimplemented, "method AllKeys not implemented")
}

func (s *ipfsLiteServer) HashOnRead(ctx context.Context, req *pb.HashOnReadRequest) (*pb.HashOnReadResponse, error) {
	s.peer.BlockStore().HashOnRead(req.GetHashOnRead())
	return &pb.HashOnReadResponse{}, nil
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
