package storage

import (
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

		ret, err := store.Load("test")
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/testme", ret)
	})

	t.Run("adds records to file", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		err = store.Store("test2", "http://afawef.com/yteyj", "testUser")
		require.NoError(t, err)

		fileData, err := os.ReadFile(tempfilepath)
		require.NoError(t, err)
		assert.NotEqual(t, testData, fileData)
	})

	t.Run("data saves across instances", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		err = store.Store("test2", "http://afawef.com/yteyj", "testUser")
		require.NoError(t, err)

		store, err = NewFileStorage(tempfilepath)
		require.NoError(t, err)

		ret, err := store.Load("test2")
		require.NoError(t, err)

		assert.Equal(t, "http://afawef.com/yteyj", ret)
	})

	t.Run("restores file before operations, when created not with NewFileStorage", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store := FileStorage{
			filepath: tempfilepath,
		}
		ret, err := store.Load("test")
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/testme", ret)

		store = FileStorage{
			filepath: tempfilepath,
		}
		err = store.Store("test2", "http://afawef.com/zxcv", "testUser")
		require.NoError(t, err)

		ret, err = store.Load("test")
		require.NoError(t, err)
		assert.Equal(t, "http://example.com/testme", ret)
	})

	t.Run("deletes data", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		err = store.Delete("test")
		require.NoError(t, err)

		_, err = store.Load("test")
		assert.Error(t, err)
	})

	t.Run("return error on when deleting key not exists", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		err = store.Delete("test123")
		require.Error(t, err)
	})

	t.Run("return user's shorts", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		records, err := store.LoadForUser("testUser")
		assert.NoError(t, err)
		assert.Len(t, records, 1)

		err = store.Store("test2", "http://example.com/testme2", "testUser")
		assert.NoError(t, err)

		records, err = store.LoadForUser("testUser")
		require.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("not return other user's shorts", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)
		err = store.Store("test2", "http://example.com/testme2", "testUser2")
		assert.NoError(t, err)
		err = store.Store("test3", "http://example.com/testme3", "testUser2")
		assert.NoError(t, err)

		records, err := store.LoadForUser("testUser")
		assert.NoError(t, err)
		assert.Len(t, records, 1)

		records, err = store.LoadForUser("testUser2")
		assert.NoError(t, err)
		assert.Len(t, records, 2)
	})

	t.Run("return empty list when user doesn't have shorts", func(t *testing.T) {
		resetFileContents(t, tempfilepath)
		store, err := NewFileStorage(tempfilepath)
		require.NoError(t, err)

		records, err := store.LoadForUser("testUser2")
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
		ret, err := store.Load("test")
		assert.Error(t, err)
		assert.Empty(t, ret)

		store = &FileStorage{
			filepath: tempfilepath,
		}
		err = store.Store("test2", "http://afawef.com/zxcv", "testUser")
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
