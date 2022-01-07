package storage

import (
	"context"
	"fmt"
)

func NewBatchInserter(store Storager, bufferSize int) *BatchInserter {
	inserter := BatchInserter{
		store:      store,
		bufferSize: bufferSize,
		buffer:     make([]Record, 0, bufferSize),
	}

	return &inserter
}

type BatchInserter struct {
	store      Storager
	bufferSize int
	buffer     []Record
}

func (b *BatchInserter) AddItem(ctx context.Context, r Record) error {
	b.buffer = append(b.buffer, r)
	if len(b.buffer) == cap(b.buffer) {
		err := b.Flush(ctx)
		if err != nil {
			return fmt.Errorf("cannot write records: %w", err)
		}
	}
	return nil
}

func (b *BatchInserter) Flush(ctx context.Context) error {
	if len(b.buffer) == 0 {
		return nil
	}

	if err := b.store.StoreBatch(ctx, b.buffer); err != nil {
		return err
	}

	b.buffer = b.buffer[:0]
	return nil
}
