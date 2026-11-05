package db

import (
	"fmt"
	"strings"
	"testing"
)

func TestPlannerBootstrapSharedRecordLimit(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	exp, err := d.CreateExploration("records", "goal")
	if err != nil {
		t.Fatal(err)
	}
	defer d.Exec(`DELETE FROM explorations WHERE id=$1`, exp)
	s := d.Exploration(exp)
	ids := []int64{}
	for i := 0; i < 20; i++ {
		id, e := s.AddNode(KindFact, map[string]any{"summary": strings.Repeat("中", 10000)}, 0, "confirmed", "test", nil)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, id)
	}
	got, err := s.PlannerBootstrap(t.Context(), []string{"fact"}, ids, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 8 || !got.HasMore {
		t.Fatal(got)
	}
	for i, r := range got.Records {
		if r.Ref != fmt.Sprintf("node:%d", ids[len(ids)-1-i]) || len([]rune(r.Summary)) > 240 {
			t.Fatal(r)
		}
	}
}

func TestPlannerBootstrapHeartbeatTimeAndSharedAsset(t *testing.T) {
	d, e := Open(testDSN(t))
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	d.SetMaxOpenConns(1)
	task, e := d.CreateTask("heartbeat", "goal", nil, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	defer d.DeleteTask(task.ID)
	es := d.Exploration(task.ExplorationID)
	asset, e := d.Assets().UpsertRootDomain(UpsertRootDomainReq{Domain: fmt.Sprintf("idle-%d.test", task.ID), TaskID: task.ID})
	if e != nil {
		t.Fatal(e)
	}
	ids := []int64{}
	for i := 0; i < 3; i++ {
		id, err := es.AddIntent(map[string]any{"summary": fmt.Sprint(i)}, 1, []int64{asset}, "test")
		if err != nil {
			t.Fatal(err)
		}
		if e = es.SetIntentState(id, "running"); e != nil {
			t.Fatal(e)
		}
		ids = append(ids, id)
	}
	// Timestamp order intentionally differs from insertion/ID order.
	for i, days := range []int{1, 30, 10} {
		if _, e = d.Exec(`INSERT INTO activity(exploration_id,node_id,worker,kind,summary,created_at) VALUES($1,$2,'worker','text','activity',now()-($3*interval '1 day'))`, es.ID(), ids[i], days); e != nil {
			t.Fatal(e)
		}
	}
	var before, after string
	if e = d.QueryRow(`SHOW jit`).Scan(&before); e != nil {
		t.Fatal(e)
	}
	out, e := es.PlannerBootstrap(t.Context(), []string{"intent", "fact"}, nil, true)
	if e != nil {
		t.Fatal(e)
	}
	if e = d.QueryRow(`SHOW jit`).Scan(&after); e != nil || after != before {
		t.Fatal("connection setting leaked", before, after, e)
	}
	if len(out.Records) != 2 {
		t.Fatal(out)
	}
	for _, r := range out.Records {
		if r.Ref == fmt.Sprintf("node:%d", ids[0]) || r.Reason != "idle_running" {
			t.Fatal("wrong idle order", out)
		}
	}
	fact, e := es.AddNode(KindFact, map[string]any{"summary": "asset changed"}, 0, "confirmed", "test", []int64{asset})
	if e != nil {
		t.Fatal(e)
	}
	out, e = es.PlannerBootstrap(t.Context(), []string{"intent", "fact"}, []int64{fact}, false)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Records) != 4 || out.Records[0].Reason != "event" {
		t.Fatal(out)
	}
	for _, r := range out.Records[1:] {
		if r.Reason != "same_asset_active" {
			t.Fatal(out)
		}
	}
}
