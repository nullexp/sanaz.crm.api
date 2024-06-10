package gin

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"time"

	limit "github.com/aviddiviner/gin-limit"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gorilla/websocket"
	wsmodel "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/gin/ws"
	httpapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"
	model "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model/openapi"
	response "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/response"
	logger "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/log"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
)

//go:embed asset/swagger
var swaggerDirectory embed.FS

var NotFound = gin.H{"code": "404", "message": "Request not found"}

const (
	OpenApiRoute                         = "/openapi.json"
	AssetSwagger                         = "asset/swagger"
	BaseApiURL                           = "/api"
	Token                                = "token"
	Authorization                        = "Authorization"
	Bearer                               = "Bearer"
	BearerSpace                          = Bearer + " "
	MissingAuthHeader                    = "Authorization header missing."
	EmptyAuthenticationIsDetected        = "Nil Authentication Is Detected"
	EmptyAuthorizationIsDetected         = "Nil Authorization Is Detected"
	UnknownAuthFormat                    = "Unknown Authorization header format."
	CouldNotParseToken                   = "Could not parse token."
	TTLExpired                           = "Ttl expired!"
	Release                              = "release"
	UnrecognizedRoute                    = "unrecognized Route."
	ArrayIsExpected                      = "An array is Expected."
	MissingParamWithName                 = "Missing param with name "
	UnknownValueParameter                = "Unknown value for parameter "
	ExpectMultipartFormData              = "Expect Multipart form data"
	MissingPart                          = "Missing Part %s"
	SingleSupport                        = "Only single data is requested"
	MessageTooLarge                      = "multipart: message too large"
	ProtocolNegotiationFailed            = "ProtocolNegotiationFailed"
	ProtocolNegotiationFailedDescription = "Must have two field: topic and dto where topic is just a string and dto is json form of object"
	TopicNotSupported                    = "TopicNotSupported"
	TopicNotSupportedDescription         = "Given topic is not defined. contact admins!"
)

type GinApp struct {
	ginDuplexHandlers []httpapi.DuplexModule
	ginDomainHandlers []httpapi.Module
	preHandlers       map[string][]httpapi.Action
	PermissionManager *PermissionManager
	authenticators    map[string]httpapi.Authenticator
	authorizers       map[string]httpapi.Authorizer
	router            *Router
	cors              []string
	logHandler        httpapi.LogHandler
	logPolicy         model.LogPolicy
	gin               *gin.Engine

	// for openapi
	ApiInfo      *openapi.Info
	Servers      []openapi.Server
	Contact      *openapi.Contact
	ExternalDocs *openapi.ExternalDocs
	Errors       []string
	openApiRoute string
	jsonOpenApi  string
}

func NewGinApp() *GinApp {
	instance := GinApp{}
	instance.ginDomainHandlers = []httpapi.Module{}
	instance.preHandlers = map[string][]httpapi.Action{}
	instance.PermissionManager = NewPermissionManager()
	instance.router = NewRouter()
	instance.authenticators = make(map[string]httpapi.Authenticator)
	instance.authorizers = map[string]httpapi.Authorizer{}
	instance.cors = []string{}
	return &instance
}

func (ginApp *GinApp) SetLogHandler(handler httpapi.LogHandler) {
	ginApp.logHandler = handler
}

func (ginApp *GinApp) GetRoute(url, method string) *httpapi.RequestDefinition {
	return ginApp.router.GetRoute(url, httpapi.HTTPMethod(method))
}

func (ginApp *GinApp) GetLogHandler() httpapi.LogHandler {
	return ginApp.logHandler
}

func (ginApp *GinApp) SetLogPolicy(policy model.LogPolicy) {
	ginApp.logPolicy = policy
}

func (ginApp *GinApp) GetLogPolicy() model.LogPolicy {
	return ginApp.logPolicy
}

func (ginApp *GinApp) SetCors(cors []string) {
	ginApp.cors = cors
}

func (ginApp *GinApp) GetCors() []string {
	return ginApp.cors
}

func (ginApp *GinApp) AppendModule(handler httpapi.Module) {
	ginApp.ginDomainHandlers = append(ginApp.ginDomainHandlers, handler)
}

func (ginApp *GinApp) AppendDuplexModule(handler httpapi.DuplexModule) {
	ginApp.ginDuplexHandlers = append(ginApp.ginDuplexHandlers, handler)
}

func (ginApp *GinApp) AppendPreHandlers(baseURL string, action httpapi.Action) {
	// Race Condition will never happends.
	if ginApp.preHandlers[baseURL] == nil {
		ginApp.preHandlers[baseURL] = []httpapi.Action{}
	}
	ginApp.preHandlers[baseURL] = append(ginApp.preHandlers[baseURL], action)
}

func (ginApp *GinApp) AppendAuthorizer(baseURL string, authorizer httpapi.Authorizer) {
	// No Race Condition will ever happens
	ginApp.authorizers[baseURL] = authorizer
}

func (ginApp *GinApp) AppendAuthenticator(baseURL string, authenticator httpapi.Authenticator) {
	// No Race Condition will ever happens
	ginApp.authenticators[baseURL] = authenticator
}

func (ginApp *GinApp) Run(ip string, port uint, mode string) error {
	ginApp.Init(mode)
	return ginApp.gin.Run(fmt.Sprintf("%s:%d", ip, port))
}

func (ginApp *GinApp) Init(mode string) {
	if ginApp.gin != nil {
		panic("GinApp seems to be already initiated!")
	}

	// TODO you should set gin mode with env not hard code
	// https://stackoverflow.com/questions/46411173/how-to-set-gin-mode-to-release-mode
	if gin.Mode() != mode {
		if mode == gin.ReleaseMode {
			gin.SetMode(gin.ReleaseMode)
		} else if mode == gin.TestMode {
			gin.SetMode(gin.TestMode)
		} else if mode == gin.DebugMode {
			gin.SetMode(gin.DebugMode)
		} else {
			panic("Invalid mode for gin.SetMode, passed value: \"" + mode + "\"")
		}
	}

	r := gin.New()
	ginApp.initDefaultHandlers(r)
	ginApp.initRouter()
	ginApp.enableOpenApiIfRequired(r)
	ginApp.initAuthentication(r)
	ginApp.initAuthorization(r)
	ginApp.initAny(r)
	ginApp.initDomainHandlers(r)
	ginApp.initDuplexHandlers(r)

	ginApp.gin = r
}

func (ginApp *GinApp) initDefaultHandlers(r *gin.Engine) {
	r.Use(gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		req := NewRequest(c)
		req.Negotiate(http.StatusNotFound, nil, NotFound)
		c.Abort()
	})

	r.Use(helmet.Default())
	r.Use(limit.MaxAllowed(100))
	c := ginApp.GetCors()
	if len(c) != 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     c,
			AllowMethods:     []string{"*", "POST", "GET", "PUT", "DELETE", "PATCH"},
			AllowHeaders:     []string{"*", "Authorization"},
			ExposeHeaders:    []string{"*"},
			AllowCredentials: true,
			MaxAge:           1 * time.Hour,
		}))
	}

	if gin.Mode() != gin.ReleaseMode {
		r.Use(gin.Logger())
	}
}

func (ginApp *GinApp) enableOpenApiIfRequired(r *gin.Engine) {
	if ginApp.jsonOpenApi == "" {
		return
	}

	r.GET(OpenApiRoute, func(ctx *gin.Context) {
		_, _ = ctx.Writer.Write([]byte(ginApp.jsonOpenApi))
		ctx.Status(http.StatusOK)
	})
	fsRoot, _ := fs.Sub(swaggerDirectory, AssetSwagger)
	r.StaticFS(ginApp.openApiRoute, http.FS(fsRoot))
}

func (ginApp *GinApp) initAuthentication(r *gin.Engine) {
	r.Use(ginApp.AuthenticationHandler)
}

func (ginApp *GinApp) AuthenticationHandler(c *gin.Context) {
	route := GetRegisteredRoute(c, BaseApiURL)
	if ginApp.router.IsFree(route, httpapi.HTTPMethod(c.Request.Method)) {
		return
	}
	var authenticator httpapi.Authenticator
	for k, v := range ginApp.authenticators {
		if strings.HasPrefix(route, k) {
			authenticator = v
			break
		}
	}
	req := NewRequest(c)
	if authenticator == nil {
		return
	}

	reqToken := ""

	if c.Query(Token) == "" {
		reqToken = c.GetHeader(Authorization)
	} else {
		reqToken = BearerSpace + c.Query(Token)
	}

	if reqToken == "" {
		req.SetUnauthorized(MissingAuthHeader, response.UnknownFormat)
		return
	}

	splitToken := strings.Split(reqToken, " ")

	if splitToken == nil || len(splitToken) != 2 || splitToken[0] != Bearer {
		req.SetUnauthorized(UnknownAuthFormat, response.UnknownFormat)
		return
	}

	var err error
	var m misc.JwtClaim
	m, err = authenticator.GetModel(splitToken[1])
	if err != nil {
		req.SetUnauthorized(CouldNotParseToken, response.UnknownFormat)
		return
	}

	if m.IsExpired() {
		req.SetUnauthorized(TTLExpired, response.SessionExpired)
		return
	}

	var valid bool
	valid, err = authenticator.CheckToken(splitToken[1])
	if err != nil {
		req.SetServerError(err.Error())
		return
	}

	if !valid {
		req.SetUnauthorized(TTLExpired, response.SessionExpired)
		return
	}

	req.Set(httpapi.KeyAuth, m)
}

func (ginApp *GinApp) initRouter() {
	for _, module := range ginApp.ginDomainHandlers {
		reqDefs := module.GetRequestHandlers()
		for _, v := range reqDefs {
			ginApp.router.Register(v, module.GetBaseURL())
		}
	}
}

func (ginApp *GinApp) initAuthorization(r *gin.Engine) {
	r.Use(ginApp.AuthorizationHandler)
}

func (ginApp *GinApp) initAny(r *gin.Engine) {
	r.Use(ginApp.AnyReq)
}

func (ginApp *GinApp) GetHandlerFunc(route, method string) gin.HandlerFunc {
	routeInfo := ginApp.gin.Routes()

	for _, v := range routeInfo {
		if v.Method == method && v.Path == route {
			return v.HandlerFunc
		}
	}
	return nil
}

func ToGinHandler(act httpapi.Action) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		act(NewRequest(ctx))
	}
}

func (ginApp *GinApp) initDomainHandlers(r *gin.Engine) {
	for _, domainHandler := range ginApp.ginDomainHandlers {

		reqData := domainHandler.GetRequestHandlers()
		baseURL := domainHandler.GetBaseURL()

		for _, v := range reqData {

			if v.Handler == nil {
				panic("nil handler is detected")
			}
			hnd := ToGinHandler(v.Handler)

			if v.AnyPermissions != nil {
				for key := range v.AnyPermissions {
					ginApp.PermissionManager.SetPermission(v.Method, baseURL+v.Route, v.AnyPermissions[key])
				}
			} else {
				ginApp.PermissionManager.SetFreePermission(v.Route, v.Method)
			}

			group := r.Group(baseURL)

			if chain := ginApp.preHandlers[baseURL]; chain != nil {
				var use func(middleware ...gin.HandlerFunc) gin.IRoutes
				if baseURL == "" {
					use = r.Use
				} else {
					use = group.Use
				}
				for _, handler := range chain {
					use(ToGinHandler(handler))
				}
			}

			for _, v := range v.Parameters {
				if misc.ReservedQueryMap[v.Definition.GetName()] {
					panic("using reserved query definition")
				}
			}

			switch v.Method {
			case http.MethodPost:
				group.POST(v.Route, hnd)
			case http.MethodDelete:
				group.DELETE(v.Route, hnd)
			case http.MethodGet:
				group.GET(v.Route, hnd)
			case http.MethodPut:
				group.PUT(v.Route, hnd)
			}
		}

	}
}

func toTopicMap(dhds []*httpapi.DuplexHandlerDefinition) (out map[string]*httpapi.DuplexHandlerDefinition) {
	out = map[string]*httpapi.DuplexHandlerDefinition{}
	for k, v := range dhds {
		out[v.Topic] = dhds[k]
	}
	return out
}

func loop(duplexCon *wsmodel.DuplexConnection, topics map[string]*httpapi.DuplexHandlerDefinition) <-chan error {
	echan := make(chan error)

	// TODO: must handle may ws features like ping pong , write/read deadline and so on
	go func(echam chan error) {
		for {
			// registering the connection
			_, jdmbyte, err := duplexCon.ReadMessage()
			if err != nil {
				echam <- err
				logger.Trace.Println(err.Error())
				return
			}
			// getting socket dto
			var jdm wsmodel.JsonDtoMessage
			if err := json.Unmarshal(jdmbyte, &jdm); err != nil {
				_ = duplexCon.SendError(ProtocolNegotiationFailed, ProtocolNegotiationFailedDescription)
				continue
			}

			// geting topic definition
			definition := topics[jdm.Topic]
			if definition == nil {
				if err := duplexCon.SendError(TopicNotSupported, TopicNotSupportedDescription); err != nil {
					return
				}
				continue
			}

			var dto any
			// checking dto
			if definition.Dto != nil {
				dtoType := reflect.TypeOf(definition.Dto).Elem()
				v := reflect.New(dtoType)
				dto = v.Interface()

				if err = json.Unmarshal([]byte(jdm.Dto), dto); err != nil {
					_ = duplexCon.SendError(response.UnknownFormat, err.Error())
					continue
				}
				verifier, ok := dto.(httpapi.Verifier)
				if ok {
					if err := verifier.Verify(); err != nil {
						_ = duplexCon.SendError(response.ValidationError, err.Error())
						continue
					}
				}
			}

			definition.Handler(wsmodel.NewDuplexMessage(definition.Topic, dto, nil))
		}
	}(echan)
	return echan
}

func (ginApp *GinApp) initDuplexHandlers(r *gin.Engine) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// TODO: must validate the cors
		return true
	}
	for _, duplexHandler := range ginApp.ginDuplexHandlers {
		handlers := duplexHandler.GetDuplexHandlers()
		baseURL := duplexHandler.GetBaseURL()
		topicMap := toTopicMap(handlers)
		r.GET(baseURL, func(c *gin.Context) {
			callerObj, _ := c.Get(httpapi.KeyAuth)
			claim := callerObj.(misc.JwtClaim)

			conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
			if err != nil {
				c.AbortWithStatus(http.StatusUpgradeRequired)
				return
			}
			defer conn.Close()
			duplexCon := wsmodel.NewDuplexConnection(conn, claim)
			duplexHandler.OnDuplexConnected(duplexCon)
			defer duplexHandler.OnDuplexDisconnected(duplexCon)

			timeToDisconnect := claim.GetExpireTime() - time.Now().Unix()
			select {

			case <-time.After(time.Duration(timeToDisconnect * int64(time.Second))):
				_ = duplexCon.SendError(wsmodel.TokenError, wsmodel.ConnectionWasForJwtExpired)
				duplexCon.Close()
			case <-time.After(time.Duration(time.Hour * 2)): // TODO: time must be injected with given policy , no connection can persist more than given time
				_ = duplexCon.SendError(wsmodel.TimeoutError, wsmodel.ConnectionWasForTooLong)
				duplexCon.Close()
			case err := <-loop(duplexCon, topicMap):
				logger.Trace.Println(err.Error())
			}
			logger.Trace.Println("done")
		})
	}
}

func GetRegisteredRoute(c *gin.Context, base string) string {
	url := fmt.Sprintf("%+v", c.Request.URL)
	url = strings.ReplaceAll(url, base, "")
	url = strings.ReplaceAll(url, "?"+c.Request.URL.RawQuery, "")

	for _, p := range c.Params {
		url = strings.Replace(url, p.Value, ":"+p.Key, 1)
	}

	return url
}

func (ginApp *GinApp) AuthorizationHandler(c *gin.Context) {
	route := GetRegisteredRoute(c, BaseApiURL)
	if ginApp.PermissionManager.IsFree(route, httpapi.HTTPMethod(c.Request.Method)) {
		return
	}

	if ginApp.router.IsFree(route, httpapi.HTTPMethod(c.Request.Method)) {
		return
	}
	var authorize httpapi.Authorizer
	for k, v := range ginApp.authorizers {
		if strings.HasPrefix(route, k) {
			authorize = v
			break
		}
	}

	if authorize == nil {
		return
	}

	req := NewRequest(c)

	auth, _ := req.Get(httpapi.KeyAuth)
	model, _ := auth.(misc.JwtClaim)

	routePerms := ginApp.PermissionManager.GetPermission(route, httpapi.HTTPMethod(c.Request.Method))

	for _, v := range routePerms {
		valid, err := authorize(model.GetSubject(), v)
		if err != nil {
			req.SetServerError(err.Error())
			return
		}
		if !valid {
			req.SetForbidden()
			return
		}
	}
}

func (ginApp *GinApp) AnyReq(c *gin.Context) {
	route := GetRegisteredRoute(c, BaseApiURL)

	fullRoute := ginApp.router.GetRoute(route, httpapi.HTTPMethod(c.Request.Method))

	req := NewRequest(c)
	if fullRoute == nil {
		return
	}
	req.Set(httpapi.MaxLimit, fullRoute.MaxLimit)

	if fullRoute.Dto != nil {

		dtoType := reflect.TypeOf(fullRoute.Dto).Elem()
		v := reflect.New(dtoType)
		copy := v.Interface()

		if err := c.ShouldBind(copy); err != nil {
			req.SetBadRequest(err.Error(), response.UnknownFormat)
			return
		}

		verifier, ok := copy.(httpapi.Verifier)
		if ok {
			if err := verifier.Verify(); err != nil {
				req.SetBadRequest(err.Error(), response.ValidationError)
				return
			}
		}
		req.Set(httpapi.KeyDTO, copy)
	}

	if fullRoute.DtoArray != nil {

		dtoType := reflect.TypeOf(fullRoute.DtoArray)
		slice := reflect.SliceOf(dtoType)

		copy := reflect.New(slice).Interface()

		if err := c.ShouldBind(copy); err != nil {
			req.SetBadRequest(err.Error(), response.UnknownFormat)
			return
		}

		dt := reflect.ValueOf(copy).Elem().Interface()

		s := reflect.ValueOf(dt)
		for i := 0; i < s.Len(); i++ {
			verifier, ok := s.Index(i).Interface().(httpapi.Verifier)
			if ok {
				if err := verifier.Verify(); err != nil {
					req.SetBadRequest(err.Error(), response.ValidationError)
					return
				}
			}
		}

		req.Set(httpapi.KeyDTO, dt)
	}

	for _, v := range fullRoute.Parameters {

		var value string

		if v.Query {
			value = c.Query(v.Definition.GetName())
		} else {
			value = c.Param(v.Definition.GetName())
		}

		if value == "" {
			if !v.Optional {
				req.SetBadRequest(MissingParamWithName+v.Definition.GetName(), response.UnknownFormat)
				return
			}
			continue
		}
		var val misc.Query
		var err error
		if val, err = ParseQuery(v.Definition, value); err != nil {
			req.SetBadRequest(err.Error(), response.UnknownFormat)
			return
		}

		req.Set(v.Definition.GetName(), val.GetOperand().Value)
		if !v.Query {
			continue
		}

		data, ok := req.Get(httpapi.KeyQuery)
		if !ok {
			req.Set(httpapi.KeyQuery, []misc.Query{val})
		} else {
			v := data.([]misc.Query)
			v = append(v, val)
			req.Set(httpapi.KeyQuery, v)
		}
	}

	if len(fullRoute.FileParts) == 0 && len(fullRoute.ValueParts) == 0 {
		return
	}

	form, err := ginApp.readForm(c.Request)
	if err != nil {
		req.SetBadRequest(ExpectMultipartFormData, response.UnknownFormat)
		return
	}

	for _, v := range fullRoute.FileParts {

		files, ok := form.File[v.GetPartName()]

		if !v.IsOptional() && !ok {
			req.SetBadRequest(fmt.Sprintf(MissingPart, v.GetPartName()), response.ValidationError)
			return
		}

		// TODO: Must write test to address this issue
		if len(files) == 0 {
			continue
		}

		out := []httpapi.FileHeader{}
		for _, f := range files {

			if len(out) == 1 && v.IsSingle() {
				req.SetBadRequest(SingleSupport, response.ValidationError)
				return
			}
			err := v.Verify(f)
			if err != nil {
				req.SetBadRequest(err.Error(), response.ValidationError)
				return
			}
			out = append(out, f)
		}
		if v.IsSingle() {
			req.SetFile(v.GetPartName(), out[0])
		} else {
			req.SetFiles(v.GetPartName(), out)
		}
	}

	for _, v := range fullRoute.ValueParts {

		values, ok := form.Value[v.GetPartName()]

		if !v.IsOptional() && !ok {
			req.SetBadRequest(fmt.Sprintf(MissingPart, v.GetPartName()), response.ValidationError)
			return
		}

		out := []any{}
		for _, dto := range values {

			if len(out) == 1 && v.IsSingle() {
				req.SetBadRequest(SingleSupport, response.ValidationError)
				return
			}
			ct := dto.Header.Get(CType)

			if ct == "" {
				ct = gin.MIMEJSON
			}

			m, _, err := mime.ParseMediaType(ct)
			if err != nil {
				req.SetBadRequest(err.Error(), response.UnknownFormat)
				return
			}
			binder := binding.Default(c.Request.Method, m)

			bodyBinder, _ := binder.(binding.BindingBody)

			dtoType := reflect.TypeOf(v.GetObject()).Elem()
			v := reflect.New(dtoType)
			copy := v.Interface()
			if err := bodyBinder.BindBody(dto.Data, &copy); err != nil {
				req.SetBadRequest(err.Error(), response.UnknownFormat)
				return
			}
			verifier, ok := copy.(httpapi.Verifier)
			if ok {
				if err := verifier.Verify(); err != nil {
					req.SetBadRequest(err.Error(), response.ValidationError)
					return
				}
			}
			out = append(out, copy)
		}
		if v.IsSingle() {
			req.Set(v.GetPartName(), out[0])
		} else {
			req.Set(v.GetPartName(), out)
		}

	}
}

func (ginApp *GinApp) LogRequests(c *gin.Context) {
	start := time.Now()

	hLog := &model.HttpLog{
		Request:         &model.Request{Time: start},
		RequestBody:     &model.Body{},
		RequestHeaders:  []*model.Header{},
		ResponseHeaders: []*model.Header{},
		ResponseBody:    &model.Body{},
		Response:        &model.Response{},
	}

	hLog.Request.Size = CalcRequestSize(c.Request)
	hLog.Request.IP = c.ClientIP()
	hLog.Request.Method = c.Request.Method
	hLog.Request.Route = GetRegisteredRoute(c, BaseApiURL)
	hLog.Request.Path = fmt.Sprintf("%+v", c.Request.URL)

	for k, v := range c.Request.Header {
		value := ""
		for _, v2 := range v {
			value += v2
		}
		h := model.Header{Name: k, Value: value}
		hLog.RequestHeaders = append(hLog.RequestHeaders, &h)
	}

	blw := &BodyLogWriter{body: bytes.NewBuffer([]byte{}), ResponseWriter: c.Writer}

	if ginApp.GetLogPolicy().LogBody {

		buf, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		hLog.RequestBody = &model.Body{Data: buf}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))
		c.Writer = blw
	}
	c.Next()

	if ginApp.GetLogPolicy().LogBody {
		bdr := model.Body{Data: blw.body.Bytes()}
		hLog.RequestBody = &bdr
	}

	hLog.Request.ExecutionNanoSecond = time.Since(start).Nanoseconds()
	hLog.Response.Size = int64(c.Writer.Size())
	hLog.Response.Status = uint(c.Writer.Status())

	it := c.Writer.Header()
	for k, v := range it {
		value := ""
		for _, v2 := range v {
			value += v2
		}
		h := model.Header{Name: k, Value: value}
		hLog.ResponseHeaders = append(hLog.ResponseHeaders, &h)
	}

	handler := ginApp.GetLogHandler()

	handler.Handle(*hLog)
}

func CalcRequestSize(r *http.Request) int64 {
	size := 0
	if r.URL != nil {
		size = len(r.URL.String())
	}

	size += len(r.Method)
	size += len(r.Proto)

	for name, values := range r.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}
	size += len(r.Host)

	// r.Form and r.MultipartForm are assumed to be included in r.URL.
	if r.ContentLength != -1 {
		size += int(r.ContentLength)
	}
	return int64(size)
}

type BodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w BodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GinBodyLogMiddleware(c *gin.Context) {
	blw := &BodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	statusCode := c.Writer.Status()
	if statusCode >= 400 {
		// ok this is an request with error, let's make a record for it
		// now print body (or log in your preferred way)
		fmt.Println("Response body: " + blw.body.String())
	}
}

func (ginApp *GinApp) TestHandle(recorder *httptest.ResponseRecorder, request *http.Request) error {
	ginApp.gin.ServeHTTP(recorder, request)
	return nil
}
