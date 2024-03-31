package specification

import (
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
)

type Executer interface {
	Execute(app Executable) any
}

type Specification interface {
	Executer

	And(Specification) Specification

	GroupAnd(Specification) Specification

	GroupOr(Specification) Specification

	Or(Specification) Specification

	Not() Specification
}

type CompositeSpecification struct {
	query protocol.Query
}

func ToSpecification(query protocol.QueryInfo) Specification {
	if query == nil || len(query.GetQuery()) == 0 {
		return nil
	}

	isAnd := query.IsAnd()

	var spec Specification

	for _, v := range query.GetQuery() {

		if spec == nil {

			spec = NewSpecification(v)

			continue

		}

		if isAnd {
			spec = spec.And(NewSpecification(v))
		} else {
			spec = spec.Or(NewSpecification(v))
		}

	}

	return spec
}

func NewSpecification(query protocol.Query) Specification {
	return CompositeSpecification{query: query}
}

func NewQuerySpecification(name string, op misc.QueryOperator, operand *misc.Operand) Specification {
	return CompositeSpecification{query: protocol.NewQuery(name, op, operand)}
}

func NewSpecificationWithModel(name string, op misc.QueryOperator, model protocol.Identity, fields []string, operand *misc.Operand) Specification {
	return CompositeSpecification{query: protocol.NewQueryWithModel(name, op, model, fields, operand)}
}

func NewEmptySpecification(name string, op misc.QueryOperator) Specification {
	return CompositeSpecification{query: protocol.NewQuery(name, op, misc.NewOperand(nil))}
}

func (c CompositeSpecification) Execute(applier Executable) any {
	return applier.Execute(c.query)
}

func (c CompositeSpecification) And(other Specification) Specification {
	return NewAndSpecification(c, other)
}

func (c CompositeSpecification) Or(left Specification) Specification {
	return NewOrSpecification(c, left)
}

func (c CompositeSpecification) GroupAnd(other Specification) Specification {
	return NewGroupAndSpecification(c, other)
}

func (c CompositeSpecification) Not() Specification {
	return NewNotSpecification(c)
}

func (c CompositeSpecification) GroupOr(other Specification) Specification {
	return NewGroupOrSpecification(c, other)
}
