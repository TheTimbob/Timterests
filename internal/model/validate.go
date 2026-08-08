package model

import (
	"fmt"
	"reflect"
	"strings"
)

// requiredTag marks a field that must be populated, e.g. `validate:"required"`.
const requiredTag = "required"

// ValidateRequired reports every field tagged as required that is empty.
//
// Driven by struct tags rather than per-type code, so a new document type is
// validated correctly without anything being written for it — the tags on the
// struct are the only place the rules live.
func ValidateRequired(doc any) error {
	missing := missingRequired(reflect.ValueOf(doc))
	if len(missing) == 0 {
		return nil
	}

	if len(missing) == 1 {
		return fmt.Errorf("%s is required", missing[0])
	}

	return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
}

func missingRequired(v reflect.Value) []string {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	var missing []string

	structType := v.Type()

	for i := range structType.NumField() {
		field := structType.Field(i)

		// Embedded structs are walked so fields on Document count too.
		if field.Anonymous {
			missing = append(missing, missingRequired(v.Field(i))...)

			continue
		}

		if !strings.Contains(field.Tag.Get("validate"), requiredTag) {
			continue
		}

		if v.Field(i).IsZero() {
			missing = append(missing, fieldName(field))
		}
	}

	return missing
}

// fieldName prefers the yaml tag, so errors name the field as it appears in the
// uploaded file rather than using the Go field name.
func fieldName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" {
		return field.Name
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return field.Name
	}

	return name
}
