package gorm

import (
	"context"

	database "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	dbspec "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"gorm.io/gorm"
)

type UuidCompoundRepository[T, J database.UuIdIdentity] struct {
	UuidMapperRepository[T, J]

	primaryKeyName string
}

func NewGormUuidCompoundRepository[T, J database.UuIdIdentity](getter database.DataContextGetter, parser QueryParser, mapper database.Mapper[T, J], withTestId bool, primaryKeyName string) UuidCompoundRepository[T, J] {
	repo := NewGormUuidRepository[T](getter, parser)

	return UuidCompoundRepository[T, J]{UuidMapperRepository: UuidMapperRepository[T, J]{repository: repo, mapper: mapper}, primaryKeyName: primaryKeyName}
}

func NewGormUuidCompoundRepositoryWithDB[T, J database.UuIdIdentity](gdb *gorm.DB, parser QueryParser, mapper database.Mapper[T, J], primaryKeyName string) UuidCompoundRepository[T, J] {
	repo := NewGormUuidRepositoryWithDB[T](gdb, parser)

	return UuidCompoundRepository[T, J]{UuidMapperRepository: UuidMapperRepository[T, J]{repository: repo, mapper: mapper}, primaryKeyName: primaryKeyName}
}

func (r UuidCompoundRepository[T, J]) GetKeys(ctx context.Context, query dbspec.Specification, page misc.Pagination, sort []misc.Sort) (out []string, err error) {
	db := r.repository.db

	parser := r.repository.parser

	db = db.Model(new(T))

	db = parser.ParseSpecification(db, query)

	db = parser.ParseSort(db, sort...)

	db = parser.ParsePage(db, page)

	err = db.WithContext(ctx).Pluck(r.primaryKeyName, &out).Error

	return
}

func (r UuidCompoundRepository[T, J]) GetAsMap(ctx context.Context, query dbspec.Specification, page misc.Pagination, sort []misc.Sort) (out map[string]J, err error) {
	out = map[string]J{}

	models, err := r.UuidMapperRepository.Get(ctx, query, page, sort)

	for _, model := range models {

		v := model

		out[model.GetUuid()] = v

	}

	return
}

func (r *UuidCompoundRepository[T, J]) SetDB(db *gorm.DB) {
	r.UuidMapperRepository.SetDB(db)
}
