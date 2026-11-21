package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ASTSymbol represents an extracted symbol from codebase AST.
type ASTSymbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"` // "function", "method", "type", "struct", "interface"
	Signature string `json:"signature,omitempty"`
	Doc       string `json:"doc,omitempty"`
	File      string `json:"file"`
	Line      int    `json:"line"`
}

// ASTSearchTool provides AST semantic search across Go files in the workspace.
type ASTSearchTool struct {
	Workspace string
}

// NewASTSearchTool creates a new ASTSearchTool rooted at workspace.
func NewASTSearchTool(workspace string) *ASTSearchTool {
	return &ASTSearchTool{Workspace: workspace}
}

func (t *ASTSearchTool) Name() string {
	return "ast_search"
}

func (t *ASTSearchTool) Description() string {
	return "Search codebase symbols (functions, structs, interfaces, methods, types) using Go Abstract Syntax Tree (AST) parser. Returns symbol signatures, file locations, line numbers, and doc comments."
}

func (t *ASTSearchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Symbol name or substring to search for (e.g. 'Runner', 'Create', 'Workspace')",
			},
			"kind": map[string]any{
				"type":        "string",
				"description": "Filter by kind: 'function', 'type', 'struct', 'interface', or '' for all",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Optional subdirectory to limit search scope",
			},
		},
		"required": []string{"query"},
	}
}

func (t *ASTSearchTool) Execute(_ context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Query string `json:"query"`
		Kind  string `json:"kind"`
		Path  string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	query := strings.ToLower(strings.TrimSpace(input.Query))
	if query == "" {
		return nil, fmt.Errorf("query parameter cannot be empty")
	}

	searchRoot := t.Workspace
	if input.Path != "" {
		cleaned := filepath.Clean(input.Path)
		if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
			return nil, fmt.Errorf("path outside workspace not allowed: %s", input.Path)
		}
		searchRoot = filepath.Join(t.Workspace, cleaned)
	}

	fset := token.NewFileSet()
	var results []ASTSymbol
	limit := 50 // cap maximum symbols returned

	err := filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip non-Go files, vendor, .git, and hidden directories
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return nil // best-effort parsing
		}

		relPath, _ := filepath.Rel(t.Workspace, path)

		ast.Inspect(node, func(n ast.Node) bool {
			if len(results) >= limit {
				return false
			}

			switch decl := n.(type) {
			case *ast.FuncDecl:
				name := decl.Name.Name
				if strings.Contains(strings.ToLower(name), query) {
					kind := "function"
					if decl.Recv != nil && len(decl.Recv.List) > 0 {
						kind = "method"
					}
					if matchKind(input.Kind, kind) {
						doc := decl.Doc.Text()
						pos := fset.Position(decl.Pos())
						results = append(results, ASTSymbol{
							Name:      name,
							Kind:      kind,
							Signature: formatFuncSignature(decl),
							Doc:       strings.TrimSpace(doc),
							File:      relPath,
							Line:      pos.Line,
						})
					}
				}

			case *ast.GenDecl:
				if decl.Tok == token.TYPE {
					for _, spec := range decl.Specs {
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						name := typeSpec.Name.Name
						if strings.Contains(strings.ToLower(name), query) {
							kind := "type"
							switch typeSpec.Type.(type) {
							case *ast.StructType:
								kind = "struct"
							case *ast.InterfaceType:
								kind = "interface"
							}
							if matchKind(input.Kind, kind) {
								doc := decl.Doc.Text()
								pos := fset.Position(typeSpec.Pos())
								results = append(results, ASTSymbol{
									Name: name,
									Kind: kind,
									Doc:  strings.TrimSpace(doc),
									File: relPath,
									Line: pos.Line,
								})
							}
						}
					}
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ast search walk error: %w", err)
	}

	return results, nil
}

func matchKind(filter, actual string) bool {
	if filter == "" {
		return true
	}
	f := strings.ToLower(strings.TrimSpace(filter))
	return f == actual
}

func formatFuncSignature(decl *ast.FuncDecl) string {
	var b strings.Builder
	b.WriteString("func ")
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		b.WriteString("(...) ")
	}
	b.WriteString(decl.Name.Name)
	b.WriteString("(...)")
	return b.String()
}
