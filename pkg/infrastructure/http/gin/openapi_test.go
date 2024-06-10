package gin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol/model/openapi"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"github.com/stretchr/testify/assert"
)

func NewTestModuleWithTag(base string, tag openapi.Tag, rdfs ...*protocol.RequestDefinition) *testModule {
	tm := testModule{}
	tm.RequestDefinitions = rdfs
	tm.BaseURL = base
	tm.Tag = tag
	return &tm
}

type testDto struct {
	Id             string    `json:"id" description:"id description" example:"1321231"`
	ValueUInt      uint      `json:"value" description:"id description" example:"32"`
	ValueInt       int       `json:"ValueInt" description:"id description" example:"12"`
	ValueFloat64   float64   `json:"valueFloat64" description:"id description" example:"64.5"`
	ValueFloat32   float32   `json:"valueFloat32" description:"id description" example:"5.5"`
	FullDate       time.Time `json:"fullDate" description:"id description" example:"1937-01-01T12:00:27.87+00:20"`
	ValueInt64     int64     `json:"valueInt64" description:"id description" example:"12"`
	ValueUIntArray []uint    `json:"valueUIntArray" description:"id description" example:"32"`
}

func TestGenerateOpenApi(t *testing.T) {
	const description = "description"
	const title = "title"
	const version = "1.0.0"
	const name = "voip team"
	const email = "voip@espad.ir"
	const url = "www.espad.ir"
	const TestPermission = "TestPermission"
	const RandomError = "RandomError"
	const TestPermission2 = "TestPermission2"

	t.Run("must parse single server modules with array dto return type", func(t *testing.T) {
		app := NewGinApp()

		app.SetInfo(openapi.Info{Version: version, Description: description, Title: title})
		app.SetContact(openapi.Contact{Name: name, Email: email, URL: url})
		app.SetErrors([]string{RandomError})
		app.SetExternalDocs(openapi.ExternalDocs{Description: description, URL: url})
		app.SetServers([]openapi.Server{
			{URL: url, Description: description},
			{URL: url, Description: description},
			{URL: url, Description: description},
		})

		app.AppendModule(NewTestModuleWithTag("/users", openapi.Tag{
			Name:         name,
			Description:  description,
			ExternalDocs: &openapi.ExternalDocs{URL: url, Description: description},
		}, &protocol.RequestDefinition{
			Route:          "/:id",
			Parameters:     []protocol.RequestParameter{protocol.ResourceIdParameter},
			MaxLimit:       100,
			Method:         http.MethodPost,
			AnyPermissions: []string{TestPermission, TestPermission2},
			Summary:        Summary,
			OperationId:    OperationId,
			Description:    description,
			Deprecated:     true,
			Dto:            &TestDto{},
			ResponseDefinitions: []protocol.ResponseDefinition{
				{Description: description, Status: http.StatusOK, Dto: []User{}},
			},
			Handler: func(req protocol.Request) {},
		}))

		data, err := app.generateSwagger()
		assert.Nil(t, err)
		if err != nil {
			return
		}

		assert.NotEmpty(t, data[Paths])
		if data[Paths] == nil {
			return
		}

		paths := data[Paths].(map[string]map[string]any)
		assert.NotEmpty(t, paths["/users/{id}"])

		post := paths["/users/{id}"]["post"].(map[string]any)

		assert.EqualValues(t, description, post[description])
		assert.EqualValues(t, OperationId, post[OperationId])
		assert.EqualValues(t, Summary, post[Summary])

		security := post[Security].([]map[string]any)
		assert.EqualValues(t, []string{}, security[0][BearerAuth])

		tags := post[Tags].([]string)
		assert.EqualValues(t, name, tags[0])

		responses := post[Responses].(map[string]any)

		statusOk := responses["200"].(map[string]any)
		assert.EqualValues(t, description, statusOk[Description])

		params := post[Parameters].([]map[string]any)

		assert.EqualValues(t, Path, params[0][In])
		assert.EqualValues(t, misc.Id, params[0][Name])
		assert.EqualValues(t, true, params[0][Required])

		js, err := json.Marshal(data)
		assert.Nil(t, err)
		str := string(js)
		_ = str
	})
}

func TestMapGinParamToOpenApiPath(t *testing.T) {
	tests := []struct {
		name            string
		route           string
		wantMappedRoute string
	}{
		{
			name:            "Must handle no param format",
			route:           "/users/somethingelse/somthing",
			wantMappedRoute: "/users/somethingelse/somthing",
		},
		{
			name:            "Must handle no param format with ending url seperator",
			route:           "/users/somethingelse/somthing/",
			wantMappedRoute: "/users/somethingelse/somthing/",
		},
		{
			name:            "Must handle single url seperator",
			route:           "/users",
			wantMappedRoute: "/users",
		},
		{
			name:            "Must handle one parameter",
			route:           "/users/:id",
			wantMappedRoute: "/users/{id}",
		},
		{
			name:            "Must handle one parameter with url seperator suffix",
			route:           "/users/:id/",
			wantMappedRoute: "/users/{id}/",
		},
		{
			name:            "Must handle multiple parameter",
			route:           "/users/:id/friends/:friendID",
			wantMappedRoute: "/users/{id}/friends/{friendID}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMappedRoute := mapGinParamToOpenApiPath(tt.route); gotMappedRoute != tt.wantMappedRoute {
				t.Errorf("mapGinParamToOpenApiPath() = %v, want %v", gotMappedRoute, tt.wantMappedRoute)
			}
		})
	}
}

type User struct {
	Dto       testDto
	Id        string `json:"id" description:"id description" example:"1321231"`
	ValueUInt uint   `json:"value" description:"id description" example:"32"`
}

type ArrayDto struct {
	Id []string `json:"id" description:"id description" example:"1321231"`
}

type ArrayOfStructDto struct {
	Arrays []ArrayDto `json:"arrays" description:"id description" example:"1321231"`
}

func TestAssignSchemaIfNotExist(t *testing.T) {
	t.Run("Must not repeat a schema component", func(t *testing.T) {
		data := map[string]any{}
		assignSchemaIfNotExist(&testDto{}, data, http.MethodPost)
		assignSchemaIfNotExist(&User{}, data, http.MethodPost)

		assert.EqualValues(t, 2, len(data))
		fmt.Println(data)
		testValue := data["PosttestDto"]
		assert.NotNil(t, testValue)
		assert.NotNil(t, data["PostUser"])

		js, err := json.Marshal(data)
		assert.Nil(t, err)
		str := string(js)
		_ = str
	})

	t.Run("Must handle array of basic values", func(t *testing.T) {
		data := map[string]any{}
		assignSchemaIfNotExist(&ArrayDto{}, data, http.MethodPost)

		assert.EqualValues(t, 1, len(data))

		assert.NotNil(t, data["PostArrayDto"])
		obj := data["PostArrayDto"].(map[string]any)

		assert.EqualValues(t, Object, obj[Type])

		props := obj[Properties].(map[string]any)

		assert.NotNil(t, props["id"])

		idObj := props["id"].(map[string]any)

		assert.EqualValues(t, Array, idObj[Type])
		assert.NotNil(t, idObj[Items])

		js, err := json.Marshal(data)
		assert.Nil(t, err)
		str := string(js)
		_ = str
	})

	t.Run("Must handle array of structs", func(t *testing.T) {
		data := map[string]any{}
		assignSchemaIfNotExist(&ArrayOfStructDto{}, data, http.MethodPost)

		assert.EqualValues(t, 2, len(data))

		assert.NotNil(t, data["PostArrayOfStructDto"])
		obj := data["PostArrayOfStructDto"].(map[string]any)

		assert.EqualValues(t, Object, obj[Type])

		props := obj[Properties].(map[string]any)

		assert.NotNil(t, props["arrays"])

		idObj := props["arrays"].(map[string]any)

		assert.EqualValues(t, Array, idObj[Type])
		assert.NotNil(t, idObj[Items])

		js, err := json.Marshal(data)
		assert.Nil(t, err)
		str := string(js)
		_ = str
	})
}
