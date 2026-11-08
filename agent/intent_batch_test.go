package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

// add_intent 原版没有批量条数上限；防止 maxItems 再次被加回来。
func TestAddIntentSchemaHasNoBatchCap(t *testing.T) {
	raw, err := json.Marshal(intentInputSchema())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"maxItems"`) {
		t.Fatalf("add_intent schema 不应再有 maxItems 限制: %s", raw)
	}
}
