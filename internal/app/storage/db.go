package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose/v3"
	"golang.org/x/sync/errgroup"
)

var _ Storager = &DBStorage{}

var recordsTableName = "shorts"
var queryTimeout = 5 * time.Second
var batchQueryTimeout = 30 * time.Second

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(databaseDSN, migrationsDir string) (*DBStorage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxIdleTime(30 * time.Second)
	db.SetConnMaxLifetime(2 * time.Minute)

	storage := &DBStorage{db}

	//migrate
	if migrationsDir != "" {
		err := goose.Up(db, migrationsDir)
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	return storage, db.Ping()
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}

func (s *DBStorage) Store(ctx context.Context, record Record) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	insertSQL := fmt.Sprintf(`INSERT INTO
		%s ("short", "original", "user_id") VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, recordsTableName)
	res, err := s.db.ExecContext(ctx, insertSQL, record.Short, record.Full, record.UserID)
	if err != nil {
		log.Println(err)
		return err
	}

	insertedRows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}
	if insertedRows == 0 {
		var oldRecord Record
		// nothing inserted get conflicted row
		selectSQL := fmt.Sprintf(`SELECT "short", "original", "user_id"
		FROM %s WHERE "original" = $1`, recordsTableName)
		row := s.db.QueryRowContext(ctx, selectSQL, record.Full)
		err := row.Scan(&oldRecord.Short, &oldRecord.Full, &oldRecord.UserID)
		if err != nil {
			log.Println(err)
			return err
		}
		return NewRecordConflictError(oldRecord)
	}
	return nil
}

func (s *DBStorage) StoreBatch(ctx context.Context, records []Record) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertSQL := fmt.Sprintf(`INSERT INTO %s ("short", "original", "user_id") VALUES ($1, $2, $3)`, recordsTableName)
	insertStmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, batchQueryTimeout)
	defer cancel()

	for _, record := range records {
		_, err := insertStmt.ExecContext(ctx, record.Short, record.Full, record.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *DBStorage) Load(ctx context.Context, short string) (Record, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	r := Record{}
	selectSQL := fmt.Sprintf("SELECT short, original, user_id, deleted FROM %s WHERE short = $1 LIMIT 1", recordsTableName)
	row := s.db.QueryRowContext(ctx, selectSQL, short)
	err := row.Scan(&r.Short, &r.Full, &r.UserID, &r.Deleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, NewRecordNotFoundError(short)
		}
		return Record{}, err
	}

	return r, nil
}

func (s *DBStorage) LoadBatch(ctx context.Context, shorts []string) ([]Record, error) {
	recordList := make([]Record, 0)
	if len(shorts) == 0 {
		return recordList, nil
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	shortsPlaceholderList, args := prepareSQLPlaceholders(1, shorts)
	shortsPlaceholderCommaList := strings.Join(shortsPlaceholderList, ",")

	selectSQL := fmt.Sprintf("SELECT short, original, user_id, deleted FROM %s WHERE short in (%s) and deleted = FALSE", recordsTableName, shortsPlaceholderCommaList)
	rows, err := s.db.QueryContext(ctx, selectSQL, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var r Record
		err = rows.Scan(&r.Short, &r.Full, &r.UserID, &r.Deleted)
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

func (s *DBStorage) LoadForUser(ctx context.Context, userID string) ([]Record, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	recordList := make([]Record, 0)
	selectSQL := fmt.Sprintf("SELECT short, original, user_id, deleted FROM %s WHERE user_id = $1 and deleted = FALSE", recordsTableName)
	rows, err := s.db.QueryContext(ctx, selectSQL, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var r Record
		err = rows.Scan(&r.Short, &r.Full, &r.UserID, &r.Deleted)
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

	deleteSQL := fmt.Sprintf("UPDATE %s SET deleted = TRUE WHERE short = $1", recordsTableName)
	_, err := s.db.ExecContext(ctx, deleteSQL, short)
	return err
}

func (s *DBStorage) DeleteBatch(ctx context.Context, shorts []string) error {
	if len(shorts) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	shortsPlaceholderList, args := prepareSQLPlaceholders(1, shorts)
	shortsPlaceholderCommaList := strings.Join(shortsPlaceholderList, ",")
	updateSQL := fmt.Sprintf("UPDATE %s SET deleted = TRUE WHERE short IN (%s)", recordsTableName, shortsPlaceholderCommaList)
	_, err := s.db.ExecContext(ctx, updateSQL, args...)
	return err
}

func (s *DBStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	return s.db.PingContext(ctx)
}

// prepareSQLPlaceholders create 2 arrays:
// 1 - with placeholders with indexes starting from `startIndex`
// 2 - with values for that placeholders
func prepareSQLPlaceholders(startIndex int, values []string) ([]string, []interface{}) {
	pIndex := startIndex
	args := make([]interface{}, 0, len(values))

	shortsPlaceholderList := make([]string, 0, len(values))
	for _, short := range values {
		shortsPlaceholderList = append(shortsPlaceholderList, "$"+strconv.Itoa(pIndex))
		args = append(args, short)
		pIndex++
	}

	return shortsPlaceholderList, args
}

func (s *DBStorage) GetStats(ctx context.Context) (*ServiceStats, error) {
	var stats ServiceStats
	eg := errgroup.Group{}
	eg.Go(func() error {
		cnt, err := s.countURLs(ctx)
		if err != nil {
			return err
		}
		stats.URLsCount = cnt
		return nil
	})
	eg.Go(func() error {
		cnt, err := s.countUsers(ctx)
		if err != nil {
			return err
		}
		stats.UsersCount = cnt
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *DBStorage) countURLs(ctx context.Context) (int, error) {
	countURLs := 0
	selectSQL := fmt.Sprintf("SELECT count(*) as cnt FROM %s WHERE deleted = FALSE", recordsTableName)
	row := s.db.QueryRowContext(ctx, selectSQL)
	err := row.Scan(&countURLs)
	if err != nil {
		return 0, err
	}

	return countURLs, nil
}

func (s *DBStorage) countUsers(ctx context.Context) (int, error) {
	countUsers := 0
	selectSQL := fmt.Sprintf("SELECT count(distinct user_id) as cnt FROM %s WHERE deleted = FALSE", recordsTableName)
	row := s.db.QueryRowContext(ctx, selectSQL)
	err := row.Scan(&countUsers)
	if err != nil {
		return 0, err
	}

	return countUsers, nil
}
