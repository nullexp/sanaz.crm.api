package specification

type NotSpecification struct {
	wrapped Specification
}

func NewNotSpecification(condition Specification) Specification {
	spec := NotSpecification{wrapped: condition}
	return spec
}

func (c NotSpecification) And(other Specification) Specification {
	return NewAndSpecification(c, other)
}

func (c NotSpecification) Not() Specification {
	return NewNotSpecification(c.wrapped)
}

func (c NotSpecification) Or(left Specification) Specification {
	return NewOrSpecification(c, left)
}

func (c NotSpecification) GroupAnd(other Specification) Specification {
	return NewGroupAndSpecification(c, other)
}

func (a NotSpecification) Execute(app Executable) any {
	return app.Not(a.wrapped)
}

func (c NotSpecification) GroupOr(other Specification) Specification {
	return NewGroupOrSpecification(c, other)
}
