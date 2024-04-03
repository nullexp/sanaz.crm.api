package gin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	httpapi "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/model/openapi"
	response "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol/response"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
)

const (
	ErrFieldNotSupported = "requested field is not supported"

	ApiInfoMustNotBeEmpty       = "api info should not be empty"
	URLMustNotBeEmpty           = "url should not be empty"
	TagIsMandetory              = "tag is mandatory"
	PathCannotHaveTwoSameMethod = "a path cannot have two  similar methods"
	RepeatedResponseStatus      = "repeated status code for responses has been detected"

	Version                 = "version"
	OpenApi                 = "openapi"
	Description             = "description"
	Summary                 = "summary"
	OperationId             = "operationId"
	Title                   = "title"
	Contact                 = "contact"
	Info                    = "info"
	Email                   = "email"
	Name                    = "name"
	URL                     = "url"
	OpenApiVersion          = "3.0.3"
	ExternalDocs            = "externalDocs"
	Servers                 = "servers"
	Tags                    = "tags"
	Paths                   = "paths"
	SchemaLocation          = "#/components/schemas/"
	Content                 = "content"
	ApplicationJson         = "application/json"
	Json                    = "json"
	Scheme                  = "scheme"
	Schema                  = "schema"
	Ref                     = "$ref"
	Type                    = "type"
	Array                   = "array"
	Items                   = "items"
	RequestBody             = "requestBody"
	Required                = "required"
	BearerAuth              = "bearerAuth"
	Responses               = "responses"
	Security                = "security"
	In                      = "in"
	Query                   = "query"
	Path                    = "path"
	String                  = "string"
	Parameters              = "parameters"
	Format                  = "format"
	Example                 = "example"
	Object                  = "object"
	Properties              = "properties"
	SecuritySchemes         = "securitySchemes"
	HTTP                    = "http"
	JWT                     = "JWT"
	SmallBearer             = "bearer"
	BearerFormat            = "bearerFormat"
	Components              = "components"
	Schemas                 = "schemas"
	Error                   = "Error"
	ErrorCode               = "ErrorCode"
	ErrorCodeMessageExample = "Supplied message for developers only"
	Enum                    = "enum"

	Time                   = "Time"
	JsonTagSeperator       = ","
	FourOFour              = "404"
	FourHundred            = "400"
	IfClientErrorOccured   = "If any client error occured, 400 with error object will returned"
	IfResousrceWasNotFound = "If any wanted resource was not found to perform required operation, 404 with error object will returned"
)

var (
	ErrApiInfoShouldNotBeEmpty     = errors.New(ApiInfoMustNotBeEmpty)
	ErrURLMustNotBeEmpty           = errors.New(URLMustNotBeEmpty)
	ErrTagIsMandetory              = errors.New(TagIsMandetory)
	ErrPathCannotHaveTwoSameMethod = errors.New(PathCannotHaveTwoSameMethod)
	ErrRepeatedResponseStatus      = errors.New(RepeatedResponseStatus)
)

func (ginApp *GinApp) SetInfo(appInfo openapi.Info) {
	ginApp.ApiInfo = &appInfo
}

func (ginApp *GinApp) SetContact(contact openapi.Contact) {
	ginApp.Contact = &contact
}

func (ginApp *GinApp) SetServers(servers []openapi.Server) {
	ginApp.Servers = servers
}

func (ginApp *GinApp) SetExternalDocs(docs openapi.ExternalDocs) {
	ginApp.ExternalDocs = &docs
}

func (ginApp *GinApp) SetErrors(errors []string) {
	ginApp.Errors = errors
}

func (ginApp *GinApp) EnableOpenApi(route string) (err error) {
	ginApp.openApiRoute = route

	ginApp.jsonOpenApi, err = ginApp.generateSwaggerAsJson()
	if err != nil {
		return err
	}
	return
}

func (ginApp *GinApp) generateSwaggerAsJson() (out string, err error) {
	data, err := ginApp.generateSwagger()
	if err != nil {
		return
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	return string(raw), err
}

func (ginApp *GinApp) generateSwagger() (out map[string]any, err error) {
	out = map[string]any{}

	// Setting open api version
	out[OpenApi] = OpenApiVersion

	// Api info is mandatory
	if ginApp.ApiInfo == nil {
		return nil, ErrApiInfoShouldNotBeEmpty
	}
	assignServerInfo(out, *ginApp.ApiInfo)

	// contact is optional
	if ginApp.Contact != nil {
		assignContactInfo(out, *ginApp.Contact)
	}

	if ginApp.ExternalDocs != nil {

		if ginApp.ExternalDocs.URL == "" {
			return nil, ErrURLMustNotBeEmpty
		}
		assignExternalDoc(out, *ginApp.ExternalDocs)
	}

	if ginApp.Servers != nil {
		err = assignServers(out, ginApp.Servers)
		if err != nil {
			return
		}
	}

	if len(ginApp.ginDomainHandlers) != 0 {
		err = assignTags(out, ginApp.ginDomainHandlers)
		if err != nil {
			return
		}

		err = assignPaths(out, ginApp.ginDomainHandlers)
		if err != nil {
			return
		}

		allErrors := response.GetErrors()
		if len(ginApp.Errors) != 0 {
			allErrors = append(allErrors, ginApp.Errors...)
		}
		err = assignComponents(out, allErrors, ginApp.ginDomainHandlers)

	}
	return
}

func assignComponents(src map[string]any, errors []string, modules []httpapi.Module) (err error) {
	components := map[string]any{}

	assignSecuritySchemas(components)

	schemas := map[string]any{}

	assignSchemas(schemas, modules)
	assignDefaultSchemas(schemas, errors)

	components[Schemas] = schemas

	src[Components] = components
	return
}

func assignSchemas(schemas map[string]any, modules []httpapi.Module) {
	for _, module := range modules {
		for _, reqHandler := range module.GetRequestHandlers() {

			for _, responseDef := range reqHandler.ResponseDefinitions {
				if responseDef.Dto != nil {
					assignSchemaIfNotExist(responseDef.Dto, schemas, http.MethodGet)
				}
			}
			if reqHandler.Dto != nil {
				assignSchemaIfNotExist(reqHandler.Dto, schemas, reqHandler.Method)
			}
		}
	}
}

func assignDefaultSchemas(schemas map[string]any, errors []string) {
	// Setting error
	schemas[Error] = map[string]any{
		Type: Object,
		Properties: map[string]any{
			"code":    map[string]any{Ref: SchemaLocation + ErrorCode},
			"message": map[string]any{Type: String, Description: ErrorCodeMessageExample},
		},
	}
	schemas[ErrorCode] = map[string]any{
		Type: String,
		Enum: errors,
	}
}

func assignSecuritySchemas(component map[string]any) {
	component[SecuritySchemes] = map[string]any{
		BearerAuth: map[string]string{
			Type:         HTTP,
			Scheme:       SmallBearer,
			BearerFormat: JWT,
		},
	}
}

func assignServerInfo(src map[string]any, info openapi.Info) {
	infoData := map[string]any{}
	infoData[Version] = info.Version
	infoData[Description] = info.Description
	infoData[Title] = info.Title
	src[Info] = infoData
}

func assignContactInfo(src map[string]any, contact openapi.Contact) {
	info := src[Info].(map[string]any)
	contactData := map[string]any{}
	contactData[Name] = contact.Name
	contactData[Email] = contact.Email
	contactData[URL] = contact.URL
	info[Contact] = contactData
}

func assignExternalDoc(src map[string]any, exDoc openapi.ExternalDocs) {
	externalDocs := map[string]any{}
	externalDocs[Description] = exDoc.Description
	externalDocs[URL] = exDoc.URL
	src[ExternalDocs] = externalDocs
}

func assignServers(src map[string]any, servers []openapi.Server) (err error) {
	serversNode := []map[string]any{}

	for _, v := range servers {

		if v.URL == "" {
			return ErrURLMustNotBeEmpty
		}

		sr := map[string]any{}
		sr[URL] = v.URL
		sr[Description] = v.Description
		serversNode = append(serversNode, sr)
	}

	src[Servers] = serversNode
	return
}

func assignTags(src map[string]any, modules []httpapi.Module) (err error) {
	tags := []map[string]any{}

	for _, v := range modules {

		tag := v.GetTag()

		out, err := getTag(tag)
		if err != nil {
			return nil
		}
		tags = append(tags, out)
	}

	src[Tags] = tags
	return
}

func getTag(tag openapi.Tag) (out map[string]any, err error) {
	if tag.Name == "" {
		err = ErrTagIsMandetory
		return
	}

	out = map[string]any{}
	out[Name] = tag.Name
	out[Description] = tag.Description

	if tag.ExternalDocs != nil {
		ed := map[string]any{}

		if tag.ExternalDocs.URL == "" {
			err = ErrURLMustNotBeEmpty
			return
		}
		ed[URL] = tag.ExternalDocs.URL
		ed[Description] = tag.ExternalDocs.Description
		out[ExternalDocs] = ed
	}
	return
}

func assignPaths(src map[string]any, modules []httpapi.Module) (err error) {
	// map of paths , /users/{id} = > get , post ,delete
	paths := map[string]map[string]any{}

	for _, v := range modules {
		for _, handler := range v.GetRequestHandlers() {
			route := v.GetBaseURL() + handler.Route
			// "/users"   / post , get
			openapiPath := mapGinParamToOpenApiPath(route)
			pathMethods, ok := paths[openapiPath]

			if !ok {
				pathMethods = map[string]any{}   // Get => def , Post => def
				paths[openapiPath] = pathMethods // register the path in all paths
			}

			// By standard, open api http methods are lower cased
			// I TRUST the developer that does not use custom non http methods
			// TODO: validate  methods
			openapiMethod := strings.ToLower(string(handler.Method))
			_, ok = pathMethods[openapiMethod]
			if ok {
				err = ErrPathCannotHaveTwoSameMethod
				return
			}
			pathMethods[openapiMethod], err = getMethod(handler, v.GetTag())

			if err != nil {
				return
			}

		}
	}

	src[Paths] = paths
	return
}

func getMethod(def *httpapi.RequestDefinition, tag openapi.Tag) (out map[string]any, err error) {
	out = map[string]any{}

	out[Tags] = []string{tag.Name}
	out[Description] = def.Description
	out[Summary] = def.Summary
	out[OperationId] = def.OperationId

	reqbody := getRequestBody(def)
	if reqbody != nil {
		out[RequestBody] = reqbody
	}

	responses, err := getResponses(def)
	if err != nil {
		return
	}
	if responses != nil {
		out[Responses] = responses
	}

	if !def.FreeRoute {
		out[Security] = getMethodSecurity()
	}

	if len(def.Parameters) != 0 {
		params := []map[string]any{}
		for _, v := range def.Parameters {
			// TODO: guard agaist repeated param name
			params = append(params, parseOpenApiQuery(v))
		}
		out[Parameters] = params
	}

	return
}

// Only or dto array is supported
func getRequestBody(def *httpapi.RequestDefinition) (out map[string]any) {
	out = map[string]any{}
	out[Required] = true
	if def.Dto != nil {
		location := getDtoComponentLocation(def.Dto, def.Method)
		out[Content] = getContentWithlocation(location)
		return
	}
	if def.DtoArray != nil {
		location := getDtoComponentLocation(def.Dto, def.Method)
		out[Content] = getContentArrayWithlocation(location)
		return
	}

	return nil
}

func getContentWithlocation(location string) map[string]any {
	return map[string]any{
		ApplicationJson: map[string]any{
			Schema: map[string]any{
				Ref: location,
			},
		},
	}
}

func getContentArrayWithlocation(location string) map[string]any {
	return map[string]any{
		ApplicationJson: map[string]any{
			Schema: map[string]any{
				Type: Array,
				Items: map[string]any{
					Ref: location,
				},
			},
		},
	}
}

func getDefaultBadRequestResponse() (out map[string]any) {
	out = map[string]any{
		Description: IfClientErrorOccured,
		Content:     getContentWithlocation(SchemaLocation + Error),
	}
	return
}

func getDefaultNotFoundResponse() (out map[string]any) {
	out = map[string]any{
		Description: IfResousrceWasNotFound,
		Content:     getContentWithlocation(SchemaLocation + Error),
	}
	return
}

func getResponseBody(def httpapi.ResponseDefinition) (out map[string]any) {
	out = map[string]any{}
	out[Description] = def.Description

	if def.Dto == nil {
		return
	}

	rt := reflect.TypeOf(def.Dto)

	kind := rt.Kind()

	if kind == reflect.Pointer {
		rt = rt.Elem()
		kind = rt.Kind()
	}
	switch kind {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		rt = rt.Elem()
		location := getDtoComponentLocation(reflect.New(rt).Interface(), http.MethodGet)
		out[Content] = getContentArrayWithlocation(location)
		return
	case reflect.Struct:
		location := getDtoComponentLocation(def.Dto, http.MethodGet)
		out[Content] = getContentWithlocation(location)
		return
	default:
		return nil
	}
}

func getResponses(def *httpapi.RequestDefinition) (out map[string]any, err error) {
	if len(def.ResponseDefinitions) == 0 {
		return
	}
	out = map[string]any{}

	for _, v := range def.ResponseDefinitions {
		key := strconv.Itoa(v.Status)
		_, ok := out[key]

		if ok {
			err = ErrRepeatedResponseStatus
			return
		}
		out[key] = getResponseBody(v)
	}

	// Adding default response body
	if _, ok := out[FourHundred]; !ok {
		out[FourHundred] = getDefaultBadRequestResponse()
	}

	if _, ok := out[FourOFour]; !ok {
		out[FourOFour] = getDefaultNotFoundResponse()
	}

	return
}

func getDtoName(dto any, method httpapi.HTTPMethod) string {
	if v, ok := (dto).(openapi.NameGetter); ok {
		return v.GetName()
	}

	return getDtoTypeName(reflect.TypeOf(dto), method)
}

func getDtoTypeName(dtoType reflect.Type, method httpapi.HTTPMethod) string {
	methodName := string(method)
	methodName = strings.ToLower(methodName)
	methodName = strings.ToUpper(methodName[:1]) + methodName[1:]

	t := dtoType
	kind := t.Kind()

	if kind == reflect.Ptr {
		t = t.Elem()
		kind = t.Kind()
	}

	if kind == reflect.Slice || kind == reflect.Array {
		t = t.Elem()
	}

	return methodName + t.Name()
}

func getDtoComponentLocation(dto any, method httpapi.HTTPMethod) string {
	return getComponentLocation(getDtoName(dto, method))
}

func getComponentLocation(name string) string {
	return SchemaLocation + name
}

func getMethodSecurity() (out []map[string]any) {
	out = make([]map[string]any, 1)
	out[0] = map[string]any{}
	out[0][BearerAuth] = []string{}
	return out
}

const (
	GinParamSignature      = ":"
	UrlSeperator           = "/"
	OpenApiParameterFormat = "{%s}"
)

// /users/:id => /users/{id}
// /:something-else => /{something-else}
// TODO: not an optimal way!
func mapGinParamToOpenApiPath(route string) (mappedRoute string) {
	if strings.HasPrefix(route, UrlSeperator) {
		mappedRoute += UrlSeperator
	}

	parts := strings.Split(route, "/")

	for k, v := range parts {
		if v == "" {
			continue
		}
		if strings.HasPrefix(v, GinParamSignature) {
			param := strings.TrimPrefix(v, GinParamSignature)
			mappedRoute += fmt.Sprintf(OpenApiParameterFormat, param)
		} else {
			mappedRoute += v
		}
		if k != len(parts)-1 {
			mappedRoute += UrlSeperator
		}
	}
	return mappedRoute
}

func parseOpenApiQuery(param httpapi.RequestParameter) (out map[string]any) {
	out = map[string]any{}
	out[Name] = param.Definition.GetName()
	out[Description] = param.Definition.GetDescription()
	out[Required] = !param.Optional

	if param.Query {
		out[In] = Query
		if len(param.Definition.GetSupportedOperators()) != 0 {
			supportedOps := "Supports: "
			for _, op := range param.Definition.GetSupportedOperators() {
				supportedOps += string(op) + " "
			}
			supportedOps += "."

			desc := ""
			if param.Definition.GetDescription() != "" {
				desc = param.Definition.GetDescription() + "." + supportedOps
				out[Description] = desc
			} else {
				out[Description] = supportedOps
			}
		}
		out[Schema] = map[string]any{Type: String}

	} else {
		out[In] = Path
		scm := getSchemaTypeFromDataType(param.Definition.GetType())
		out[Schema] = map[string]any{Type: scm.Kind, Format: scm.Format}
	}

	return out
}

type paramSchemaType struct {
	Format string
	Kind   string
}

const (
	Integer  = "integer"
	Number   = "number"
	Boolean  = "boolean"
	Int32    = "int32"
	Int64    = "int64"
	DateTime = "date-time"
	Float    = "float"
	Double   = "double"
	Password = "password"
)

func getSchemaTypeFromDataType(dt misc.DataType) paramSchemaType {
	switch dt {
	case misc.DataTypeString:
		return paramSchemaType{Kind: String}
	case misc.DataTypeBoolean:
		return paramSchemaType{Kind: Boolean}
	case misc.DataTypeInteger:
		fallthrough
	case misc.DataTypeUInteger:
		return paramSchemaType{Kind: Integer, Format: Int32}
	case misc.DataTypeULong:
		fallthrough
	case misc.DataTypeLong:
		return paramSchemaType{Kind: Integer, Format: Int64}
	case misc.DataTypeDouble:
		return paramSchemaType{Kind: Integer, Format: Int64}
	case misc.DataTypeTime:
		return paramSchemaType{Kind: String, Format: DateTime}
	case misc.DataTypeBase64:
		return paramSchemaType{Kind: String, Format: DateTime}
	}
	return paramSchemaType{}
}

func GetSchemaField(desc schemaTypeDescription, method httpapi.HTTPMethod) (out map[string]any) {
	out = map[string]any{}

	if desc.IsStruct {
		if desc.IsArray {
			out[Type] = Array
			dtoName := getDtoTypeName(desc.StructType, method)
			out[Items] = map[string]any{Ref: getComponentLocation(dtoName)}

		} else {
			dtoName := getDtoTypeName(desc.StructType, method)
			out[Ref] = getComponentLocation(dtoName)
		}
		return
	}

	if desc.IsArray {
		out[Type] = Array
		out[Items] = map[string]string{
			Type:        desc.Type,
			Format:      desc.Format,
			Example:     desc.Example,
			Description: desc.Description,
		}
	} else {
		out[Type] = desc.Type
		out[Format] = desc.Format
		out[Example] = desc.Example
		out[Description] = desc.Description
	}

	return
}

func assignSchemaIfNotExist(dto any, schemas map[string]any, method httpapi.HTTPMethod) {
	embededStructs := []schemaTypeDescription{}
	typ := reflect.TypeOf(dto).Elem()

	name := getDtoName(dto, method)

	if _, ok := schemas[name]; ok {
		return
	}
	objectDescripton := map[string]any{}
	objectDescripton[Type] = Object
	objectDescripton[Properties] = map[string]any{}
	for i := 0; i < typ.NumField(); i++ {

		f := typ.Field(i)
		description := getSchemaTypeFromStructField(f)

		if description.IsStruct {
			embededStructs = append(embededStructs, description)
		}

		fieldName := f.Name
		jsonTag, ok := f.Tag.Lookup(Json)

		if ok {
			fieldName = strings.Split(jsonTag, JsonTagSeperator)[0]
		}

		objectDescripton[Properties].(map[string]any)[fieldName] = GetSchemaField(description, method)
	}

	schemas[name] = objectDescripton
	for _, v := range embededStructs {
		assignSchemaIfNotExist(reflect.New(v.StructType).Interface(), schemas, method)
	}
}

// Will embed any structs
func GetBareSchema(dto any) (objectDescripton map[string]any) {
	typ := reflect.TypeOf(dto).Elem()

	objectDescripton = map[string]any{}
	objectDescripton[Type] = Object
	objectDescripton[Properties] = map[string]any{}
	for i := 0; i < typ.NumField(); i++ {

		f := typ.Field(i)
		description := getSchemaTypeFromStructField(f)

		fieldName := f.Name
		jsonTag, ok := f.Tag.Lookup(Json)

		if ok {
			fieldName = strings.Split(jsonTag, JsonTagSeperator)[0]
		}

		if description.IsStruct {
			objectDescripton[Properties].(map[string]any)[fieldName] = GetBareSchema(reflect.New(description.StructType).Interface())
		} else {
			objectDescripton[Properties].(map[string]any)[fieldName] = GetSchemaField(description, "")
		}

	}

	return objectDescripton
}

type schemaTypeDescription struct {
	Format string // openapi format
	Type   string // openapi type
	// meta data for decision making
	IsStruct    bool
	StructType  reflect.Type
	Description string
	Example     string
	IsArray     bool
}

var setInteger32 = func(out *schemaTypeDescription) {
	out.Type = Integer
	out.Format = Int32
}

var setDateTime = func(out *schemaTypeDescription) {
	out.Type = String
	out.Format = DateTime
}

var setInteger64 = func(out *schemaTypeDescription) {
	out.Type = Integer
	out.Format = Int64
}

var setNumberFloat = func(out *schemaTypeDescription) {
	out.Type = Number
	out.Format = Float
}

var setNumberDouble = func(out *schemaTypeDescription) {
	out.Type = Number
	out.Format = Double
}

var setBareString = func(out *schemaTypeDescription) {
	out.Type = String
}

// TODO: support password
var SetStringPassword = func(out *schemaTypeDescription) {
	out.Type = String
	out.Format = Double
}

var setBoolean = func(out *schemaTypeDescription) {
	out.Type = Boolean
}

// TODO: support password
var typeConverter = map[reflect.Kind]func(out *schemaTypeDescription){
	reflect.Bool:    setBoolean,
	reflect.Int:     setInteger32,
	reflect.Int8:    setInteger32,
	reflect.Int16:   setInteger32,
	reflect.Int32:   setInteger32,
	reflect.Int64:   setInteger64,
	reflect.Uint:    setInteger32,
	reflect.Uint8:   setInteger32,
	reflect.Uint16:  setInteger32,
	reflect.Uint32:  setInteger32,
	reflect.Uint64:  setInteger64,
	reflect.Float32: setNumberFloat,
	reflect.Float64: setNumberDouble,
	reflect.String:  setBareString,
}

func getSchemaTypeFromStructField(field reflect.StructField) (out schemaTypeDescription) {
	out.Description, _ = field.Tag.Lookup(Description)
	out.Example, _ = field.Tag.Lookup(Example)

	fieldType := field.Type
	kind := fieldType.Kind()
	if kind == reflect.Array || kind == reflect.Slice {
		out.IsArray = true
		fieldType = field.Type.Elem()
		kind = fieldType.Kind()
	}

	if kind == reflect.Pointer {
		fieldType = field.Type.Elem()
		kind = fieldType.Kind()
	}

	// time.Time is reserved and is not considered a normal schema type
	if kind == reflect.Struct && fieldType.Name() != Time {

		out.IsStruct = true
		out.StructType = fieldType

		return
	}

	if fieldType.Name() == Time {
		setDateTime(&out)
		return
	}

	setter, ok := typeConverter[kind]

	if !ok {
		panic(ErrFieldNotSupported)
	}
	setter(&out)
	return
}
