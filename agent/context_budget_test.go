package agent

import (
	"context"
	"strings"
	"testing"
)

// 减代码验收：工具输出不再被 30000 字符上限截断/报错。
func TestToolOutputHasNoBudgetCap(t *testing.T) {
	if got := readOutputBudget(context.Background()); got < 1<<20 {
		t.Fatalf("readOutputBudget = %d, want effectively unlimited", got)
	}
	big := strings.Repeat("x", 200_000)
	res, err := contextJSON(context.Background(), map[string]any{"body": big})
	if err != nil {
		t.Fatalf("contextJSON returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("contextJSON reported tool error: %s", res.Flatten())
	}
	if len(res.Flatten()) < 200_000 {
		t.Fatalf("output truncated: got %d chars", len(res.Flatten()))
	}
}

// 减代码验收：每个来源的内容预算不再被 (总预算-1024*N)/6 压缩。
func TestReadContentBudgetKeepsRequested(t *testing.T) {
	if got := readContentBudget(context.Background(), 12000, 5); got != 12000 {
		t.Fatalf("readContentBudget = %d, want 12000", got)
	}
}
