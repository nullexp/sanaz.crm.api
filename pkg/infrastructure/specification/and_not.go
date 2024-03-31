package specification

// AndSpecification
type AndNotSpecification[T any] struct {
	CompositeSpecification[T]
	leftCondition  Specification[T]
	rightCondition Specification[T]
}

func NewAndNotSpecification[T any](left, right Specification[T]) AndNotSpecification[T] {
	return AndNotSpecification[T]{leftCondition: left, rightCondition: right}
}

func (and AndNotSpecification[T]) IsSatisfiedBy(value T) bool {
	return !(and.leftCondition.IsSatisfiedBy(value) && and.rightCondition.IsSatisfiedBy(value))
}

func (c AndNotSpecification[T]) AndNot(other Specification[T]) Specification[T] {
	return NewAndNotSpecification[T](c, other)
}

func (c AndNotSpecification[T]) Not() Specification[T] {
	return NewNotSpecification[T](c)
}

func (c AndNotSpecification[T]) OrNot(other Specification[T]) Specification[T] {
	return NewOrNotSpecification[T](c, other)
}

func (c AndNotSpecification[T]) Or(other Specification[T]) Specification[T] {
	return NewOrSpecification[T](c, other)
}

func (c AndNotSpecification[T]) And(other Specification[T]) Specification[T] {
	return NewAndNotSpecification[T](c, other)
}
