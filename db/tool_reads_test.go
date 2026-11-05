package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestToolReadProjectionSizes(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	for _, size := range []int{10, 100, 1000} {
		exp, err := d.CreateExploration("projection size", "goal")
		if err != nil {
			t.Fatal(err)
		}
		defer d.Exec(`DELETE FROM explorations WHERE id=$1`, exp)
		if _, err := d.Exec(`INSERT INTO exploration_nodes(exploration_id,kind,state,payload)
  SELECT $1,'fact','confirmed',jsonb_build_object('summary','fixture '||n,'detail',repeat('大段证据',2048)) FROM generate_series(1,$2::int)n`, exp, size); err != nil {
			t.Fatal(err)
		}
		es := d.Exploration(exp)
		started := time.Now()
		raw, _, _, err := es.ListByKindPageWithSources(KindFact, 0, 20, "fixture")
		if err != nil {
			t.Fatal(err)
		}
		fullTime := time.Since(started)
		started = time.Now()
		page, err := es.ToolNodePage(context.Background(), KindFact, "fixture", "", 0, 0, 20)
		if err != nil {
			t.Fatal(err)
		}
		projectTime := time.Since(started)
		oldBytes, _ := json.Marshal(raw)
		newBytes, _ := json.Marshal(page)
		if len(newBytes) >= len(oldBytes) {
			t.Fatal("projection did not shrink evidence-heavy page")
		}
		t.Logf("rows=%d full_page=%dB/%s projected_page=%dB/%s (database projection, not exact model Tokens)", size, len(oldBytes), fullTime, len(newBytes), projectTime)
	}
}

func TestLatestWorkerOutputBeyondThousandActivities(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	exp, err := d.CreateExploration("output", "goal")
	if err != nil {
		t.Fatal(err)
	}
	defer d.Exec(`DELETE FROM explorations WHERE id=$1`, exp)
	es := d.Exploration(exp)
	id, err := es.AddIntent(map[string]any{"summary": "run"}, 1, nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	if out, err := es.LatestWorkerOutput(context.Background(), id); err != nil || out != nil {
		t.Fatalf("empty: %+v %v", out, err)
	}
	if _, err := d.Exec(`INSERT INTO activity(exploration_id,node_id,kind,detail) SELECT $1,$2,'text','phase-'||n FROM generate_series(1,1100)n`, exp, id); err != nil {
		t.Fatal(err)
	}
	out, err := es.LatestWorkerOutput(context.Background(), id)
	if err != nil || out == nil || out.Detail != "phase-1100" {
		t.Fatalf("latest text: %+v %v", out, err)
	}
	if _, err := d.Exec(`INSERT INTO activity(exploration_id,node_id,kind,detail) VALUES($1,$2,'result','最终结论'),($1,$2,'text','later text')`, exp, id); err != nil {
		t.Fatal(err)
	}
	out, err = es.LatestWorkerOutput(context.Background(), id)
	if err != nil || out == nil || out.Detail != "最终结论" {
		t.Fatalf("latest result: %+v %v", out, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := es.LatestWorkerOutput(ctx, id); err == nil {
		t.Fatal("query failure disguised as no output")
	}
}

