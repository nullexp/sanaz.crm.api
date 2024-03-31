package gin

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"

	fle "gitlab.espadev.ir/espad-go/infrastructure/file"
	"gitlab.espadev.ir/espad-go/infrastructure/misc"
)

var ErrMessageTooLarge = errors.New("multipart: message too large")

type Form struct {
	Value map[string][]*ValuePart
	File  map[string][]*FileHeader
}

func (f *Form) RemoveAll() error {
	var err error
	for _, fhs := range f.File {
		for _, fh := range fhs {
			if fh.tmpfile != "" {
				e := os.Remove(fh.tmpfile)
				if e != nil && err == nil {
					err = e
				}
			}
		}
	}
	return err
}

type FileHeader struct {
	Filename string
	Header   textproto.MIMEHeader
	Size     int64

	content []byte
	tmpfile string
}

func (fh *FileHeader) GetFilename() string {
	return fh.Filename
}

func (fh *FileHeader) GetHeader() textproto.MIMEHeader {
	return fh.Header
}

func (fh *FileHeader) GetSize() int64 {
	return fh.Size
}

func (fh *FileHeader) OpenFile() (fle.File, error) {
	var readerCloser io.ReadCloser
	var err error
	if b := fh.content; b != nil {
		readerCloser = io.NopCloser(bytes.NewReader(b))
	} else {
		readerCloser, err = os.Open(fh.tmpfile)
		if err != nil {
			return nil, err
		}
	}
	return memoryFile{readerCloser: readerCloser, mimeType: fh.GetHeader().Get(CType), fileName: fh.Filename}, nil
}

type memoryFile struct {
	fileName     string
	mimeType     string
	readerCloser io.ReadCloser
}

func (m memoryFile) Read(p []byte) (n int, err error) {
	return m.readerCloser.Read(p)
}

func (m memoryFile) Close() error {
	return m.readerCloser.Close()
}

func (m memoryFile) GetFilename() string {
	return m.fileName
}

func (m memoryFile) GetMimeType() string {
	return m.mimeType
}

func (m memoryFile) GetLastModifiedDate() time.Time {
	return time.Now()
}

func (ginApp *GinApp) multipartReader(r *http.Request) (*multipart.Reader, error) {
	v := r.Header.Get("Content-Type")
	if v == "" {
		return nil, http.ErrNotMultipart
	}
	d, params, err := mime.ParseMediaType(v)
	if err != nil || !(d == "multipart/form-data" || d == "multipart/mixed") {
		return nil, http.ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}
	return multipart.NewReader(r.Body, boundary), nil
}

type ValuePart struct {
	Data   []byte
	Header textproto.MIMEHeader
}

func (ginApp *GinApp) readForm(req *http.Request) (form *Form, err error) {
	form = &Form{}
	form.Value = map[string][]*ValuePart{}
	form.File = map[string][]*FileHeader{}

	defer func() {
		if err != nil {
			_ = form.RemoveAll()
		}
	}()

	// Set 10 MB as max memory
	r, err := ginApp.multipartReader(req)
	if err != nil {
		return form, err
	}
	maxMemory := int64(10 * misc.MB)
	maxValueBytes := int64(5 * misc.MB)
	for {
		p, err := r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return form, err
		}

		name := p.FormName()
		if name == "" {
			continue
		}
		filename := p.FileName()

		var b bytes.Buffer

		if filename == "" {
			// value, store as string in memory
			n, err := io.CopyN(&b, p, maxValueBytes+1)
			if err != nil && err != io.EOF {
				return form, err
			}
			maxValueBytes -= n
			if maxValueBytes < 0 {
				return form, errors.New(MessageTooLarge)
			}

			form.Value[name] = append(form.Value[name], &ValuePart{Data: b.Bytes(), Header: p.Header})
			continue
		}

		// file, store in memory or on disk
		fh := &FileHeader{
			Filename: filename,
			Header:   p.Header,
		}
		n, err := io.CopyN(&b, p, maxMemory+1)
		if err != nil && err != io.EOF {
			return form, err
		}
		if n > maxMemory {
			// too big, write to disk and flush buffer
			file, err := os.CreateTemp("", "multipart-")
			if err != nil {
				return form, err
			}
			size, err := io.Copy(file, io.MultiReader(&b, p))
			if cerr := file.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				os.Remove(file.Name())
				return form, err
			}
			fh.tmpfile = file.Name()
			fh.Size = size
		} else {
			fh.content = b.Bytes()
			fh.Size = int64(len(fh.content))
			maxMemory -= n
			maxValueBytes -= n
		}
		form.File[name] = append(form.File[name], fh)
	}

	return form, nil
}
