package protocol

type Mapper[T, J Identity] interface {
	MapToEntity(model J) T
	MapToPartialEntity(model interface{}) T
	MapToModel(T) J

	MapToEntities(model []J) []T
	MapToModels(identity []T) []J
}
