package model

import (
	"fmt"
	"github.com/stoewer/go-strcase"
	"reflect"
	"strings"
)

func SetField(item interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(item).Elem()
	if !v.CanAddr() {
		return fmt.Errorf("cannot assign to the item passed, item must be a pointer in order to assign")
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

func parseModelName(q interface{}) string {
	a := reflect.TypeOf(q).String()
	b := a[strings.LastIndex(a, ",")+1:]
	c := b[strings.LastIndex(b, ".")+1:]
	return strcase.SnakeCase(c)
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
