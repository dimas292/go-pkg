package validator

import (
	"reflect"
)

// sanitizeFields sanitizes all exported string fields of a struct pointer.
// It recursively processes embedded structs.
func sanitizeFields(v interface{}) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			if field.CanSet() {
				field.SetString(SanitizeString(field.String()))
			}
		case reflect.Struct:
			// Recursively sanitize embedded structs
			if field.CanAddr() {
				sanitizeFields(field.Addr().Interface())
			}
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.String {
				sanitized := SanitizeString(field.Elem().String())
				field.Elem().SetString(sanitized)
			}
		}
	}
}
