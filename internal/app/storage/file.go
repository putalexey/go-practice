package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
)

var _ Storager = &FileStorage{}

type FileStorage struct {
	records  RecordMap
	filepath string
}

func NewFileStorage(filepath string) (*FileStorage, error) {
	storage := &FileStorage{
		records:  make(RecordMap),
		filepath: filepath,
	}
	if err := storage.restore(); err != nil {
		return nil, err
	}
	return storage, nil
}

func (s *FileStorage) restore() error {
	file, err := os.OpenFile(s.filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	if s.records == nil {
		s.records = make(RecordMap)
	}
	decoder := json.NewDecoder(file)
	for {
		var record Record
		if err := decoder.Decode(&record); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		s.records[record.Short] = record
	}
	return nil
}

func (s *FileStorage) Store(_ context.Context, record Record) error {
	if s.records == nil {
		if err := s.restore(); err != nil {
			return err
		}
	}
	s.records[record.Short] = record
	return s.saveToFile()
}

func (s *FileStorage) StoreBatch(_ context.Context, records []Record) error {
	if s.records == nil {
		if err := s.restore(); err != nil {
			return err
		}
	}
	for _, record := range records {
		s.records[record.Short] = record
	}
	return s.saveToFile()
}

func (s *FileStorage) Load(_ context.Context, short string) (Record, error) {
	if s.records == nil {
		if err := s.restore(); err != nil {
			return Record{}, err
		}
	}

	if record, ok := s.records[short]; ok {
		return record, nil
	}
	return Record{}, NewRecordNotFoundError(short)
}

func (s *FileStorage) LoadBatch(_ context.Context, shorts []string) ([]Record, error) {
	if s.records == nil {
		if err := s.restore(); err != nil {
			return nil, err
		}
	}

	recordsList := make([]Record, 0)
	for _, short := range shorts {
		r, ok := s.records[short]
		if !ok {
			return nil, NewRecordNotFoundError(short)
		}
		recordsList = append(recordsList, r)
	}
	return recordsList, nil
}

func (s *FileStorage) LoadForUser(_ context.Context, userID string) ([]Record, error) {
	if s.records == nil {
		if err := s.restore(); err != nil {
			return nil, err
		}
	}

	recordList := make([]Record, 0)
	for _, record := range s.records {
		if record.UserID == userID {
			recordList = append(recordList, record)
		}
	}

	return recordList, nil
}

func (s *FileStorage) Delete(_ context.Context, short string) error {
	if _, ok := s.records[short]; !ok {
		return NewRecordNotFoundError(short)
	}
	delete(s.records, short)
	return s.saveToFile()
}

func (s *FileStorage) DeleteBatch(ctx context.Context, shorts []string) error {
	// check all shorts exists
	filteredShorts := make([]string, 0, len(shorts))
	for _, short := range shorts {
		if _, ok := s.records[short]; !ok {
			return NewRecordNotFoundError(short)
		}
	}
	// delete them
	for _, short := range filteredShorts {
		delete(s.records, short)
	}
	return s.saveToFile()
}

func (s *FileStorage) Ping(_ context.Context) error {
	return nil
}

func (s *FileStorage) saveToFile() error {
	file, err := os.Create(s.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range s.records {
		if err = encoder.Encode(record); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStorage) GetStats(ctx context.Context) (*ServiceStats, error) {
	stats := ServiceStats{
		URLsCount:  s.countURLs(ctx),
		UsersCount: s.countUsers(ctx),
	}
	return &stats, nil
}

func (s *FileStorage) countURLs(_ context.Context) int {
	return len(s.records)
}

func (s *FileStorage) countUsers(_ context.Context) int {
	usersID := make(map[string]bool)
	for _, record := range s.records {
		if _, ok := usersID[record.UserID]; !ok {
			usersID[record.UserID] = true
		}
	}
	return len(usersID)
}
