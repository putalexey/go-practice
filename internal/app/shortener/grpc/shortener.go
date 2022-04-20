package grpc

import (
	"context"
	"github.com/putalexey/go-practicum/internal/app/proto"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/putalexey/go-practicum/internal/app/urlgenerator"
	"google.golang.org/grpc"
	"net"
)

func NewGRPCShortener(
	ctx context.Context,
	baseURL string,
	store storage.Storager,
	urlGenerator urlgenerator.URLGenerator,
	batchDeleter *storage.BatchDeleter,
) (*ShortenerGRPCServer, error) {
	s := grpc.NewServer()
	server := ShortenerGRPCServer{
		grpcServer: s,
		ctx:        ctx,
	}
	proto.RegisterShortenerServer(s, server)
	return &server, nil
}

type ShortenerGRPCServer struct {
	proto.UnimplementedShortenerServer
	ctx        context.Context
	grpcServer *grpc.Server
}

func (s *ShortenerGRPCServer) Serve() error {
	listen, err := net.Listen("tcp", ":3030")
	if err != nil {
		return err
	}

	go func() {
		<-s.ctx.Done()
		s.grpcServer.GracefulStop()
	}()
	return s.grpcServer.Serve(listen)
}
