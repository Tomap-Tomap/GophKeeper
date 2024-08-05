package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

// DBFiler is an interface that describes objects for reading files.
type DBFiler interface {
	Write(b []byte) (int, error)
	GetChunk() ([]byte, error)
	Close() error
}

// FileStorage is a struct that holds the chunk size for file operations.
type FileStorage struct {
	fileSystem fs.StatFS
	chunkSize  int
}

// NewFileStorage creates a new FileStorage with the given chunk size.
func NewFileStorage(folder string, chunkSize int) *FileStorage {
	return &FileStorage{os.DirFS(folder).(fs.StatFS), chunkSize}
}

// CreateDBFile create a file and returns a DBFiler for it.
// It returns an error if the file cannot be create.
func (fs *FileStorage) CreateDBFile(fileName string) (DBFiler, error) {
	_, err := fs.fileSystem.Stat(fileName)

	switch {
	case err == nil:
		return nil, fmt.Errorf("file %s is exists", fileName)
	case !errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("checking file %s exist: %w", fileName, err)
	}

	filePath := fmt.Sprintf("%s%s%s", fs.fileSystem, string(os.PathSeparator), fileName)

	file, err := os.Create(filePath)

	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return NewDBFile(file, fs.chunkSize), nil
}

// GetDBFile opens a file and returns a DBFiler for it.
// It returns an error if the file cannot be opened or is not of type *os.File.
func (fs *FileStorage) GetDBFile(fileName string) (DBFiler, error) {
	file, err := fs.fileSystem.Open(fileName)

	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", fileName, err)
	}

	osFile, ok := file.(*os.File)
	if !ok {
		return nil, errors.New("file is not of type *os.File")
	}

	return NewDBFile(osFile, fs.chunkSize), nil
}

// DeleteDBFile deletes DBFiler.
func (fs *FileStorage) DeleteDBFile(fileName string) error {
	filePath := fmt.Sprintf("%s%s%s", fs.fileSystem, string(os.PathSeparator), fileName)

	err := os.Remove(filePath)

	if os.IsNotExist(err) {
		return nil
	}

	return err
}

// GetChunkSize returns the chunk size used by the FileStorage instance.
// This chunk size is typically used to determine the size of data chunks
// when reading from or writing to files in the storage system.
func (fs *FileStorage) GetChunkSize() int {
	return fs.chunkSize
}

// DBFile is a struct that wraps an os.File and provides synchronized access to it.
type DBFile struct {
	sync.RWMutex
	file   *os.File
	buffer []byte
}

// NewDBFile creates a new DBFile with the given os.File and chunk size.
func NewDBFile(file *os.File, chunkSize int) *DBFile {
	return &DBFile{
		file:   file,
		buffer: make([]byte, chunkSize),
	}
}

// Write writes the given byte slice to the file.
// It returns the number of bytes written and any error encountered.
func (dbf *DBFile) Write(b []byte) (int, error) {
	return dbf.file.Write(b)
}

// GetChunk returns a chunk of data from the file.
// It returns the byte slice containing the chunk and any error encountered.
func (dbf *DBFile) GetChunk() ([]byte, error) {
	dbf.RWMutex.Lock()
	n, err := dbf.file.Read(dbf.buffer)
	dbf.RWMutex.Unlock()

	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return dbf.buffer[:n], nil
}

// Close closes the file.
// It returns any error encountered.
func (dbf *DBFile) Close() error {
	return dbf.file.Close()
}
