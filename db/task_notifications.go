package db

import (
	"context"
	"encoding/json"
)

// Cursors are PostgreSQL visibility snapshots, not timestamps or sequence IDs:
// a transaction that commits after a page was read must still count as unread.
// ARTEX only tracks findings; SERIES also counts new assets and pending asset
// approvals, which have no host here (no approval pipeline in ARTEX).
type TaskNotificationQuery struct {
	TaskID   string `json:"task_id"`
	Findings string `json:"findings"`
}

type TaskNotificationCounts struct {
	TaskID   string `json:"task_id"`
	Findings int    `json:"findings"`
}

type TaskNotificationSummary struct {
	Snapshot   string                   `json:"snapshot"`
	ObservedAt int64                    `json:"observed_at"`
	Items      []TaskNotificationCounts `json:"items"`
}

// TaskNotifications returns, for each requested task, how many findings were
// created after the caller's snapshot. One statement, one visibility snapshot for
// the whole page: no payloads and no per-task round trips.
func (d *DB) TaskNotifications(ctx context.Context, queries []TaskNotificationQuery) (*TaskNotificationSummary, error) {
	input, err := json.Marshal(queries)
	if err != nil {
		return nil, err
	}
	var raw []byte
	if err := d.QueryRowContext(ctx, taskNotificationsSQL, string(input)).Scan(&raw); err != nil {
		return nil, err
	}
	var result TaskNotificationSummary
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Rows with notification_xid IS NULL are legacy rows written before the column
// existed: they are treated as already read so enabling the badge never floods a
// project with historical findings.
const taskNotificationsSQL = `WITH requested AS MATERIALIZED (
 SELECT q.* FROM jsonb_to_recordset($1::jsonb) AS q(task_id text,findings text)
 JOIN tasks t ON t.id=q.task_id::bigint WHERE t.deleted_at IS NULL
), finding_counts AS (
 SELECT q.task_id,count(*) n FROM requested q JOIN findings f ON f.task_id=q.task_id::bigint
 WHERE NULLIF(q.findings,'') IS NOT NULL AND f.notification_xid IS NOT NULL
 AND NOT pg_visible_in_snapshot(f.notification_xid,NULLIF(q.findings,'')::pg_snapshot) GROUP BY q.task_id
)
SELECT jsonb_build_object('snapshot',pg_current_snapshot()::text,
 'observed_at',floor(extract(epoch FROM statement_timestamp())*1000000)::bigint,
 'items',COALESCE((SELECT jsonb_agg(jsonb_build_object('task_id',q.task_id,
 'findings',COALESCE(f.n,0)))
 FROM requested q LEFT JOIN finding_counts f USING(task_id)),'[]'::jsonb))`
