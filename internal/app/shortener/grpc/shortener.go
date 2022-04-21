package grpc

import (
	"context"
	"errors"
	"github.com/putalexey/go-practicum/internal/app/proto"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/putalexey/go-practicum/internal/app/urlgenerator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/url"
)

func NewGRPCShortener(
	ctx context.Context,
	store storage.Storager,
	urlGenerator urlgenerator.URLGenerator,
	batchDeleter storage.BatchDeleter,
) *ShortenerGRPCServer {
	s := grpc.NewServer()
	server := ShortenerGRPCServer{
		grpcServer:   s,
		ctx:          ctx,
		store:        store,
		urlGenerator: urlGenerator,
		batchDeleter: batchDeleter,
	}
	proto.RegisterShortenerServer(s, &server)
	return &server
}

type ShortenerGRPCServer struct {
	proto.UnimplementedShortenerServer
	ctx          context.Context
	grpcServer   *grpc.Server
	store        storage.Storager
	urlGenerator urlgenerator.URLGenerator
	batchDeleter storage.BatchDeleter
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

func (s *ShortenerGRPCServer) CreateShort(ctx context.Context, in *proto.CreateShortRequest) (*proto.CreateShortResponse, error) {
	responseStatus := proto.CreateShortResponse_Created
	record := storage.Record{
		Short:  s.urlGenerator.GenerateShort(in.Url),
		Full:   in.Url,
		UserID: in.UserId,
	}

	err := s.store.Store(ctx, record)
	if err != nil {
		var conflictError *storage.RecordConflictError
		if errors.As(err, &conflictError) {
			responseStatus = proto.CreateShortResponse_Conflict
			record = conflictError.OldRecord
		} else {
			log.Println("ERROR:", err)
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	}

	response := proto.CreateShortResponse{
		Status: responseStatus,
		Short: &proto.Short{
			ShortUrl:    s.urlGenerator.GetURL(record.Short),
			OriginalUrl: record.Full,
			UserId:      record.UserID,
		}}
	return &response, nil
}

func (s *ShortenerGRPCServer) CreateShortBatch(ctx context.Context, in *proto.CreateShortBatchRequest) (*proto.CreateShortBatchResponse, error) {
	results := make([]*proto.CreateShortBatchResponse_CreateShortBatchResponseItem, 0, 10)
	batchInserter := storage.NewBatchInserter(s.store, 10)
	for _, item := range in.Shorts {
		if !isValidURL(item.OriginalUrl) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid url: %s", item.OriginalUrl)
		}

		//r := storage.NewRecord(item.OriginalUrl, in.UserId)
		r := storage.Record{
			Short:  s.urlGenerator.GenerateShort(item.OriginalUrl),
			Full:   item.OriginalUrl,
			UserID: in.UserId,
		}

		if err2 := batchInserter.AddItem(ctx, r); err2 != nil {
			log.Println("ERROR:", err2)
			return nil, status.Errorf(codes.Internal, err2.Error())
		}

		responseItem := proto.CreateShortBatchResponse_CreateShortBatchResponseItem{
			CorrelationId: item.CorrelationId,
			Short: &proto.Short{
				ShortUrl:    s.urlGenerator.GetURL(r.Short),
				OriginalUrl: r.Full,
				UserId:      r.UserID,
			},
		}
		results = append(results, &responseItem)
	}

	if err := batchInserter.Flush(ctx); err != nil {
		log.Println("ERROR:", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	response := proto.CreateShortBatchResponse{Results: results}
	return &response, nil
}

func (s *ShortenerGRPCServer) GetShortsForCurrentUser(ctx context.Context, in *proto.ListShortsRequest) (*proto.ListShortsResponse, error) {
	records, err := s.store.LoadForUser(ctx, in.UserId)
	if err != nil {
		log.Println("ERROR:", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	responseItems := make([]*proto.Short, 0, len(records))

	for _, record := range records {
		responseItems = append(responseItems, &proto.Short{
			ShortUrl:    s.urlGenerator.GetURL(record.Short),
			OriginalUrl: record.Full,
			UserId:      record.UserID,
		})
	}
	response := proto.ListShortsResponse{
		Shorts: responseItems,
	}

	return &response, nil
}

func (s *ShortenerGRPCServer) DeleteUserShorts(_ context.Context, in *proto.DeleteShortBatchRequest) (*proto.DeleteShortBatchResponse, error) {
	s.batchDeleter.QueueItems(in.Shorts, in.UserId)
	return &proto.DeleteShortBatchResponse{}, nil
}

func (s *ShortenerGRPCServer) InternalStats(ctx context.Context, _ *proto.Empty) (*proto.InternalStatsResponse, error) {
	stats, err := s.store.GetStats(ctx)
	if err != nil {
		log.Println("ERROR:", err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &proto.InternalStatsResponse{
		UrlsCount:  int32(stats.URLsCount),
		UsersCount: int32(stats.UsersCount),
	}, nil
}

func isValidURL(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	return err == nil
}
