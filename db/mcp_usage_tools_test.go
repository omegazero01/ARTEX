package db

import "testing"

// TestMCPToolsDetailedCarriesCalls checks the tools tab shows per-tool call counts
// from the ledger, not just the cached names/descriptions.
func TestMCPToolsDetailedCarriesCalls(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer d.Close()

	const name = "zz-test-mcp-tools-calls"
	// A previous aborted run may have left the row behind.
	_, _ = d.Exec(`DELETE FROM mcp_servers WHERE name=$1`, name)
	m := &MCPServer{Name: name, Transport: "stdio", Command: "true", Args: []byte("[]"), Enabled: true}
	id, err := d.SaveMCP(m)
	if err != nil {
		t.Fatalf("save mcp: %v", err)
	}
	// defer (not t.Cleanup): t.Cleanup runs after defer d.Close(), leaving rows behind.
	defer func() {
		_, _ = d.Exec(`DELETE FROM mcp_usage WHERE server_id=$1`, id)
		_ = d.DeleteMCP(id)
	}()

	// Cache rows carry the model-facing prefix while the ledger stores the bare name.
	if err := d.SaveMCPTools(id, []MCPTool{
		{Name: "mcp__" + name + "__query_http_flow", Description: "query flows"},
		{Name: "mcp__" + name + "__get_mitm_filter", Description: "mitm filter"},
	}); err != nil {
		t.Fatalf("save tools: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := d.InsertMCPUsage(&MCPUsage{ServerID: id, ServerName: name, ToolName: "query_http_flow", AgentKey: "worker", TaskID: 993}); err != nil {
			t.Fatalf("insert usage: %v", err)
		}
	}

	tools, err := d.MCPToolsDetailed(id)
	if err != nil {
		t.Fatalf("tools detailed: %v", err)
	}
	got := map[string]int{}
	for _, tool := range tools {
		got[tool.Name] = tool.Calls
	}
	prefixed := "mcp__" + name + "__query_http_flow"
	if got[prefixed] != 3 {
		t.Errorf("%s calls: want 3, got %d (%v)", prefixed, got[prefixed], got)
	}
	if got["mcp__"+name+"__get_mitm_filter"] != 0 {
		t.Errorf("get_mitm_filter calls: want 0, got %v", got)
	}
}
