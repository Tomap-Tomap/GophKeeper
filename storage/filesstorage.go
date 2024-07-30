package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

// DBFiler интерфейс описывающий объекты для чтения файлов
type DBFiler interface {
	Write(b []byte) (int, error)
	GetChunck() ([]byte, error)
	Close() error
}

// FileStorage структура хранилища файлов
type FileStorage struct {
	fileSystem fs.StatFS
	chunkSize  int
}

// NewFileStorage аллоцирует новую структуру FileStorage
func NewFileStorage(folder string, chunkSize int) *FileStorage {
	return &FileStorage{os.DirFS(folder).(fs.StatFS), chunkSize}
}

// CreateDBFile создает DBFile
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

// GetDBFile полуачает уже созданный DBFile
func (fs *FileStorage) GetDBFile(fileName string) (DBFiler, error) {
	file, err := fs.fileSystem.Open(fileName)

	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", fileName, err)
	}
	return NewDBFile(file.(*os.File), fs.chunkSize), nil
}

// DBFile структура файла базы данных
type DBFile struct {
	sync.RWMutex
	file   *os.File
	buffer []byte
}

// NewDBFile создает новый DBFile
func NewDBFile(file *os.File, chunkSize int) *DBFile {
	return &DBFile{
		file:   file,
		buffer: make([]byte, chunkSize),
	}
}

// Write записывает данные в файл
func (dbf *DBFile) Write(b []byte) (int, error) {
	return dbf.file.Write(b)
}

// GetChunck возвращает часть данных файла
func (dbf *DBFile) GetChunck() ([]byte, error) {
	dbf.RWMutex.Lock()
	n, err := dbf.file.Read(dbf.buffer)
	dbf.RWMutex.Unlock()

	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return dbf.buffer[:n], nil
}

// Close закрывает DBFile
func (dbf *DBFile) Close() error {
	return dbf.file.Close()
}
