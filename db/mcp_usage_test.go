package db

import (
	"testing"
)

// TestMCPUsageLedger writes a few ledger rows and checks the three aggregates the
// MCP page relies on: per-server totals, per-tool totals, and the recent-call list.
// Rows are namespaced by a unique server name so the shared dev DB stays usable and
// the test cleans up after itself.
func TestMCPUsageLedger(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer d.Close()

	const (
		serverName = "zz-test-mcp-ledger"
		serverID   = int64(990001)
	)
	cleanup := func() {
		_, _ = d.Exec(`DELETE FROM mcp_usage WHERE server_name=$1`, serverName)
	}
	cleanup()
	defer cleanup()

	rows := []*MCPUsage{
		{ServerID: serverID, ServerName: serverName, ToolName: "query_http_flow", AgentKey: "worker", TaskID: 991, ExplorationID: 5, IntentID: 7},
		{ServerID: serverID, ServerName: serverName, ToolName: "query_http_flow", AgentKey: "worker", TaskID: 991, ExplorationID: 5},
		{ServerID: serverID, ServerName: serverName, ToolName: "get_mitm_filter", AgentKey: "planner", TaskID: 992, ExplorationID: 6},
		{ServerID: serverID, ServerName: serverName, ToolName: "http_fuzzer", AgentKey: "chatbot", SessionID: "conv-1"},
	}
	for _, r := range rows {
		if err := d.InsertMCPUsage(r); err != nil {
			t.Fatalf("insert %s: %v", r.ToolName, err)
		}
	}

	stats, err := d.MCPServerUsageStats()
	if err != nil {
		t.Fatal(err)
	}
	st, ok := stats[serverID]
	if !ok {
		t.Fatalf("server %d absent from MCPServerUsageStats", serverID)
	}
	if st.Calls != 4 {
		t.Errorf("calls: want 4, got %d", st.Calls)
	}
	if st.Tasks != 2 { // task 991 twice, 992 once; chat rows have no task
		t.Errorf("tasks: want 2, got %d", st.Tasks)
	}
	if len(st.Agents) != 3 {
		t.Errorf("agents: want 3 distinct, got %v", st.Agents)
	}
	if st.LastUsed == nil {
		t.Error("last_used must be set")
	}

	tools, err := d.MCPToolUsageStats(serverID)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 3 {
		t.Fatalf("tool rows: want 3, got %d (%+v)", len(tools), tools)
	}
	if tools[0].ToolName != "query_http_flow" || tools[0].Calls != 2 {
		t.Errorf("busiest tool: want query_http_flow x2, got %+v", tools[0])
	}

	calls, err := d.RecentMCPCalls(serverID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("recent calls: want 2, got %d", len(calls))
	}
	for _, c := range calls {
		if c.ToolName == "" {
			t.Errorf("recent call missing tool name: %+v", c)
		}
	}
}
