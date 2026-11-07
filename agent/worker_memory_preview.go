package agent

import (
	"fmt"
	"unicode/utf8"

	"github.com/Autumn-27/norma/memory"
)

func workerMemoryPreview(store *memory.Store, query string) string {
	if store == nil {
		return ""
	}
	items := store.Relevant(query, 3)
	if len(items) == 0 {
		return ""
	}
	budget := 1600
	var text string
	for _, m := range items {
		if budget <= 0 {
			break
		}
		preview := firstLine(m.Content, min(500, budget))
		line := fmt.Sprintf("\n记忆 %s（非实时证据，完整内容用RecallMemory）：%s", firstLine(m.Name, 80), preview)
		if utf8.RuneCountInString(line) > budget {
			continue
		}
		text += line
		budget -= utf8.RuneCountInString(line)
	}
	return "\n【相关经验预览：不代表当前目标事实】" + text
}
