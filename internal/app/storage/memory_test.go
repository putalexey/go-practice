package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemoryStorage_Delete(t *testing.T) {
	tests := []struct {
		name    string
		records map[string]string
		short   string
		wantErr bool
	}{
		{
			name:    "deletes value from storage",
			records: map[string]string{"key1": "http://example.com"},
			short:   "key1",
			wantErr: false,
		},
		{
			name:    "not found in nil db",
			records: nil,
			short:   "key1",
			wantErr: true,
		},
		{
			name:    "returns error on not found",
			records: map[string]string{},
			short:   "key1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				records: tt.records,
			}

			if err := s.Delete(tt.short); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_Load(t *testing.T) {
	tests := []struct {
		name    string
		records map[string]string
		short   string
		want    string
		wantErr bool
	}{
		{
			name:    "returns full url",
			records: map[string]string{"key1": "http://example.com"},
			short:   "key1",
			want:    "http://example.com",
			wantErr: false,
		},
		{
			name:    "not found in nil db",
			records: nil,
			short:   "key1",
			wantErr: true,
		},
		{
			name:    "returns error when not found",
			records: map[string]string{},
			short:   "key1",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				records: tt.records,
			}
			got, err := s.Load(tt.short)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemoryStorage_Store(t *testing.T) {
	type args struct {
		short string
		full  string
	}
	tests := []struct {
		name    string
		records map[string]string
		args    args
		wantErr bool
	}{
		{
			name:    "stores value",
			records: map[string]string{},
			args: args{
				short: "key1",
				full:  "http://example.com",
			},
			wantErr: false,
		},
		{
			name:    "not panic if created with nil records",
			records: nil,
			args: args{
				short: "key1",
				full:  "http://example.com",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				records: tt.records,
			}
			if err := s.Store(tt.args.short, tt.args.full); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNewMemoryStorage(t *testing.T) {
	type args struct {
		records map[string]string
	}
	tests := []struct {
		name string
		args args
		want *MemoryStorage
	}{
		{
			name: "creates memory storage with filled records",
			args: args{records: map[string]string{"key1": "value"}},
			want: &MemoryStorage{records: map[string]string{"key1": "value"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewMemoryStorage(tt.args.records), "NewMemoryStorage(%v)", tt.args.records)
		})
	}
}
