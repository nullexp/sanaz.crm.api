package specification

// AndSpecification
type GrouporSpecification struct {
	rightCondition, leftCondition Specification
}

func NewGroupOrSpecification(leftCondition, rightCondition Specification) Specification {
	spec := GrouporSpecification{leftCondition: leftCondition, rightCondition: rightCondition}
	return spec
}

func (c GrouporSpecification) And(other Specification) Specification {
	return NewAndSpecification(c, other)
}

func (c GrouporSpecification) GroupAnd(other Specification) Specification {
	return NewGroupAndSpecification(c, other)
}

func (c GrouporSpecification) GroupOr(other Specification) Specification {
	return NewGroupOrSpecification(c, other)
}

func (c GrouporSpecification) Not() Specification {
	return NewNotSpecification(c)
}

func (c GrouporSpecification) Or(left Specification) Specification {
	return NewOrSpecification(c, left)
}

func (a GrouporSpecification) Execute(app Executable) any {
	return app.GroupOr(a.leftCondition, a.rightCondition)
}
