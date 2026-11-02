package server

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/Autumn-27/artex/agent"
	"github.com/Autumn-27/artex/db"
	actool "github.com/Autumn-27/norma/tool"
)

// meteredMCPTool records one ledger row per MCP tool call and then delegates.
// Arguments and results are never copied into the ledger: attribution only.
type meteredMCPTool struct {
	actool.CoreTool
	recorder   interface{ InsertMCPUsage(*db.MCPUsage) error }
	serverID   int64
	serverName string
	toolName   string
	agentKey   string
	ri         agent.RunInfo
}

// meterMCPTool wraps t with the usage ledger. A nil recorder returns t unchanged,
// so the ledger can be switched off without touching the call path.
func meterMCPTool(t actool.CoreTool, recorder interface{ InsertMCPUsage(*db.MCPUsage) error }, serverID int64, serverName, agentKey string, ri agent.RunInfo) actool.CoreTool {
	if recorder == nil || t == nil {
		return t
	}
	name := strings.TrimPrefix(t.Name(), "mcp__"+serverName+"__")
	return &meteredMCPTool{CoreTool: t, recorder: recorder, serverID: serverID, serverName: serverName, toolName: name, agentKey: agentKey, ri: ri}
}

func (m *meteredMCPTool) Call(ctx context.Context, input json.RawMessage, tc *actool.ToolContext) (actool.Result, error) {
	// Best effort: a ledger failure must never block the tool call itself.
	if err := m.recorder.InsertMCPUsage(&db.MCPUsage{
		ServerID:      m.serverID,
		ServerName:    m.serverName,
		ToolName:      m.toolName,
		AgentKey:      m.agentKey,
		TaskID:        m.ri.TaskID,
		ExplorationID: m.ri.ExplorationID,
		IntentID:      m.ri.IntentID,
		SessionID:     m.ri.SessionID,
	}); err != nil {
		log.Printf("[mcpusage] insert %s/%s: %v", m.serverName, m.toolName, err)
	}
	return m.CoreTool.Call(ctx, input, tc)
}
