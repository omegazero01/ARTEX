package server

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"github.com/Autumn-27/artex/db"
)

func TestBuildFindingsReportZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reports.zip")
	a := &db.DBFinding{ID: 1, Report: "**漏洞名称：** 注入A\n正文A"}
	b := &db.DBFinding{ID: 2, Report: "**漏洞名称：** 注入A\n正文B"}
	empty := &db.DBFinding{ID: 3}
	if err := buildFindingsReportZip(path, []*db.DBFinding{a, b, empty}); err != nil {
		t.Fatalf("zip: %v", err)
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer zr.Close()
	if len(zr.File) != 2 {
		t.Fatalf("entries = %d", len(zr.File))
	}
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["注入A.md"] || !names["注入A-2.md"] {
		t.Fatalf("names = %v", names)
	}
}
