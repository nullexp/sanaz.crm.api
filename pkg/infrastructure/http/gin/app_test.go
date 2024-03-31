package gin

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ldez/mimetype"
	"github.com/stretchr/testify/assert"
	"gitlab.espadev.ir/espad-go/infrastructure/http/protocol"
	"gitlab.espadev.ir/espad-go/infrastructure/http/protocol/model"
	mlp "gitlab.espadev.ir/espad-go/infrastructure/http/protocol/model/multipart"
	"gitlab.espadev.ir/espad-go/infrastructure/http/protocol/model/openapi"
	"gitlab.espadev.ir/espad-go/infrastructure/http/protocol/utility"
	"gitlab.espadev.ir/espad-go/infrastructure/log"
	"gitlab.espadev.ir/espad-go/infrastructure/misc"
)

func init() {
	log.Initialize(false)
}

type testModule struct {
	RequestDefinitions []*protocol.RequestDefinition
	BaseURL            string
	Tag                openapi.Tag
}

func (tm *testModule) GetRequestHandlers() []*protocol.RequestDefinition {
	return tm.RequestDefinitions
}

func (tm *testModule) GetBaseURL() string {
	return tm.BaseURL
}

func (tm *testModule) GetTag() openapi.Tag {
	return tm.Tag
}

func NewTestModule(base string, rdfs ...*protocol.RequestDefinition) *testModule {
	tm := testModule{}
	tm.RequestDefinitions = rdfs
	tm.BaseURL = base
	return &tm
}

func TestModuleInit(t *testing.T) {
	app := NewGinApp()
	var a protocol.Api = app

	var module protocol.Module = NewTestModule("", &protocol.RequestDefinition{
		Route:  "/test",
		Method: http.MethodGet,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusOK, nil)
		},
	})

	a.AppendModule(module)

	app.Init(gin.TestMode)
	handler := app.GetHandlerFunc("/test", http.MethodGet)

	if handler == nil {
		t.Error("Handler is nil")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	handler(c)

	t.Run("Check status ok", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func (t TokenInfo) GetExpireTime() int64 {
	return t.ExpireTime
}

func (t TokenInfo) GetSubject() string {
	return t.Subject
}

func (t TokenInfo) GetIssuer() string {
	return t.Issuer
}

func (t TokenInfo) GetAudience() []string {
	return t.Audience
}

func (t TokenInfo) GetIssuedAt() int64 {
	return t.IssuedAt
}

func (t TokenInfo) GetIdentity() string {
	return t.Identity
}

func (t TokenInfo) IsExpired() bool {
	return time.Now().Unix() > t.ExpireTime
}

type TokenInfo struct {
	ExpireTime int64    `json:"expireTime"`
	Subject    string   `json:"sub"` // User Info
	Issuer     string   `json:"iss"` // Issuer is the api identity
	Audience   []string `json:"aud"` // Apis which can process this token
	IssuedAt   int64    `json:"iat"` // The time it has been issued
	Identity   string   `json:"jti"` // Token Identity
}

type testAuthenticator struct {
	getModel   func(token string) (misc.JwtClaim, error)
	checkToken func(token string) (out bool, outE error)
}

func NewTestAuthenticator(gm func(token string) (misc.JwtClaim, error), ct func(token string) (out bool, outE error)) *testAuthenticator {
	return &testAuthenticator{getModel: gm, checkToken: ct}
}

func NewOkTestAuthenticator() *testAuthenticator {
	return NewTestAuthenticator(func(token string) (misc.JwtClaim, error) {
		return TokenInfo{ExpireTime: time.Now().AddDate(1, 0, 0).Unix()}, nil
	}, func(token string) (bool, error) { return true, nil })
}

func NewOkTestAuthenticatorWithToken(m TokenInfo) *testAuthenticator {
	return NewTestAuthenticator(func(token string) (misc.JwtClaim, error) {
		return m, nil
	}, func(token string) (bool, error) { return true, nil })
}

func NewFailedTestAuthenticator() *testAuthenticator {
	return NewTestAuthenticator(func(token string) (misc.JwtClaim, error) {
		return TokenInfo{ExpireTime: time.Now().AddDate(1, 0, 0).Unix()}, nil
	}, func(token string) (bool, error) { return false, nil })
}

func (t *testAuthenticator) GetModel(token string) (misc.JwtClaim, error) {
	return t.getModel(token)
}

func (t *testAuthenticator) CheckToken(token string) (out bool, outE error) {
	return t.checkToken(token)
}

func TestAuthentication(t *testing.T) {
	app := NewGinApp()
	var a protocol.Api = app

	var smodule protocol.Module = NewTestModule("/success", &protocol.RequestDefinition{
		Route:  "/data",
		Method: http.MethodGet,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	})
	a.AppendModule(smodule)
	auth := NewOkTestAuthenticator()
	a.AppendAuthenticator("/success", auth)

	var fmodule protocol.Module = NewTestModule("/failed", &protocol.RequestDefinition{
		Route:  "/data",
		Method: http.MethodGet,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	})
	a.AppendModule(fmodule)
	auth = NewFailedTestAuthenticator()
	a.AppendAuthenticator("/failed", auth)

	var noAuthRequired protocol.Module = NewTestModule("/noauth", &protocol.RequestDefinition{
		Route:  "/data",
		Method: http.MethodGet,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	})
	a.AppendModule(noAuthRequired)

	var freeRouteModule protocol.Module = NewTestModule("/freeroute", &protocol.RequestDefinition{
		Route:  "/nonfree",
		Method: http.MethodGet,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	},
		&protocol.RequestDefinition{
			Route:     "/free",
			FreeRoute: true,
			Method:    http.MethodGet,
			Handler: func(req protocol.Request) {
				req.ReturnStatus(http.StatusNoContent, nil)
			},
		})
	a.AppendModule(freeRouteModule)
	auth = NewFailedTestAuthenticator()
	a.AppendAuthenticator("/freeroute", auth)

	app.Init(gin.TestMode)

	t.Run("Check base auth status ok", func(t *testing.T) {
		// Testing success auth
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/success/data", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Check base auth stastus failed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/failed/data", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Check no auth required", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/noauth/data", nil)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Check free route with auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/freeroute/free", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, "/freeroute/nonfree", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthoriation(t *testing.T) {
	app := NewGinApp()
	var a protocol.Api = app

	baseRoute := "/test"
	var smodule protocol.Module = NewTestModule(baseRoute, &protocol.RequestDefinition{
		Route:  "/data",
		Method: http.MethodGet,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	})
	a.AppendModule(smodule)
	info := TokenInfo{ExpireTime: time.Now().AddDate(1, 0, 0).Unix(), Subject: "1", Identity: uuid.NewString()}
	auth := NewOkTestAuthenticatorWithToken(info)
	a.AppendAuthenticator(baseRoute, auth)
	a.AppendAuthorizer(baseRoute, func(identity string, permission string) (bool, error) {
		return true, nil
	})

	failbaseRoute := "/fail"
	var module protocol.Module = NewTestModule(failbaseRoute, &protocol.RequestDefinition{
		Route:          "/data",
		Method:         http.MethodGet,
		AnyPermissions: []string{"UserManagement", "SpeceficManagement"},
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	}, &protocol.RequestDefinition{
		Route:          "/data2",
		Method:         http.MethodGet,
		AnyPermissions: []string{"UserManagement"},
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	},
		&protocol.RequestDefinition{
			Route:  "/freeauthor",
			Method: http.MethodGet,
			Handler: func(req protocol.Request) {
				req.ReturnStatus(http.StatusNoContent, nil)
			},
		},
	)
	a.AppendModule(module)
	info = TokenInfo{ExpireTime: time.Now().AddDate(1, 0, 0).Unix(), Subject: "1", Identity: uuid.NewString()}
	auth = NewOkTestAuthenticatorWithToken(info)
	a.AppendAuthenticator(failbaseRoute, auth)
	a.AppendAuthorizer(failbaseRoute, func(identity string, permission string) (bool, error) {
		return false, nil
	})

	app.Init(gin.TestMode)

	t.Run("Expect no authorization error for no permission", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, baseRoute+"/data", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run(" Expect  authorization error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, failbaseRoute+"/data", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, failbaseRoute+"/data2", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Expect no auth error for combination", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, failbaseRoute+"/freeauthor", nil)
		req.Header.Add("Authorization", "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestPreHandlers(t *testing.T) {
	t.Parallel()
	app := NewGinApp()
	var a protocol.Api = app

	baseRoute := "/test"

	a.AppendPreHandlers(baseRoute, func(req protocol.Request) {
		req.SetServerError("error happedn")
	})

	var smodule protocol.Module = NewTestModule(baseRoute, &protocol.RequestDefinition{
		Route:     "/data",
		Method:    http.MethodGet,
		FreeRoute: true,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	})
	a.AppendModule(smodule)

	unrp := "/unrelated"
	var unrel protocol.Module = NewTestModule(unrp, &protocol.RequestDefinition{
		Route:     "/data",
		Method:    http.MethodGet,
		FreeRoute: true,
		Handler: func(req protocol.Request) {
			req.ReturnStatus(http.StatusNoContent, nil)
		},
	})
	a.AppendModule(unrel)

	mlp := "/multiplesetget"
	var mlpm protocol.Module = NewTestModule(mlp, &protocol.RequestDefinition{
		Route:     "/data",
		Method:    http.MethodGet,
		FreeRoute: true,
		Handler: func(req protocol.Request) {
			v, e := req.Get(mlp)
			if e && v != nil && v == 3 {
				req.ReturnStatus(http.StatusNoContent, nil)
			} else {
				req.ReturnStatus(http.StatusInternalServerError, nil)
			}
		},
	})
	a.AppendModule(mlpm)
	a.AppendPreHandlers(mlp, func(req protocol.Request) {
		req.Set(mlp, 1)
	})
	a.AppendPreHandlers(mlp, func(req protocol.Request) {
		c, _ := req.Get(mlp)
		d, _ := c.(int)
		d = d + 2
		req.Set(mlp, d)
	})

	app.Init(gin.TestMode)

	t.Run("Check basic working", func(t *testing.T) {
		// Testing success auth
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/data", new(bytes.Buffer))
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Unrelated path must pass through", func(t *testing.T) {
		// Testing success auth
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/unrelated/data", new(bytes.Buffer))
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Multiple instance running", func(t *testing.T) {
		// Testing success auth
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, mlp+"/data", new(bytes.Buffer))
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

const (
	Dto            = "Dto"
	NoDTO          = "No dto has been found"
	Accept         = "Accept"
	AppJson        = "application/json"
	MultipartMixed = "multipart/mixed"
	AppJsonUtf8    = "application/json; charset=utf-8"
	AppOctedStream = "application/octet-stream"
	AppXml         = "application/xml"
	AppXmlUtf8     = "application/xml; charset=utf-8"
	AppText        = "application/text"
)

type TestDto struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

type TestSocketDto struct {
	Topic string `json:"topic"`
	Dto   any    `json:"dto"`
}

func (t *TestDto) GetPartName() string {
	return strconv.Itoa(int(t.Id))
}

func (t *TestDto) GetContentType() string {
	return gin.MIMEJSON
}

func (t *TestDto) GetObject() interface{} {
	return t
}

func (t *TestDto) Verify() error {
	if t.Id != Id {
		return errors.New("Expect ID")
	}
	return nil
}

type anotherDto struct {
	Passion string `json:"passion"`
}

func (*anotherDto) Verify() error {
	return nil
}

const Id uint = 1

func getTestDto() *TestDto {
	return &TestDto{Id: Id, Name: Name}
}

type object struct {
	Data             []byte
	DataMode         bool
	Reader           io.Reader
	ReadSeeker       io.ReadSeeker
	Name             string
	MimeType         string
	LastModifiedDate time.Time
}

func (f *object) Read(p []byte) (n int, err error) {
	if f.DataMode && f.ReadSeeker == nil {
		f.ReadSeeker = bytes.NewReader(f.Data)
	}
	if f.DataMode {
		return f.ReadSeeker.Read(p)
	} else {
		return f.Reader.Read(p)
	}
}

func (f *object) Seek(offset int64, whence int) (int64, error) {
	if f.DataMode && f.ReadSeeker == nil {
		f.ReadSeeker = bytes.NewReader(f.Data)
	}
	return f.ReadSeeker.Seek(offset, whence)
}

func (f *object) Close() error {
	return nil
}

func (f object) GetFilename() string {
	return f.Name
}

func (f object) GetMimeType() string {
	return f.MimeType
}

func (f object) GetLastModifiedDate() time.Time {
	return f.LastModifiedDate
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func TestDataHandler(t *testing.T) {
	app := NewGinApp()
	var a protocol.Api = app

	baseRoute := "/test"
	var smodule protocol.Module = NewTestModule(baseRoute, &protocol.RequestDefinition{
		Route:     "/data",
		Method:    http.MethodPost,
		Dto:       &TestDto{},
		FreeRoute: true,
		Handler: func(req protocol.Request) {
			raw, _ := req.GetDTO()
			td := raw.(*TestDto)
			req.Negotiate(http.StatusOK, nil, td)
		},
	},
		&protocol.RequestDefinition{
			Route:     "/show-server-error",
			Method:    http.MethodPut,
			Dto:       &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				raw, _ := req.GetDTO()
				td := raw.(*TestDto)
				req.Negotiate(http.StatusOK, errors.New("Server Error"), td)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/operation-error",
			Method:    http.MethodPut,
			Dto:       &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				td := req.MustGetDTO().(*TestDto)
				req.Negotiate(http.StatusOK, misc.UserOperationError{ErrorCode: ErrorCode}, td)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/basic-fileInfo-multipart",
			Method:    http.MethodPost,
			FileParts: []protocol.MultipartFileDefinition{&mlp.FileDefinition{Name: protocol.KeyDTO, Single: true, Optional: false, MaxSize: 20 * misc.MB, MinSize: 10}},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				td := req.MustGetFile(protocol.KeyDTO)
				mf, _ := td.OpenFile()
				defer mf.Close()

				rf := object{Reader: mf, DataMode: false, Name: td.GetFilename(), MimeType: td.GetHeader().Get("Content-Type"), LastModifiedDate: time.Now()}
				req.WriteFile(http.StatusOK, nil, &rf)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/check-unique-dto",
			Method:    http.MethodPost,
			Dto:       &TestDto{Name: "tom"},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				raw, _ := req.GetDTO()
				td := raw.(*TestDto)
				td.Id = td.Id + 1
				req.Negotiate(http.StatusOK, nil, td)
			},
		},
		&protocol.RequestDefinition{
			Route:      "/basic-data-multipart",
			Method:     http.MethodPost,
			ValueParts: []protocol.MultipartValueDefinition{&mlp.DataDefinition{Name: protocol.KeyDTO, Single: true, Optional: false, Object: getTestDto()}},
			FreeRoute:  true,
			Handler: func(req protocol.Request) {
				raw, _ := req.Get(protocol.KeyDTO)
				td, ok := raw.(*TestDto)
				if !ok {
					req.SetBadRequest("Unknown data", "InvalidData")
					return
				}
				req.Negotiate(http.StatusOK, nil, td)
			},
		},
		&protocol.RequestDefinition{
			Route:      "/multiple-singledata-multipart",
			Method:     http.MethodPost,
			ValueParts: []protocol.MultipartValueDefinition{&mlp.DataDefinition{Name: protocol.KeyDTO, Single: false, Optional: false, Object: getTestDto()}},
			FreeRoute:  true,
			Handler: func(req protocol.Request) {
				all, ok := utility.GetByType[TestDto](protocol.KeyDTO, req)
				if !ok {
					req.SetBadRequest("Unknown data", "InvalidData")
					return
				}
				req.Negotiate(http.StatusOK, nil, all)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/simple-multipart-mixed",
			Method:    http.MethodPost,
			Dto:       &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				dt := req.MustGetDTO().(*TestDto)
				req.Negotiate(http.StatusOK, nil, dt)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/multiple-multipart-mixed",
			Method:    http.MethodPost,
			Dto:       &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				dt := req.MustGetDTO().(*TestDto)
				out := []interface{}{dt, getTestDto()}
				req.Negotiate(http.StatusOK, nil, out)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/data-array",
			Method:    http.MethodPost,
			DtoArray:  &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				raw, _ := req.GetDTO()
				tdr, ok := raw.([]*TestDto)
				_ = ok
				req.Negotiate(http.StatusOK, nil, tdr)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/force-multipart",
			Method:    http.MethodPost,
			DtoArray:  &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				tdr := req.MustGetDTO().([]*TestDto)
				req.ReturnMultpartMixed(http.StatusOK, nil, mlp.NewJsonPart(tdr, "data"))
			},
		},
		&protocol.RequestDefinition{
			Route:     "/force-multipart-fileInfo",
			Method:    http.MethodPost,
			DtoArray:  &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				tdr := req.MustGetDTO().([]*TestDto)
				dt, _ := json.Marshal(tdr[0])
				dt1, _ := json.Marshal(tdr[1])
				f1 := mlp.NewFilePart(&object{Data: dt, DataMode: true, Name: protocol.KeyDTO, MimeType: mimetype.ApplicationOctetStream, LastModifiedDate: time.Now()}, protocol.KeyDTO)
				f2 := mlp.NewFilePart(&object{Data: dt1, DataMode: true, Name: protocol.KeyDTO, MimeType: mimetype.ApplicationOctetStream, LastModifiedDate: time.Now()}, protocol.KeyDTO)
				req.ReturnMultpartMixed(http.StatusOK, nil, f1, f2)
			},
		},
		&protocol.RequestDefinition{
			Route:     "/basic-range-multipart",
			Method:    http.MethodGet,
			DtoArray:  &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				text := "abcdefghij"
				data := []byte(text)
				f := &object{Data: data, DataMode: true, Name: protocol.KeyDTO, MimeType: mimetype.ApplicationOctetStream, LastModifiedDate: time.Now()}
				req.RangeFile(http.StatusOK, nil, f)
			},
		},
	)

	a.AppendModule(smodule)

	app.Init(gin.TestMode)

	t.Run("Must return error if dto is empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/data", nil)
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Must return error if dto cant be verified", func(t *testing.T) {
		dto, _ := json.Marshal(&anotherDto{Passion: Name})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/data", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Must return valid object", func(t *testing.T) {
		w := httptest.NewRecorder()
		dto, _ := json.Marshal(getTestDto())
		req, _ := http.NewRequest(http.MethodPost, "/test/data", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		output := TestDto{}
		err := json.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)
	})

	t.Run("Must return server error", func(t *testing.T) {
		w := httptest.NewRecorder()
		dto, _ := json.Marshal(getTestDto())
		req, _ := http.NewRequest(http.MethodPut, baseRoute+"/show-server-error", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Must return user operation error", func(t *testing.T) {
		w := httptest.NewRecorder()
		dto, _ := json.Marshal(getTestDto())
		req, _ := http.NewRequest(http.MethodPut, baseRoute+"/operation-error", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		output := model.RequestError{}
		err := json.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, ErrorCode, output.Code)
	})

	t.Run("Must try negotiate content-type", func(t *testing.T) {
		w := httptest.NewRecorder()
		dto, _ := json.Marshal(getTestDto())
		req, _ := http.NewRequest(http.MethodPost, "/test/data", bytes.NewReader(dto))
		req.Header.Add(misc.HeaderContentType, AppJson)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		t.Log(w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Must negotiate return type ", func(t *testing.T) {
		w := httptest.NewRecorder()
		dto, _ := json.Marshal(getTestDto())
		req, _ := http.NewRequest(http.MethodPost, "/test/data", bytes.NewReader(dto))
		req.Header.Add(Accept, AppXml)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		output := TestDto{}
		err := xml.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)
	})

	t.Run("Must negotiate return type with html", func(t *testing.T) {
		w := httptest.NewRecorder()
		dto, _ := json.Marshal(getTestDto())
		req, _ := http.NewRequest(http.MethodPost, "/test/data", bytes.NewReader(dto))
		req.Header.Add(Accept, mimetype.TextHTML)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Must handle single multipart fileInfo  and return a fileInfo", func(t *testing.T) {
		var buffer bytes.Buffer
		mwriter := multipart.NewWriter(&buffer)
		writer, err := mwriter.CreateFormFile(protocol.KeyDTO, "some.txt")
		assert.Empty(t, err)
		data, _ := json.Marshal(getTestDto())
		_, err = writer.Write([]byte(data))
		assert.Empty(t, err)
		mwriter.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/basic-fileInfo-multipart", &buffer)
		req.Header.Add(Accept, AppOctedStream)
		req.Header.Add(misc.HeaderContentType, mwriter.FormDataContentType())
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, w.Header().Get(misc.HeaderContentType), AppOctedStream)
		t.Log(w.Body.String())
		output := TestDto{}
		err = json.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)
	})

	t.Run("Must handle multipart single data with negotiation and return data", func(t *testing.T) {
		var buffer bytes.Buffer
		mwriter := multipart.NewWriter(&buffer)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(protocol.KeyDTO)))
		h.Set(misc.HeaderContentType, AppXml)

		writer, err := mwriter.CreatePart(h)
		assert.Empty(t, err)
		data, _ := xml.Marshal(getTestDto())
		_, err = writer.Write([]byte(data))

		assert.Empty(t, err)
		mwriter.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/basic-data-multipart", &buffer)
		req.Header.Add(Accept, AppXml)
		req.Header.Add(misc.HeaderContentType, mwriter.FormDataContentType())
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, AppXmlUtf8, w.Header().Get(misc.HeaderContentType))
		output := TestDto{}
		err = xml.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)

		var buffer2 bytes.Buffer
		mwriter = multipart.NewWriter(&buffer2)
		h = make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(protocol.KeyDTO)))
		h.Set(misc.HeaderContentType, AppXml)

		writer, err = mwriter.CreatePart(h)
		assert.Empty(t, err)
		data, _ = xml.Marshal(getTestDto())
		_, err = writer.Write([]byte(data))

		assert.Empty(t, err)
		mwriter.Close()
		w = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodPost, "/test/basic-data-multipart", &buffer2)
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, mwriter.FormDataContentType())
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, AppJsonUtf8, w.Header().Get(misc.HeaderContentType))
		output = TestDto{}
		err = json.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)
	})

	t.Run("Must not keep httpapi.KeyDTO definition state", func(t *testing.T) {
		w := httptest.NewRecorder()
		dto, _ := json.Marshal(getTestDto())
		req, _ := http.NewRequest(http.MethodPost, "/test/check-unique-dto", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		output := TestDto{}
		err := json.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id+1, output.Id)

		w = httptest.NewRecorder()
		dto, _ = json.Marshal(getTestDto())
		req, _ = http.NewRequest(http.MethodPost, "/test/check-unique-dto", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		output = TestDto{}
		err = json.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id+1, output.Id)
	})

	t.Run("Must handle multipart multiple send data with single name", func(t *testing.T) {
		var buffer bytes.Buffer
		mwriter := multipart.NewWriter(&buffer)

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(protocol.KeyDTO)))
		h.Set(misc.HeaderContentType, AppXml)
		writer, err := mwriter.CreatePart(h)
		assert.Empty(t, err)
		data, _ := xml.Marshal(getTestDto())
		_, err = writer.Write([]byte(data))
		assert.Empty(t, err)

		h = make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(protocol.KeyDTO)))
		h.Set(misc.HeaderContentType, AppJson)
		writer, err = mwriter.CreatePart(h)
		assert.Empty(t, err)
		data, _ = json.Marshal(getTestDto())
		_, err = writer.Write([]byte(data))
		assert.Empty(t, err)

		mwriter.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/multiple-singledata-multipart", &buffer)
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, mwriter.FormDataContentType())
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, AppJsonUtf8, w.Header().Get(misc.HeaderContentType))
		output := []*TestDto{}
		err = json.Unmarshal(w.Body.Bytes(), &output)
		t.Log(w.Body.String())
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output[0].Id)
		assert.Equal(t, Name, output[0].Name)
		assert.Equal(t, Id, output[1].Id)
		assert.Equal(t, Name, output[1].Name)
	})

	t.Run("multipart mixed must return single data succesfully", func(t *testing.T) {
		data, _ := json.Marshal(getTestDto())
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/simple-multipart-mixed", bytes.NewReader(data))
		req.Header.Add(Accept, MultipartMixed)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		ty, p, err := mime.ParseMediaType(w.Header().Get(misc.HeaderContentType))
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, p)
		assert.NotEmpty(t, p["boundary"])

		assert.Equal(t, MultipartMixed, ty)
		assert.Contains(t, w.Header().Get(misc.HeaderContentType), MultipartMixed)
		output := TestDto{}

		reader := multipart.NewReader(w.Body, p["boundary"])
		t.Log(w.Body.String())
		np, err := reader.NextPart()
		assert.Equal(t, nil, err)

		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(np)
		t.Log(buf.String())
		err = json.Unmarshal(buf.Bytes(), &output)
		t.Log(w.Body.String())
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)
	})

	t.Run("multipart mixed must return multiple data succesfully", func(t *testing.T) {
		data, _ := json.Marshal(getTestDto())
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/multiple-multipart-mixed", bytes.NewReader(data))
		req.Header.Add(Accept, MultipartMixed)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		ty, p, err := mime.ParseMediaType(w.Header().Get(misc.HeaderContentType))
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, p)
		assert.NotEmpty(t, p["boundary"])

		assert.Equal(t, MultipartMixed, ty)
		assert.Contains(t, w.Header().Get(misc.HeaderContentType), MultipartMixed)
		output := TestDto{}

		reader := multipart.NewReader(w.Body, p["boundary"])
		t.Log(w.Body.String())
		np, err := reader.NextPart()
		assert.Equal(t, nil, err)
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(np)
		t.Log(buf.String())
		err = json.Unmarshal(buf.Bytes(), &output)
		t.Log(w.Body.String())
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)

		np, err = reader.NextPart()
		assert.Equal(t, nil, err)
		buf = new(bytes.Buffer)
		_, _ = buf.ReadFrom(np)
		t.Log(buf.String())
		err = json.Unmarshal(buf.Bytes(), &output)
		t.Log(w.Body.String())
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)

		_, err = reader.NextPart()
		assert.NotEqual(t, nil, err)
	})

	t.Run("Must handle dto array and return array", func(t *testing.T) {
		w := httptest.NewRecorder()
		out := []protocol.Verifier{getTestDto(), getTestDto()}
		dto, _ := json.Marshal(out)
		req, _ := http.NewRequest(http.MethodPost, "/test/data-array", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		output := []*TestDto{}
		err := json.Unmarshal(w.Body.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(out), len(output))

		assert.Equal(t, Id, output[0].Id)
		assert.Equal(t, Name, output[1].Name)

		assert.Equal(t, Id, output[0].Id)
		assert.Equal(t, Name, output[1].Name)
	})

	t.Run("Must verify dto array", func(t *testing.T) {
		w := httptest.NewRecorder()
		out := []protocol.Verifier{&TestDto{Id: 2, Name: "Jack"}, getTestDto()}
		dto, _ := json.Marshal(out)
		req, _ := http.NewRequest(http.MethodPost, "/test/data-array", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Must verify dto array atleast has one element", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/data-array", nil)
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Must verify dto array correct type", func(t *testing.T) {
		w := httptest.NewRecorder()
		out := []protocol.Verifier{&anotherDto{Passion: ""}}
		dto, _ := json.Marshal(out)
		req, _ := http.NewRequest(http.MethodPost, "/test/data-array", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Must verify dto array single correct type", func(t *testing.T) {
		w := httptest.NewRecorder()
		out := []protocol.Verifier{&anotherDto{Passion: ""}, getTestDto()}
		dto, _ := json.Marshal(out)
		req, _ := http.NewRequest(http.MethodPost, "/test/data-array", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Must server force return multipart data", func(t *testing.T) {
		dtos := []protocol.Verifier{getTestDto(), getTestDto()}
		data, _ := json.Marshal(dtos)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/force-multipart", bytes.NewReader(data))
		req.Header.Add(Accept, MultipartMixed)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		ty, p, err := mime.ParseMediaType(w.Header().Get(misc.HeaderContentType))
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, p)
		assert.NotEmpty(t, p["boundary"])

		assert.Equal(t, MultipartMixed, ty)

		output := []*TestDto{}
		reader := multipart.NewReader(w.Body, p["boundary"])

		np, err := reader.NextPart()
		assert.Equal(t, nil, err)
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(np)
		err = json.Unmarshal(buf.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output[0].Id)
		assert.Equal(t, Name, output[0].Name)

		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output[1].Id)
		assert.Equal(t, Name, output[1].Name)

		assert.Equal(t, len(dtos), len(output))

		_, err = reader.NextPart()
		assert.NotEqual(t, nil, err)
	})

	t.Run("Must server force return multipart fileInfo", func(t *testing.T) {
		dtos := []protocol.Verifier{getTestDto(), getTestDto()}
		data, _ := json.Marshal(dtos)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/force-multipart-fileInfo", bytes.NewReader(data))
		req.Header.Add(Accept, MultipartMixed)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		ty, p, err := mime.ParseMediaType(w.Header().Get(misc.HeaderContentType))
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, p)
		assert.NotEmpty(t, p["boundary"])

		assert.Equal(t, MultipartMixed, ty)

		output := TestDto{}
		reader := multipart.NewReader(w.Body, p["boundary"])

		np, err := reader.NextPart()
		assert.Equal(t, nil, err)
		assert.Equal(t, mimetype.ApplicationOctetStream, np.Header.Get(misc.HeaderContentType))
		h := np.Header.Get(misc.HeaderContentDisposition)
		_, prs, err := mime.ParseMediaType(h)
		assert.Equal(t, nil, err)
		assert.Equal(t, protocol.KeyDTO, prs[misc.HeaderFileName])
		assert.Equal(t, protocol.KeyDTO, prs[misc.HeaderName])
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(np)
		assert.Equal(t, nil, err)
		err = json.Unmarshal(buf.Bytes(), &output)

		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)

		np, err = reader.NextPart()
		assert.Equal(t, nil, err)
		assert.Equal(t, mimetype.ApplicationOctetStream, np.Header.Get(misc.HeaderContentType))
		_, prs, err = mime.ParseMediaType(np.Header.Get(misc.HeaderContentDisposition))
		assert.Equal(t, nil, err)
		assert.Equal(t, protocol.KeyDTO, prs[misc.HeaderFileName])
		assert.Equal(t, protocol.KeyDTO, prs[misc.HeaderName])
		buf = new(bytes.Buffer)
		_, _ = buf.ReadFrom(np)
		err = json.Unmarshal(buf.Bytes(), &output)
		assert.Equal(t, nil, err)
		assert.Equal(t, Id, output.Id)
		assert.Equal(t, Name, output.Name)

		_, err = reader.NextPart()
		assert.NotEqual(t, nil, err)
	})

	t.Run("Must handle big size multipart", func(t *testing.T) {
		var buffer bytes.Buffer
		mwriter := multipart.NewWriter(&buffer)
		writer, err := mwriter.CreateFormFile(protocol.KeyDTO, "some.txt")
		assert.Empty(t, err)
		data := make([]byte, 12*misc.MB)
		_, _ = rand.Read(data)
		_, err = writer.Write(data)
		assert.Empty(t, err)
		mwriter.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/basic-fileInfo-multipart", &buffer)
		ttt := AppOctedStream + `, application/json`
		req.Header.Add(Accept, ttt)
		req.Header.Add(misc.HeaderContentType, mwriter.FormDataContentType())
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, w.Header().Get(misc.HeaderContentType), AppOctedStream)

		rs := bytes.Compare(w.Body.Bytes(), data)
		assert.Equal(t, 0, rs)
	})

	t.Run("Must handle Range fileInfo", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/basic-range-multipart", nil)
		ttt := AppOctedStream + `, application/json`
		req.Header.Add(Accept, ttt)
		req.Header.Add(misc.HeaderRange, "bytes=3-5")

		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusPartialContent, w.Code)
		assert.Equal(t, w.Header().Get(misc.HeaderContentType), AppOctedStream)
		expect := []byte("def")
		t.Log(w.Body.String())
		rs := bytes.Compare(w.Body.Bytes(), expect)
		assert.Equal(t, 0, rs)
	})
}

func BenchmarkDtoArrayProcessing(b *testing.B) {
	app := NewGinApp()
	var a protocol.Api = app

	baseRoute := "/test"
	var smodule protocol.Module = NewTestModule(baseRoute,
		&protocol.RequestDefinition{
			Route:     "/data-array",
			Method:    http.MethodPost,
			DtoArray:  &TestDto{},
			FreeRoute: true,
			Handler: func(req protocol.Request) {
				raw, _ := req.GetDTO()
				tdr, ok := raw.([]*TestDto)
				_ = ok
				req.Negotiate(http.StatusOK, nil, tdr)
			},
		},
	)

	a.AppendModule(smodule)

	app.Init(gin.TestMode)

	out := []protocol.Verifier{getTestDto()}
	dto, _ := json.Marshal(out)
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/data-array", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
	}
}

func BenchmarkSingleProcessing(b *testing.B) {
	app := NewGinApp()
	var a protocol.Api = app

	baseRoute := "/test"
	var smodule protocol.Module = NewTestModule(baseRoute, &protocol.RequestDefinition{
		Route:     "/data",
		Method:    http.MethodPost,
		Dto:       &TestDto{},
		FreeRoute: true,
		Handler: func(req protocol.Request) {
			raw, _ := req.GetDTO()
			td := raw.(*TestDto)
			req.Negotiate(http.StatusOK, nil, td)
		},
	},
	)

	a.AppendModule(smodule)

	app.Init(gin.TestMode)

	dto, _ := json.Marshal(getTestDto())
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test/data", bytes.NewReader(dto))
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderContentType, AppJson)
		_ = app.TestHandle(w, req)
	}
}

// Duplex tests

type testDuplexModule struct {
	defs []*protocol.DuplexHandlerDefinition
	base string
}

func (t testDuplexModule) GetDuplexHandlers() []*protocol.DuplexHandlerDefinition {
	return t.defs
}

func (t testDuplexModule) GetBaseURL() string {
	return t.base
}

func (t testDuplexModule) OnDuplexConnected(_ protocol.DuplexConnection) {
	fmt.Println("duplex connected")
}

func (t testDuplexModule) OnDuplexDisconnected(_ protocol.DuplexConnection) {
	fmt.Println("duplex disconnected")
}

func NewTestDuplexModule(base string, dhds ...*protocol.DuplexHandlerDefinition) *testDuplexModule {
	tm := testDuplexModule{}
	tm.defs = dhds
	return &tm
}

func TestDuplexCanRun(t *testing.T) {
	app := NewGinApp()
	var a protocol.Api = app

	var module protocol.DuplexModule = NewTestDuplexModule("", &protocol.DuplexHandlerDefinition{
		Topic: "test",
		Dto:   nil,
		Handler: func(msg protocol.DuplexMessage) {
			fmt.Println("new message!")
		},
	})

	a.AppendDuplexModule(module)
	getError := func() <-chan error {
		channel := make(chan error)

		go func() {
			err := a.Run("", 9095, gin.TestMode)
			// if there is no error, Run will block the code
			channel <- err
		}()
		return channel
	}

	select {
	case v := <-time.After(20 * time.Millisecond):
		t.Log(v)
		break

	case e := <-getError():
		t.Error(e)
	}
}
