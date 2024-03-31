package gorm

import (
	"context"

	database "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	dbspec "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"gorm.io/gorm"
)

type UuidMapperRepository[T, J database.Identity] struct {
	repository UuidRepository[T]

	mapper database.Mapper[T, J]
}

func NewGormUuidMapperRepository[T, J database.Identity](getter database.DataContextGetter, parser QueryParser, mapper database.Mapper[T, J], withTestId bool) UuidMapperRepository[T, J] {
	repo := NewGormUuidRepository[T](getter, parser)

	return UuidMapperRepository[T, J]{repository: repo, mapper: mapper}
}

func NewGormUuidMapperRepositoryWithDB[T, J database.Identity](gdb *gorm.DB, parser QueryParser, mapper database.Mapper[T, J]) UuidMapperRepository[T, J] {
	repo := NewGormUuidRepositoryWithDB[T](gdb, parser)

	return UuidMapperRepository[T, J]{repository: repo, mapper: mapper}
}

func (r UuidMapperRepository[T, J]) Get(ctx context.Context, query dbspec.Specification, page misc.Pagination, sort []misc.Sort) (out []J, err error) {
	entities, err := r.repository.Get(ctx, query, page, sort)
	if err != nil {
		return nil, err
	}

	return r.mapper.MapToModels(entities), nil
}

func (r *UuidMapperRepository[T, J]) Insert(ctx context.Context, model *J) (err error) {
	mappedEntity := r.mapper.MapToEntity(*model)

	err = r.repository.Insert(ctx, &mappedEntity)

	*model = r.mapper.MapToModel(mappedEntity) // Setting the Id

	return
}

func (r UuidMapperRepository[T, J]) GetById(ctx context.Context, id string) (out J, err error) {
	outEntity, err := r.repository.GetById(ctx, id)
	if err != nil {
		return
	}

	out = r.mapper.MapToModel(outEntity)

	return
}

func (r UuidMapperRepository[T, J]) Exist(ctx context.Context, query dbspec.Specification) (exist bool, err error) {
	return r.repository.Exist(ctx, query)
}

func (r UuidMapperRepository[T, J]) Sum(ctx context.Context, column string, query dbspec.Specification) (sum float64, err error) {
	return r.repository.Sum(ctx, column, query)
}

func (r UuidMapperRepository[T, J]) Average(ctx context.Context, column string, query dbspec.Specification) (average float64, err error) {
	return r.repository.Average(ctx, column, query)
}

func (r UuidMapperRepository[T, J]) Count(ctx context.Context, query dbspec.Specification) (count int64, err error) {
	return r.repository.Count(ctx, query)
}

func (r UuidMapperRepository[T, J]) Update(ctx context.Context, model *J) (err error) {
	entity := r.mapper.MapToEntity(*model)

	err = r.repository.Update(ctx, &entity)

	*model = r.mapper.MapToModel(entity) // Update time is changed

	return
}

func (r UuidMapperRepository[T, J]) PartialUpdate(ctx context.Context, model database.UuIdIdentity) (err error) {
	err = r.repository.PartialUpdate(ctx, model)
	return
}

func (r UuidMapperRepository[T, J]) Delete(ctx context.Context, id string) (err error) {
	return r.repository.Delete(ctx, id)
}

func (r UuidMapperRepository[T, J]) DeleteBySpecification(ctx context.Context, query dbspec.Specification) (err error) {
	return r.repository.DeleteBySpecification(ctx, query)
}

func (r UuidMapperRepository[T, J]) GetSingle(ctx context.Context, query dbspec.Specification) (out J, err error) {
	outEntity, err := r.repository.GetSingle(ctx, query)
	if err != nil {
		return
	}

	out = r.mapper.MapToModel(outEntity)

	return
}

func (r *UuidMapperRepository[T, J]) SetDB(db *gorm.DB) {
	r.repository.SetDB(db)
}

func (r UuidMapperRepository[T, J]) DistinctCount(ctx context.Context, column string, query dbspec.Specification) (distinctCount int64, err error) {
	return r.repository.DistinctCount(ctx, column, query)
}

func (r UuidMapperRepository[T, J]) DistinctSum(ctx context.Context, column string, query dbspec.Specification) (distinctSum float64, err error) {
	return r.repository.DistinctSum(ctx, column, query)
}

func (r UuidMapperRepository[T, J]) UpdateField(ctx context.Context, fieldName string, fieldValue any, query dbspec.Specification) (err error) {
	return r.repository.UpdateField(ctx, fieldName, fieldValue, query)
}
