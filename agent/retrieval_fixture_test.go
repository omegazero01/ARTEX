package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Autumn-27/norma/llm"
)

func retrievalFixture(pages int) llm.CompletionRequest {
	req := llm.CompletionRequest{System: []string{"system"}}
	for i := 0; i < pages; i++ {
		id := fmt.Sprintf("read-%d", i)
		in := fmt.Sprintf(`{"items":[{"ref":"node:1","field":"/detail","offset":%d}]}`, i*4000)
		body, _ := json.Marshal(map[string]any{"items": []any{map[string]any{"ref": "node:1", "node_id": 1, "payload": map[string]any{"value": strings.Repeat("中证据", 1333), "offset": i * 4000, "next_offset": (i + 1) * 4000}, "next": map[string]any{"items": []any{map[string]any{"ref": "node:1", "field": "/detail", "offset": (i + 1) * 4000}}}}}})
		req.Messages = append(req.Messages, llm.Message{Role: llm.RoleAssistant, Content: []llm.ContentBlock{{Type: llm.BlockToolUse, ID: id, Name: "context_get", Input: json.RawMessage(in)}}}, llm.Message{Role: llm.RoleUser, Content: []llm.ContentBlock{llm.ToolResultText(id, string(body), false)}})
	}
	return req
}
