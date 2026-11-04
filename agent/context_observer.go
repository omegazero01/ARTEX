package agent

import (
	"context"
	"log"
	"time"

	"github.com/Autumn-27/norma/llm"
	"github.com/neouks/series/db"
)

// contextObserverKey 挂在 ctx 上，用来把 agent 内部的活动事件回传给上层记录。
// 这个能力与请求预算无关，因此预算代码删除后单独保留在本文件。

type contextObserverKey struct{}

func withContextObserver(ctx context.Context, emit func(db.Activity)) context.Context {
	return context.WithValue(ctx, contextObserverKey{}, emit)
}

// logContextUsage 记录每次模型调用的 token 用量与耗时，便于事后核对上下文规模。
func logContextUsage(ctx context.Context, usage llm.Usage, elapsed time.Duration, failed bool) {
	ri := RunInfoFrom(ctx)
	if ri.AgentKey != "planner" && ri.AgentKey != "worker" {
		return
	}
	log.Printf("[context-usage] task=%d role=%s input_tokens=%d output_tokens=%d cache_read_tokens=%d cache_write_tokens=%d elapsed_ms=%d failed=%t", ri.TaskID, ri.AgentKey, usage.InputTokens, usage.OutputTokens, usage.CacheReadTokens, usage.CacheWriteTokens, elapsed.Milliseconds(), failed)
}
