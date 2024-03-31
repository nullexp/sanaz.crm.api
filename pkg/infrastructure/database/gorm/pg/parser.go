package pg

import (
	"fmt"

	database "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	dbspec "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"gorm.io/gorm"
)

type statelessParser struct{}

func NewParser() statelessParser {
	return statelessParser{}
}

func (statelessParser) ParsePage(db *gorm.DB, page misc.Pagination) *gorm.DB {
	if page != nil {
		return db.Limit(int(page.GetLimit())).Offset(int(page.GetSkip()))
	}

	return db
}

const (
	SqlDescending = "DESC"

	SqlAscending = "ASC"
)

func (statelessParser) ParseSort(db *gorm.DB, sorts ...misc.Sort) *gorm.DB {
	parse := func(asc bool) string {
		if asc {
			return SqlAscending
		}

		return SqlDescending
	}

	for _, s := range sorts {

		q := fmt.Sprintf("%s %s", s.GetName(), parse(s.IsAscending()))

		db = db.Order(q)

	}

	return db
}

const (
	SqlIsNull = "IS NULL"

	SqlIsNotNull = "IS NOT NULL"
)

type SqlOperator string

const (
	SqlOperatorEqual SqlOperator = "="

	SqlOperatorNotEqual SqlOperator = "!="

	SqlEqualOrMoreThan SqlOperator = ">="

	SqlMoreThan SqlOperator = ">"

	SqlLessThan SqlOperator = "<"

	SqlLessOrEqualThan SqlOperator = "<="

	SqlOperatorIn SqlOperator = "IN"

	SqlOperatorNotIn SqlOperator = "NOT IN"

	SqlOperatorEmpty SqlOperator = "ILIKE ''"

	SqlOperatorNotEmpty SqlOperator = "NOT ILIKE ''"

	SqlOperatorLike SqlOperator = "ILIKE"
)

const (
	SqlParameter = "(?)"

	SqlParameterMultiple = "(?)"
)

type ParameterExpectation int

const (
	ParameterExpectationZero ParameterExpectation = iota

	ParameterExpectationSingle

	ParameterExpectationMultiple
)

func getOperatorInfo(op misc.QueryOperator) (SqlOperator, ParameterExpectation) {
	switch op {

	case misc.QueryOperatorEqual:

		return SqlOperatorEqual, ParameterExpectationSingle

	case misc.QueryOperatorNotEqual:

		return SqlOperatorNotEqual, ParameterExpectationSingle

	case misc.QueryOperatorMoreThan:

		return SqlMoreThan, ParameterExpectationSingle

	case misc.QueryOperatorEqualOrMoreThan:

		return SqlEqualOrMoreThan, ParameterExpectationSingle

	case misc.QueryOperatorLessThan:

		return SqlLessThan, ParameterExpectationSingle

	case misc.QueryOperatorEqualOrLessThan:

		return SqlLessOrEqualThan, ParameterExpectationSingle

	case misc.QueryOperatorContain:

		return SqlOperatorIn, ParameterExpectationMultiple

	case misc.QueryOperatorNotContain:

		return SqlOperatorNotIn, ParameterExpectationMultiple

	case misc.QueryOperatorEmpty:

		return SqlOperatorEmpty, ParameterExpectationZero

	case misc.QueryOperatorNotEmpty:

		return SqlOperatorNotEmpty, ParameterExpectationZero

	case misc.QueryOperatorLike:

		return SqlOperatorLike, ParameterExpectationSingle

	}

	panic("unknown Operator")
}

func (statelessParser) ParseSpecification(db *gorm.DB, spec dbspec.Specification) *gorm.DB {
	if spec != nil {

		apl := &applyer{DB: db}

		q := spec.Execute(apl)

		if len(apl.Vars) != 0 {
			return db.Where(q, apl.Vars...)
		} else {
			return db.Where(q)
		}

	}

	return db
}

type applyer struct {
	Vars []any

	DB *gorm.DB

	LastQuery database.Query
}

func (a *applyer) Execute(query database.Query) any {
	parseSpec := func(spec dbspec.Specification) {
		nsession := a.DB.Session(&gorm.Session{})

		if len(query.GetFields()) != 0 {
			nsession.Select(query.GetFields())
		}

		napl := applyer{DB: nsession, Vars: []any{}}

		q := spec.Execute(&napl)

		st := nsession.Model(napl.LastQuery.GetModel()).Select(napl.LastQuery.GetFields())

		if len(napl.Vars) != 0 {
			st = st.Where(q, napl.Vars...)
		} else {
			st = st.Where(q)
		}

		a.Vars = append(a.Vars, st)
	}

	op, prexpt := getOperatorInfo(query.GetOperator())

	if prexpt == ParameterExpectationZero && query.GetOperand() != nil {
		panic(query.GetOperator() + " does not expect operand")
	}

	if (prexpt == ParameterExpectationSingle || prexpt == ParameterExpectationMultiple) && query.GetOperand() == nil {
		panic(query.GetOperator() + " expect  operand")
	}

	if v := query.GetOperand(); v != nil {
		if spec, ok := v.Value.(dbspec.Specification); ok {
			parseSpec(spec)
		} else if specs, ok := v.Value.([]dbspec.Specification); ok {
			for _, v := range specs {
				parseSpec(v)
			}
		} else if q, ok := v.Value.(database.Query); ok {
			parseSpec(dbspec.NewSpecification(q))
		} else if v.Value != nil {

			if query.GetOperator() == misc.QueryOperatorLike {
				v.Value = fmt.Sprintf("%%%s%%", v.Value)
			}

			a.Vars = append(a.Vars, v.Value)

		} else {

			if op == SqlOperatorEqual {
				return fmt.Sprintf("(%s %s)", query.GetName(), SqlIsNull)
			}

			if op == SqlOperatorNotEqual {
				return fmt.Sprintf("(%s %s)", query.GetName(), SqlIsNotNull)
			}

			panic(query.GetOperator() + " is not supported for empty specification")

		}
	}

	a.LastQuery = query

	if query.GetModel() != nil {
		a.DB = a.DB.Model(query.GetModel())
	}

	if len(query.GetFields()) != 0 {
		a.DB = a.DB.Select(query.GetFields())
	}

	sp := SqlParameter

	if prexpt == ParameterExpectationMultiple || prexpt == ParameterExpectationSingle {

		if op == SqlOperatorLike {
			return fmt.Sprintf("(%s::text %s %s)", query.GetName(), op, sp)
		}

		return fmt.Sprintf("(%s %s %s)", query.GetName(), op, sp)

	}

	return fmt.Sprintf("(%s %s)", query.GetName(), op)
}

func (a *applyer) And(left, right dbspec.Specification) any {
	return " " + left.Execute(a).(string) + " AND " + right.Execute(a).(string)
}

func (a *applyer) Not(wrapped dbspec.Specification) any {
	return " NOT (" + wrapped.Execute(a).(string) + ")"
}

func (a *applyer) Or(left, right dbspec.Specification) any {
	return " " + left.Execute(a).(string) + " OR " + right.Execute(a).(string)
}

func (a *applyer) GroupAnd(left, right dbspec.Specification) any {
	return " " + left.Execute(a).(string) + " AND (" + right.Execute(a).(string) + ")"
}

func (a *applyer) GroupOr(left, right dbspec.Specification) any {
	return " " + left.Execute(a).(string) + " OR (" + right.Execute(a).(string) + ")"
}
