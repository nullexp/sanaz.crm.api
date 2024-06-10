package gin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	httpapi "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"github.com/stretchr/testify/assert"
)

func TestRequestHandler(t *testing.T) {
	type page struct {
		Skip  uint `json:"skip"`
		Limit uint `json:"limit"`
	}

	type cursorPage struct {
		Cursor string `json:"cursor" form:"cursor"`
		Limit  uint   `json:"limit" form:"limit"`
		After  bool   `json:"after" form:"after"`
	}

	type sort struct {
		Name        string `json:"name"`
		IsAscending bool   `json:"isAscending"`
		Order       int    `json:"order"`
	}

	type query struct {
		Name     string `json:"name"`
		Operator string `json:"operator"`
		Value    any    `json:"value"`
	}

	app := NewGinApp()
	var a httpapi.Api = app

	baseRoute := "/test"
	var smodule httpapi.Module = NewTestModule(baseRoute,
		&httpapi.RequestDefinition{
			Route:     "/pagination-cursor",
			Method:    http.MethodGet,
			FreeRoute: true,
			Handler: func(req httpapi.Request) {
				raw, ok := req.GetCursorPagination()
				if !ok {
					req.SetBadRequest("unwanted value", "ErrUnwantedValue")
					return
				}
				td := cursorPage{Limit: raw.GetLimit(), Cursor: raw.GetCursor(), After: raw.After()}
				req.Negotiate(http.StatusOK, nil, td)
			},
		},
		&httpapi.RequestDefinition{
			Route:     "/pagination-default-cursor",
			Method:    http.MethodGet,
			FreeRoute: true,
			Handler: func(req httpapi.Request) {
				raw, _ := req.GetCursorPagination()
				td := cursorPage{Limit: raw.GetLimit(), Cursor: raw.GetCursor(), After: raw.After()}
				req.Negotiate(http.StatusOK, nil, td)
			},
		},
		&httpapi.RequestDefinition{
			Route:     "/pagination",
			Method:    http.MethodGet,
			FreeRoute: true,
			Handler: func(req httpapi.Request) {
				raw, ok := req.GetPagination()
				if !ok {
					req.SetBadRequest("unwanted value", "ErrUnwantedValue")
					return
				}
				td := page{Skip: raw.GetSkip(), Limit: raw.GetLimit()}
				req.Negotiate(http.StatusOK, nil, td)
			},
		},
		&httpapi.RequestDefinition{
			Route:     "/pagination-default",
			Method:    http.MethodGet,
			FreeRoute: true,
			Handler: func(req httpapi.Request) {
				raw, _ := req.GetPagination()
				td := page{Skip: raw.GetSkip(), Limit: raw.GetLimit()}
				req.Negotiate(http.StatusOK, nil, td)
			},
		},
		&httpapi.RequestDefinition{
			Route:  "/caller-default",
			Method: http.MethodGet,
			Handler: func(req httpapi.Request) {
				caller := req.MustGetCaller()
				req.Negotiate(http.StatusOK, nil, map[string]interface{}{"id": caller.GetSubject()})
			},
		},
		&httpapi.RequestDefinition{
			Route:     "/sort-default",
			Method:    http.MethodGet,
			FreeRoute: true,
			Handler: func(req httpapi.Request) {
				srots := req.GetSort()
				out := []sort{}
				for _, v := range srots {
					out = append(out, sort{Order: v.GetOrder(), Name: v.GetName(), IsAscending: v.IsAscending()})
				}
				req.Negotiate(http.StatusOK, nil, out)
			},
		},
		&httpapi.RequestDefinition{
			Route:  "/query-default",
			Method: http.MethodGet,
			Parameters: []httpapi.RequestParameter{
				{Query: true, Optional: true, Definition: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorEqual}, misc.DataTypeUInteger)},
				{Query: true, Optional: true, Definition: misc.NewQueryDefinition("username", []misc.QueryOperator{misc.QueryOperatorEqual}, misc.DataTypeString)},
				{Query: true, Optional: true, Definition: misc.NewQueryDefinition("valid", []misc.QueryOperator{misc.QueryOperatorEqual}, misc.DataTypeBoolean)},
			},
			FreeRoute: true,
			Handler: func(req httpapi.Request) {
				queries := req.GetQuery()
				out := []query{}
				for _, v := range queries {
					d := v.GetOperand().Value
					out = append(out, query{Name: v.GetName(), Operator: string(v.GetOperator()), Value: d})
				}
				req.Negotiate(http.StatusOK, nil, out)
			},
		},
	)
	a.AppendModule(smodule)
	auth := NewOkTestAuthenticatorWithSubject("1")
	a.AppendAuthenticator(baseRoute, auth)
	app.Init(gin.TestMode)

	t.Run("Pagination: Must return simple valid pagination", func(t *testing.T) {
		skip := uint(5)
		limit := uint(15)
		query := fmt.Sprintf("?limit=%d&skip=%d", limit, skip)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/pagination"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		p := &page{}
		err := json.Unmarshal(w.Body.Bytes(), &p)
		assert.Equal(t, nil, err)
		assert.Equal(t, skip, p.Skip)
		assert.Equal(t, limit, p.Limit)
	})

	t.Run("Pagination: Must return default if query is invalid", func(t *testing.T) {
		offest := -1
		limit := -10
		query := fmt.Sprintf("?limit=%d&skip=%d", limit, offest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/pagination-default"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		p := &page{}
		err := json.Unmarshal(w.Body.Bytes(), &p)
		assert.Equal(t, nil, err)
		assert.Equal(t, uint(0), p.Skip)
		assert.Equal(t, uint(10), p.Limit)
	})

	t.Run("Cursor Pagination: Must return simple valid cursor pagination", func(t *testing.T) {
		limit := uint(5)
		after := true
		cursor := "cursor"
		query := fmt.Sprintf("?limit=%d&after=%v&cursor=%s", limit, after, cursor)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/pagination-cursor"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		p := &cursorPage{}
		err := json.Unmarshal(w.Body.Bytes(), &p)
		assert.Equal(t, nil, err)
		assert.Equal(t, after, p.After)
		assert.Equal(t, cursor, p.Cursor)
		assert.Equal(t, limit, p.Limit)
	})

	t.Run("Pagination: Must return default cursor if query is invalid", func(t *testing.T) {
		limit := -10
		query := fmt.Sprintf("?limit=%d", limit)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/pagination-default-cursor"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		p := &cursorPage{}
		err := json.Unmarshal(w.Body.Bytes(), &p)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, p.After)
		assert.Equal(t, "", p.Cursor)
		assert.Equal(t, uint(10), p.Limit)
	})

	t.Run("GetCaller: Must return valid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/caller-default", nil)
		req.Header.Add(Accept, AppJson)
		req.Header.Add(misc.HeaderAuthorization, "Bearer somerandomText")
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		dto := map[string]interface{}{}
		err := json.Unmarshal(w.Body.Bytes(), &dto)
		assert.Equal(t, nil, err)
		assert.Equal(t, dto["id"], "1")
	})

	t.Run("Sort: Must handle valid", func(t *testing.T) {
		query := "?sort=-name,-id,description,+phone"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/sort-default"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		sorts := []sort{}
		err := json.Unmarshal(w.Body.Bytes(), &sorts)
		assert.Equal(t, nil, err)

		assert.Equal(t, false, sorts[0].IsAscending)
		assert.Equal(t, 1, sorts[0].Order)
		assert.Equal(t, "name", sorts[0].Name)

		assert.Equal(t, false, sorts[1].IsAscending)
		assert.Equal(t, 2, sorts[1].Order)
		assert.Equal(t, "id", sorts[1].Name)

		assert.Equal(t, true, sorts[2].IsAscending)
		assert.Equal(t, 3, sorts[2].Order)
		assert.Equal(t, "description", sorts[2].Name)

		assert.Equal(t, true, sorts[3].IsAscending)
		assert.Equal(t, 4, sorts[3].Order)
		assert.Equal(t, "phone", sorts[3].Name)
	})

	t.Run("Sort: Must ignore invalid comma", func(t *testing.T) {
		query := "?sort=-name,-id,description,+phone,"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/sort-default"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		sorts := []sort{}
		err := json.Unmarshal(w.Body.Bytes(), &sorts)
		assert.Equal(t, nil, err)

		assert.Equal(t, false, sorts[0].IsAscending)
		assert.Equal(t, 1, sorts[0].Order)
		assert.Equal(t, "name", sorts[0].Name)

		assert.Equal(t, false, sorts[1].IsAscending)
		assert.Equal(t, 2, sorts[1].Order)
		assert.Equal(t, "id", sorts[1].Name)

		assert.Equal(t, true, sorts[2].IsAscending)
		assert.Equal(t, 3, sorts[2].Order)
		assert.Equal(t, "description", sorts[2].Name)

		assert.Equal(t, true, sorts[3].IsAscending)
		assert.Equal(t, 4, sorts[3].Order)
		assert.Equal(t, "phone", sorts[3].Name)
		assert.Equal(t, 4, len(sorts))
	})

	t.Run("Sort: Must ignore empty sort", func(t *testing.T) {
		query := "?sort="
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/sort-default"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		sorts := []sort{}
		err := json.Unmarshal(w.Body.Bytes(), &sorts)
		assert.Equal(t, nil, err)

		assert.Equal(t, 0, len(sorts))
	})

	t.Run("Sort: Must ignore empty sort with comma", func(t *testing.T) {
		query := "?sort=,,"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/sort-default"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		sorts := []sort{}
		err := json.Unmarshal(w.Body.Bytes(), &sorts)
		assert.Equal(t, nil, err)

		assert.Equal(t, 0, len(sorts))
	})

	t.Run("Sort: Must ignore empty sort with minus", func(t *testing.T) {
		query := "?sort=-+-"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/sort-default"+query, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		sorts := []sort{}
		err := json.Unmarshal(w.Body.Bytes(), &sorts)
		assert.Equal(t, nil, err)

		assert.Equal(t, 0, len(sorts))
	})

	t.Run("Query Param: Must handle correct query parameter", func(t *testing.T) {
		q := "?"
		for k := range misc.ReservedQueryMap {
			new := fmt.Sprintf("%s=%s&", k, Data)
			q += new
		}
		q += "id=10&username=somedata&valid=true"
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/query-default"+q, nil)
		req.Header.Add(Accept, AppJson)
		_ = app.TestHandle(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		queries := []query{}
		err := json.Unmarshal(w.Body.Bytes(), &queries)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(queries))
		queryMap := map[string]query{}
		for k, v := range queries {
			queryMap[v.Name] = queries[k]
			_, ok := misc.ReservedQueryMap[v.Name]
			assert.Equal(t, false, ok)
		}
		assert.Equal(t, float64(10), queryMap["id"].Value)
		assert.Equal(t, "somedata", queryMap["username"].Value)
		assert.Equal(t, true, queryMap["valid"].Value)

		assert.Equal(t, string(QueryOperatorEqual), queryMap["id"].Operator)
		assert.Equal(t, string(QueryOperatorEqual), queryMap["username"].Operator)
		assert.Equal(t, string(QueryOperatorEqual), queryMap["valid"].Operator)
	})
}

func NewOkTestAuthenticatorWithSubject(subject string) *testAuthenticator {
	return NewTestAuthenticator(func(token string) (misc.JwtClaim, error) {
		return TokenInfo{ExpireTime: time.Now().AddDate(1, 0, 0).Unix(), Subject: subject}, nil
	}, func(token string) (bool, error) { return true, nil })
}
