package db

import (
	"fmt"
	"testing"
)

func TestMixedContextEqualTimeAndIDPagination(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	task, err := d.CreateTask("mixed ties", "goal", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteTask(task.ID)
	es := d.Exploration(task.ExplorationID)
	intent, err := es.AddIntent(map[string]any{"summary": "work"}, 1, nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := d.QueryRow(`SELECT greatest((SELECT coalesce(max(id),0) FROM exploration_nodes),(SELECT coalesce(max(id),0) FROM activity))+1000000`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO exploration_nodes(id,exploration_id,kind,state,payload,created_at) VALUES($1,$2,'fact','confirmed','{"summary":"UNIQUE_TIE_MARKER"}','2026-01-01T00:00:00Z')`, id, task.ExplorationID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO activity(id,exploration_id,node_id,kind,detail,created_at) VALUES($1,$2,$3,'text','UNIQUE_TIE_MARKER','2026-01-01T00:00:00Z')`, id, task.ExplorationID, intent); err != nil {
		t.Fatal(err)
	}
	q := ContextSearch{Types: []string{"fact", "activity"}, Query: "UNIQUE_TIE_MARKER", Limit: 1}
	hits, next, err := es.SearchMixedContext(t.Context(), q, nil)
	if err != nil || len(hits) != 2 || next == nil || hits[0].Ref != fmt.Sprintf("node:%d", id) {
		t.Fatal(hits, next, err)
	}
	hits, next, err = es.SearchMixedContext(t.Context(), q, next)
	if err != nil || len(hits) != 1 || next != nil || hits[0].Ref != fmt.Sprintf("activity:%d", id) {
		t.Fatal(hits, next, err)
	}
}
