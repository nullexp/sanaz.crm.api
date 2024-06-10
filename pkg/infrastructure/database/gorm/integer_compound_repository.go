package gorm

import (
	"context"

	database "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	dbspec "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol/specification"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"gorm.io/gorm"
)

type IntegerCompoundRepository[T, J database.IdIdentity] struct {
	MapperIntegerRepository[T, J]

	primaryKeyName string
}

func NewGormIntegerCompoundRepository[T, J database.IdIdentity](getter database.DataContextGetter, parser QueryParser, mapper database.Mapper[T, J], withTestId bool, primaryKeyName string) IntegerCompoundRepository[T, J] {
	repo := NewGormIntegerRepository[T](getter, parser)

	return IntegerCompoundRepository[T, J]{MapperIntegerRepository: MapperIntegerRepository[T, J]{IntegerRepository: repo, mapper: mapper}, primaryKeyName: primaryKeyName}
}

func NewGormIntegerCompoundRepositoryWithDB[T, J database.IdIdentity](gdb *gorm.DB, parser QueryParser, mapper database.Mapper[T, J], primaryKeyName string) IntegerCompoundRepository[T, J] {
	repo := NewGormIntegerRepositoryWithDB[T](gdb, parser)

	return IntegerCompoundRepository[T, J]{MapperIntegerRepository: MapperIntegerRepository[T, J]{IntegerRepository: repo, mapper: mapper}, primaryKeyName: primaryKeyName}
}

func (r IntegerCompoundRepository[T, J]) GetKeys(ctx context.Context, query dbspec.Specification, page misc.Pagination, sort []misc.Sort) (out []int64, err error) {
	db := r.IntegerRepository.db

	parser := r.IntegerRepository.parser

	db = db.Model(new(T))

	db = parser.ParseSpecification(db, query)

	db = parser.ParseSort(db, sort...)

	db = parser.ParsePage(db, page)

	err = db.WithContext(ctx).Pluck(r.primaryKeyName, &out).Error

	return
}

func (r IntegerCompoundRepository[T, J]) GetAsMap(ctx context.Context, query dbspec.Specification, page misc.Pagination, sort []misc.Sort) (out map[int64]J, err error) {
	out = map[int64]J{}

	models, err := r.MapperIntegerRepository.Get(ctx, query, page, sort)

	for _, model := range models {

		v := model

		out[model.GetId()] = v

	}

	return
}

func (r *IntegerCompoundRepository[T, J]) SetDB(db *gorm.DB) {
	r.IntegerRepository.SetDB(db)
}
