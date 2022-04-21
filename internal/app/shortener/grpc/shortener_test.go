package grpc

import (
	"context"
	"errors"
	"github.com/putalexey/go-practicum/internal/app/proto"
	"github.com/putalexey/go-practicum/internal/app/storage"
	storageTesting "github.com/putalexey/go-practicum/internal/app/storage/testing"
	"github.com/putalexey/go-practicum/internal/app/urlgenerator"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func makeStore(fs ...func(*storageTesting.MockStorager)) *storageTesting.MockStorager {
	m := new(storageTesting.MockStorager)
	for _, f := range fs {
		f(m)
	}
	return m
}

func makeURLGenerator() urlgenerator.URLGenerator {
	return &urlgenerator.SequenceGenerator{BaseURL: "http://localhost"}
}

func makeBatchDeleter(fs ...func(*storageTesting.MockBatchDeleter)) *storageTesting.MockBatchDeleter {
	m := new(storageTesting.MockBatchDeleter)
	for _, f := range fs {
		f(m)
	}
	return m
}

func TestNewGRPCShortener(t *testing.T) {
	ctx := context.Background()
	t.Run("creates grpc server", func(t *testing.T) {
		store := makeStore()
		urlGenerator := makeURLGenerator()
		batchDeleter := makeBatchDeleter()
		server := NewGRPCShortener(ctx, store, urlGenerator, batchDeleter)
		if server == nil {
			t.Error("NewGRPCShortener() not created")
		}
	})
}

func TestShortenerGRPCServer_CreateShort(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		store        storage.Storager
		urlGenerator urlgenerator.URLGenerator
		batchDeleter storage.BatchDeleter
	}
	type args struct {
		in *proto.CreateShortRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.CreateShortResponse
		wantErr bool
	}{
		{
			name: "creates shorts",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("Store", mock.Anything, mock.Anything).
						Return(nil)
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{in: &proto.CreateShortRequest{
				UserId: "test-user",
				Url:    "http://original.com",
			}},
			want: &proto.CreateShortResponse{
				Status: proto.CreateShortResponse_Created,
				Short: &proto.Short{
					ShortUrl:    "http://localhost/0",
					OriginalUrl: "http://original.com",
					UserId:      "test-user",
				},
			},
			wantErr: false,
		},
		{
			name: "return with status \"Conflict\" when record exists",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("Store", mock.Anything, mock.Anything).
						Return(storage.NewRecordConflictError(storage.Record{
							Short:   "0",
							Full:    "http://original.com",
							UserID:  "test-user",
							Deleted: false,
						}))
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{in: &proto.CreateShortRequest{
				UserId: "test-user",
				Url:    "http://original.com",
			}},
			want: &proto.CreateShortResponse{
				Status: proto.CreateShortResponse_Conflict,
				Short: &proto.Short{
					ShortUrl:    "http://localhost/0",
					OriginalUrl: "http://original.com",
					UserId:      "test-user",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShortenerGRPCServer{
				ctx:          ctx,
				store:        tt.fields.store,
				urlGenerator: tt.fields.urlGenerator,
				batchDeleter: tt.fields.batchDeleter,
			}
			got, err := s.CreateShort(ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateShort() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShortenerGRPCServer_CreateShortBatch(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		store        storage.Storager
		urlGenerator urlgenerator.URLGenerator
		batchDeleter storage.BatchDeleter
	}
	type args struct {
		in *proto.CreateShortBatchRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.CreateShortBatchResponse
		wantErr bool
	}{
		{
			name: "creates shorts",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("StoreBatch", mock.Anything, mock.Anything).
						Return(nil)
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{in: &proto.CreateShortBatchRequest{
				UserId: "test-user",
				Shorts: []*proto.CreateShortBatchRequest_CreateShortBatchItem{
					{CorrelationId: "c1", OriginalUrl: "http://original.com"},
					{CorrelationId: "c2", OriginalUrl: "http://original2.com"},
				},
			}},
			want: &proto.CreateShortBatchResponse{
				Results: []*proto.CreateShortBatchResponse_CreateShortBatchResponseItem{
					{
						CorrelationId: "c1",
						Short: &proto.Short{
							ShortUrl:    "http://localhost/0",
							OriginalUrl: "http://original.com",
							UserId:      "test-user",
						},
					},
					{
						CorrelationId: "c2",
						Short: &proto.Short{
							ShortUrl:    "http://localhost/1",
							OriginalUrl: "http://original2.com",
							UserId:      "test-user",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "returns error, when storage not available",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("StoreBatch", mock.Anything, mock.Anything).
						Return(errors.New("failed"))
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{in: &proto.CreateShortBatchRequest{
				UserId: "test-user",
				Shorts: []*proto.CreateShortBatchRequest_CreateShortBatchItem{
					{CorrelationId: "c1", OriginalUrl: "http://original.com"},
					{CorrelationId: "c2", OriginalUrl: "http://original2.com"},
				},
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns error, when trying to add invalid url",
			fields: fields{
				store:        makeStore(),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{in: &proto.CreateShortBatchRequest{
				UserId: "test-user",
				Shorts: []*proto.CreateShortBatchRequest_CreateShortBatchItem{
					{CorrelationId: "c1", OriginalUrl: "http//original.com"},
					{CorrelationId: "c2", OriginalUrl: "http://original2.com"},
				},
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns error, when storage not available and buffer of inserter is full",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("StoreBatch", mock.Anything, mock.Anything).
						Return(errors.New("failed"))
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{in: &proto.CreateShortBatchRequest{
				UserId: "test-user",
				Shorts: []*proto.CreateShortBatchRequest_CreateShortBatchItem{
					{CorrelationId: "c1", OriginalUrl: "http://original.com"},
					{CorrelationId: "c2", OriginalUrl: "http://original2.com"},
					{CorrelationId: "c3", OriginalUrl: "http://original3.com"},
					{CorrelationId: "c4", OriginalUrl: "http://original4.com"},
					{CorrelationId: "c5", OriginalUrl: "http://original5.com"},
					{CorrelationId: "c6", OriginalUrl: "http://original6.com"},
					{CorrelationId: "c7", OriginalUrl: "http://original7.com"},
					{CorrelationId: "c8", OriginalUrl: "http://original8.com"},
					{CorrelationId: "c9", OriginalUrl: "http://original9.com"},
					{CorrelationId: "c10", OriginalUrl: "http://original10.com"},
					// buffer size is 10, adding 11th element error will be thrown
					{CorrelationId: "c11", OriginalUrl: "http://original11.com"},
				},
			}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShortenerGRPCServer{
				ctx:          ctx,
				store:        tt.fields.store,
				urlGenerator: tt.fields.urlGenerator,
				batchDeleter: tt.fields.batchDeleter,
			}
			got, err := s.CreateShortBatch(ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShortBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateShortBatch() got = %v, want %v", got, tt.want)
			}
			tt.fields.store.(*storageTesting.MockStorager).AssertExpectations(t)
		})
	}
}

func TestShortenerGRPCServer_DeleteUserShorts(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		store        storage.Storager
		urlGenerator urlgenerator.URLGenerator
		batchDeleter storage.BatchDeleter
	}
	type args struct {
		in *proto.DeleteShortBatchRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.DeleteShortBatchResponse
		wantErr bool
	}{
		{
			name: "queues deleting",
			fields: fields{
				store:        makeStore(),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(func(m *storageTesting.MockBatchDeleter) {
					m.On("QueueItems", mock.Anything, mock.Anything).Return()
				}),
			},
			args: args{
				in: &proto.DeleteShortBatchRequest{
					UserId: "test-user",
					Shorts: []string{"1", "2", "3"},
				},
			},
			want:    &proto.DeleteShortBatchResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShortenerGRPCServer{
				ctx:          ctx,
				store:        tt.fields.store,
				urlGenerator: tt.fields.urlGenerator,
				batchDeleter: tt.fields.batchDeleter,
			}
			got, err := s.DeleteUserShorts(ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUserShorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteUserShorts() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShortenerGRPCServer_GetShortsForCurrentUser(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		store        storage.Storager
		urlGenerator urlgenerator.URLGenerator
		batchDeleter storage.BatchDeleter
	}
	type args struct {
		in *proto.ListShortsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.ListShortsResponse
		wantErr bool
	}{
		{
			name: "returns stats",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("LoadForUser", mock.Anything, "test-user").
						Return(
							[]storage.Record{{Short: "1", Full: "http://original.com/", UserID: "test-user", Deleted: false}},
							nil,
						)
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{
				in: &proto.ListShortsRequest{UserId: "test-user"},
			},
			want: &proto.ListShortsResponse{Shorts: []*proto.Short{
				{
					ShortUrl:    "http://localhost/1",
					OriginalUrl: "http://original.com/",
					UserId:      "test-user",
				},
			}},
			wantErr: false,
		},
		{
			name: "error on storage unavailable",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("LoadForUser", mock.Anything, mock.Anything).
						Return(
							nil,
							errors.New("failed"),
						)
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{
				in: &proto.ListShortsRequest{UserId: "test-user"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty array",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("LoadForUser", mock.Anything, mock.Anything).
						Return(
							[]storage.Record{},
							nil,
						)
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{
				in: &proto.ListShortsRequest{UserId: "test-user"},
			},
			want:    &proto.ListShortsResponse{Shorts: []*proto.Short{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShortenerGRPCServer{
				ctx:          ctx,
				store:        tt.fields.store,
				urlGenerator: tt.fields.urlGenerator,
				batchDeleter: tt.fields.batchDeleter,
			}
			got, err := s.GetShortsForCurrentUser(ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShortsForCurrentUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetShortsForCurrentUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShortenerGRPCServer_InternalStats(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		store        storage.Storager
		urlGenerator urlgenerator.URLGenerator
		batchDeleter storage.BatchDeleter
	}
	type args struct {
		in *proto.Empty
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.InternalStatsResponse
		wantErr bool
	}{
		{
			name: "returns stats",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("GetStats", mock.Anything).
						Return(&storage.ServiceStats{
							URLsCount:  2,
							UsersCount: 1,
						}, nil)
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{
				in: &proto.Empty{},
			},
			want:    &proto.InternalStatsResponse{UrlsCount: 2, UsersCount: 1},
			wantErr: false,
		},
		{
			name: "returns error, when storage not available",
			fields: fields{
				store: makeStore(func(m *storageTesting.MockStorager) {
					m.On("GetStats", mock.Anything, mock.Anything).
						Return(nil, errors.New("failed"))
				}),
				urlGenerator: makeURLGenerator(),
				batchDeleter: makeBatchDeleter(),
			},
			args: args{
				in: &proto.Empty{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShortenerGRPCServer{
				ctx:          ctx,
				store:        tt.fields.store,
				urlGenerator: tt.fields.urlGenerator,
				batchDeleter: tt.fields.batchDeleter,
			}
			got, err := s.InternalStats(ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("InternalStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InternalStats() got = %v, want %v", got, tt.want)
			}
			tt.fields.store.(*storageTesting.MockStorager).AssertExpectations(t)
			tt.fields.batchDeleter.(*storageTesting.MockBatchDeleter).AssertExpectations(t)
		})
	}
}

func Test_isValidURL(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "empty url is invalid", args: args{uri: ""}, want: false},
		{name: "true on valid url", args: args{uri: "http://example.com"}, want: true},
		{name: "false on urls with error", args: args{uri: "http//example.com"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURL(tt.args.uri); got != tt.want {
				t.Errorf("isValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
