package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemoryStorage_Delete(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		records RecordMap
		short   string
		wantErr bool
	}{
		{
			name:    "deletes value from storage",
			records: RecordMap{"key1": {Short: "key1", Full: "http://example.com", UserID: "testUser"}},
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
			records: RecordMap{},
			short:   "key1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				records: tt.records,
			}

			if err := s.Delete(ctx, tt.short); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStorage_Load(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		records RecordMap
		short   string
		want    string
		wantErr bool
	}{
		{
			name:    "returns full url",
			records: RecordMap{"key1": {Short: "key1", Full: "http://example.com", UserID: "testUser"}},
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
			records: RecordMap{},
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
			got, err := s.Load(ctx, tt.short)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got.Full)
		})
	}
}

func TestMemoryStorage_Store(t *testing.T) {
	ctx := context.Background()
	type args struct {
		short  string
		full   string
		userID string
	}
	tests := []struct {
		name    string
		records RecordMap
		args    args
		wantErr bool
	}{
		{
			name:    "stores value",
			records: RecordMap{},
			args: args{
				short:  "key1",
				full:   "http://example.com",
				userID: "testUser",
			},
			wantErr: false,
		},
		{
			name:    "not panic if created with nil records",
			records: nil,
			args: args{
				short:  "key1",
				full:   "http://example.com",
				userID: "testUser",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemoryStorage{
				records: tt.records,
			}
			r, err := NewRecord(tt.args.full, tt.args.userID)
			require.NoError(t, err)

			if err := s.Store(ctx, r); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	t.Run("return user's shorts", func(t *testing.T) {
		store := &MemoryStorage{records: defaultRecords()}

		records, err := store.LoadForUser(ctx, "testUser")
		assert.NoError(t, err)
		assert.Len(t, records, 1)

		err = store.Store(ctx, Record{"test2", "http://example.com/testme2", "testUser", false})
		assert.NoError(t, err)

		records, err = store.LoadForUser(ctx, "testUser")
		assert.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("not return other user's shorts", func(t *testing.T) {
		store := &MemoryStorage{records: defaultRecords()}
		err := store.Store(ctx, Record{"test2", "http://example.com/testme2", "testUser2", false})
		assert.NoError(t, err)
		err = store.Store(ctx, Record{"test3", "http://example.com/testme3", "testUser2", false})
		assert.NoError(t, err)

		records, err := store.LoadForUser(ctx, "testUser")
		assert.NoError(t, err)
		assert.Len(t, records, 1)

		records, err = store.LoadForUser(ctx, "testUser2")
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("return empty list when user doesn't have shorts", func(t *testing.T) {
		store := &MemoryStorage{records: defaultRecords()}

		records, err := store.LoadForUser(ctx, "testUser_not_exists")
		assert.NoError(t, err)
		assert.NotNil(t, records)
		assert.Len(t, records, 0)
	})

}

func defaultRecords() RecordMap {
	return RecordMap{
		"test": Record{
			Short:  "test",
			Full:   "http://example.com/testme",
			UserID: "testUser",
		},
	}
}

func TestNewMemoryStorage(t *testing.T) {
	type args struct {
		records RecordMap
	}
	tests := []struct {
		name string
		args args
		want *MemoryStorage
	}{
		{
			name: "creates memory storage with filled records",
			args: args{records: RecordMap{"key1": {Short: "key1", Full: "value", UserID: "testUser"}}},
			want: &MemoryStorage{records: RecordMap{"key1": {Short: "key1", Full: "value", UserID: "testUser"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewMemoryStorage(tt.args.records), "NewMemoryStorage(%v)", tt.args.records)
		})
	}
}
