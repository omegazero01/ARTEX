package db

import (
	"fmt"
	"testing"
)

// TestTaskNotificationsFindingCursors checks the badge semantics against a real
// database: a finding created after the caller's snapshot counts as unread, a
// legacy row (notification_xid IS NULL) never does, and the returned snapshot
// advances so a second poll reports zero.
func TestTaskNotificationsFindingCursors(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer d.Close()

	task, err := d.CreateTask("zz-notify-task", "goal", nil, 0, 0)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	defer func() {
		_, _ = d.Exec(`DELETE FROM findings WHERE task_id=$1`, task.ID)
		_, _ = d.Exec(`DELETE FROM tasks WHERE id=$1`, task.ID)
	}()

	first, err := d.AddFinding(task.ID, 0, "sqli", "first", SeverityHigh, "s", "e", "worker", nil)
	if err != nil {
		t.Fatalf("add finding: %v", err)
	}
	_ = first

	// First poll has no cursor: nothing is unread, and it hands back the baseline.
	base, err := d.TaskNotifications(t.Context(), []TaskNotificationQuery{{TaskID: fmt.Sprint(task.ID)}})
	if err != nil {
		t.Fatalf("baseline poll: %v", err)
	}
	if len(base.Items) != 1 || base.Items[0].Findings != 0 {
		t.Fatalf("baseline should report 0 unread, got %#v", base.Items)
	}
	if base.Snapshot == "" {
		t.Fatal("baseline poll returned an empty snapshot")
	}

	// Legacy row: written before the column existed → must stay "read".
	if _, err := d.Exec(`INSERT INTO findings(task_id,vulnclass,name,severity,summary,evidence,worker,asset_ids)
		VALUES($1,'legacy','legacy','low','s','e','worker','[]')`, task.ID); err != nil {
		t.Fatalf("legacy insert: %v", err)
	}

	// New finding after the baseline → exactly one unread.
	if _, err := d.AddFinding(task.ID, 0, "rce", "second", SeverityCritical, "s", "e", "worker", nil); err != nil {
		t.Fatalf("add finding: %v", err)
	}
	next, err := d.TaskNotifications(t.Context(), []TaskNotificationQuery{{TaskID: fmt.Sprint(task.ID), Findings: base.Snapshot}})
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(next.Items) != 1 || next.Items[0].Findings != 1 {
		t.Fatalf("want 1 unread finding, got %#v", next.Items)
	}

	// Polling again with the fresh snapshot must report zero.
	after, err := d.TaskNotifications(t.Context(), []TaskNotificationQuery{{TaskID: fmt.Sprint(task.ID), Findings: next.Snapshot}})
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(after.Items) != 1 || after.Items[0].Findings != 0 {
		t.Fatalf("want 0 after advancing the cursor, got %#v", after.Items)
	}
}

// Deleted tasks must not leak counts into the page.
func TestTaskNotificationsSkipsDeletedTasks(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer d.Close()

	task, err := d.CreateTask("zz-notify-deleted", "goal", nil, 0, 0)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	defer func() {
		_, _ = d.Exec(`DELETE FROM findings WHERE task_id=$1`, task.ID)
		_, _ = d.Exec(`DELETE FROM tasks WHERE id=$1`, task.ID)
	}()

	if _, err := d.AddFinding(task.ID, 0, "sqli", "x", SeverityHigh, "s", "e", "worker", nil); err != nil {
		t.Fatalf("add finding: %v", err)
	}
	if _, err := d.Exec(`UPDATE tasks SET deleted_at=now() WHERE id=$1`, task.ID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	got, err := d.TaskNotifications(t.Context(), []TaskNotificationQuery{{TaskID: fmt.Sprint(task.ID)}})
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if len(got.Items) != 0 {
		t.Fatalf("deleted task leaked into counts: %#v", got.Items)
	}
}
