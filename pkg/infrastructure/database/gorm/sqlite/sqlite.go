package sqlite

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TODO: must be in a utility/common package

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}

	defer source.Close()

	os.Remove(dst)

	os.RemoveAll(dst)

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	defer destination.Close()

	nBytes, err := io.Copy(destination, source)

	return nBytes, err
}

// TODO: must be in a utility/common package

func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)

	if err != nil {
		return
	}

	err = out.Sync()

	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}

	err = os.Chmod(dst, si.Mode())

	if err != nil {
		return
	}

	return
}

// TODO: must be in a utility/common package

// CopyDir recursively copies a directory tree, attempting to preserve permissions.

// Source directory must exist, destination directory must *not* exist.

// Symlinks are ignored and skipped.

func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)

	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)

	if err != nil && !os.IsNotExist(err) {
		return
	}

	err = os.MkdirAll(dst, si.Mode())

	if err != nil {
		return
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {

		srcPath := filepath.Join(src, entry.Name())

		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {

			err = CopyDir(srcPath, dstPath)

			if err != nil {
				return
			}

		} else {

			err = CopyFile(srcPath, dstPath)

			if err != nil {
				return
			}

		}

	}

	return
}

type SqliteDatabase struct {
	Database *gorm.DB

	DatabaseLocation string

	DatabaseName string

	Memory bool

	InnerMux *sync.Mutex
}

func (md *SqliteDatabase) Backup(location string) error {
	source := md.getDbLocation()

	_, err := Copy(source, location)
	if err != nil {

		log.Println(err)

		return err

	}

	_, err = Copy(source+"-shm", location+"-shm")

	if err != nil {
		return err
	}

	_, err = Copy(source+"-wal", location+"-wal")

	if err != nil {
		return err
	}

	return nil
}

func NewSqliteDatabase(memory bool, location, dbName string) *SqliteDatabase {
	md := SqliteDatabase{}

	md.DatabaseLocation = location

	md.DatabaseName = dbName

	md.Memory = memory

	md.InnerMux = &sync.Mutex{}

	return &md
}

func (md *SqliteDatabase) GetGorm() *gorm.DB {
	md.InnerMux.Lock()

	defer md.InnerMux.Unlock()

	return md.Database
}

func (md *SqliteDatabase) open(dir string, memory bool) error {
	md.InnerMux.Lock()

	defer md.InnerMux.Unlock()

	fullDirectory := dir + "/" + md.DatabaseName + ".db"

	if memory {
		fullDirectory = "file::memory:?mode=memory&cache=private"
	}

	db, err := gorm.Open(sqlite.Open(fullDirectory), &gorm.Config{SkipDefaultTransaction: true})

	db.Exec("PRAGMA foreign_keys = ON")

	db.Exec("PRAGMA journal_mode = WAL  ")

	if err != nil {
		return err
	}

	preloadPlugin := &PreloadAllPlugin{}

	err = preloadPlugin.Initialize(db)

	if err != nil {
		return err
	}

	md.Database = db

	return nil
}

func (md *SqliteDatabase) Open() error {
	return md.open(md.getDbLocation(), md.Memory)
}

func (md *SqliteDatabase) getDbLocation() string {
	return md.DatabaseLocation
}
