package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
	"time"
)

var _ Storager = &DBStorage{}

type DBStorage struct {
	// temporary extend MemoryStorage to pass tests
	MemoryStorage
	db *sql.DB
}

func NewDBStorage(databaseDSN string) (*DBStorage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	storage := &DBStorage{
		MemoryStorage: MemoryStorage{
			records: make(RecordMap),
		},
		db: db,
	}
	return storage, nil
}

//func (s *DBStorage) Store(short, full, userID string) error {
//	//s.db.Exec("INSERT INTO {}")
//	//s.records[short] = Record{Short: short, Full: full, UserID: userID}
//	return nil
//}
//
//func (s *DBStorage) Load(short string) (string, error) {
//	//if record, ok := s.records[short]; ok {
//	//	return record.Full, nil
//	//}
//	return "", fmt.Errorf("record \"%s\" not found", short)
//}
//
//func (s *DBStorage) LoadForUser(userID string) ([]Record, error) {
//	recordList := make([]Record, 0)
//	//for _, record := range s.records {
//	//	if record.UserID == userID {
//	//		recordList = append(recordList, record)
//	//	}
//	//}
//
//	return recordList, nil
//}
//
//func (s *DBStorage) Delete(short string) error {
//	//if _, ok := s.records[short]; !ok {
//	//	return fmt.Errorf("record \"%s\" not found", short)
//	//}
//	//delete(s.records, short)
//	return nil
//}

func (s *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
