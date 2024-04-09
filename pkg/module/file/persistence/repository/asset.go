package repository

import (
	"context"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"git.omidgolestani.ir/clinic/crm.api/pkg/module/file/model/entity"
)

type Asset interface {
	Insert(context.Context, *entity.Asset) error
	Update(context.Context, *entity.Asset) error
	Delete(context.Context, string) error
	GetById(context.Context, string) (entity.Asset, error)
	Get(context.Context, specification.Specification, misc.Pagination, []misc.Sort) ([]entity.Asset, error)
	GetSingle(context.Context, specification.Specification) (entity.Asset, error)
	Exist(context.Context, specification.Specification) (bool, error)
	Count(context.Context, specification.Specification) (count int64, err error)
	GetAsMap(context.Context, specification.Specification, misc.Pagination, []misc.Sort) (map[string]entity.Asset, error)
}
type AssetRepoFactory interface {
	NewAsset(protocol.DataContextGetter) Asset
}
