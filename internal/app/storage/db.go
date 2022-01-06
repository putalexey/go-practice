package storage

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose/v3"
	"time"
)

var _ Storager = &DBStorage{}

var recordsTableName = "shorts"
var queryTimeout = 5 * time.Second

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(databaseDSN, migrationsDir string) (*DBStorage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	storage := &DBStorage{
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

func (s *DBStorage) Store(ctx context.Context, record Record) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	insertSQL := fmt.Sprintf(`INSERT INTO %s ("short", "original", "user_id") VALUES ($1, $2, $3)`, recordsTableName)
	_, err := s.db.ExecContext(ctx, insertSQL, record.Short, record.Full, record.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) Load(ctx context.Context, short string) (Record, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	r := Record{}
	selectSQL := fmt.Sprintf("SELECT short, original, user_id from %s WHERE short = $1 LIMIT 1", recordsTableName)
	row := s.db.QueryRowContext(ctx, selectSQL, short)
	err := row.Scan(&r.Short, &r.Full, &r.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return Record{}, fmt.Errorf("record \"%s\" not found: %w", short, err)
		}
		return Record{}, err
	}

	return r, nil
}

func (s *DBStorage) LoadForUser(ctx context.Context, userID string) ([]Record, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	recordList := make([]Record, 0)
	selectSQL := fmt.Sprintf("SELECT short, original, user_id from %s WHERE user_id = $1", recordsTableName)
	rows, err := s.db.QueryContext(ctx, selectSQL, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var r Record
		err := rows.Scan(&r.Short, &r.Full, &r.UserID)
		if err != nil {
			return nil, err
		}
		recordList = append(recordList, r)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return recordList, nil
}

func (s *DBStorage) Delete(ctx context.Context, short string) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE short = $1", recordsTableName)
	_, err := s.db.ExecContext(ctx, deleteSQL, short)
	return err
}

func (s *DBStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	return s.db.PingContext(ctx)
}
