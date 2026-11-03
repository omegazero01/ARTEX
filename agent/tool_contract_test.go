package agent

import (
	"encoding/json"
	"reflect"
	"testing"

	actool "github.com/Autumn-27/norma/tool"
	"github.com/Autumn-27/artex/traffic"
)

// Ported from SERIES agent/tool_contract_test.go, adapted to tools ARTEX ships.

func trafficTool(t *testing.T, name string) actool.CoreTool {
	t.Helper()
	for _, tool := range traffic.SeedToolMetas() {
		if tool.Name() == name {
			return tool
		}
	}
	t.Fatalf("missing tool %s", name)
	return nil
}

func TestBuiltinSchemaCannotBeReplacedByStoredVersion(t *testing.T) {
	base := trafficTool(t, "traffic_search")
	stored := map[string]any{
		"description": "旧说明：用 page 翻页，只传 limit 即可",
		"type":        "array",
		"properties": map[string]any{
			"limit":   map[string]any{"type": "string", "default": "bad"},
			"page":    map[string]any{"type": "integer", "default": 0},
			"removed": map[string]any{"type": "string"},
		},
	}
	tool := ResolveBuiltinTool(base, stored["description"].(string), stored)
	if tool.Description() != base.Description() {
		t.Fatalf("old protocol replaced code contract: %q", tool.Description())
	}
	if tool.InputSchema()["type"] != "object" {
		t.Fatal("type overridden")
	}
	props := tool.InputSchema()["properties"].(map[string]any)
	if props["cursor"] == nil || props["method"] == nil || props["page"] != nil || props["removed"] != nil {
		t.Fatalf("schema drift: %v", keysOf(props))
	}
	if _, ok := props["limit"].(map[string]any)["default"].(string); ok {
		t.Fatal("invalid default retained")
	}
}

func TestStoredScalarDefaultRetainedAndBoundsEnforced(t *testing.T) {
	base := trafficTool(t, "traffic_search")
	valid := map[string]any{"properties": map[string]any{"limit": map[string]any{"default": 5}}}
	tool := ResolveBuiltinTool(base, "旧说明", valid)
	defs := scalarDefaults(tool.InputSchema())
	var got int
	if err := json.Unmarshal(defs["limit"], &got); err != nil {
		t.Fatal(err)
	}
	if got != 5 {
		t.Fatalf("valid default dropped: %d", got)
	}
	if tool.Description() == base.Description() {
		t.Fatal("missing customized-default note")
	}
	invalid := map[string]any{"properties": map[string]any{"limit": map[string]any{"default": 999}}}
	tool = ResolveBuiltinTool(base, "旧说明", invalid)
	if string(scalarDefaults(tool.InputSchema())["limit"]) == "999" {
		t.Fatal("out-of-range default retained")
	}
}

func TestNonContractToolKeepsUserDescription(t *testing.T) {
	base := actool.Build(actool.Spec{
		Name:        "list_tasks",
		Description: "code description",
		Schema:      map[string]any{"type": "object", "properties": map[string]any{"limit": map[string]any{"type": "integer"}}},
	})
	stored := map[string]any{"description": "用户改过的说明", "properties": map[string]any{"limit": map[string]any{"default": 7}}}
	tool := ResolveBuiltinTool(base, stored["description"].(string), stored)
	if tool.Description() != "用户改过的说明" {
		t.Fatalf("user description lost: %q", tool.Description())
	}
	if string(scalarDefaults(tool.InputSchema())["limit"]) != "7" {
		t.Fatal("user default lost")
	}
}

func TestAssetQueryContractSurvivesStoredMetadata(t *testing.T) {
	base := NewToolSet(nil, "worker").listAssets()
	stored := map[string]any{"description": "只传 limit 即可", "properties": map[string]any{
		"dsl":   map[string]any{"description": "可省略", "default": ""},
		"limit": map[string]any{"description": "可单独使用", "default": 50},
	}}
	resolved := ResolveBuiltinTool(base, "只传 limit 即可", stored)
	want, _ := json.Marshal(base.InputSchema())
	got, _ := json.Marshal(resolved.InputSchema())
	if !reflect.DeepEqual(want, got) || resolved.Description() != base.Description() {
		t.Fatal("stored metadata replaced asset query instructions")
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
