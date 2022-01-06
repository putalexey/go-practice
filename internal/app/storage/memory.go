package storage

import (
	"context"
	"fmt"
)

var _ Storager = &MemoryStorage{}

type MemoryStorage struct {
	records RecordMap
}

func NewMemoryStorage(records RecordMap) *MemoryStorage {
	return &MemoryStorage{records: records}
}

func (s *MemoryStorage) Store(_ context.Context, record Record) error {
	if s.records == nil {
		s.records = make(RecordMap)
	}
	s.records[record.Short] = record //Record{Short: short, Full: full, UserID: userID}
	return nil
}

func (s *MemoryStorage) Load(_ context.Context, short string) (Record, error) {
	if r, ok := s.records[short]; ok {
		return r, nil
	}
	return Record{}, fmt.Errorf("record \"%s\" not found", short)
}

func (s *MemoryStorage) LoadForUser(_ context.Context, userID string) ([]Record, error) {
	recordsList := make([]Record, 0)
	for _, record := range s.records {
		if record.UserID == userID {
			recordsList = append(recordsList, record)
		}
	}
	return recordsList, nil
}

func (s *MemoryStorage) Delete(_ context.Context, short string) error {
	if _, ok := s.records[short]; !ok {
		return fmt.Errorf("record \"%s\" not found", short)
	}
	delete(s.records, short)
	return nil
}

func (s *MemoryStorage) Ping(_ context.Context) error {
	return nil
}
