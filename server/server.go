package server

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
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
	err = grpcServer.Serve(lis)
	if err != nil {
		return err
	}
	return nil
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

	respBlock := pb.Block{
		RawData: node.RawData(),
		Cid:     node.Cid().String(),
	}

	respLinks := []*pb.Link{}
	for _, link := range node.Links() {
		respLink := pb.Link{
			Name: link.Name,
			Size: int64(link.Size),
			Cid:  link.Cid.String(),
		}
		respLinks = append(respLinks, &respLink)
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

	respNode := pb.Node{
		Block: &respBlock,
		Links: respLinks,
		Stat:  &respStat,
		Size:  int64(size),
	}

	return &pb.AddFileResponse{Node: &respNode}, nil
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

	buffer, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &pb.GetFileResponse{Data: buffer}, nil
}

func (s *ipfsLiteServer) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNode not implemented")
}

func (s *ipfsLiteServer) GetNodes(req *pb.GetNodesRequest, srv pb.IpfsLite_GetNodesServer) error {
	return status.Errorf(codes.Unimplemented, "method GetNodes not implemented")
}

func (s *ipfsLiteServer) ResolveLink(ctx context.Context, req *pb.ResolveLinkRequest) (*pb.ResolveLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResolveLink not implemented")
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
