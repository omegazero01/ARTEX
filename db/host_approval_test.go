package db

import (
	"testing"
)

func TestTaskGrantDoesNotConsumeTaskSequence(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	task, err := d.CreateTaskWithOptions("sequence ownership", "goal", TaskCreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteTask(task.ID)
	var before, after int64
	if err := d.QueryRow(`SELECT last_value FROM tasks_id_seq`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO task_asset_grants(task_id,kind,value,source) SELECT $1,'host','grant-'||n||'.test','manual' FROM generate_series(1,100) n`, task.ID); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT last_value FROM tasks_id_seq`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("grants consumed task sequence: %d -> %d", before, after)
	}
}
