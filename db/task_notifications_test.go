package db

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func TestTaskNotificationsBatchIsolation(t *testing.T) {
	d, e := Open(testDSN(t))
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	queries := make([]TaskNotificationQuery, 0, 20)
	for range 20 {
		task, e := d.CreateTaskWithOptions("notification batch", "", TaskCreateOptions{})
		if e != nil {
			t.Fatal(e)
		}
		defer d.DeleteTask(task.ID)
		queries = append(queries, TaskNotificationQuery{TaskID: strconv.FormatInt(task.ID, 10)})
	}
	base, e := d.TaskNotifications(t.Context(), queries, true)
	if e != nil {
		t.Fatal(e)
	}
	for i := range queries {
		queries[i].Findings = base.Snapshot
	}
	if _, e = d.Exec(`INSERT INTO findings(task_id) SELECT $1::bigint FROM generate_series(1,120)`, queries[0].TaskID); e != nil {
		t.Fatal(e)
	}
	started := time.Now()
	r, e := d.TaskNotifications(t.Context(), queries, true)
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(r)
	t.Logf("20-task batch, 120 new findings: one SQL statement, %d response bytes, %s", len(raw), time.Since(started))
	for _, row := range r.Items {
		want := 0
		if row.TaskID == queries[0].TaskID {
			want = 120
		}
		if row.Findings != want {
			t.Fatalf("isolation %+v", row)
		}
	}
}

