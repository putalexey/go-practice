package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

func resetFileContents(t *testing.T, tempfilepath string) {
	err := os.WriteFile(tempfilepath, testData, 0666)
	require.NoError(t, err)
}

var testData = []byte("{\"short\":\"test\",\"full\":\"http://example.com/testme\",\"user_id\":\"testUser\"}")

func TestFileStorage(t *testing.T) {
	ctx := context.Background()
	var err error
	tempfilepath := GetFilePath()
	defer os.Remove(tempfilepath)

	t.Run("creates new file on path", func(t *testing.T) {
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)
		assert.FileExists(t, tempfilepath)
		assert.IsType(t, &FileStorage{}, store)
	})

	t.Run("reads from file", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		ret, err := store.Load(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/testme", ret.Full)
	})

	t.Run("adds records to file", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		r, err := NewRecord("http://afawef.com/yteyj", "testUser")
		require.NoError(t, err)

		err = store.Store(ctx, r)
		require.NoError(t, err)

		fileData, err := os.ReadFile(tempfilepath)
		require.NoError(t, err)
		assert.NotEqual(t, testData, fileData)
	})

	t.Run("data saves across instances", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		r, err := NewRecord("http://afawef.com/yteyj", "testUser")
		require.NoError(t, err)

		err = store.Store(ctx, r)
		require.NoError(t, err)

		store, err = NewFileStorage(tempfilepath)
		require.NoError(t, err)

		ret, err := store.Load(ctx, r.Short)
		require.NoError(t, err)

		assert.Equal(t, "http://afawef.com/yteyj", ret.Full)
	})

	t.Run("restores file before operations, when created not with NewFileStorage", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store := FileStorage{
			filepath: tempfilepath,
		}
		ret, err := store.Load(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/testme", ret.Full)

		store = FileStorage{
			filepath: tempfilepath,
		}

		r, err := NewRecord("http://afawef.com/yteyj", "testUser")
		require.NoError(t, err)

		err = store.Store(ctx, r)
		require.NoError(t, err)

		ret, err = store.Load(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/testme", ret.Full)
	})

	t.Run("deletes data", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		err = store.Delete(ctx, "test")
		require.NoError(t, err)

		_, err = store.Load(ctx, "test")
		assert.Error(t, err)
	})

	t.Run("return error on when deleting key not exists", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		err = store.Delete(ctx, "test123")
		require.Error(t, err)
	})

	t.Run("return user's shorts", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		records, err := store.LoadForUser(ctx, "testUser")
		assert.NoError(t, err)
		assert.Len(t, records, 1)

		r, err := NewRecord("http://example.com/testme2", "testUser")
		require.NoError(t, err)

		err = store.Store(ctx, r)
		assert.NoError(t, err)

		records, err = store.LoadForUser(ctx, "testUser")
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("not return other user's shorts", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		r1, err := NewRecord("http://example.com/testme2", "testUser2")
		require.NoError(t, err)

		err = store.Store(ctx, r1)
		assert.NoError(t, err)

		r2, err := NewRecord("http://example.com/testme3", "testUser2")
		require.NoError(t, err)

		err = store.Store(ctx, r2)
		assert.NoError(t, err)

		records, err := store.LoadForUser(ctx, "testUser")
		assert.NoError(t, err)
		assert.Len(t, records, 1)

		records, err = store.LoadForUser(ctx, "testUser2")
		assert.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("return empty list when user doesn't have shorts", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		records, err := store.LoadForUser(ctx, "testUser2")
		assert.NoError(t, err)
		assert.NotNil(t, records)
		assert.Len(t, records, 0)
	})

	err = os.WriteFile(tempfilepath, []byte("{\"short\":\"asd\",\"full\":\"http://example.com/tes"), 0666)
	require.NoError(t, err)
	t.Run("returns errors when fails to read file", func(t *testing.T) {
		store, err := NewFileStorage(tempfilepath)
		assert.Error(t, err)
		assert.Nil(t, store)

		store = &FileStorage{
			filepath: tempfilepath,
		}
		_, err = store.Load(ctx, "test")
		assert.Error(t, err)

		store = &FileStorage{
			filepath: tempfilepath,
		}

		r, err := NewRecord("http://afawef.com/zxcv", "testUser")
		require.NoError(t, err)

		err = store.Store(ctx, r)
		assert.Error(t, err)
	})
}

func GetFilePath() string {
	randString := ""
	for i := 0; i < 5; i++ {
		randString += strconv.Itoa(int(rand.Intn(10)))
	}
	return os.TempDir() + "/testfile_" + randString
}
