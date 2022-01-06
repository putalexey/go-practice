package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose/v3"
	"time"
)

var _ Storager = &DBStorage{}

type DBStorage struct {
	// temporary extend MemoryStorage to pass tests
	MemoryStorage
	db *sql.DB
}

func NewDBStorage(databaseDSN, migrationsDir string) (*DBStorage, error) {
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

	//migrate
	if migrationsDir != "" {
		err := goose.Up(db, migrationsDir)
		if err != nil {
			return nil, err
		}
	}

	return storage, nil
}

//func (s *DBStorage) Store(_ context.Context, record Record) error {
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

func (s *DBStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return s.db.PingContext(ctx)
}
