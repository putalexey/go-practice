package storage

import (
	"context"
	"github.com/google/uuid"
)

type Record struct {
	Short  string `json:"short"`
	Full   string `json:"full"`
	UserID string `json:"user_id"`
}

func NewRecord(full, userID string) (Record, error) {
	recordID, err := uuid.NewUUID()
	if err != nil {
		return Record{}, err
	}

	r := Record{
		Short:  recordID.String(),
		Full:   full,
		UserID: userID,
	}
	return r, nil
}

type RecordMap map[string]Record

type Storager interface {
	Store(ctx context.Context, r Record) error
	Load(ctx context.Context, short string) (Record, error)
	LoadForUser(ctx context.Context, userID string) ([]Record, error)
	Delete(ctx context.Context, short string) error
	Ping(ctx context.Context) error
}
