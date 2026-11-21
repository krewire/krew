package tools_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krew/pkg/tools"
)

func TestASTSearchTool(t *testing.T) {
	tempDir := t.TempDir()

	sampleCode := `package sample

// Engine performs calculations.
type Engine struct {
	ID string
}

// Compute processes inputs.
func (e *Engine) Compute(val int) int {
	return val * 2
}

// HelperFunc standalone function.
func HelperFunc() string {
	return "ok"
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "sample.go"), []byte(sampleCode), 0644); err != nil {
		t.Fatal(err)
	}

	astTool := tools.NewASTSearchTool(tempDir)

	// 1. Search for function "Helper"
	args, _ := json.Marshal(map[string]any{"query": "Helper"})
	res, err := astTool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("ast search failed: %v", err)
	}
	syms, ok := res.([]tools.ASTSymbol)
	if !ok || len(syms) == 0 {
		t.Fatalf("expected symbols, got: %v", res)
	}
	if syms[0].Name != "HelperFunc" || syms[0].Kind != "function" {
		t.Errorf("unexpected symbol: %+v", syms[0])
	}

	// 2. Search for struct "Engine"
	argsStruct, _ := json.Marshal(map[string]any{"query": "Engine", "kind": "struct"})
	resStruct, err := astTool.Execute(context.Background(), argsStruct)
	if err != nil {
		t.Fatalf("ast search struct failed: %v", err)
	}
	symsStruct := resStruct.([]tools.ASTSymbol)
	if len(symsStruct) == 0 || symsStruct[0].Name != "Engine" {
		t.Fatalf("expected struct Engine, got: %v", resStruct)
	}

	// 3. Search for method "Compute"
	argsMethod, _ := json.Marshal(map[string]any{"query": "Compute", "kind": "method"})
	resMethod, err := astTool.Execute(context.Background(), argsMethod)
	if err != nil {
		t.Fatalf("ast search method failed: %v", err)
	}
	symsMethod := resMethod.([]tools.ASTSymbol)
	if len(symsMethod) == 0 || symsMethod[0].Name != "Compute" {
		t.Fatalf("expected method Compute, got: %v", resMethod)
	}
}
