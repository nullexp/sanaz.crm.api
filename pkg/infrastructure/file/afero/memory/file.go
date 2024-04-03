package memory

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/afero/utility"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
	"github.com/spf13/afero"
)

type FileStorage struct {
	fileSystem     afero.Fs
	thumbStoreLock *sync.RWMutex
	dir            string
}

func NewFileStorage(dir string) protocol.FileStorage {

	u := FileStorage{}
	u.fileSystem = afero.NewMemMapFs()
	u.fileSystem.Mkdir(dir, os.ModeDir)
	u.thumbStoreLock = &sync.RWMutex{}
	u.dir = dir

	return u
}

func (u FileStorage) Store(rc io.ReadCloser, name string) error {

	if strings.TrimSpace(name) == "" {
		return protocol.ErrFileNameIsEmpty
	}
	u.remove(name)
	defer rc.Close()
	return u.saveFile(rc, u.dir+name)
}

func (u FileStorage) saveFile(reader io.Reader, name string) error {

	file, err := u.fileSystem.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}

func (u FileStorage) retrieve(name string) (io.ReadCloser, time.Time, error) {

	file, err := u.fileSystem.Open(name)
	if err != nil {
		return nil, time.Time{}, utility.NormalizeError(err)
	}
	s, e := file.Stat()
	var modTime = time.Now()
	if e != nil {
		modTime = s.ModTime()
	}
	return file, modTime, nil
}

func (u FileStorage) Retrieve(name string) (io.ReadCloser, time.Time, error) {
	return u.retrieve(u.dir + name)
}

func (u FileStorage) GetLastModifiedDate(name string) (time.Time, error) {

	stat, err := u.fileSystem.Stat(u.dir + name)
	if err != nil {
		return time.Time{}, err
	}
	return stat.ModTime(), err
}

func (u FileStorage) Exist(name string) bool {

	_, err := u.fileSystem.Stat(u.dir + name)
	err = utility.NormalizeError(err)
	return err == nil
}

func (u FileStorage) remove(name string) error {

	return u.fileSystem.Remove(name)
}

func (u FileStorage) Remove(name string) error {
	u.remove(u.dir + name)
	u.fileSystem.RemoveAll(profileThumbsDir + name)
	return nil
}
