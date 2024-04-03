package protocol

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/textproto"

	fileProtocol "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/file/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/model"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/model/openapi"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
)

type (
	Api interface {
		Run(ip string, port uint, mode string) error
		AppendDuplexModule(mod DuplexModule)
		AppendModule(mod Module)
		AppendPreHandlers(string, Action)
		GetRoute(url, method string) *RequestDefinition
		AppendAuthorizer(baseURL string, authorizer Authorizer)
		AppendAuthenticator(baseURL string, authorizer Authenticator)
		SetCors(cors []string)
		SetLogHandler(LogHandler)
		SetLogPolicy(model.LogPolicy)
		TestHandle(*httptest.ResponseRecorder, *http.Request) error
		// OpenAPI
		SetExternalDocs(openapi.ExternalDocs)
		SetInfo(openapi.Info)
		SetContact(openapi.Contact)
		SetServers([]openapi.Server)
		EnableOpenApi(route string) error
		SetErrors([]string)
	}

	LogHandler interface {
		Handle(model.HttpLog)
	}

	Module interface {
		GetRequestHandlers() []*RequestDefinition
		GetBaseURL() string
		GetTag() openapi.Tag
	}

	Request interface {
		// GetJson and other stuff
		SetServerError(msg string)
		SetForbidden()
		SetUnauthorized(msg string, code string)
		SetBadRequest(msg string, code string)
		SetNotFound(msg string, code string)
		ReturnStatus(int, error)
		Set(key string, value interface{})
		SetFile(key string, f FileHeader)
		SetFiles(key string, f []FileHeader)
		Get(key string) (interface{}, bool)
		MustGet(key string) interface{}
		GetDTO() (interface{}, bool)
		MustGetDTO() interface{}
		GetFile(partName string) (FileHeader, bool)
		MustGetFile(partName string) FileHeader
		GetFiles(partName string) ([]FileHeader, bool)
		MustGetFiles(partName string) []FileHeader
		Negotiate(stausCode int, err error, dto interface{})
		RangeFile(status int, err error, file fileProtocol.SeekerFile)
		WriteFile(status int, err error, file fileProtocol.File)
		ReturnMultipartMixed(status int, err error, out ...Multipart)
		GetPagination() (misc.Pagination, bool)
		GetCursorPagination() (misc.CursorPagination, bool)
		GetCaller() (misc.Caller, bool)
		MustGetCaller() misc.Caller
		GetSort() []misc.Sort
		GetQuery() []misc.Query
		GetDefaultQuery() (string, bool)
		IsAndQuery() bool
	}

	Verifier interface {
		Verify() error
	}

	Action     func(req Request)
	Authorizer func(identity string, permission string) (bool, error)

	MultipartDefinition interface {
		IsOptional() bool
		GetPartName() string
		IsSingle() bool
		GetObject() interface{}
	}

	MultipartFileDefinition interface {
		MultipartDefinition
		Verify(FileHeader) error
	}

	MultipartValueDefinition interface {
		MultipartDefinition
		Verify() error
	}

	Multipart interface {
		io.ReadCloser
		GetPartName() string
		GetMimeType() string
	}

	FileMultipart interface {
		fileProtocol.File
		Multipart
	}

	FileHeader interface {
		GetHeader() textproto.MIMEHeader
		GetSize() int64
		GetFilename() string
		OpenFile() (fileProtocol.File, error)
	}
)

const (
	KeyRole  = "Role"
	KeyQuery = "Query"
	KeyAuth  = "Auth"
	KeyDTO   = "DTO"
	KeyFile  = "File"
	MaxLimit = "MaxLimit"
)
