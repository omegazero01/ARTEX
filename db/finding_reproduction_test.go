package db

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestDecodeFindingReproductionRejectsModelBindingIDs(t *testing.T) {
	raw := json.RawMessage(`{"steps":[{"title":"登录","action":"POST /login","result":"200","binding_ids":["7"]}]}`)
	if _, err := DecodeFindingReproduction(raw); err == nil || !strings.Contains(err.Error(), "binding_ids") {
		t.Fatalf("模型传入 binding_ids 必须被拒绝，实际: %v", err)
	}
}

func TestDecodeFindingReproductionRequiresStepFields(t *testing.T) {
	raw := json.RawMessage(`{"steps":[{"title":"登录","action":"POST /login"}]}`)
	if _, err := DecodeFindingReproduction(raw); err == nil {
		t.Fatal("缺少 result 应报错")
	}
	if out, err := DecodeFindingReproduction(json.RawMessage(`null`)); err != nil || out != nil {
		t.Fatalf("null 应返回 nil,nil: %v %v", out, err)
	}
}

func TestFindingReproductionResolveUsesBindings(t *testing.T) {
	r := &FindingReproduction{Steps: []FindingReproductionStep{{
		Title: "登录", Action: "POST /login", Result: "200",
		TrafficRefs: []TrafficRef{{TrafficID: "t-1"}},
	}}}
	bindings := []FindingTrafficBinding{
		{ID: 42, Snapshot: TrafficEvidenceSnapshot{SourceTrafficID: "t-1"}},
	}
	raw, err := r.Resolve(bindings)
	if err != nil {
		t.Fatal(err)
	}
	var got FindingReproduction
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Steps[0].BindingIDs) != 1 || got.Steps[0].BindingIDs[0] != "42" {
		t.Fatalf("binding_ids=%v", got.Steps[0].BindingIDs)
	}
	if got.Steps[0].TrafficRefs != nil {
		t.Fatal("traffic_refs 必须在服务端解析后清空")
	}
	if len(r.Steps[0].BindingIDs) != 0 {
		t.Fatal("Resolve 不能修改调用方输入")
	}
}

func TestFindingReproductionResolveRejectsUnboundTraffic(t *testing.T) {
	r := &FindingReproduction{Steps: []FindingReproductionStep{{
		Title: "登录", Action: "POST /login", Result: "200",
		TrafficRefs: []TrafficRef{{TrafficID: "other-task"}},
	}}}
	if _, err := r.Resolve(nil); err == nil || !strings.Contains(err.Error(), "未绑定") {
		t.Fatalf("未绑定的流量必须报错，实际: %v", err)
	}
}

// 真实库验证：复现记录写进 findings.reproduction，并同步到探索节点 payload。
func TestRecordFindingPersistsReproduction(t *testing.T) {
	dsn := os.Getenv("ARTEX_PG_DSN")
	if dsn == "" {
		t.Skip("ARTEX_PG_DSN 未设置")
	}
	d, err := Open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// t.Cleanup 是后进先出：先注册关闭，后注册数据清理，清理才会真正执行。
	t.Cleanup(func() { d.Close() })
	expID, err := d.CreateExploration("reproduction-test", "verify structured reproduction")
	if err != nil {
		t.Fatalf("exploration: %v", err)
	}
	st := d.Exploration(expID)
	reproduction := &FindingReproduction{
		EntryPoint:    "https://example.test/login",
		Prerequisites: []string{"一个有效账号"},
		Steps: []FindingReproductionStep{{
			Title: "登录接口 SQL 注入", Action: "POST /login 提交 ' OR 1=1--", Result: "返回 200 并登录成功", Proves: "参数未过滤",
			Request: "POST /login HTTP/1.1\nHost: example.test", Response: "HTTP/1.1 200 OK",
		}},
	}
	out, err := st.RecordFinding(context.Background(), RecordFindingInput{
		Reproduction: reproduction, VulnClass: "sqli", Name: "登录注入", Severity: "high",
		Summary: "登录参数可注入", Evidence: "curl ...", Worker: "worker",
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	t.Cleanup(func() {
		_, _ = d.Exec(`DELETE FROM findings WHERE id=$1`, out.FindingID)
		_, _ = d.Exec(`DELETE FROM exploration_nodes WHERE id=$1`, out.NodeID)
		_, _ = d.Exec(`DELETE FROM explorations WHERE id=$1`, expID)
	})
	var raw string
	if err := d.QueryRow(`SELECT COALESCE(reproduction::text,'') FROM findings WHERE id=$1`, out.FindingID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var saved FindingReproduction
	if err := json.Unmarshal([]byte(raw), &saved); err != nil {
		t.Fatalf("saved json: %v (%s)", err, raw)
	}
	if saved.EntryPoint != reproduction.EntryPoint || len(saved.Steps) != 1 || saved.Steps[0].Title != "登录接口 SQL 注入" {
		t.Fatalf("saved=%+v", saved)
	}
	var payload string
	if err := d.QueryRow(`SELECT payload::text FROM exploration_nodes WHERE id=$1`, out.NodeID).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(payload, "登录接口 SQL 注入") {
		t.Fatalf("节点 payload 未同步复现记录: %s", payload)
	}

	// 引用未绑定的流量：必须整笔失败，且不落库。
	bad := *reproduction
	bad.Steps = []FindingReproductionStep{{Title: "t", Action: "a", Result: "r", TrafficRefs: []TrafficRef{{TrafficID: "not-bound"}}}}
	if _, err := st.RecordFinding(context.Background(), RecordFindingInput{
		Reproduction: &bad, VulnClass: "sqli", Name: "bad", Severity: "low", Summary: "s", Evidence: "e", Worker: "worker",
	}); err == nil || !strings.Contains(err.Error(), "未绑定") {
		t.Fatalf("未绑定流量必须失败，实际: %v", err)
	}
	var leaked int
	if err := d.QueryRow(`SELECT count(*) FROM findings WHERE name='bad'`).Scan(&leaked); err != nil {
		t.Fatal(err)
	}
	if leaked != 0 {
		t.Fatalf("失败的上报不应落库，实际 %d 行", leaked)
	}
}
