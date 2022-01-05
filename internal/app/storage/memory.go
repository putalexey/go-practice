package storage

import (
	"fmt"
)

var _ Storager = &MemoryStorage{}

type MemoryStorage struct {
	records RecordMap
}

func NewMemoryStorage(records RecordMap) *MemoryStorage {
	return &MemoryStorage{records: records}
}

func (s *MemoryStorage) Store(short, full, userID string) error {
	if s.records == nil {
		s.records = make(RecordMap)
	}
	s.records[short] = Record{Short: short, Full: full, UserID: userID}
	return nil
}

func (s *MemoryStorage) Load(short string) (string, error) {
	if r, ok := s.records[short]; ok {
		return r.Full, nil
	}
	return "", fmt.Errorf("record \"%s\" not found", short)
}

func (s *MemoryStorage) LoadForUser(userID string) ([]Record, error) {
	recordsList := make([]Record, 0)
	for _, record := range s.records {
		if record.UserID == userID {
			recordsList = append(recordsList, record)
		}
	}
	return recordsList, nil
}

func (s *MemoryStorage) Delete(short string) error {
	if _, ok := s.records[short]; !ok {
		return fmt.Errorf("record \"%s\" not found", short)
	}
	delete(s.records, short)
	return nil
}
