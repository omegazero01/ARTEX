package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// UserCancelled reports whether this intent carries the user-cancellation lock.
// It is stored in the payload so the node keeps its outputs and lineage without a
// schema change, and so every claim path can filter it with one predicate.
func (n *Node) UserCancelled() bool {
	if n == nil {
		return false
	}
	var p struct {
		Cancelled bool `json:"cancelled_by_user"`
	}
	if json.Unmarshal(n.Payload, &p) != nil {
		return false
	}
	return p.Cancelled
}

// StopIntentWithReason marks an open or paused intent as stopped on behalf of the
// user: the intent, its session and all outputs are kept, cancelled_by_user is
// written into the payload, and every claim query skips it from then on.
func (s *ExplorationStore) StopIntentWithReason(id int64, reason string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var state string
	var raw []byte
	if err := tx.QueryRow(`SELECT state, payload FROM exploration_nodes
		WHERE id=$1 AND exploration_id=$2 AND kind='intent' FOR UPDATE`, id, s.expID).Scan(&state, &raw); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("intent not found")
		}
		return err
	}
	payload := map[string]any{}
	_ = json.Unmarshal(raw, &payload)
	if state == "stopped" && payload["cancelled_by_user"] == true {
		return tx.Commit() // already user-cancelled: no duplicate write
	}
	if state != "open" && state != "paused" {
		return fmt.Errorf("%w: intent state %s cannot be cancelled", ErrIntentStateConflict, state)
	}
	if strings.TrimSpace(reason) == "" {
		reason = "用户取消等待运行"
	}
	payload["cancelled_by_user"] = true
	payload["cancel_reason"] = reason
	payload["cancelled_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	merged, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE exploration_nodes
		SET state='stopped', payload=$3, blocked_reason=NULL, completed_at=now(),
		    content_version=content_version+1
		WHERE id=$1 AND exploration_id=$2`, id, s.expID, string(merged)); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM side_question_sessions WHERE intent_id=$1`, id); err != nil {
		return err
	}
	return tx.Commit()
}
