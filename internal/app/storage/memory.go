package storage

import (
	"errors"
	"fmt"
)

type MemoryStorage struct {
	records map[string]string
}

func NewMemoryStorage(records map[string]string) *MemoryStorage {
	return &MemoryStorage{records: records}
}

func (s *MemoryStorage) Store(short, full string) error {
	if s.records == nil {
		s.records = make(map[string]string)
	}
	s.records[short] = full
	return nil
}

func (s *MemoryStorage) Load(short string) (string, error) {
	if full, ok := s.records[short]; ok {
		return full, nil
	}
	return "", errors.New(fmt.Sprintf("record \"%s\" not found", short))
}

func (s *MemoryStorage) Delete(short string) error {
	if _, ok := s.records[short]; !ok {
		return errors.New(fmt.Sprintf("record \"%s\" not found", short))
	}
	delete(s.records, short)
	return nil
}
