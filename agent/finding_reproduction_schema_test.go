package agent

import (
	"testing"
)

// report_finding 必须向模型暴露 reproduction 结构，且 steps 是必填数组。
func TestReportFindingSchemaExposesReproduction(t *testing.T) {
	tool := NewToolSet(nil, "worker").addFinding()
	props, _ := tool.InputSchema()["properties"].(map[string]any)
	reproduction, ok := props["reproduction"].(map[string]any)
	if !ok {
		t.Fatal("report_finding 缺少 reproduction 参数")
	}
	fields, _ := reproduction["properties"].(map[string]any)
	steps, _ := fields["steps"].(map[string]any)
	if steps == nil || steps["type"] != "array" {
		t.Fatalf("steps 应为数组: %v", steps)
	}
	item, _ := steps["items"].(map[string]any)
	required, _ := item["required"].([]string)
	if len(required) != 3 {
		t.Fatalf("steps 必填字段应为 title/action/result，实际 %v", required)
	}
}
