package report

import (
	"strings"
	"testing"

	"github.com/neouks/series/db"
)

func TestFindingReportTitle(t *testing.T) {
	f := &db.DBFinding{ID: 7, Name: "长标题", Report: "**漏洞名称：** 某系统 SQL 注入\n\n## 影响\n..."}
	if got := FindingReportTitle(f); got != "某系统 SQL 注入" {
		t.Fatalf("title = %q", got)
	}
	// 报告里没有“漏洞名称”时回退到 Name。
	g := &db.DBFinding{ID: 9, Name: "回退标题", Report: "## 影响\n无"}
	if got := FindingReportTitle(g); got != "回退标题" {
		t.Fatalf("fallback = %q", got)
	}
	if got := FindingReportTitle(nil); got != "" {
		t.Fatalf("nil = %q", got)
	}
}

func TestFindingReportFilename(t *testing.T) {
	f := &db.DBFinding{ID: 7, Report: "**漏洞名称：** 某系统 SQL 注入\n正文"}
	if got := FindingReportFilename(f); got != "某系统_SQL_注入.md" {
		t.Fatalf("filename = %q", got)
	}
	// 无标题时用 ID 兜底。
	g := &db.DBFinding{ID: 12}
	if got := FindingReportFilename(g); got != "长12" && !strings.HasSuffix(got, ".md") {
		t.Fatalf("fallback filename = %q", got)
	}
	if got := FindingReportFilename(nil); got != "finding.md" {
		t.Fatalf("nil filename = %q", got)
	}
}

func TestFindingsReportMarkdownSkipsEmptyReport(t *testing.T) {
	a := &db.DBFinding{ID: 1, Report: "# A\n内容A"}
	b := &db.DBFinding{ID: 2, Report: "   "}
	c := &db.DBFinding{ID: 3, Report: "# C\n内容C"}
	got := FindingsReportMarkdown([]*db.DBFinding{a, b, c})
	if strings.Contains(got, "内容A") == false || strings.Contains(got, "内容C") == false {
		t.Fatalf("missing content: %q", got)
	}
	if strings.Count(got, "---") != 1 {
		t.Fatalf("separator count wrong: %q", got)
	}
}
