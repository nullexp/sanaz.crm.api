package protocol

import "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"

type RequestParameter struct {
	Definition misc.QueryDefinition
	Query      bool
	Optional   bool
}

var ResourceIdParameter = RequestParameter{
	Definition: misc.NewQueryDefinition(misc.Id, []misc.QueryOperator{misc.QueryOperatorEqual}, misc.DataTypeString),
	Query:      false,
	Optional:   false,
}
