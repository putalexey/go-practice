package storage

import (
	"context"
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

func (s *MemoryStorage) StoreBatch(_ context.Context, records []Record) error {
	if s.records == nil {
		s.records = make(RecordMap)
	}

	for _, record := range records {
		s.records[record.Short] = record
	}
	return nil
}

func (s *MemoryStorage) Load(_ context.Context, short string) (Record, error) {
	if r, ok := s.records[short]; ok {
		return r, nil
	}
	return Record{}, NewRecordNotFoundError(short)
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
		return NewRecordNotFoundError(short)
	}
	delete(s.records, short)
	return nil
}

func (s *MemoryStorage) DeleteBatchForUser(_ context.Context, shorts []string, userID string) error {
	// check all shorts exists
	for _, short := range shorts {
		v, ok := s.records[short]
		if !ok {
			return NewRecordNotFoundError(short)
		}
		if v.UserID != userID {
			return ErrAccessDenied
		}
	}
	// delete them
	for _, short := range shorts {
		delete(s.records, short)
	}
	return nil
}

func (s *MemoryStorage) Ping(_ context.Context) error {
	return nil
}
