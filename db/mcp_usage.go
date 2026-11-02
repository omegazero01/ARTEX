package db

import (
	"database/sql"
	"sort"
	"strings"
	"time"
)

// MCPUsage is one dynamic MCP tool invocation. Arguments and results are not
// retained; the ledger only carries attribution dimensions and a server-name
// snapshot so history remains readable after a server is renamed or deleted.
type MCPUsage struct {
	ServerID      int64     `json:"server_id"`
	ServerName    string    `json:"server_name"`
	ToolName      string    `json:"tool_name"`
	AgentKey      string    `json:"agent_key"`
	TaskID        int64     `json:"task_id"`
	ExplorationID int64     `json:"exploration_id"`
	IntentID      int64     `json:"intent_id"`
	SessionID     string    `json:"session_id"`
	TS            time.Time `json:"ts"`
}

func (d *DB) InsertMCPUsage(u *MCPUsage) error {
	var ts any
	if u != nil && !u.TS.IsZero() {
		ts = u.TS
	}
	_, err := d.Exec(`
	INSERT INTO mcp_usage(ts,server_id,server_name,tool_name,agent_key,task_id,exploration_id,intent_id,session_id)
	VALUES (COALESCE($1,now()),NULLIF($2,0),$3,$4,NULLIF($5,''),$6,$7,$8,NULLIF($9,''))`, ts,
		u.ServerID, u.ServerName, u.ToolName, u.AgentKey, nullIfZero(u.TaskID),
		nullIfZero(u.ExplorationID), nullIfZero(u.IntentID), u.SessionID)
	return err
}

// MCPUsageStat is one aggregate row, either per server or per tool.
type MCPUsageStat struct {
	ServerID   int64      `json:"server_id"`
	ServerName string     `json:"server_name"`
	ToolName   string     `json:"tool_name,omitempty"`
	Calls      int        `json:"calls"`
	Tasks      int        `json:"tasks"`
	Agents     []string   `json:"agents"`
	LastUsed   *time.Time `json:"last_used,omitempty"`
}

// MCPCall is one row of the recent-calls list; still no arguments or results.
type MCPCall struct {
	TS        time.Time `json:"ts"`
	ToolName  string    `json:"tool_name"`
	AgentKey  string    `json:"agent_key"`
	TaskID    int64     `json:"task_id"`
	SessionID string    `json:"session_id"`
}

// MCPServerUsageStats aggregates the whole ledger per server, busiest first.
func (d *DB) MCPServerUsageStats() (map[int64]MCPUsageStat, error) {
	rows, err := d.Query(`SELECT COALESCE(server_id,0),MAX(server_name),COUNT(*) calls,
COUNT(DISTINCT task_id) FILTER (WHERE task_id IS NOT NULL) tasks,
COALESCE(STRING_AGG(DISTINCT agent_key, ','),''),MAX(ts)
FROM mcp_usage GROUP BY server_id ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]MCPUsageStat{}
	for rows.Next() {
		var st MCPUsageStat
		var agents string
		var last sql.NullTime
		if err := rows.Scan(&st.ServerID, &st.ServerName, &st.Calls, &st.Tasks, &agents, &last); err != nil {
			return nil, err
		}
		st.Agents = splitAgents(agents)
		if last.Valid {
			t := last.Time
			st.LastUsed = &t
		}
		out[st.ServerID] = st
	}
	return out, rows.Err()
}

func splitAgents(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	agents := strings.Split(raw, ",")
	sort.Strings(agents)
	return agents
}

// MCPToolUsageStats aggregates one server's ledger per tool, busiest first.
func (d *DB) MCPToolUsageStats(serverID int64) ([]MCPUsageStat, error) {
	rows, err := d.Query(`SELECT server_id,MAX(server_name),tool_name,COUNT(*) calls,
COUNT(DISTINCT task_id) FILTER (WHERE task_id IS NOT NULL) tasks,
COALESCE(STRING_AGG(DISTINCT agent_key, ','),''),MAX(ts)
FROM mcp_usage WHERE server_id=$1 GROUP BY server_id,tool_name ORDER BY COUNT(*) DESC,tool_name`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MCPUsageStat{}
	for rows.Next() {
		var st MCPUsageStat
		var agents string
		var last sql.NullTime
		if err := rows.Scan(&st.ServerID, &st.ServerName, &st.ToolName, &st.Calls, &st.Tasks, &agents, &last); err != nil {
			return nil, err
		}
		st.Agents = splitAgents(agents)
		if last.Valid {
			t := last.Time
			st.LastUsed = &t
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// RecentMCPCalls returns the newest calls of one server, capped at 200 rows.
func (d *DB) RecentMCPCalls(serverID int64, limit int) ([]MCPCall, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := d.Query(`SELECT ts,tool_name,COALESCE(agent_key,''),COALESCE(task_id,0),COALESCE(session_id,'')
FROM mcp_usage WHERE server_id=$1 ORDER BY ts DESC,id DESC LIMIT $2`, serverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MCPCall{}
	for rows.Next() {
		var c MCPCall
		if err := rows.Scan(&c.TS, &c.ToolName, &c.AgentKey, &c.TaskID, &c.SessionID); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
