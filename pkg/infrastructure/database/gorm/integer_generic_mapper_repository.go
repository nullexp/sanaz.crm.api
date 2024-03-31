package gorm

import (
	"context"

	database "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	dbspec "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"gorm.io/gorm"
)

type MapperIntegerRepository[T, J database.Identity] struct {
	IntegerRepository IntegerRepository[T]

	mapper database.Mapper[T, J]
}

func NewGormMapperIntegerRepository[T, J database.Identity](getter database.DataContextGetter, parser QueryParser, mapper database.Mapper[T, J], withTestId bool) MapperIntegerRepository[T, J] {
	repo := NewGormIntegerRepository[T](getter, parser)

	return MapperIntegerRepository[T, J]{IntegerRepository: repo, mapper: mapper}
}

func NewGormMapperIntegerRepositoryWithDB[T, J database.Identity](gdb *gorm.DB, parser QueryParser, mapper database.Mapper[T, J]) MapperIntegerRepository[T, J] {
	repo := NewGormIntegerRepositoryWithDB[T](gdb, parser)

	return MapperIntegerRepository[T, J]{IntegerRepository: repo, mapper: mapper}
}

func (r MapperIntegerRepository[T, J]) Get(ctx context.Context, query dbspec.Specification, page misc.Pagination, sort []misc.Sort) (out []J, err error) {
	entities, err := r.IntegerRepository.Get(ctx, query, page, sort)
	if err != nil {
		return nil, err
	}

	return r.mapper.MapToModels(entities), nil
}

func (r *MapperIntegerRepository[T, J]) Insert(ctx context.Context, model *J) (err error) {
	mappedEntity := r.mapper.MapToEntity(*model)

	err = r.IntegerRepository.Insert(ctx, &mappedEntity)

	*model = r.mapper.MapToModel(mappedEntity) // Setting the Id

	return
}

func (r MapperIntegerRepository[T, J]) GetById(ctx context.Context, id int64) (out J, err error) {
	outEntity, err := r.IntegerRepository.GetById(ctx, id)
	if err != nil {
		return
	}

	out = r.mapper.MapToModel(outEntity)

	return
}

func (r MapperIntegerRepository[T, J]) Exist(ctx context.Context, query dbspec.Specification) (exist bool, err error) {
	return r.IntegerRepository.Exist(ctx, query)
}

func (r MapperIntegerRepository[T, J]) Count(ctx context.Context, query dbspec.Specification) (count int64, err error) {
	return r.IntegerRepository.Count(ctx, query)
}

func (r MapperIntegerRepository[T, J]) Update(ctx context.Context, model *J) (err error) {
	entity := r.mapper.MapToEntity(*model)

	err = r.IntegerRepository.Update(ctx, &entity)

	*model = r.mapper.MapToModel(entity) // Update time is changed

	return
}

func (r MapperIntegerRepository[T, J]) PartialUpdate(ctx context.Context, model database.IdIdentity) (err error) {
	err = r.IntegerRepository.PartialUpdate(ctx, model)
	return
}

func (r MapperIntegerRepository[T, J]) Delete(ctx context.Context, id int64) (err error) {
	return r.IntegerRepository.Delete(ctx, id)
}

func (r MapperIntegerRepository[T, J]) DeleteBySpecification(ctx context.Context, query dbspec.Specification) (err error) {
	return r.IntegerRepository.DeleteBySpecification(ctx, query)
}

func (r MapperIntegerRepository[T, J]) GetSingle(ctx context.Context, query dbspec.Specification) (out J, err error) {
	outEntity, err := r.IntegerRepository.GetSingle(ctx, query)
	if err != nil {
		return
	}

	out = r.mapper.MapToModel(outEntity)

	return
}

func (r *MapperIntegerRepository[T, J]) SetDB(db *gorm.DB) {
	r.IntegerRepository.SetDB(db)
}

func (r MapperIntegerRepository[T, J]) DistinctSum(ctx context.Context, column string, query dbspec.Specification) (distinctSum float64, err error) {
	return r.IntegerRepository.DistinctSum(ctx, column, query)
}

func (r MapperIntegerRepository[T, J]) DistinctCount(ctx context.Context, column string, query dbspec.Specification) (distinctCount int64, err error) {
	return r.IntegerRepository.DistinctCount(ctx, column, query)
}

func (r MapperIntegerRepository[T, J]) Sum(ctx context.Context, column string, query dbspec.Specification) (sum float64, err error) {
	return r.IntegerRepository.Sum(ctx, column, query)
}

func (r MapperIntegerRepository[T, J]) Average(ctx context.Context, column string, query dbspec.Specification) (average float64, err error) {
	return r.IntegerRepository.Average(ctx, column, query)
}

func (r MapperIntegerRepository[T, J]) UpdateField(ctx context.Context, fieldName string, fieldValue any, query dbspec.Specification) (err error) {
	return r.IntegerRepository.UpdateField(ctx, fieldName, fieldValue, query)
}
