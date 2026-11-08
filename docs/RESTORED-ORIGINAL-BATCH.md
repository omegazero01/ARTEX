# 恢复原版：意图批量条数 + pentest 上下文读取

> 基准：ARTEX 原版 `f48967a`（v0.3.15 官方快照）
> 变更分支：`rebuild/from-series`
> 一个提交完成两处能力恢复 + 两处残留清理。

## 一、add_intent 取消 4 条硬上限（恢复原版）

原版 `add_intent`（规划者新增探索方向的工具）支持一次提交任意条数；SERIES 在参数 schema 里加了 `maxItems: 4`，处理函数也加了 `>4` 直接报错的硬校验。本次移除两处：

- `agent/tools.go` `intentInputSchema()`：删掉 `intents` 上的 `maxItems: 4`，描述改为“1条以上”。
- `agent/tools.go` `addIntent()`：删掉 `if len(items) > 4 { return Errorf("一轮最多新增 4 条意图…") }`。
- 工具描述里的“（最多 4 条，比逐条调用省往返）”恢复为原版“（比逐条调用省往返）”。

保留：`duplicates`（重复方向提示）、逐条 `errors`、取消/终止能力——这些是 SERIES 的增强，不属于限制。
保留：planner 提示词里“最多 4 个最高价值”的建议——原版 `f48967a:agent/planner.go:335` 本来就有这句，是建议不是硬限制。

老库更新：`tools` 表是首次插入后不再覆盖，新增启动迁移 `refreshAddIntentToolDefinition`（标记 `add_intent_no_batch_cap_v1`），仅在 schema 仍含 `properties.intents.maxItems=4` 时刷新 schema，并把描述里的“最多 4 条”替换掉；不覆盖用户改过的其他描述。

## 二、pentest 补绑 context_get（恢复统一上下文检索的完整读取链）

SERIES 的 `MigrateContextTools`（db/tool_context_catalog.go，标记 `tool_context_catalog_v1`）按旧工具绑定推导新工具：`context_get` 只从 `node_detail` / `get_worker_trace` / `get_worker_output` / `expand_digest` 继承角色。pentest 原版只有 `list_findings`、`list_assets` 和流量工具，没有这些旧工具，于是老库漏绑了 `context_get`；而代码默认工具集给 pentest 带了这个工具，新库正常、老库缺失。

本次新增启动迁移 `MigratePentestContextGetBinding`（标记 `pentest_context_get_binding_v1`）：把 `pentest` 追加到 `tools.agents`（`key='context_get' AND system`，已存在则跳过），advisory lock 防并发。

同时补 `detail` 读取能力（迁移 `MigratePentestContextDetailCap`，标记 `pentest_context_detail_cap_v1`）：`context_get` 默认 `view=detail`，没有该能力时读正文仍会被拒；`context_search` 也只在有 `detail` 能力时才返回可直接读取的 `read` 引用。老库能力表由旧工具推导，pentest 原版只有 `list_findings`（映射到 `finding`），因此漏了 `detail`。

影响：`context_search` 命中结果里的 `read` 引用现在 pentest 能真正读全文，取证链不再断在“只能看摘要”。

## 三、两处残留清理

- 删除 `settings.tool_context_prompt_migration_report`：这是旧提示词迁移的扫描报告，每次启动重复打印“pentest references list_assets,list_findings”的过期结论。
- 删除 `tools` 表 `list_finding_deletion_feedback` 行：禁报功能已删除，代码无此工具实现，但老库里仍留着一条启用中的绑定（mainagent/planner/worker），会出现在工具列表里。
- 迁移函数 `MigrateRestoreOriginalCleanup`（标记 `restore_original_cleanup_v1`），仅删上面两项。

## 四、验证

- `go build ./...`、`go vet ./...` 通过。
- `go test ./agent -run TestAddIntentSchemaHasNoBatchCap` 通过（新增回归用例，防止 `maxItems` 再被加回）。
- 重启 8899 后查库：
  - `tools.schema` 的 `add_intent` 不再含 `maxItems`；
  - `tools` 表 `context_get.agents` 含 `pentest`；
  - `list_finding_deletion_feedback` 行不存在；
  - `settings.tool_context_prompt_migration_report` 不存在；
  - 启动日志不再打印旧提示词迁移报告。
