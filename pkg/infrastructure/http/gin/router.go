package gin

import (
	"net/http"

	httpapi "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol"
)

type Router struct {
	Routes          []*httpapi.RequestDefinition
	RoutesPostMap   map[string]*httpapi.RequestDefinition
	RoutesPutMap    map[string]*httpapi.RequestDefinition
	RoutesGetMap    map[string]*httpapi.RequestDefinition
	RoutesDeleteMap map[string]*httpapi.RequestDefinition
}

func NewRouter() *Router {
	return &Router{Routes: []*httpapi.RequestDefinition{}, RoutesDeleteMap: map[string]*httpapi.RequestDefinition{}, RoutesGetMap: map[string]*httpapi.RequestDefinition{}, RoutesPutMap: map[string]*httpapi.RequestDefinition{}, RoutesPostMap: map[string]*httpapi.RequestDefinition{}}
}

func (router *Router) Register(rt *httpapi.RequestDefinition, baseURL string) {
	fUrl := baseURL + rt.Route
	switch rt.Method {
	case "POST":
		router.RoutesPostMap[fUrl] = rt
	case "DELETE":
		router.RoutesDeleteMap[fUrl] = rt
	case "GET":
		router.RoutesGetMap[fUrl] = rt
	case "PUT":
		router.RoutesPutMap[fUrl] = rt
	}

	router.Routes = append(router.Routes, rt)
}

func (router *Router) GetRoutes() []*httpapi.RequestDefinition {
	return router.Routes
}

func (router *Router) GetRoute(route string, method httpapi.HTTPMethod) *httpapi.RequestDefinition {
	switch method {
	case http.MethodPost:
		return router.RoutesPostMap[route]
	case http.MethodDelete:
		return router.RoutesDeleteMap[route]
	case http.MethodGet:
		return router.RoutesGetMap[route]
	case http.MethodPut:
		return router.RoutesPutMap[route]
	}
	panic(UnknownRoute)
}

const UnknownRoute = "Given route and method is not registred"

func (router *Router) IsFree(route string, method httpapi.HTTPMethod) bool {
	r := router.GetRoute(route, method)
	if r == nil {
		return false
	}

	return r.FreeRoute
}
