package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	actool "github.com/Autumn-27/norma/tool"
	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
	"github.com/Autumn-27/artex/traffic"
)

// Live wiring check: even if the persisted row carries an obsolete protocol
// (old description + "page" parameter), the resolved tool the agent receives must
// come from the current code contract.
func TestToolResolveKeepsCodeContract(t *testing.T) {
	dsn, _, err := db.DSN()
	if err != nil {
		t.Skipf("no database config (%v) — skipping", err)
	}
	pg, err := db.Open(dsn)
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	// Close AFTER the restore cleanup (t.Cleanup is LIFO); a deferred Close
	// would run first and make the restore silently fail.
	t.Cleanup(func() { pg.Close() })
	wireTools(pg, map[string]actool.CoreTool{})
	t.Cleanup(func() { agent.ToolResolve = nil })

	old, err := pg.GetTool("traffic_get")
	if err != nil || old == nil {
		t.Fatalf("load traffic_get: %v", err)
	}
	agents, _ := json.Marshal(old.Agents)
	poison := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":   map[string]any{"type": "string"},
			"page": map[string]any{"type": "integer"},
		},
	}
	raw, _ := json.Marshal(poison)
	if err := pg.UpdateTool("traffic_get", "旧说明：用 page 翻页", raw, agents, old.Enabled); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pg.UpdateTool("traffic_get", old.Description, old.Schema, agents, old.Enabled)
	})

	resolved := agent.ToolResolve(context.Background(), "worker", traffic.SeedToolMetas())
	var got actool.CoreTool
	for _, tool := range resolved {
		if tool.Name() == "traffic_get" {
			got = tool
		}
	}
	if got == nil {
		t.Fatal("traffic_get missing after resolve")
	}
	if strings.Contains(got.Description(), "page 翻页") {
		t.Fatalf("stale description survived: %q", got.Description())
	}
	props := got.InputSchema()["properties"].(map[string]any)
	if props["part"] == nil || props["offset"] == nil || props["page"] != nil {
		t.Fatalf("stale schema survived: %v", props)
	}
}
