package specification

// OrSpecification
type OrSpecification[T any] struct {
	CompositeSpecification[T]
	leftCondition  Specification[T]
	rightCondition Specification[T]
}

func NewOrSpecification[T any](left, right Specification[T]) OrSpecification[T] {
	return OrSpecification[T]{leftCondition: left, rightCondition: right}
}

func (Or OrSpecification[T]) IsSatisfiedBy(value T) bool {
	return Or.leftCondition.IsSatisfiedBy(value) || Or.rightCondition.IsSatisfiedBy(value)
}

func (c OrSpecification[T]) OrNot(other Specification[T]) Specification[T] {
	return NewOrNotSpecification[T](c, other)
}

func (c OrSpecification[T]) Not() Specification[T] {
	return NewNotSpecification[T](c)
}

func (c OrSpecification[T]) Or(other Specification[T]) Specification[T] {
	return NewOrSpecification[T](c, other)
}

func (c OrSpecification[T]) And(other Specification[T]) Specification[T] {
	return NewAndSpecification[T](c, other)
}
