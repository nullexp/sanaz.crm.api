package gin

import (
	"net/http"

	httpapi "gitlab.espadev.ir/espad-go/infrastructure/http/protocol"
)

type PermissionManager struct {
	innerPostRouteMap   map[string][]string
	innerDeleteRouteMap map[string][]string
	innerPutRouteMap    map[string][]string
	innerGetRouteMap    map[string][]string
	innerFreeRouteMap   map[string][]httpapi.HTTPMethod
}

func NewPermissionManager() *PermissionManager {
	pm := PermissionManager{}

	pm.innerPostRouteMap = map[string][]string{}
	pm.innerDeleteRouteMap = map[string][]string{}
	pm.innerPutRouteMap = map[string][]string{}
	pm.innerGetRouteMap = map[string][]string{}
	pm.innerFreeRouteMap = map[string][]httpapi.HTTPMethod{}
	return &pm
}

func (pm *PermissionManager) GetGetPermission(route string) []string {
	val, ok := pm.innerGetRouteMap[route]

	if !ok {
		return []string{}
	}

	return val
}

func (pm *PermissionManager) GetPostPermission(route string) []string {
	val, ok := pm.innerPostRouteMap[route]

	if !ok {
		return []string{}
	}

	return val
}

func (pm *PermissionManager) GetDeletePermission(route string) []string {
	val, ok := pm.innerDeleteRouteMap[route]

	if !ok {
		return []string{}
	}

	return val
}

func (pm *PermissionManager) GetPutPermission(route string) []string {
	val, ok := pm.innerPutRouteMap[route]

	if !ok {
		return []string{}
	}

	return val
}

func (pm *PermissionManager) SetPostPermission(route string, perm string) {
	pm.innerPostRouteMap[route] = append(pm.innerPostRouteMap[route], perm)
}

func (pm *PermissionManager) GetPermission(route string, method httpapi.HTTPMethod) []string {
	routePerms := []string{}
	switch method {
	case http.MethodGet:
		routePerms = pm.GetGetPermission(route)
	case http.MethodPost:
		routePerms = pm.GetPostPermission(route)
	case http.MethodDelete:
		routePerms = pm.GetDeletePermission(route)
	case http.MethodPut:
		routePerms = pm.GetPutPermission(route)
	default:
		return routePerms
	}
	return routePerms
}

func (pm *PermissionManager) SetPutPermission(route string, perm string) {
	pm.innerPutRouteMap[route] = append(pm.innerPutRouteMap[route], perm)
}

func (pm *PermissionManager) SetGetPermission(route string, perm string) {
	pm.innerGetRouteMap[route] = append(pm.innerGetRouteMap[route], perm)
}

func (pm *PermissionManager) SetDeletePermission(route string, perm string) {
	pm.innerDeleteRouteMap[route] = append(pm.innerDeleteRouteMap[route], perm)
}

func (pm *PermissionManager) SetPermission(method httpapi.HTTPMethod, route string, perm string) {
	switch method {
	case "POST":
		pm.SetPostPermission(route, perm)
	case "DELETE":
		pm.SetDeletePermission(route, perm)
	case "GET":
		pm.SetGetPermission(route, perm)
	case "PUT":
		pm.SetPutPermission(route, perm)
	}
}

func (pm *PermissionManager) SetFreePermission(route string, method httpapi.HTTPMethod) {
	if pm.innerFreeRouteMap[route] == nil {
		pm.innerFreeRouteMap[route] = make([]httpapi.HTTPMethod, 0)
	}
	pm.innerFreeRouteMap[route] = append(pm.innerFreeRouteMap[route], method)
}

func (pm *PermissionManager) IsFree(route string, method httpapi.HTTPMethod) bool {
	val, ok := pm.innerFreeRouteMap[route]

	if !ok {
		return false
	}

	for _, v := range val {
		if v == method {
			return true
		}
	}
	return false
}
