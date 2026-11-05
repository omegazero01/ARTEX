package db

import (
	"encoding/json"
	"strings"
	"testing"
)

const reproductionFixture = `{"entry_point":"GET https://example.test/orders","prerequisites":["测试账号A"],"steps":[{"title":"正常对照","action":"读取自己的订单","result":"返回200及账号A订单","request":"GET /orders?id=1 HTTP/1.1\r\nHost: example.test\r\n\r\n","response":"HTTP/1.1 200 OK\r\n\r\n中文🙂"},{"title":"命令验证","action":"查看已有输出","result":"已保存输出","command":"printf test","output":"test\n\u0060\u0060\u0060\n"}]}`

func TestReproductionInputValidation(t *testing.T) {
	for _, input := range []string{`{}`, `[]`, `{"steps":"wrong"}`, `{"steps":[{"title":"x","action":"y"}]}`, strings.Replace(reproductionFixture, `"result":"已保存输出"`, `"result":"已保存输出","binding_ids":["123"]`, 1), strings.Replace(reproductionFixture, `"result":"已保存输出"`, `"result":"已保存输出","unknown":true`, 1), reproductionFixture + `{}`} {
		if _, err := DecodeFindingReproduction(json.RawMessage(input)); err == nil {
			t.Fatalf("accepted malformed reproduction: %s", input)
		}
	}
	r, err := DecodeFindingReproduction(json.RawMessage(reproductionFixture))
	if err != nil || r.Steps[0].Response != "HTTP/1.1 200 OK\r\n\r\n中文🙂" {
		t.Fatal(r, err)
	}
	for _, raw := range []json.RawMessage{nil, json.RawMessage(`null`)} {
		if r, err := DecodeFindingReproduction(raw); err != nil || r != nil {
			t.Fatal(r, err)
		}
	}
	r.Steps = make([]FindingReproductionStep, 51)
	if r.ValidateInput() == nil {
		t.Fatal("unbounded steps accepted")
	}
	large := json.RawMessage(`{"steps":[{"title":"large","action":"read","result":"ok","output":"` + strings.Repeat("x", 1024*1024) + `"}]}`)
	if _, err := DecodeFindingReproduction(large); err == nil {
		t.Fatal("unbounded evidence accepted")
	}
}

func TestReproductionAtomicRoundTripAndDetailProjection(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	task, err := d.CreateTask("reproduction", "fixture", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteTask(task.ID)
	s := d.Exploration(task.ExplorationID)
	r, _ := DecodeFindingReproduction(json.RawMessage(reproductionFixture))
	in := RecordFindingInput{TaskID: task.ID, Reproduction: r, VulnClass: "TEST", Severity: "low", Summary: "immutable reproduction", Evidence: "legacy evidence"}
	recorded, err := s.RecordFinding(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}
	defer d.DeleteFinding(recorded.FindingID)
	f, err := d.GetFinding(recorded.FindingID)
	if err != nil {
		t.Fatal(err)
	}
	n, err := s.GetNode(recorded.NodeID)
	if err != nil {
		t.Fatal(err)
	}
	var p map[string]json.RawMessage
	json.Unmarshal(n.Payload, &p)
	if string(p["reproduction"]) != string(f.Reproduction) || !strings.Contains(string(f.Reproduction), "中文") {
		t.Fatal("dual write drift", string(n.Payload), string(f.Reproduction))
	}
	listed, err := d.ListFindings(10)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range listed {
		if len(row.Reproduction) > 0 {
			t.Fatal("list loaded reproduction")
		}
	}
	exported, err := d.ListFindingsForExport(FindingFilter{}, []int64{f.ID})
	if err != nil || len(exported) != 1 || string(exported[0].Reproduction) != string(f.Reproduction) {
		t.Fatal(exported, err)
	}
	// Resolution fails after provisional node/finding inserts; transaction must
	// roll them all back, and must not mutate the caller's original steps.
	in.Reproduction.Steps[0].TrafficRefs = []TrafficRef{{TrafficID: "unbound"}}
	var before int
	d.QueryRow(`SELECT count(*) FROM exploration_nodes WHERE exploration_id=$1`, task.ExplorationID).Scan(&before)
	if _, err = s.RecordFinding(t.Context(), in); err == nil {
		t.Fatal("unbound step accepted")
	}
	var after int
	d.QueryRow(`SELECT count(*) FROM exploration_nodes WHERE exploration_id=$1`, task.ExplorationID).Scan(&after)
	if before != after || len(in.Reproduction.Steps[0].BindingIDs) != 0 {
		t.Fatal("rollback leaked state")
	}
	// Standalone evidence survives removal of the originating task and graph.
	if err = d.DeleteTask(task.ID); err != nil {
		t.Fatal(err)
	}
	afterDelete, err := d.GetFinding(f.ID)
	if err != nil || afterDelete == nil || string(afterDelete.Reproduction) != string(f.Reproduction) {
		t.Fatal(afterDelete, err)
	}
}
