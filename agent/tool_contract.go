package agent

import (
	"encoding/json"
	"fmt"
	"reflect"

	actool "github.com/Autumn-27/norma/tool"
)

// ResolveBuiltinTool keeps executable structure code-owned. Persisted metadata
// is not rewritten here; valid scalar defaults and non-protocol prose remain editable.
// Ported from SERIES agent/tool_contract.go.
func ResolveBuiltinTool(t actool.CoreTool, description string, saved map[string]any) actool.CoreTool {
	raw, _ := json.Marshal(t.InputSchema())
	var schema map[string]any
	_ = json.Unmarshal(raw, &schema)
	// Asset selector and pagination instructions are executable contract metadata.
	// Persisted descriptions/defaults may advertise obsolete limit-only calls.
	contractOwned := builtinContractOwned(t.Name())
	if t.Name() != "list_assets" && t.Name() != "check_target_access" {
		mergeToolMetadataMode(schema, saved, !contractOwned)
	}
	if t.Name() == "asset_search" {
		// The untested mode requires an explicit selector and a separate grant.
		// An old saved default would otherwise be injected again by this inner
		// decorator after the outer role/coverage schema removed the parameter.
		if prop, ok := nestedMap(schema, "properties", "untested"); ok {
			delete(prop, "default")
		}
	}
	// These descriptions contain versioned pagination/write protocols. Old saved
	// text must not replace the current contract (e.g. "all findings").
	if contractOwned {
		description = t.Description()
		if !reflect.DeepEqual(scalarDefaults(schema), scalarDefaults(t.InputSchema())) {
			description += "\n已配置自定义缺省值，实际缺省值以当前参数 Schema 的 default 为准。"
		}
	}
	return DecorateTool(t, description, schema)
}

// builtinContractOwned lists tools whose description/schema is a protocol
// contract. The SERIES list is intersected with the tools ARTEX actually ships;
// names ARTEX does not have are dropped, ARTEX-only tools stay editable.
func builtinContractOwned(name string) bool {
	switch name {
	case "cancel_intent", "kill_work", "steer_work":
		return true
	case "graph_overview", "traffic_get", "traffic_search", "update_finding_report", "get_finding_traffic", "bind_finding_traffic", "add_intent", "get_finding_retest_context", "insert_assets", "create_custom_tool", "update_custom_tool", "list_facts", "list_findings", "list_task_findings", "list_assets", "node_detail", "get_task_node_detail", "get_worker_output", "get_worker_trace", "search_all_worker_traces", "get_task_worker_trace", "list_task_worker_traces", "search_task_worker_traces", "get_task_graph", "add_task_hint", "expand_digest", "record_fact", "report_finding":
		return true
	}
	return false
}

// Protocol descriptions are code-owned at every depth. Retain valid configurable
// defaults, but never restore stale field instructions from a previously seeded schema.
func mergeToolMetadataMode(code, saved map[string]any, descriptions bool) {
	mergeToolMetadataAt(code, saved, descriptions, 0)
}

func mergeToolMetadataAt(code, saved map[string]any, descriptions bool, depth int) {
	if code == nil || saved == nil {
		return
	}
	if desc, ok := saved["description"].(string); ok && descriptions {
		code["description"] = desc
	}
	// injectDefaults only applies top-level scalar fields. Advertising saved
	// nested/array defaults would promise behavior the execution layer lacks.
	typ, _ := code["type"].(string)
	scalar := typ == "string" || typ == "integer" || typ == "number" || typ == "boolean"
	if value, ok := saved["default"]; ok && depth == 1 && scalar {
		raw, err := json.Marshal(value)
		if err == nil && validateBusinessInput(code, raw) == nil {
			code["default"] = value
		}
	}
	props, _ := code["properties"].(map[string]any)
	oldProps, _ := saved["properties"].(map[string]any)
	for key, value := range props {
		child, _ := value.(map[string]any)
		old, _ := oldProps[key].(map[string]any)
		mergeToolMetadataAt(child, old, descriptions, depth+1)
	}
	items, _ := code["items"].(map[string]any)
	oldItems, _ := saved["items"].(map[string]any)
	mergeToolMetadataAt(items, oldItems, descriptions, depth+1)
}

func nestedMap(m map[string]any, keys ...string) (map[string]any, bool) {
	cur := m
	for _, k := range keys {
		next, ok := cur[k].(map[string]any)
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}

// Business contracts live at the tool boundary. In particular the SDK's basic
// type checker does not enforce our ID, batch and response-budget limits.
// JSON normalization also makes code schemas ([]string/int) and stored schemas
// ([]any/float64) behave identically.
func normalizeToolSchema(schema map[string]any) map[string]any {
	b, _ := json.Marshal(schema)
	var normalized map[string]any
	_ = json.Unmarshal(b, &normalized)
	return normalized
}

func validateBusinessInput(schema map[string]any, input json.RawMessage) error {
	if err := actool.ValidateInput(schema, input); err != nil {
		return err
	}
	var value any
	if err := json.Unmarshal(input, &value); err != nil {
		return err
	}
	return validateBusinessBounds(schema, value, "input")
}

func validateBusinessBounds(schema map[string]any, value any, path string) error {
	if values, ok := schema["enum"].([]any); ok {
		found := false
		for _, allowed := range values {
			found = found || reflect.DeepEqual(allowed, value)
		}
		if !found {
			// Report the schema's choices, not the rejected input (which may
			// contain secrets). Role-filtered enums need actionable diagnostics.
			choices, err := json.Marshal(values)
			if err == nil && len(choices) <= 512 {
				return fmt.Errorf("%s: unsupported value; allowed values: %s", path, choices)
			}
			return fmt.Errorf("%s: unsupported value; see the parameter's enum for allowed values", path)
		}
	}
	check := func(n float64, low, high string) error {
		if min, ok := schema[low].(float64); ok && n < min {
			return fmt.Errorf("%s: %s must be >= %g", path, low, min)
		}
		if max, ok := schema[high].(float64); ok && n > max {
			return fmt.Errorf("%s: %s must be <= %g", path, high, max)
		}
		return nil
	}
	switch v := value.(type) {
	case float64:
		return check(v, "minimum", "maximum")
	case []any:
		if err := check(float64(len(v)), "minItems", "maxItems"); err != nil {
			return err
		}
		item, _ := schema["items"].(map[string]any)
		for i, child := range v {
			if err := validateBusinessBounds(item, child, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case map[string]any:
		props, _ := schema["properties"].(map[string]any)
		for key, child := range v {
			property, _ := props[key].(map[string]any)
			if err := validateBusinessBounds(property, child, path+"."+key); err != nil {
				return err
			}
		}
	}
	return nil
}
