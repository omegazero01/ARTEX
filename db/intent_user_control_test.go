package db

import (
	"os"
	"testing"
)

// 需要真实 PostgreSQL：ARTEX_PG_DSN=postgres://user:pass@127.0.0.1:5432/artex?sslmode=disable
func TestStopIntentWithReasonLocksClaim(t *testing.T) {
	dsn := os.Getenv("ARTEX_PG_DSN")
	if dsn == "" {
		t.Skip("ARTEX_PG_DSN 未设置")
	}
	d, err := Open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	expID, err := d.CreateExploration("stop-intent-test", "verify user cancel lock")
	if err != nil {
		t.Fatalf("exploration: %v", err)
	}
	st := d.Exploration(expID)
	id, err := st.AddIntent(map[string]any{"summary": "待取消意图"}, 1, nil, "planner")
	if err != nil {
		t.Fatalf("intent: %v", err)
	}
	if err := st.StopIntentWithReason(id, "用户明确要求停止"); err != nil {
		t.Fatalf("stop: %v", err)
	}
	n, err := st.GetNode(id)
	if err != nil || n == nil {
		t.Fatalf("get: %v", err)
	}
	if n.State != "stopped" || !n.UserCancelled() {
		t.Fatalf("state=%s cancelled=%v", n.State, n.UserCancelled())
	}
	ok, err := st.ClaimIntent(id, "worker")
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if ok {
		t.Fatal("user-cancelled intent must not be claimable")
	}
	// 用户手动重开：状态回 open 且清除取消锁。
	changed, err := st.ReopenIntent(id)
	if err != nil || !changed {
		t.Fatalf("reopen: %v %v", changed, err)
	}
	n2, err := st.GetNode(id)
	if err != nil || n2 == nil {
		t.Fatalf("get2: %v", err)
	}
	if n2.State != "open" || n2.UserCancelled() {
		t.Fatalf("after reopen state=%s cancelled=%v", n2.State, n2.UserCancelled())
	}
}
