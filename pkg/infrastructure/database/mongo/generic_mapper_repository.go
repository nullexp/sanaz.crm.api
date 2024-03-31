package mongo

import (
	"context"

	database "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
)

type MapperRepository[T, J database.EntityBased] struct {
	repository Repository[T]

	mapper database.Mapper[T, J]
}

func NewMapperRepository[T, J database.EntityBased](getter database.DataContextGetter, collectionName string, mapper database.Mapper[T, J], withTestId bool) MapperRepository[T, J] {
	repo := NewRepository[T](getter, collectionName)

	return MapperRepository[T, J]{repository: repo, mapper: mapper}
}

func (r MapperRepository[T, J]) Read(ctx context.Context, filter DynamicFilter) (out []J, err error) {
	entities, err := r.repository.Read(ctx, filter)
	if err != nil {
		return nil, err
	}

	return r.mapper.MapToModels(entities), nil
}

func (r MapperRepository[T, J]) Create(ctx context.Context, doc *J) error {
	mappedEntity := r.mapper.MapToEntity(*doc)

	err := r.repository.Create(ctx, &mappedEntity)

	*doc = r.mapper.MapToModel(mappedEntity) // Setting the Id

	return err
}

func (r MapperRepository[T, J]) Update(ctx context.Context, filter DynamicFilter, update UpdateField) error {
	err := r.repository.Update(ctx, filter, update)

	return err
}

func (r MapperRepository[T, J]) Delete(ctx context.Context, filter DynamicFilter) error {
	err := r.repository.Delete(ctx, filter)

	return err
}
