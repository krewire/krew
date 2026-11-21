package ai

import (
	"reflect"
	"strings"
)

// GenerateSchema produces a standard JSON Schema map for any Go type via reflection.
func GenerateSchema(t reflect.Type) map[string]any {
	return generateSchemaWithVisited(t, make(map[reflect.Type]bool))
}

func generateSchemaWithVisited(t reflect.Type, visited map[reflect.Type]bool) map[string]any {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	schema := make(map[string]any)

	switch t.Kind() {
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		schema["items"] = generateSchemaWithVisited(t.Elem(), visited)
	case reflect.Map:
		schema["type"] = "object"
	case reflect.Struct:
		schema["type"] = "object"
		if visited[t] {
			return schema
		}
		visited[t] = true

		properties := make(map[string]any)
		var required []string

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			fieldName := field.Name
			var omitempty bool
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				}
				for _, p := range parts[1:] {
					if p == "omitempty" {
						omitempty = true
					}
				}
			}

			fieldSchema := generateSchemaWithVisited(field.Type, visited)
			if desc := field.Tag.Get("description"); desc != "" {
				fieldSchema["description"] = desc
			}

			properties[fieldName] = fieldSchema

			// Check required status
			valTag := field.Tag.Get("validate")
			if strings.Contains(valTag, "required") || (!omitempty && field.Type.Kind() != reflect.Pointer) {
				required = append(required, fieldName)
			}
		}

		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}
	default:
		schema["type"] = "string"
	}

	return schema
}
