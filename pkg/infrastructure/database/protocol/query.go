package protocol

import "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"

type Query interface {
	GetName() string
	SetName(string)
	GetOperator() misc.QueryOperator
	GetOperand() *misc.Operand
	GetModel() Identity
	GetFields() []string
}

type QueryInfo interface {
	GetQuery() []Query
	IsAnd() bool
}

func NewQueryInfo(queries []Query, isAnd bool) QueryInfo {
	return basicQueryInfo{queries: queries, isAnd: isAnd}
}

type basicQueryInfo struct {
	queries []Query
	isAnd   bool
}

func (b basicQueryInfo) GetQuery() []Query {
	return b.queries
}

func (b basicQueryInfo) IsAnd() bool {
	return b.isAnd
}

type basicQuery struct {
	name    string
	op      misc.QueryOperator
	operand *misc.Operand
	model   Identity
	fields  []string
}

func NewQueryWithModel(name string, op misc.QueryOperator, model Identity, fields []string, operand *misc.Operand) Query {
	return &basicQuery{name: name, op: op, operand: operand, fields: fields, model: model}
}

func NewQuery(name string, op misc.QueryOperator, operand *misc.Operand) Query {
	return &basicQuery{name: name, op: op, operand: operand, fields: []string{}, model: nil}
}

func NewEmptyQuery(name string, op misc.QueryOperator) Query {
	return &basicQuery{name: name, op: op, operand: nil, fields: []string{}, model: nil}
}

func (b basicQuery) GetModel() Identity {
	return b.model
}

func (b basicQuery) GetFields() []string {
	return b.fields
}

func (b basicQuery) GetName() string {
	return b.name
}

func (b *basicQuery) SetName(name string) {
	b.name = name
}

func (b basicQuery) GetOperator() misc.QueryOperator {
	return b.op
}

func (b basicQuery) GetOperand() *misc.Operand {
	return b.operand
}

func MapQueryWithFields(q misc.Query, identity Identity, fields []string) Query {
	return NewQueryWithModel(q.GetName(), q.GetOperator(), identity, fields, q.GetOperand())
}

func MapQueryWithModel(q misc.Query, identity Identity) Query {
	return MapQueryWithFields(q, identity, []string{})
}

func MapQueriesWithModel(qs []misc.Query, identity Identity) []Query {
	return MapQueriesWithFields(qs, identity, []string{})
}

func MapQueries(qs []misc.Query) []Query {
	return MapQueriesWithFields(qs, nil, []string{})
}

func MapQueryInfo(qf misc.QueryInfo) QueryInfo {
	if qf == nil {
		return NewQueryInfo([]Query{}, true)
	}

	queries := MapQueriesWithFields(qf.GetQuery(), nil, []string{})
	return NewQueryInfo(queries, qf.IsAnd())
}

func MapQueriesWithFields(qs []misc.Query, identity Identity, fields []string) (out []Query) {
	out = []Query{}
	for _, v := range qs {
		out = append(out, MapQueryWithFields(v, identity, fields))
	}
	return out
}
