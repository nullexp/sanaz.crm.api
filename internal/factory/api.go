package factory

import (
	"errors"

	ginapi "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/gin"
	http "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/http/protocol"
)

const Test = "test"
const Gin = "gin"

var ErrNotImplemented = errors.New("not implemented")
var ErrMissingParameter = errors.New("missing parameter")

func NewApi(name string, param ...any) http.Api {

	if name == "" {
		name = Test
	}

	switch name {
	case Test:
		fallthrough
	case Gin:
		return ginapi.NewGinApp()
	}
	panic(ErrNotImplemented)
}
