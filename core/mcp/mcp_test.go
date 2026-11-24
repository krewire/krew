package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	kiwconfig "github.com/krewire/krew/core/config"
)

func setupTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	tmpDir := t.TempDir()

	// Write dummy krewire.yaml
	yamlContent := `title: "TestApp"
project:
  name: "testapp"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "krewire.yaml"), []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("failed to write krewire.yaml: %v", err)
	}

	// Create a dummy page
	pagesDir := filepath.Join(tmpDir, "pages")
	_ = os.MkdirAll(pagesDir, 0o755)
	_ = os.WriteFile(filepath.Join(pagesDir, "index.kiw"), []byte("<h1>Home</h1>"), 0o644)

	config, _ := kiwconfig.Load(tmpDir)
	return NewServer(tmpDir, config), tmpDir
}

func TestInitialize(t *testing.T) {
	srv, _ := setupTestServer(t)
	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	res := srv.HandleRequest(context.Background(), req)
	if res == nil || res.Error != nil {
		t.Fatalf("initialize failed: %+v", res)
	}

	result, ok := res.Result.(map[string]any)
	if !ok {
		t.Fatalf("result is not map[string]any: %T", res.Result)
	}
	if result["protocolVersion"] != ProtocolVersion {
		t.Errorf("expected protocolVersion %s, got %v", ProtocolVersion, result["protocolVersion"])
	}
}

func TestPing(t *testing.T) {
	srv, _ := setupTestServer(t)
	req := Request{
		JSONRPC: "2.0",
		ID:      "ping-1",
		Method:  "ping",
	}

	res := srv.HandleRequest(context.Background(), req)
	if res == nil || res.Error != nil {
		t.Fatalf("ping failed: %+v", res)
	}
}

func TestResources(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	// 1. resources/list
	listReq := Request{JSONRPC: "2.0", ID: 1, Method: "resources/list"}
	res := srv.HandleRequest(ctx, listReq)
	if res == nil || res.Error != nil {
		t.Fatalf("resources/list failed: %+v", res)
	}

	// 2. resources/read info
	params, _ := json.Marshal(map[string]string{"uri": "krewire://project/info"})
	readReq := Request{JSONRPC: "2.0", ID: 2, Method: "resources/read", Params: params}
	res = srv.HandleRequest(ctx, readReq)
	if res == nil || res.Error != nil {
		t.Fatalf("resources/read failed: %+v", res)
	}
	m := res.Result.(map[string]any)
	contents := m["contents"].([]map[string]any)
	if len(contents) == 0 || !strings.Contains(contents[0]["text"].(string), "TestApp") {
		t.Errorf("resources/read unexpected content: %v", contents)
	}

	// 3. resources/read routes
	params, _ = json.Marshal(map[string]string{"uri": "krewire://project/routes"})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 3, Method: "resources/read", Params: params})
	if res == nil || res.Error != nil {
		t.Fatalf("resources/read routes failed: %+v", res)
	}

	// 4. resources/read metrics & health
	params, _ = json.Marshal(map[string]string{"uri": "krewire://runtime/metrics"})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 4, Method: "resources/read", Params: params})
	if res == nil || res.Error != nil {
		t.Fatalf("resources/read metrics failed: %+v", res)
	}

	params, _ = json.Marshal(map[string]string{"uri": "krewire://runtime/health"})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 5, Method: "resources/read", Params: params})
	if res == nil || res.Error != nil {
		t.Fatalf("resources/read health failed: %+v", res)
	}
}

func TestTools(t *testing.T) {
	srv, root := setupTestServer(t)
	ctx := context.Background()

	// 1. tools/list
	res := srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 1, Method: "tools/list"})
	if res == nil || res.Error != nil {
		t.Fatalf("tools/list failed: %+v", res)
	}

	// 2. kiw_inspect_project
	callParams, _ := json.Marshal(map[string]any{
		"name":      "kiw_inspect_project",
		"arguments": map[string]any{},
	})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 2, Method: "tools/call", Params: callParams})
	if res == nil || res.Error != nil {
		t.Fatalf("kiw_inspect_project failed: %+v", res)
	}

	// 3. kiw_list_routes
	callParams, _ = json.Marshal(map[string]any{
		"name":      "kiw_list_routes",
		"arguments": map[string]any{},
	})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 3, Method: "tools/call", Params: callParams})
	if res == nil || res.Error != nil {
		t.Fatalf("kiw_list_routes failed: %+v", res)
	}

	// 4. kiw_make_scaffold
	callParams, _ = json.Marshal(map[string]any{
		"name": "kiw_make_scaffold",
		"arguments": map[string]any{
			"kind": "component",
			"name": "Badge",
		},
	})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 4, Method: "tools/call", Params: callParams})
	if res == nil || res.Error != nil {
		t.Fatalf("kiw_make_scaffold failed: %+v", res)
	}
	m := res.Result.(map[string]any)
	if m["isError"] == true {
		t.Fatalf("kiw_make_scaffold returned error: %v", m["content"])
	}

	// Verify file was created in tmpDir
	badgeFile := filepath.Join(root, "components", "Badge.kiw")
	if _, err := os.Stat(badgeFile); os.IsNotExist(err) {
		t.Fatalf("expected component file %s does not exist", badgeFile)
	}

	// 5. kiw_context
	callParams, _ = json.Marshal(map[string]any{
		"name":      "kiw_context",
		"arguments": map[string]any{},
	})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 5, Method: "tools/call", Params: callParams})
	if res == nil || res.Error != nil {
		t.Fatalf("kiw_context failed: %+v", res)
	}
	m = res.Result.(map[string]any)
	if m["isError"] == true {
		t.Fatalf("kiw_context returned error: %v", m["content"])
	}

	// 6. kiw_verify
	callParams, _ = json.Marshal(map[string]any{
		"name": "kiw_verify",
		"arguments": map[string]any{
			"task_id":     "mcp-test-task",
			"save_ledger": true,
		},
	})
	res = srv.HandleRequest(ctx, Request{JSONRPC: "2.0", ID: 6, Method: "tools/call", Params: callParams})
	if res == nil || res.Error != nil {
		t.Fatalf("kiw_verify failed: %+v", res)
	}
}

func TestStdioTransport(t *testing.T) {
	srv, _ := setupTestServer(t)

	inputJSON := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"ping"}` + "\n"
	in := bytes.NewBufferString(inputJSON)
	var out bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := RunStdio(ctx, srv, in, &out); err != nil {
		t.Fatalf("RunStdio failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 responses, got %d: %s", len(lines), out.String())
	}

	var resp1, resp2 Response
	if err := json.Unmarshal([]byte(lines[0]), &resp1); err != nil {
		t.Fatalf("unmarshal resp1 error: %v", err)
	}
	if resp1.ID != float64(1) {
		t.Errorf("expected id 1, got %v", resp1.ID)
	}

	if err := json.Unmarshal([]byte(lines[1]), &resp2); err != nil {
		t.Fatalf("unmarshal resp2 error: %v", err)
	}
	if resp2.ID != float64(2) {
		t.Errorf("expected id 2, got %v", resp2.ID)
	}
}

func TestHTTPTransport(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv)
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	// 1. Direct POST message
	msgBody := `{"jsonrpc":"2.0","id":"http-1","method":"initialize"}`
	resp, err := http.Post(server.URL+"/mcp/message", "application/json", strings.NewReader(msgBody))
	if err != nil {
		t.Fatalf("POST /mcp/message failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var jsonResp Response
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if jsonResp.ID != "http-1" {
		t.Errorf("expected id 'http-1', got %v", jsonResp.ID)
	}
}
