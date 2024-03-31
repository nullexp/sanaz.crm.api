package specification

// AndSpecification
type AndSpecification struct {
	rightCondition, leftCondition Specification
}

func NewAndSpecification(leftCondition, rightCondition Specification) Specification {
	spec := AndSpecification{leftCondition: leftCondition, rightCondition: rightCondition}

	return spec
}

func (c AndSpecification) GroupAnd(other Specification) Specification {
	return NewGroupAndSpecification(c, other)
}

func (c AndSpecification) And(other Specification) Specification {
	return NewAndSpecification(c, other)
}

func (c AndSpecification) Not() Specification {
	return NewNotSpecification(c)
}

func (c AndSpecification) Or(left Specification) Specification {
	return NewOrSpecification(c, left)
}

func (a AndSpecification) Execute(app Executable) any {
	return app.And(a.leftCondition, a.rightCondition)
}

func (c AndSpecification) GroupOr(other Specification) Specification {
	return NewGroupOrSpecification(c, other)
}
