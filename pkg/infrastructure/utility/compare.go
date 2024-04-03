package utility

import (
	"errors"
	"reflect"
)

var (
	CompareErrorCurrentCannotBeNil = errors.New("Compare failed, the `current` argument cannot be nil!")
	CompareErrorNewCannotBeNil     = errors.New("Compare failed, the `newOne` argument cannot be nil!")
)

func Compare[T comparable](current T, new T) (map[string]any, error) {
	currentValue := reflect.ValueOf(current)
	newValue := reflect.ValueOf(new)

	if currentValue.Kind() == reflect.Ptr {
		currentValue = currentValue.Elem()
		newValue = newValue.Elem()
	}

	if currentValue.Kind() == reflect.Invalid {
		return nil, CompareErrorCurrentCannotBeNil
	}

	if newValue.Kind() == reflect.Invalid {
		return nil, CompareErrorNewCannotBeNil
	}

	if currentValue.Kind() != reflect.Struct {
		return nil, errors.New("Compare failed, arguments expected to be struct or pointer of a struct!")
	}

	cmp := compareStruct(currentValue, newValue)
	return cmp, nil
}

func compareStruct(currentValue reflect.Value, newValue reflect.Value) map[string]any {
	cmp := make(map[string]any)
	currentType := currentValue.Type()
	var fieldKind reflect.Kind
	var fieldName string
	for fcou := 0; fcou < currentValue.NumField(); fcou++ {
		fieldType := currentType.Field(fcou)
		currentField := currentValue.Field(fcou)
		newField := newValue.Field(fcou)
		fieldName = fieldType.Name
		fieldKind = fieldType.Type.Kind()
		if fieldKind == reflect.Struct {
			strcutDiff := compareStruct(currentField, newField)
			if strcutDiff != nil {
				cmp[fieldName] = strcutDiff
			}
			continue
		}
		if !fieldType.Anonymous &&
			fieldKind == reflect.Ptr ||
			fieldKind == reflect.Interface ||
			fieldKind == reflect.Slice {
			// Derefing the field
			if fieldKind == reflect.Ptr {
				currentField = currentField.Elem()
				fieldKind = currentField.Type().Kind()
				newField = newField.Elem()
			}
			if fieldKind == reflect.Struct {
				structCmp := compareStruct(currentField, newField)
				if structCmp != nil {
					cmp[fieldName] = structCmp
				}
			}
			//Skipped for now
			//if fieldKind == reflect.Slice {
			//}
			//if fieldKind == reflect.Interface {
			//}
			continue
		}

		if currentField.Comparable() {
			if !currentField.Equal(newField) {
				cmp[fieldName] = newField.Interface()
			}
		}
	}
	if len(cmp) > 0 {
		return cmp
	}
	return nil
}
