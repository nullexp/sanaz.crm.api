package memory

import (
	"bytes"
	"image/png"
	"io"
	"testing"

	"github.com/o1egl/govatar"
	"github.com/stretchr/testify/assert"
)

func getImage() []byte {
	const username = "username"
	img, err := govatar.GenerateForUsername(govatar.MALE, username)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	imgbytes := buf.Bytes()
	return imgbytes
}

func toReadCloser(b []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(b))
}

const fileName = "filename.png"

func TestStoreRetrieve(t *testing.T) {
	storage := NewImageStorage()
	img := getImage()

	err := storage.Store(toReadCloser(img), fileName)
	assert.Nil(t, err)

	f, _, err := storage.Retrieve(fileName)
	assert.Nil(t, err)

	data, err := io.ReadAll(f)
	assert.Nil(t, err)
	assert.Equal(t, img, data)

	thumb, _, err := storage.RetrieveThumbnail(fileName)
	assert.Nil(t, err)

	thumbData, err := io.ReadAll(thumb)
	assert.Nil(t, err)
	assert.NotEqual(t, img, thumbData)

	err = storage.Remove(fileName)
	assert.Nil(t, err)

}
