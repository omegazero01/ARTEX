# 恢复 ARTEX 原版读取方式

- 日期：2026-11-07
- 分支：`rebuild/from-series`
- 决定：按用户选择 B，连自动注入一起去掉，Planner 恢复“先查后建”

## 删掉的三样

1. **工具返回总预算 2.4 万字符**：删除 `agent/tool_budget.go` 里的常量与所有裁剪分支；`ProjectToolDetail` 不再循环减半，`budgetRows` 不再截断，`node_detail` 步骤元数据、摘要元数据、`insert_assets` 详情不再按预算均分。保留每个工具自己的 `max_chars`/`offset` 窗口参数。
2. **Worker 启动时注入 20 条上游事实**：删除 `agent/worker_context.go` 的 `workerLaunchContext`；Worker 现在只有任务目标、本条意图和记忆预览（`workerMemoryPreview` 单独保留在 `agent/worker_memory_preview.go`），要看历史自己调 `context_search`/`context_get`/`list_facts`。
3. **Planner 本轮图引导**：删除 `agent/planner_bootstrap.go`（`plannerBootstrapRunes`、`plannerRetrievalRule`、`plannerBootstrap`、`refreshPlannerBootstrap`）；`Plan` 不再拼事件计数、相关记录和读取入口。`SetTaskContextReadAccess` 保留为空实现，避免改 server 调用。

## 恢复原版的提示词

`agent/planner.go` 的默认规划提示第 1 步改为“历史进度不在本轮输入里，先按需查询已有意图、资产和探索图，确认方向不重复、不遗漏再下发”；新增常量 `plannerQueryRule` 作为系统提示补充；唤醒引导去掉“结合图引导”的表述。

## 测试

删除只验证已删机制的用例：`planner_bootstrap_test.go`、`planner_context_metrics_test.go`、`context_window_metrics_test.go`、`planner_scenario_metrics_test.go`、`hint_trigger_test.go`、`planner_user_cancel_test.go`、`context_performance_test.go`；从 `context_review_test.go` 删除两个依赖 Worker 预载的用例；从 `tool_detail_contract_test.go` 删除两个预算断言用例；`planner_proactive_test.go` 删除“无查询前置”的用例；`tool_assets_budget_test.go` 去掉预算断言。

验证结果：`go build ./...`、`go vet ./...` 通过；`go test ./agent` 通过。`go test ./server` 需要 `SERIES_PG_DSN`，未配置时按环境失败，与本次改动无关。

## 没动的

- `graph_overview` 的态势窗口上限（这是 SERIES 新工具，本身不在原版）；
- 工具参数里的 `max_chars`/`offset` 单次窗口（显式参数，不传时按工具默认）；
- 资产归属校验与任务隔离；
- 记忆预览（SERIES 记忆功能，不属于读取范围限制）。
