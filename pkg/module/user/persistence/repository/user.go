package repository

import (
	"context"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol/specification"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"github.com/nullexp/sanaz.crm.api/pkg/module/user/model/entity"
)

type User interface {
	Insert(context.Context, *entity.User) error
	Update(context.Context, *entity.User) error
	Delete(context.Context, string) error
	GetById(context.Context, string) (entity.User, error)
	Get(context.Context, specification.Specification, misc.Pagination, []misc.Sort) ([]entity.User, error)
	GetSingle(context.Context, specification.Specification) (entity.User, error)
	Exist(context.Context, specification.Specification) (bool, error)
	Count(context.Context, specification.Specification) (count int64, err error)
	GetAsMap(context.Context, specification.Specification, misc.Pagination, []misc.Sort) (map[string]entity.User, error)
}
type UserRepoFactory interface {
	NewUser(protocol.DataContextGetter) User
}
