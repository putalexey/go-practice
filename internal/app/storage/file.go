package storage

import (
	"context"
	"encoding/json"
	"fmt"
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
		if err := decoder.Decode(&record); err == io.EOF {
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
	s.records[record.Short] = record //Record{Short: short, Full: full, UserID: userID}
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
	return Record{}, fmt.Errorf("record \"%s\" not found", short)
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
		return fmt.Errorf("record \"%s\" not found", short)
	}
	delete(s.records, short)
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
