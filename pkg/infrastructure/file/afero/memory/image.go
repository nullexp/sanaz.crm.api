package memory

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/afero/utility"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	"github.com/spf13/afero"
)

type ImageStorage struct {
	fileSystem     afero.Fs
	thumbStoreLock *sync.RWMutex
}

const profileDir = "./profile/"
const profileThumbsDir = "./profile/thumbs/"

func NewImageStorage() protocol.ImageStorage {

	u := ImageStorage{}
	u.fileSystem = afero.NewMemMapFs()
	u.fileSystem.Mkdir(profileDir, os.ModeDir)
	u.fileSystem.Mkdir(profileThumbsDir, os.ModeDir)
	u.thumbStoreLock = &sync.RWMutex{}

	return u
}

func (u ImageStorage) Store(rc io.ReadCloser, name string) error {

	if strings.TrimSpace(name) == "" {
		return protocol.ErrFileNameIsEmpty
	}
	u.remove(name)
	defer rc.Close()
	return u.saveFile(rc, profileDir+name)
}

func (u ImageStorage) saveFile(reader io.Reader, name string) error {

	file, err := u.fileSystem.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}

func (u ImageStorage) retrieve(name string) (io.ReadCloser, time.Time, error) {

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

func (u ImageStorage) Retrieve(name string) (io.ReadCloser, time.Time, error) {
	return u.retrieve(profileDir + name)
}

func (u ImageStorage) GetLastModifiedDate(name string) (time.Time, error) {

	stat, err := u.fileSystem.Stat(profileDir + name)
	if err != nil {
		return time.Time{}, err
	}
	return stat.ModTime(), err
}

func (u ImageStorage) Exist(name string) bool {

	_, err := u.fileSystem.Stat(profileDir + name)
	err = utility.NormalizeError(err)
	return err == nil
}

func (u ImageStorage) remove(name string) error {

	return u.fileSystem.Remove(name)
}

func (u ImageStorage) Remove(name string) error {
	u.remove(profileDir + name)
	u.fileSystem.RemoveAll(profileThumbsDir + name)
	return nil
}

const pngExtension = "png"

func (u ImageStorage) RetrieveThumbnail(name string, params ...any) (io.ReadCloser, time.Time, error) {

	wsize := 64
	var sizeName = "64-"
	if len(params) != 0 {
		v, ok := params[0].(int)
		if ok && v >= 64 {
			wsize = v
			sizeName = strconv.Itoa(wsize) + "-"
		}
	}

	getThmbDir := func(name string) string {
		return profileThumbsDir + name + "/"
	}

	currentThumbFile, t, err := u.retrieve(getThmbDir(name) + sizeName + name)
	if err == nil {
		return currentThumbFile, t, nil
	}

	originalFile, _, err := u.retrieve(profileDir + name)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer originalFile.Close()

	data, err := io.ReadAll(originalFile)
	if err != nil {
		return nil, time.Time{}, err
	}

	getDefault := func() (io.ReadCloser, time.Time, error) {
		return io.NopCloser(bytes.NewReader(data)), t, nil
	}

	if !filetype.IsImage(data) {
		return getDefault()
	}

	img, f, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return getDefault()
	}

	dest := imaging.Resize(img, wsize, 0, imaging.NearestNeighbor)
	buf := new(bytes.Buffer)
	// only supporting png and jpeg
	if f == pngExtension {
		err = png.Encode(buf, dest)
	} else {
		err = jpeg.Encode(buf, dest, nil)
	}

	if err != nil {
		return getDefault()
	}
	newData := buf.Bytes()

	u.thumbStoreLock.Lock()
	u.fileSystem.Mkdir(getThmbDir(name), os.ModeDir)
	u.saveFile(io.NopCloser(bytes.NewReader(newData)), getThmbDir(name)+sizeName+name)
	u.thumbStoreLock.Unlock()

	return io.NopCloser(bytes.NewReader(newData)), time.Now(), nil
}
