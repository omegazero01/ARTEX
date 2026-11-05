# 移除：禁报规则（finding suppression）

日期：2026-11-05。按用户决定整块删除：ARTEX 原版（f48967a）没有这套机制，
它是 SERIES 为压误报后加的，行为是“命中规则直接拒绝记录漏洞”。

## 删除的内容

**后端**
- `db/finding_suppression.go`（167 行）：规则匹配与校验
- `db/finding_deletion_feedback.go`（105 行）：删除反馈记录
- `server/finding_suppression.go`（450 行）：规则服务、模型复核、列表/撤销接口、启动时的工具绑定播种
- `agent/finding_deletion_feedback.go`（167 行）：`list_finding_deletion_feedback` 工具
- `agent/finding_feedback_context.go`（192 行）：把“你删过这些漏洞”注入提示词
- 对应测试 6 个文件（约 1,200 行）

**接线**
- `db/finding_traffic.go`：`RecordFindingInput` 去掉 `SuppressionCheck` / `SuppressionTargets`，
  记录漏洞时不再调用 `ValidateFindingSuppressionTx`，不再写“禁报校验通过/未完成”活动记录
- `evidence/store.go`：记录路径去掉同一道校验
- `db/findings.go`：删除漏洞不再生成反馈行和禁报规则（`DeleteFindingWithFeedback` 变回 `DeleteFinding`）
- `server/server.go`：去掉规则服务字段、两条规则路由、`suppressionInfrastructureError` 包装；
  删除漏洞接口不再解析“删除原因”
- `server/orchestration.go`：去掉启动时的禁报工具绑定播种
- `agent/{tools,tools_insert,tool_contract,tool_history}.go`：去掉 `list_finding_deletion_feedback` 工具与契约
- `agent/noa_asset_context.go`：去掉每轮刷新删除反馈
- `db/task_archives.go`、`db/task_archives_restore.go`：归档不再包含/恢复这两张表
- `db/schema.sql`：删除 `finding_deletion_feedback`、`finding_suppression_rules` 两张表及索引

**前端**
- `components/finding-suppression-rules.tsx`（规则列表/撤销）
- 任务详情漏洞页签里的规则入口
- `api.ts` 两个接口、`types.ts` 的 `FindingSuppressionRule`
- mock：删除两个测试文件与 mock 路由、内存规则表
- 删除漏洞的调用不再传“删除原因”（后端已忽略该字段）

## 行为变化

- 模型再次上报你删过的漏洞时**不再被拒绝**，会正常入库；
- 任务运行时不再收到“漏洞被删除”的唤醒——保留：删除时仍会唤醒 planner（按删除前查到的任务号），
  只是不再附带反馈 ID 与规则数据；
- 归档包与恢复逻辑不再携带规则数据；老库里的两张表不会被自动删除。

## 恢复方式

`git revert` 本提交即可恢复全部禁报功能（数据表需重新执行 schema 创建）。
