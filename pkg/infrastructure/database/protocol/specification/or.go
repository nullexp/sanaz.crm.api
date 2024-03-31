package specification

type OrSpecification struct {
	leftCondition, rightCondition Specification
}

func NewOrSpecification(leftCondition, rightCondition Specification) Specification {
	spec := OrSpecification{leftCondition: leftCondition, rightCondition: rightCondition}
	return spec
}

func (c OrSpecification) And(other Specification) Specification {
	return NewAndSpecification(c, other)
}

func (c OrSpecification) Or(left Specification) Specification {
	return NewOrSpecification(c, left)
}

func (c OrSpecification) GroupAnd(other Specification) Specification {
	return NewGroupAndSpecification(c, other)
}

func (c OrSpecification) Not() Specification {
	return NewNotSpecification(c)
}

func (a OrSpecification) Execute(app Executable) any {
	return app.Or(a.leftCondition, a.rightCondition)
}

func (c OrSpecification) GroupOr(other Specification) Specification {
	return NewGroupOrSpecification(c, other)
}
