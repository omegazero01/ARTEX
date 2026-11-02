package server

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
	actool "github.com/Autumn-27/norma/tool"
)

// TestMCPUsageLedgerLive drives a REAL stdio MCP server through the metered
// wrapper and asserts the ledger row the MCP page reads. Skipped unless
// ARTEX_MCP_LIVE_CMD points at the server binary and ARTEX_PG_DSN is set.
//
//	ARTEX_MCP_LIVE_CMD='E:/tools/yakit/.../yak.exe' \
//	ARTEX_PG_DSN='postgres://...' go test ./server -run TestMCPUsageLedgerLive -v
func TestMCPUsageLedgerLive(t *testing.T) {
	cmd := os.Getenv("ARTEX_MCP_LIVE_CMD")
	dsn := os.Getenv("ARTEX_PG_DSN")
	if cmd == "" || dsn == "" {
		t.Skip("需要 ARTEX_MCP_LIVE_CMD 与 ARTEX_PG_DSN")
	}
	d, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer d.Close()

	m := &db.MCPServer{
		Name:      "zz-yakit-live",
		Transport: "stdio",
		Command:   cmd,
		Args:      json.RawMessage(`["mcp","--transport","stdio"]`),
		Enabled:   true,
	}
	id, err := d.SaveMCP(m)
	if err != nil {
		t.Fatalf("save mcp: %v", err)
	}
	defer func() {
		_, _ = d.Exec(`DELETE FROM mcp_usage WHERE server_id=$1`, id)
		_ = d.DeleteMCP(id)
	}()
	m.ID = id

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cl, err := connectMCP(ctx, m)
	if err != nil {
		t.Fatalf("connect mcp: %v", err)
	}
	defer cl.Close()

	tools, err := cl.Tools(ctx)
	if err != nil || len(tools) == 0 {
		t.Fatalf("tools/list: %v (n=%d)", err, len(tools))
	}
	t.Logf("tools discovered: %d", len(tools))

	want := "mcp__" + m.Name + "__get_mitm_filter"
	var target actool.CoreTool
	for _, tt := range tools {
		if tt.Name() == want {
			target = tt
			break
		}
	}
	if target == nil {
		t.Fatalf("tool %s not exposed by server", want)
	}

	wrapped := meterMCPTool(target, d, id, m.Name, "worker", agent.RunInfo{TaskID: 990002})
	res, callErr := wrapped.Call(ctx, json.RawMessage(`{}`), &actool.ToolContext{MaxOutputChars: 4000})
	t.Logf("real call: err=%v result=%+v", callErr, res)

	stats, err := d.MCPToolUsageStats(id)
	if err != nil {
		t.Fatalf("tool stats: %v", err)
	}
	if len(stats) != 1 || stats[0].ToolName != "get_mitm_filter" || stats[0].Calls != 1 {
		t.Fatalf("ledger tool stats mismatch: %+v", stats)
	}
	if len(stats[0].Agents) != 1 || stats[0].Agents[0] != "worker" {
		t.Fatalf("ledger agents mismatch: %+v", stats[0].Agents)
	}

	serverStats, err := d.MCPServerUsageStats()
	if err != nil {
		t.Fatalf("server stats: %v", err)
	}
	if st, ok := serverStats[id]; !ok || st.Calls != 1 || st.ServerName != m.Name {
		t.Fatalf("server stats mismatch: %+v", serverStats[id])
	}

	calls, err := d.RecentMCPCalls(id, 5)
	if err != nil || len(calls) != 1 {
		t.Fatalf("recent calls: %v n=%d", err, len(calls))
	}
	if calls[0].ToolName != "get_mitm_filter" || calls[0].AgentKey != "worker" || calls[0].TaskID != 990002 {
		t.Fatalf("recent call attribution mismatch: %+v", calls[0])
	}
}
