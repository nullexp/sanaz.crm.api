package gin

import (
	"strings"
	"time"

	genericErrors "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error"
	errorProtocol "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/error/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
)

func ParseQuery(def misc.QueryDefinition, qvalue string) (misc.Query, error) {
	smallCount := strings.Count(qvalue, "(")
	biggerCount := strings.Count(qvalue, ")")
	if smallCount != biggerCount {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "malformed query parentheses")
	}

	if smallCount >= 2 {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "unsupported format")
	}

	endWithParentheses := strings.HasSuffix(qvalue, ")")
	startWithParentheses := strings.HasPrefix(qvalue, "(")

	hasPerantheses := smallCount == 1

	if startWithParentheses && endWithParentheses {
		qvalue = strings.TrimPrefix(qvalue, "(")
		qvalue = strings.TrimSuffix(qvalue, ")")
		hasPerantheses = false
	}

	if !hasPerantheses {
		commaSeperated := strings.Split(qvalue, ",")
		if len(commaSeperated) == 1 {
			return parseSimpleEqual(def, commaSeperated[0])
		}
		return parseMultipleEqual(def, commaSeperated)
	}

	if !endWithParentheses {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "must end with parenthesis")
	}

	qvalue = strings.TrimSuffix(qvalue, ")")
	parts := strings.Split(qvalue, "(")

	value := ""
	if len(parts) > 1 {
		value = parts[1]
	}

	op, pamType, err := misc.ParseQueryOperator(parts[0])
	if err != nil {
		return nil, err
	}

	commaSeparated := strings.Split(value, ",")

	if pamType == misc.ParameterExpectationZero && value != "" {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "expect no operator")
	}

	if pamType == misc.ParameterExpectationSingle && (value == "" || len(commaSeparated) != 1) {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "expect operator")
	}

	if pamType == misc.ParameterExpectationMultiple && value == "" {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "expect operator")
	}
	if err != nil {
		return nil, err
	}

	if pamType == misc.ParameterExpectationZero {
		return misc.NewEmptyQuery(def.GetName(), op), nil
	}

	if len(commaSeparated) >= 2 {
		return parseMultiple(def, op, commaSeparated)
	} else {
		return parseSimpleOperands(def, op, value)
	}
}

func parseSimpleEqual(def misc.QueryDefinition, qvalue string) (misc.Query, error) {
	v, ok := misc.ParseValue(def.GetType(), qvalue)

	if !ok {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "invalid data type")
	}
	out := misc.NewQuery(def.GetName(), misc.QueryOperatorEqual, misc.NewOperand(v))
	return out, nil
}

func parseMultipleEqual(def misc.QueryDefinition, qvalue []string) (misc.Query, error) {
	return parseMultiple(def, misc.QueryOperatorContain, qvalue)
}

func parseSimpleOperands(def misc.QueryDefinition, op misc.QueryOperator, rawValue string) (misc.Query, error) {
	v, ok := misc.ParseValue(def.GetType(), rawValue)

	if !ok {
		return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "invalid data type")
	}
	out := misc.NewQuery(def.GetName(), op, misc.NewOperand(v))
	return out, nil
}

func parseMultipleGenericOperands[T any](def misc.QueryDefinition, op misc.QueryOperator, qvalue []string) (misc.Query, error) {
	ops := []T{}
	for _, v := range qvalue {
		v, ok := misc.ParseValue(def.GetType(), v)

		if !ok {
			return nil, errorProtocol.WrapUserOperationError(genericErrors.GenericValidationErrorKey, "invalid data type")
		}
		ops = append(ops, v.(T))
	}

	out := misc.NewQuery(def.GetName(), op, misc.NewOperand(ops))
	return out, nil
}

func parseMultiple(def misc.QueryDefinition, op misc.QueryOperator, qvalue []string) (misc.Query, error) {
	switch def.GetType() {
	case misc.DataTypeBoolean:
		return parseMultipleGenericOperands[bool](def, op, qvalue)
	case misc.DataTypeString:
		return parseMultipleGenericOperands[string](def, op, qvalue)
	case misc.DataTypeDouble:
		return parseMultipleGenericOperands[string](def, op, qvalue)

	case misc.DataTypeInteger:
		return parseMultipleGenericOperands[int](def, op, qvalue)

	case misc.DataTypeLong:
		return parseMultipleGenericOperands[int64](def, op, qvalue)

	case misc.DataTypeTime:
		return parseMultipleGenericOperands[time.Time](def, op, qvalue)

	case misc.DataTypeBase64:
		return parseMultipleGenericOperands[string](def, op, qvalue)

	case misc.DataTypeSearch:
		return parseMultipleGenericOperands[string](def, op, qvalue)

	case misc.DataTypePage:
		return parseMultipleGenericOperands[string](def, op, qvalue)

	case misc.DataTypeStaticPage:
		return parseMultipleGenericOperands[string](def, op, qvalue)

	case misc.DataTypeUInteger:
		return parseMultipleGenericOperands[uint](def, op, qvalue)

	case misc.DataTypeULong:
		return parseMultipleGenericOperands[uint64](def, op, qvalue)

	}
	return parseMultipleGenericOperands[any](def, op, qvalue)
}
