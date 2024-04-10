package repository

import (
	"context"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/model/entity"
)

type Image interface {
	Insert(context.Context, *entity.Image) error
	Update(context.Context, *entity.Image) error
	Delete(context.Context, string) error
	GetById(context.Context, string) (entity.Image, error)
	Get(context.Context, specification.Specification, misc.Pagination, []misc.Sort) ([]entity.Image, error)
	GetSingle(context.Context, specification.Specification) (entity.Image, error)
	Exist(context.Context, specification.Specification) (bool, error)
	Count(context.Context, specification.Specification) (count int64, err error)
	GetAsMap(context.Context, specification.Specification, misc.Pagination, []misc.Sort) (map[string]entity.Image, error)
}
type ImageRepoFactory interface {
	NewImage(protocol.DataContextGetter) Image
}
