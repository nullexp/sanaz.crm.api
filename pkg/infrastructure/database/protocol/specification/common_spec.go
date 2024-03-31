package specification

import (
	"errors"
	"strings"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
)

func GetIdExistSpecification(id any) Specification {
	return NewQuerySpecification(misc.Id, misc.QueryOperatorEqual, misc.NewOperand(id))
}

func GetIdContainSpecification(ids []any) Specification {
	return NewQuerySpecification(misc.Id, misc.QueryOperatorContain, misc.NewOperand(ids))
}

func GenerateSearchSpecification(searchQuery, field string) Specification {
	var spec Specification = NewQuerySpecification(field, misc.QueryOperatorLike, misc.NewOperand(searchQuery))

	parts := strings.Split(searchQuery, " ")

	for _, p := range parts {

		trimmed := strings.Trim(p, " ")

		if trimmed == "" {
			continue
		}

		spec = spec.Or(NewQuerySpecification(field, misc.QueryOperatorEqual, misc.NewOperand(trimmed)))

	}

	return spec
}

var ErrFieldsMustNotBeEmpty = errors.New("fields should not be empty")

func GenerateSearchSpecifications(searchQuery string, fields []string) Specification {
	if len(fields) == 0 {
		panic(ErrFieldsMustNotBeEmpty)
	}

	appendOrSpec := func(src Specification, dest Specification) (out Specification) {
		if dest == nil {
			out = src
		} else {
			out = src.Or(dest)
		}

		return out
	}

	var out Specification

	for _, field := range fields {
		out = appendOrSpec(GenerateSearchSpecification(searchQuery, field), out)
	}

	return out
}
