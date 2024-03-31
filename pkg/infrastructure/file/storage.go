package file

import (
	"errors"
	"io"
	"time"
)

type FileStorage interface {
	Store(io.ReadCloser, string) error
	Retrive(name string) (io.ReadCloser, time.Time, error)
	GetLastModifiedDate(name string) (time.Time, error)
	Exist(name string) bool
	Remove(name string) error
}

type ImageStorage interface {
	FileStorage
	RetriveThumbnail(string, ...any) (io.ReadCloser, time.Time, error)
}

var (
	ErrFileNotExist    = errors.New("file does not exist")
	ErrFileNameIsEmpty = errors.New("file name can't be empty")
)
