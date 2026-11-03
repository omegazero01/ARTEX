package report

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Autumn-27/artex/db"
)

func TestReproductionMarkdownRendersSavedEvidence(t *testing.T) {
	raw, _ := json.Marshal(db.FindingReproduction{
		EntryPoint:    "https://example.test/login",
		Prerequisites: []string{"有效账号"},
		Steps: []db.FindingReproductionStep{{
			Title: "登录注入", Action: "POST /login 提交 ' OR 1=1--", Result: "登录成功", Proves: "参数未过滤",
			Request: "POST /login HTTP/1.1\nHost: example.test", Response: "HTTP/1.1 200 OK", BindingIDs: []string{"7"},
		}},
	})
	out := ReproductionMarkdown(raw)
	for _, want := range []string{"## 分步复现", "**入口**：https://example.test/login", "- 有效账号", "### 步骤 1：登录注入", "**请求**", "**响应**", "**实际结果**：登录成功", "**证明范围**：参数未过滤", "**关联流量证据编号**：7"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestReproductionMarkdownMarksMissingResponse(t *testing.T) {
	raw, _ := json.Marshal(db.FindingReproduction{Steps: []db.FindingReproductionStep{{
		Title: "请求", Action: "GET /x", Result: "未记录", Request: "GET /x HTTP/1.1",
	}}})
	out := ReproductionMarkdown(raw)
	if !strings.Contains(out, "响应原文未记录") || !strings.Contains(out, "未关联流量证据") {
		t.Fatalf("raw:\n%s", out)
	}
}

func TestReproductionMarkdownHandlesFencesInsideEvidence(t *testing.T) {
	raw, _ := json.Marshal(db.FindingReproduction{Steps: []db.FindingReproductionStep{{
		Title: "回显", Action: "cat", Result: "看到内容", Command: "cat out.md", Output: "```\nfenced\n```",
	}}})
	out := ReproductionMarkdown(raw)
	if !strings.Contains(out, "````") {
		t.Fatalf("围栏应加长以保住原文:\n%s", out)
	}
}

func TestReproductionMarkdownInvalidJSON(t *testing.T) {
	if out := ReproductionMarkdown(json.RawMessage(`{"steps":`)); !strings.Contains(out, "格式无效") {
		t.Fatalf("out=%q", out)
	}
}

// 没有人类提交报告时，技术导出的“详细报告”回退到结构化复现记录。
func TestSingleFindingMarkdownFallsBackToReproduction(t *testing.T) {
	raw, _ := json.Marshal(db.FindingReproduction{Steps: []db.FindingReproductionStep{{
		Title: "步骤一", Action: "POST /a", Result: "200",
	}}})
	f := &db.DBFinding{ID: 1, Severity: "high", VulnClass: "sqli", Status: "pending", Reproduction: raw, CreatedAt: time.Now()}
	out := SingleFindingMarkdown(f, time.Now())
	if !strings.Contains(out, "## 分步复现") || !strings.Contains(out, "### 步骤 1：步骤一") {
		t.Fatalf("out:\n%s", out)
	}
}
