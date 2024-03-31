package gin

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	gah "github.com/timewasted/go-accept-headers"
	"gitlab.espadev.ir/espad-go/infrastructure/file"
	httpapi "gitlab.espadev.ir/espad-go/infrastructure/http/protocol"
	"gitlab.espadev.ir/espad-go/infrastructure/http/protocol/model"
	mutipartmodel "gitlab.espadev.ir/espad-go/infrastructure/http/protocol/model/multipart"
	"gitlab.espadev.ir/espad-go/infrastructure/http/protocol/response"
	logger "gitlab.espadev.ir/espad-go/infrastructure/log"
	"gitlab.espadev.ir/espad-go/infrastructure/misc"
)

const (
	AcceptAll = "*/*"
)

type request struct {
	ctx *gin.Context
}

func NewRequest(ctx *gin.Context) httpapi.Request {
	return &request{ctx}
}

const serverErrorLog = "Server err of %s - %s , msg: %s"

func (req *request) SetServerError(msg string) {
	c := req.ctx
	logger.Error.Printf(serverErrorLog, c.ClientIP(), c.Request.URL, msg)
	req.ctx.JSON(http.StatusInternalServerError, model.RequestError{Message: msg, Code: response.ServerError})
	req.ctx.Abort()
}

// SetForbidden will set http.Forbidden status code with given data
func (req *request) SetForbidden() {
	req.negotiate(http.StatusForbidden, model.RequestError{Code: response.AccessDenied})
	req.ctx.Abort()
}

func (req *request) SetUnauthorized(msg string, code string) {
	req.negotiate(http.StatusUnauthorized, model.RequestError{Message: msg, Code: code})
	req.ctx.Abort()
}

func (req *request) SetBadRequest(msg string, code string) {
	req.negotiate(http.StatusBadRequest, model.RequestError{Message: msg, Code: code})

	req.ctx.Abort()
}

func (req *request) SetStatus(status int) {
	req.ctx.Status(status)
}

func (req *request) SetNotFound(message string, code string) {
	req.negotiate(http.StatusNotFound, model.RequestError{Message: message, Code: code})
	req.ctx.Abort()
}

func (req *request) Set(key string, value interface{}) {
	req.ctx.Set(key, value)
}

func (req *request) SetFile(key string, value httpapi.FileHeader) {
	req.ctx.Set(key, value)
}

func (req *request) SetFiles(key string, value []httpapi.FileHeader) {
	req.ctx.Set(key, value)
}

func (req *request) Get(key string) (interface{}, bool) {
	val, ok := req.ctx.Get(key)

	if !ok {
		return val, ok
	}

	casted, ok := val.(misc.Operand)

	if !ok {
		return val, true
	} else {
		return casted.Value, true
	}
}

func (req *request) MustGet(key string) interface{} {
	v, ok := req.Get(key)

	if !ok {
		panic("missing key value")
	}

	return v
}

const (
	PleaseReadTheErrorCode = "Please read the error code"
	ServerErrorOccured     = "Server error occured, please contact administrator"
)

const (
	GenericNotFound = "GenericNotFound"
	DataWasNotFound = "Data Was not found"
)

func (req *request) handleError(err error) bool {
	if err != nil {
		if ok, oe := misc.ToUserOperationError(err); ok {

			if oe.IsNotFoundError() {
				req.SetNotFound(DataWasNotFound, oe.GetOperationErrorCode())
			} else {
				req.SetBadRequest(PleaseReadTheErrorCode, oe.GetOperationErrorCode())
			}
			return true
		}

		req.SetServerError(ServerErrorOccured)
		logger.Error.Println(err, err.Error())
		return true
	}
	return false
}

func (req *request) Negotiate(stausCode int, err error, dto interface{}) {
	if req.handleError(err) {
		return
	}

	if dto == nil || (reflect.ValueOf(dto).Kind() == reflect.Ptr && reflect.ValueOf(dto).IsNil()) {
		req.SetNotFound(DataWasNotFound, GenericNotFound)
		return
	}

	req.negotiate(stausCode, dto)
}

func (req *request) ReturnStatus(stausCode int, err error) {
	if req.handleError(err) {
		return
	}

	req.ctx.Status(stausCode)
}

const NotOffered = "the accepted formats are not offered by the server or format is not supplied"

func (req *request) setStatusNotAccetable() {
	_ = req.ctx.AbortWithError(http.StatusNotAcceptable, errors.New(NotOffered))
}

const Data = "Data"

const AcceptHeader = "Accept"

func (req *request) getAccept() gah.AcceptSlice {
	accept := req.ctx.GetHeader(AcceptHeader)
	if accept == "" {
		a := gah.Accept{
			Type:       "application",
			Subtype:    "json",
			Q:          1.0,
			Extensions: make(map[string]string),
		}
		return []gah.Accept{a}
	}

	accepts := gah.Parse(accept)
	return accepts
}

func (req *request) getMixedContentType(b string) string {
	// We must quote the boundary if it contains any of the
	// tspecials characters defined by RFC 2045, or space.
	if strings.ContainsAny(b, `()<>@,;:\"/[]?= `) {
		b = `"` + b + `"`
	}
	return "multipart/mixed; boundary=" + b
}

const (
	ContentDisposition            = "Content-Disposition"
	AttachmentNamePlaceholder     = `attachment; name="%s"`
	AttachmentWithFilePlaceholder = AttachmentNamePlaceholder + `;filename="%s";`
)

var escaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func (req *request) escapeQuotes(s string) string {
	return escaper.Replace(s)
}

func (req *request) writeMultipart(vm []httpapi.Multipart, mixed bool) error {
	mwriter := multipart.NewWriter(req.ctx.Writer)

	ctType := mwriter.FormDataContentType()
	if mixed {
		ctType = req.getMixedContentType(mwriter.Boundary())
	}
	req.ctx.Header("Content-Type", ctType)

	for _, v := range vm {

		h := make(textproto.MIMEHeader)

		fpart, isFile := v.(file.File)

		dispositionContent := fmt.Sprintf(AttachmentNamePlaceholder, req.escapeQuotes(v.GetPartName()))
		if isFile {
			dispositionContent = fmt.Sprintf(AttachmentWithFilePlaceholder, req.escapeQuotes(v.GetPartName()), req.escapeQuotes(fpart.GetFilename()))
		}
		h.Set(ContentDisposition, dispositionContent)
		h.Set(CType, v.GetMimeType())
		writer, err := mwriter.CreatePart(h)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, v)
		if err != nil {
			return err
		}
		v.Close()
	}

	return mwriter.Close()
}

const MMixed = "multipart/mixed"

func (req *request) sendMultipartMixed(code int, data interface{}) {
	parts := []httpapi.Multipart{}

	drr, arrayBased := data.([]interface{})
	append := func(v interface{}) {
		f, isFile := v.(file.File)
		if isFile {
			parts = append(parts, newFilePart(f, "file"))
		} else {
			parts = append(parts, newValuePart(v, "data", gin.MIMEJSON))
		}
	}
	if arrayBased {
		for _, v := range drr {
			append(v)
		}
	} else {
		append(data)
	}

	if len(parts) == 0 {
		req.SetStatus(http.StatusNoContent)
		return
	}

	err := req.writeMultipart(parts, true)
	if err != nil {
		req.SetServerError(err.Error())
		return
	}
	req.SetStatus(code)
}

func (req *request) ReturnMultpartMixed(code int, err error, out ...httpapi.Multipart) {
	if req.handleError(err) {
		return
	}

	err = req.writeMultipart(out, true)
	if err != nil {
		req.SetServerError(err.Error())
		return
	}
	req.SetStatus(code)
}

func (req *request) negotiate(code int, data interface{}) {
	accepter := req.getAccept()

	for _, v := range accepter {
		mime := v.Type + "/" + v.Subtype
		switch mime {

		case MMixed:
			if code >= http.StatusBadRequest {
				req.ctx.JSON(code, data)
				return
			}
			req.sendMultipartMixed(code, data)
			return

		case AcceptAll:
			fallthrough
		case binding.MIMEHTML:
			fallthrough
		case binding.MIMEJSON:
			req.ctx.JSON(code, data)
			return

		case binding.MIMEXML:
			req.ctx.XML(code, data)
			return

		case binding.MIMEYAML:
			req.ctx.YAML(code, data)
			return

		}
	}

	req.setStatusNotAccetable()
}

func (req *request) GetDTO() (interface{}, bool) {
	return req.ctx.Get(httpapi.KeyDTO)
}

const NoDtoExit = "No dto has been given"

func (req *request) MustGetDTO() interface{} {
	dto, exist := req.ctx.Get(httpapi.KeyDTO)

	if !exist {
		panic(NoDtoExit)
	}
	return dto
}

func (req *request) GetFile(partName string) (httpapi.FileHeader, bool) {
	f, ok := req.ctx.Get(partName)

	if !ok {
		return nil, ok
	}

	file, correctCast := f.(httpapi.FileHeader)
	return file, correctCast
}

func (req *request) MustGetFile(partName string) httpapi.FileHeader {
	f := req.ctx.MustGet(partName)

	file, correctCast := f.(httpapi.FileHeader)

	if !correctCast {
		panic("File \"" + partName + "\" does not exist")
	}
	return file
}

func (req *request) GetFiles(partName string) ([]httpapi.FileHeader, bool) {
	f, ok := req.ctx.Get(partName)

	if !ok {
		return nil, ok
	}

	file, correctCast := f.([]httpapi.FileHeader)
	return file, correctCast
}

func (req *request) MustGetFiles(partName string) []httpapi.FileHeader {
	f := req.ctx.MustGet(partName)

	file, correctCast := f.([]httpapi.FileHeader)

	if !correctCast {
		panic("File \"" + partName + "\" does not exist")
	}
	return file
}

const CType = "Content-Type"

func (req *request) RangeFile(status int, err error, file file.SeekerFile) {
	if req.handleError(err) {
		return
	}
	req.SetStatus(status)
	req.ctx.Header(CType, file.GetMimeType())
	defer file.Close()
	http.ServeContent(req.ctx.Writer, req.ctx.Request, file.GetFilename(), file.GetLastModifiedDate(), file)
}

func (req *request) WriteFile(status int, err error, file file.File) {
	if req.handleError(err) {
		return
	}
	req.ctx.Header(CType, file.GetMimeType())
	req.SetStatus(status)
	_, err = io.Copy(req.ctx.Writer, file)
	defer file.Close()
	if err != nil {
		req.SetServerError(err.Error())
	}
}

func newValuePart(obj interface{}, partName, contentType string) httpapi.Multipart {
	return mutipartmodel.NewJsonPart(obj, partName)
}

func newFilePart(f file.File, partName string) httpapi.Multipart {
	return mutipartmodel.NewFilePart(f, partName)
}

type page struct {
	Skip  int `json:"skip" form:"skip"`
	Limit int `json:"limit" form:"limit"`
}

type cursorPage struct {
	Cursor string `json:"cursor" form:"cursor"`
	Limit  int    `json:"limit" form:"limit"`
	After  bool   `json:"after" form:"after"`
}

func (req *request) GetPagination() (misc.Pagination, bool) {
	p := page{Skip: 0, Limit: 10}
	err := req.ctx.ShouldBindWith(&p, binding.Query)
	if err != nil {
		return misc.NewPage(0, 10), false
	}

	maxLimit := req.MustGet(httpapi.MaxLimit).(int)

	if maxLimit != 0 {
		if p.Limit < 1 || p.Limit > maxLimit {
			p.Limit = maxLimit
		}
	}

	if p.Limit == -1 {
		return nil, false
	}

	return misc.NewPage(p.Skip, p.Limit), true
}

func (req *request) GetCursorPagination() (misc.CursorPagination, bool) {
	p := cursorPage{Limit: 10}
	err := req.ctx.ShouldBindWith(&p, binding.Query)
	if err != nil {
		return misc.NewCursorPagination(10, "", false), false
	}

	maxLimit := req.MustGet(httpapi.MaxLimit).(int)

	if maxLimit != 0 {
		if p.Limit < 1 || p.Limit > maxLimit {
			p.Limit = maxLimit
		}
	}

	if p.Limit == -1 {
		return nil, false
	}

	return misc.NewCursorPagination(p.Limit, p.Cursor, p.After), true
}

func (req *request) GetCaller() (misc.Caller, bool) {
	auth, ok := req.Get(httpapi.KeyAuth)
	if !ok {
		return nil, ok
	}
	model, ok := auth.(misc.JwtClaim)
	return model, ok
}

func (req *request) MustGetCaller() misc.Caller {
	caller, ok := req.GetCaller()
	if !ok {
		panic("Error finiding caller")
	}
	return caller
}

type sort struct {
	Name      string
	Ascending bool
	Order     int
}

func (s sort) GetName() string {
	return s.Name
}

func (s sort) IsAscending() bool {
	return s.Ascending
}

func (s sort) GetOrder() int {
	return s.Order
}

func (req *request) GetSort() []misc.Sort {
	out := []misc.Sort{}

	sortQuery := req.ctx.Query(misc.QuerySort)

	if sortQuery == "" {
		return out
	}

	parts := strings.Split(sortQuery, ",")

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		if part == "" {
			continue
		}
		s := sort{}
		part = strings.TrimSpace(part)
		rawArr := strings.Split(part, "-")

		if len(rawArr) == 2 {
			s.Ascending = false
			s.Name = rawArr[1]
			s.Order = i + 1
			if strings.ContainsAny(s.Name, "-+") {
				continue
			}
			out = append(out, s)
			continue
		}

		rawArr = strings.Split(part, "+")

		if len(rawArr) == 2 || len(rawArr) == 1 {
			s.Ascending = true
			if len(rawArr) == 1 {
				s.Name = rawArr[0]
			} else {
				s.Name = rawArr[1]
			}
			s.Order = i + 1
			if strings.ContainsAny(s.Name, "-+") {
				continue
			}
			out = append(out, s)
		}

	}
	return out
}

type QueryOperator string

const (
	QueryOperatorEqual           QueryOperator = "eq"
	QueryOperatorNotEqual        QueryOperator = "neq"
	QueryOperatorMoreThan        QueryOperator = "mr"
	QueryOperatorEqualOrMoreThan QueryOperator = "mroeq"
	QueryOperatorLessThan        QueryOperator = "ls"
	QueryOperatorEqualOrLessThan QueryOperator = "lsoeq"
	QueryOperatorContain         QueryOperator = "cn"
	QueryOperatorNotContain      QueryOperator = "ncn"
	QueryOperatorEmpty           QueryOperator = "empt"
	QueryOperatorNotEmpty        QueryOperator = "nempt"
)

func (req *request) GetQuery() []misc.Query {
	value, ok := req.Get(httpapi.KeyQuery)
	if !ok {
		return []misc.Query{}
	}

	return value.([]misc.Query)
}

func (req *request) IsAndQuery() bool {
	isOr := req.ctx.Query(misc.QueryOr)
	if isOr == "true" || isOr == "True" {
		return false
	}

	return true
}

func (req *request) GetDefaultQuery() (string, bool) {
	def := req.ctx.Query(misc.QueryDefault)
	if def == "" {
		return "", false
	}

	return def, true
}
