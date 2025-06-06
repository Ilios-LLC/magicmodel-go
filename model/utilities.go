package model

import (
	"fmt"
	"github.com/stoewer/go-strcase"
	"reflect"
	"strings"
)

func SetField(item interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(item)

	// Check if item is a pointer
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot assign to the item passed, item must be a pointer in order to assign")
	}

	// Get the element the pointer refers to
	v = v.Elem()

	if !v.CanAddr() {
		return fmt.Errorf("cannot assign to the item passed, item must be a pointer to an addressable value")
	}

	fieldNames := map[string]int{}
	for i := 0; i < v.NumField(); i++ {
		typeField := v.Type().Field(i)
		jname := typeField.Name
		fieldNames[jname] = i
	}

	fieldNum, ok := fieldNames[fieldName]
	if !ok {
		return fmt.Errorf("field %s does not exist within the provided item", fieldName)
	}
	fieldVal := v.Field(fieldNum)
	fieldVal.Set(reflect.ValueOf(value))
	return nil
}

func generateNewModelInfo(v reflect.Value) {

}

func parseModelName(q interface{}) (string, error) {
	t := reflect.TypeOf(q)

	// Unwrap pointer
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Unwrap slice if needed
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("expected a struct or slice of structs, got %s", t.Kind())
	}

	if t.Name() == "" {
		return "unnamed_struct", fmt.Errorf("cannot use an unnamed struct")
	}

	return strcase.SnakeCase(t.Name()), nil
}

func getFieldValue(value reflect.Value, fieldPath string) (reflect.Value, bool) {
	fields := strings.Split(fieldPath, ".")

	for _, field := range fields {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}

		value = value.FieldByName(field)
		if !value.IsValid() {
			return reflect.Value{}, false
		}
	}

	return value, true
}

func validateInput(q interface{}, operation, structName string) error {
	val := reflect.ValueOf(q)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("the %s operation encountered an error: expected a non-nil pointer to a struct, got %T", operation, q)
	}

	checkPayload := val.Elem()
	if checkPayload.Kind() != reflect.Struct {
		return fmt.Errorf("the %s operation encountered an error: expected a pointer to a struct, got pointer to %s", operation, checkPayload.Kind())
	}

	modelType := reflect.TypeOf((*Model)(nil)).Elem()
	hasModel := false

	for i := 0; i < checkPayload.NumField(); i++ {
		field := checkPayload.Type().Field(i)
		if field.Anonymous && field.Type == modelType {
			hasModel = true
			break
		}
	}

	if !hasModel {
		return fmt.Errorf(fmt.Sprintf(`the %s operation encountered an error: struct %s must embed model.Model (e.g., model.Model `, operation, structName) + "`yaml:\",inline\"`" + `)`)
	}
	return nil
}

func validateInputSlice(q interface{}, operation, structName string) error {
	val := reflect.ValueOf(q)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("the %s operation encountered an error: expected a non-nil pointer to a slice, got %T", operation, q)
	}

	sliceVal := val.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("the %s operation encountered an error: expected a pointer to a slice, got pointer to %s", operation, sliceVal.Kind())
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem() // unwrap *T if slice is []*T
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("the %s operation encountered an error: slice elements must be structs, got %s", operation, elemType.Kind())
	}

	modelType := reflect.TypeOf((*Model)(nil)).Elem()
	hasModel := false

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if field.Anonymous && field.Type == modelType {
			hasModel = true
			break
		}
	}

	if !hasModel {
		return fmt.Errorf(
			"the %s operation encountered an error: struct %s must embed model.Model (e.g., model.Model `yaml:\",inline\"`)",
			operation, structName)
	}

	return nil
}
