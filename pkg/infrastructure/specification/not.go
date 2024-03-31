package specification

type NotSpecification[T any] struct {
	CompositeSpecification[T]
	condition Specification[T]
}

func NewNotSpecification[T any](condition Specification[T]) NotSpecification[T] {
	return NotSpecification[T]{condition: condition, CompositeSpecification: CompositeSpecification[T]{Value: condition}}
}

func (ns NotSpecification[T]) IsSatisfiedBy(value T) bool {
	return !(ns.condition.IsSatisfiedBy(value))
}

func (c NotSpecification[T]) OrNot(other Specification[T]) Specification[T] {
	return NewOrNotSpecification[T](c, other)
}

func (c NotSpecification[T]) Not() Specification[T] {
	return NewNotSpecification[T](c)
}

func (c NotSpecification[T]) Or(other Specification[T]) Specification[T] {
	return NewOrSpecification[T](c, other)
}

func (c NotSpecification[T]) AndNot(other Specification[T]) Specification[T] {
	return NewAndNotSpecification[T](c, other)
}

func (c NotSpecification[T]) And(other Specification[T]) Specification[T] {
	return NewAndSpecification[T](c, other)
}
