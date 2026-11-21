package ai

import (
	"reflect"
	"testing"
)

type sampleUser struct {
	ID      string   `json:"id" description:"Unique user identifier"`
	Age     int      `json:"age"`
	IsAdmin bool     `json:"is_admin,omitempty"`
	Tags    []string `json:"tags"`
	Ignored string   `json:"-"`
}

func TestGenerateSchema_PrimitivesAndStructs(t *testing.T) {
	schema := GenerateSchema(reflect.TypeOf(sampleUser{}))

	if schema["type"] != "object" {
		t.Fatalf("expected type object, got %v", schema["type"])
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties map, got %T", schema["properties"])
	}

	// Verify ID
	idProp, ok := props["id"].(map[string]any)
	if !ok || idProp["type"] != "string" || idProp["description"] != "Unique user identifier" {
		t.Errorf("invalid id schema: %v", idProp)
	}

	// Verify Age
	ageProp, ok := props["age"].(map[string]any)
	if !ok || ageProp["type"] != "integer" {
		t.Errorf("invalid age schema: %v", ageProp)
	}

	// Verify IsAdmin
	adminProp, ok := props["is_admin"].(map[string]any)
	if !ok || adminProp["type"] != "boolean" {
		t.Errorf("invalid is_admin schema: %v", adminProp)
	}

	// Verify Tags
	tagsProp, ok := props["tags"].(map[string]any)
	if !ok || tagsProp["type"] != "array" {
		t.Errorf("invalid tags schema: %v", tagsProp)
	}

	// Verify Ignored is omitted
	if _, exists := props["Ignored"]; exists {
		t.Errorf("ignored field should not be present in schema")
	}
}
