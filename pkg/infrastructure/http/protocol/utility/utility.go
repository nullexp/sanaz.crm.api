package utility

import "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"

func GetByType[T any](key string, r protocol.Request) ([]*T, bool) {
	raw, ok := r.Get(key)

	if !ok {
		return nil, ok
	}

	out := []*T{}
	rawArray, _ := raw.([]interface{})
	for _, v := range rawArray {

		t, ok := v.(*T)
		if !ok {
			return nil, ok
		}
		out = append(out, t)
	}
	return out, true
}
