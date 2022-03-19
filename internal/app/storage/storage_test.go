package storage_test

import (
	"context"
	"fmt"
	"github.com/putalexey/go-practicum/internal/app/storage"
)

func ExampleMemoryStorage_Store() {
	ctx := context.Background()
	store := storage.NewMemoryStorage(nil)
	store.Store(ctx, storage.Record{
		Short:  "100",
		Full:   "http://example.com",
		UserID: "1",
	})
	r, err := store.Load(ctx, "100")
	if err != nil {
		panic(err)
	}
	fmt.Println(r.Full)

	// Output:
	// http://example.com
}

func ExampleMemoryStorage_Delete() {
	ctx := context.Background()
	store := storage.NewMemoryStorage(nil)
	store.Store(ctx, storage.Record{
		Short:  "100",
		Full:   "http://example.com",
		UserID: "1",
	})
	r, _ := store.Load(ctx, "100")
	fmt.Println(r.Full)

	store.Delete(ctx, "100")
	_, err := store.Load(ctx, "100")
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// http://example.com
	// record "100" not found
}

func ExampleMemoryStorage_DeleteBatch() {
	ctx := context.Background()
	store := storage.NewMemoryStorage(storage.RecordMap{
		"100": storage.Record{Short: "100", Full: "http://example.com", UserID: "1"},
		"101": storage.Record{Short: "101", Full: "http://example.com/1", UserID: "1"},
	})

	r, _ := store.Load(ctx, "100")
	fmt.Println(r.Full)
	r, _ = store.Load(ctx, "101")
	fmt.Println(r.Full)

	store.DeleteBatch(ctx, []string{"100", "101"})
	records, err := store.LoadForUser(ctx, "1")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(records)

	// Output:
	// http://example.com
	// http://example.com/1
	// []
}
