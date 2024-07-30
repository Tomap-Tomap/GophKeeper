//go:build unit

package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fStorage = NewFileStorage("testdata", 2)

func BenchmarkSaveFile(b *testing.B) {
	testBatch1 := make([]byte, 0, 1024)
	testBatch2 := make([]byte, 0, 1024)

	for i := 0; i < 1024; i++ {
		testBatch1 = append(testBatch1, byte(i))
		testBatch2 = append(testBatch2, byte(i))
	}

	b.Run("save with buffer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fileData := bytes.Buffer{}
			fileData.Write(testBatch1)
			fileData.Write(testBatch2)

			file, _ := os.Create("test")
			fileData.WriteTo(file)
			file.Close()
			os.Remove("test")
		}
	})

	b.Run("save with out buffer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			file, _ := os.Create("test")
			file.Write(testBatch1)
			file.Write(testBatch2)
			file.Close()
			os.Remove("test")
		}
	})
}

func TestFileStorage_CreateDBFile(t *testing.T) {
	t.Run("test file exists", func(t *testing.T) {
		require.FileExists(t, fmt.Sprintf("%s/testfile", fStorage.fileSystem))
		dbf, err := fStorage.CreateDBFile("testfile")
		require.ErrorContains(t, err, "file testfile is exists")
		assert.Nil(t, dbf)
	})

	t.Run("test checking file exist error", func(t *testing.T) {
		dbf, err := fStorage.CreateDBFile("/")
		require.ErrorContains(t, err, "checking file / exist")
		assert.Nil(t, dbf)
	})

	t.Run("test error", func(t *testing.T) {
		dbf, err := fStorage.CreateDBFile("test/error")
		require.ErrorContains(t, err, "create file")
		assert.Nil(t, dbf)
	})

	t.Run("positive test", func(t *testing.T) {
		pathToFile := fmt.Sprintf("%s/positiveFile", fStorage.fileSystem)
		require.NoFileExists(t, pathToFile)
		dbf, err := fStorage.CreateDBFile("positiveFile")
		require.NoError(t, err)
		dbfs, ok := dbf.(*DBFile)
		require.True(t, ok)
		assert.Equal(t, pathToFile, dbfs.file.Name())
		err = os.Remove(pathToFile)
		require.NoError(t, err)
	})
}

func TestFileStorage_GetDBFile(t *testing.T) {
	t.Run("error test", func(t *testing.T) {
		dbf, err := fStorage.GetDBFile("errorFile")
		require.ErrorContains(t, err, "open file errorFile")
		assert.Nil(t, dbf)
	})

	t.Run("positive test", func(t *testing.T) {
		dbf, err := fStorage.GetDBFile("testfile")
		require.NoError(t, err)
		dbfs, ok := dbf.(*DBFile)
		require.True(t, ok)
		assert.Equal(t, fmt.Sprintf("%s/testfile", fStorage.fileSystem), dbfs.file.Name())
	})
}

func TestDBFile_Write(t *testing.T) {
	t.Run("error test", func(t *testing.T) {
		fileName := "testdata/errorfile"
		file, err := os.Create(fileName)
		require.NoError(t, err)
		defer func() {
			err = os.Remove(fileName)
			require.NoError(t, err)
		}()

		dbf := NewDBFile(file, 0)
		err = dbf.Close()
		require.NoError(t, err)

		_, err = dbf.Write([]byte("test"))
		require.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		fileName := "testdata/positivefile"
		file, err := os.Create(fileName)
		require.NoError(t, err)
		defer func() {
			err = os.Remove(fileName)
			require.NoError(t, err)
		}()

		dbf := NewDBFile(file, 0)
		defer func() {
			err = dbf.Close()
			require.NoError(t, err)
		}()

		n, err := dbf.Write([]byte("test"))
		require.NoError(t, err)
		require.Equal(t, 4, n)
	})
}

func TestDBFile_GetChunck(t *testing.T) {
	t.Run("EOF test", func(t *testing.T) {
		fileName := "testdata/errorfile"
		file, err := os.Create(fileName)
		require.NoError(t, err)
		defer func() {
			err = os.Remove(fileName)
			require.NoError(t, err)
		}()

		dbf := NewDBFile(file, 1)
		defer func() {
			err = dbf.Close()
			require.NoError(t, err)
		}()

		data, err := dbf.GetChunck()
		require.ErrorIs(t, err, io.EOF)
		assert.Nil(t, data)
	})

	t.Run("negative test", func(t *testing.T) {
		fileName := "testdata/errorfile"
		file, err := os.Create(fileName)
		require.NoError(t, err)
		defer func() {
			err = os.Remove(fileName)
			require.NoError(t, err)
		}()

		dbf := NewDBFile(file, 0)
		err = dbf.Close()
		require.NoError(t, err)

		data, err := dbf.GetChunck()
		require.Error(t, err)
		require.NotErrorIs(t, err, io.EOF)
		assert.Nil(t, data)
	})

	t.Run("positive test", func(t *testing.T) {
		fileName := "testdata/testfile"
		file, err := os.Open(fileName)
		require.NoError(t, err)

		dbf := NewDBFile(file, 2)
		defer func() {
			err = dbf.Close()
			require.NoError(t, err)
		}()

		data, err := dbf.GetChunck()
		require.NoError(t, err)
		require.Equal(t, []byte("12"), data)
		data, err = dbf.GetChunck()
		require.NoError(t, err)
		require.Equal(t, []byte("34"), data)
	})
}
