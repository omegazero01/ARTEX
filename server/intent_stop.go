package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Autumn-27/artex/agent"
)

// cancelIntentFromMain 让主代理按用户要求取消一条意图：与 UI 取消共用同一条
// 停止路径（等执行真停、写 cancelled_by_user 锁、通知规划者），并带上任务操作闸门。
func (s *Server) cancelIntentFromMain(ctx context.Context, taskID, intentID int64, reason string) (agent.IntentStopResult, error) {
	key := strconv.FormatInt(taskID, 10)
	if !s.engine.beginTaskOperation(key) {
		return agent.IntentStopResult{}, fmt.Errorf("任务正在删除，无法取消意图")
	}
	defer s.engine.decInflight(key)
	return s.engine.StopIntent(ctx, taskID, intentID, reason)
}
