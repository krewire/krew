package ai

import (
	"context"
	"testing"
)

type addArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

func addFunc(_ context.Context, args addArgs) (int, error) {
	return args.A + args.B, nil
}

func TestNewTool_ReflectionAndExecution(t *testing.T) {
	tool, err := NewTool("add_numbers", "Add two numbers together", addFunc)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	if tool.Name != "add_numbers" {
		t.Errorf("expected name 'add_numbers', got %q", tool.Name)
	}
	if tool.Parameters["type"] != "object" {
		t.Errorf("expected parameters type 'object', got %v", tool.Parameters["type"])
	}

	ctx := context.Background()
	result, err := tool.Handler(ctx, []byte(`{"a": 10, "b": 25}`))
	if err != nil {
		t.Fatalf("tool handler returned error: %v", err)
	}
	if sum, ok := result.(int); !ok || sum != 35 {
		t.Errorf("expected sum 35, got %v (%T)", result, result)
	}
}

func TestToolRegistry(t *testing.T) {
	reg := NewToolRegistry()
	err := reg.RegisterFunc("multiply", "Multiplies numbers", func(args struct{ X, Y int }) (int, error) {
		return args.X * args.Y, nil
	})
	if err != nil {
		t.Fatalf("failed to register func: %v", err)
	}

	if len(reg.Tools()) != 1 {
		t.Errorf("expected 1 registered tool, got %d", len(reg.Tools()))
	}

	res, err := reg.Execute(context.Background(), "multiply", []byte(`{"X": 6, "Y": 7}`))
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if val, ok := res.(int); !ok || val != 42 {
		t.Errorf("expected 42, got %v", res)
	}

	// Unknown tool execution
	if _, err := reg.Execute(context.Background(), "non_existent", nil); err == nil {
		t.Errorf("expected error for unknown tool")
	}
}
