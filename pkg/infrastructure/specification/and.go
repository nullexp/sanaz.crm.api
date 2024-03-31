package specification

// AndSpecification
type AndSpecification[T any] struct {
	CompositeSpecification[T]
	leftCondition  Specification[T]
	rightCondition Specification[T]
}

func NewAndSpecification[T any](left, right Specification[T]) AndSpecification[T] {
	return AndSpecification[T]{leftCondition: left, rightCondition: right}
}

func (and AndSpecification[T]) IsSatisfiedBy(value T) bool {
	return and.leftCondition.IsSatisfiedBy(value) && and.rightCondition.IsSatisfiedBy(value)
}

func (c AndSpecification[T]) AndNot(other Specification[T]) Specification[T] {
	return NewAndNotSpecification[T](c, other)
}

func (c AndSpecification[T]) Not() Specification[T] {
	return NewNotSpecification[T](c)
}

func (c AndSpecification[T]) OrNot(other Specification[T]) Specification[T] {
	return NewOrNotSpecification[T](c, other)
}

func (c AndSpecification[T]) Or(other Specification[T]) Specification[T] {
	return NewOrSpecification[T](c, other)
}

func (c AndSpecification[T]) And(other Specification[T]) Specification[T] {
	return NewAndSpecification[T](c, other)
}
