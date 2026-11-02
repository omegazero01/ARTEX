package server

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
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
	// Optional: reuse an MCP server already configured in this ARTEX instance so the
	// call lands in the same ledger row the UI reads.
	var cleanup func()
	if rawID := os.Getenv("ARTEX_MCP_LIVE_SERVER_ID"); rawID != "" {
		wantID, convErr := strconv.ParseInt(rawID, 10, 64)
		if convErr != nil {
			t.Fatalf("ARTEX_MCP_LIVE_SERVER_ID: %v", convErr)
		}
		servers, listErr := d.ListMCP()
		if listErr != nil {
			t.Fatalf("list mcp: %v", listErr)
		}
		for _, s := range servers {
			if s.ID == wantID {
				m = s
				break
			}
		}
		if m.ID != wantID {
			t.Fatalf("mcp server %d not configured", wantID)
		}
		cleanup = func() {} // keep the configured server and its ledger rows
	} else {
		id, err := d.SaveMCP(m)
		if err != nil {
			t.Fatalf("save mcp: %v", err)
		}
		m.ID = id
		cleanup = func() {
			_, _ = d.Exec(`DELETE FROM mcp_usage WHERE server_id=$1`, id)
			_ = d.DeleteMCP(id)
		}
	}
	defer cleanup()
	id := m.ID

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

	before, err := d.MCPToolUsageStats(id)
	if err != nil {
		t.Fatalf("baseline tool stats: %v", err)
	}
	beforeCalls := 0
	for _, st := range before {
		beforeCalls += st.Calls
	}

	wrapped := meterMCPTool(target, d, id, m.Name, "worker", agent.RunInfo{TaskID: 990002})
	res, callErr := wrapped.Call(ctx, json.RawMessage(`{}`), &actool.ToolContext{MaxOutputChars: 4000})
	t.Logf("real call: err=%v result=%+v", callErr, res)

	stats, err := d.MCPToolUsageStats(id)
	if err != nil {
		t.Fatalf("tool stats: %v", err)
	}
	var gotTool *db.MCPUsageStat
	afterCalls := 0
	for i := range stats {
		afterCalls += stats[i].Calls
		if stats[i].ToolName == "get_mitm_filter" {
			gotTool = &stats[i]
		}
	}
	if afterCalls != beforeCalls+1 {
		t.Fatalf("ledger calls: want %d, got %d (%+v)", beforeCalls+1, afterCalls, stats)
	}
	if gotTool == nil || gotTool.Calls < 1 {
		t.Fatalf("ledger tool stats missing get_mitm_filter: %+v", stats)
	}
	if len(gotTool.Agents) != 1 || gotTool.Agents[0] != "worker" {
		t.Fatalf("ledger agents mismatch: %+v", gotTool.Agents)
	}

	serverStats, err := d.MCPServerUsageStats()
	if err != nil {
		t.Fatalf("server stats: %v", err)
	}
	if st, ok := serverStats[id]; !ok || st.Calls < 1 || st.ServerName != m.Name {
		t.Fatalf("server stats mismatch: %+v", serverStats[id])
	}

	calls, err := d.RecentMCPCalls(id, 5)
	if err != nil || len(calls) < 1 {
		t.Fatalf("recent calls: %v n=%d", err, len(calls))
	}
	if calls[0].ToolName != "get_mitm_filter" || calls[0].AgentKey != "worker" || calls[0].TaskID != 990002 {
		t.Fatalf("recent call attribution mismatch: %+v", calls[0])
	}
}
