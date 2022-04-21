package testing

import (
	"context"
	"github.com/putalexey/go-practicum/internal/app/storage"
	"github.com/stretchr/testify/mock"
)

var _ storage.BatchDeleter = &MockBatchDeleter{}
var _ storage.Storager = &MockStorager{}

type MockBatchDeleter struct {
	mock.Mock
}

func (d *MockBatchDeleter) QueueItems(shorts []string, userID string) {
	d.Called(shorts, userID)
}

func (d *MockBatchDeleter) Start() {
	d.Called()
}

type MockStorager struct {
	mock.Mock
}

func (m *MockStorager) Store(ctx context.Context, r storage.Record) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockStorager) StoreBatch(ctx context.Context, records []storage.Record) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

func (m *MockStorager) Load(ctx context.Context, short string) (storage.Record, error) {
	args := m.Called(ctx, short)
	v, _ := args.Get(0).(storage.Record)
	return v, args.Error(1)
}

func (m *MockStorager) LoadBatch(ctx context.Context, shorts []string) ([]storage.Record, error) {
	args := m.Called(ctx, shorts)
	v, _ := args.Get(0).([]storage.Record)
	return v, args.Error(1)
}

func (m *MockStorager) LoadForUser(ctx context.Context, userID string) ([]storage.Record, error) {
	args := m.Called(ctx, userID)
	v, _ := args.Get(0).([]storage.Record)
	return v, args.Error(1)
}

func (m *MockStorager) Delete(ctx context.Context, short string) error {
	args := m.Called(ctx, short)
	return args.Error(0)
}

func (m *MockStorager) DeleteBatch(ctx context.Context, shorts []string) error {
	args := m.Called(ctx, shorts)
	return args.Error(0)
}

func (m *MockStorager) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorager) GetStats(ctx context.Context) (*storage.ServiceStats, error) {
	args := m.Called(ctx)
	v, _ := args.Get(0).(*storage.ServiceStats)
	return v, args.Error(1)
}
