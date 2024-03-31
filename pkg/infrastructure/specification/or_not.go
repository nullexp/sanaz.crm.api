package specification

// OrSpecification
type OrNotSpecification[T any] struct {
	CompositeSpecification[T]
	leftCondition  Specification[T]
	rightCondition Specification[T]
}

func NewOrNotSpecification[T any](left, right Specification[T]) OrNotSpecification[T] {
	return OrNotSpecification[T]{leftCondition: left, rightCondition: right}
}

func (or OrNotSpecification[T]) IsSatisfiedBy(value T) bool {
	return !(or.leftCondition.IsSatisfiedBy(value) || or.rightCondition.IsSatisfiedBy(value))
}

func (c OrNotSpecification[T]) OrNot(other Specification[T]) Specification[T] {
	return NewOrNotSpecification[T](c, other)
}

func (c OrNotSpecification[T]) Not() Specification[T] {
	return NewNotSpecification[T](c)
}

func (c OrNotSpecification[T]) Or(other Specification[T]) Specification[T] {
	return NewOrSpecification[T](c, other)
}

func (c OrNotSpecification[T]) AndNot(other Specification[T]) Specification[T] {
	return NewAndNotSpecification[T](c, other)
}

func (c OrNotSpecification[T]) And(other Specification[T]) Specification[T] {
	return NewAndSpecification[T](c, other)
}
