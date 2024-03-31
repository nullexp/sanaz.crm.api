package protocol

import (
	"log"

	gmodel "github.com/jinzhu/copier"
)

type defaultMapper[T, J Identity] struct{}

func NewDefaultMapper[T, J Identity]() Mapper[T, J] {
	return defaultMapper[T, J]{}
}

func (d defaultMapper[T, J]) MapToPartialEntity(model interface{}) (out T) {
	err := gmodel.Copy(&out, model)
	if err != nil {
		log.Println(err)
	}

	return
}

func (defaultMapper[T, J]) MapToEntity(model J) (out T) {
	err := gmodel.Copy(&out, model)
	if err != nil {
		log.Println(err)
	}

	return
}

func (defaultMapper[T, J]) MapToModel(entity T) (out J) {
	err := gmodel.Copy(&out, entity)
	if err != nil {
		log.Println(err)
	}

	return
}

func (d defaultMapper[T, J]) MapToEntities(models []J) (out []T) {
	out = []T{}

	for k := range models {
		out = append(out, d.MapToEntity(models[k]))
	}

	return out
}

func (d defaultMapper[T, J]) MapToModels(identities []T) (out []J) {
	out = []J{}

	for k := range identities {
		out = append(out, d.MapToModel(identities[k]))
	}

	return
}
