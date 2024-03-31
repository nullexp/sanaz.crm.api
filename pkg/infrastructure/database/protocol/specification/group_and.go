package specification

// AndSpecification
type GroupAndSpecification struct {
	rightCondition, leftCondition Specification
}

func NewGroupAndSpecification(leftCondition, rightCondition Specification) Specification {
	spec := GroupAndSpecification{leftCondition: leftCondition, rightCondition: rightCondition}
	return spec
}

func (c GroupAndSpecification) And(other Specification) Specification {
	return NewAndSpecification(c, other)
}

func (c GroupAndSpecification) GroupAnd(other Specification) Specification {
	return NewGroupAndSpecification(c, other)
}

func (c GroupAndSpecification) Not() Specification {
	return NewNotSpecification(c)
}

func (c GroupAndSpecification) Or(left Specification) Specification {
	return NewOrSpecification(c, left)
}

func (a GroupAndSpecification) Execute(app Executable) any {
	return app.GroupAnd(a.leftCondition, a.rightCondition)
}

func (c GroupAndSpecification) GroupOr(other Specification) Specification {
	return NewGroupOrSpecification(c, other)
}
