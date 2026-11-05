# 第 10 项：删除资产出处登记与审批残留死代码

- 日期：2026-11-06
- 分支：`rebuild/from-series`
- 对照原版：`f48967a`（ARTEX v0.3.15 官方快照）
- 结论：这不是“资产来源限制”，而是给资产记录“出自哪个任务/会话/工具调用”的出处链。原版没有，唯一读取方（资产审批列表）已在第 9 项删除，属于只写不读的死代码。

## 删了什么

| 层 | 文件 | 改动 |
| --- | --- | --- |
| 数据库层 | `db/asset_origins.go`（99 行） | 整个删除：`AssetOrigin` 类型、`setRegistrationOrigin`、`approvalOrigins` |
| 数据库层 | `db/asset_approval_groups.go`（167 行） | 只保留 `resolveApprovalSelection`（资产审批与封禁共用的分组解析），删除 `TaskAssetApprovalGroup`、`ListTaskAssetApprovalGroups` |
| 数据库层 | `db/asset_templates.go` | `RegisterAgentAssetWithOrigin` 不再调用 `setRegistrationOrigin`，出处不写库 |
| 数据库层 | `db/schema.sql` | `sync_task_asset_links()` 触发器不再写 `source_origin` 列；列本身与历史数据保留 |
| 后端 | `server/task_assets.go` | 删除三个没有路由的处理函数：`listTaskAssetApprovals`、`updateTaskAssetApprovals`、`mutateTaskAssetApprovals` |
| 后端测试 | `server/task_asset_approval_types_test.go`、`server/task_asset_blocking_test.go`、`server/asset_templates_test.go`、`db/asset_templates_test.go` | 删除只测审批接口/分组列表的用例；`db/asset_templates_test.go` 保留模板默认值与证据测试，只去掉分组与出处断言 |
| Agent 测试 | `agent/asset_origin_test.go`、`agent/sdk_session_test.go` | 删除出处链用例；SDK 会话测试不再校验 `source_origin` |
| 前端 | `web/src/lib/types.ts`、`web/src/lib/mock/handler.ts` | 删除 `AssetOrigin` 类型、`origins` 字段与 mock 出处数据 |

## 没删什么

- `task_asset_links.source_origin` 列：保留列定义与历史数据，只停止写入。归档/恢复代码未引用该列，不受影响。
- `QueryDSLInScope` / `GetByIDsInScope`：按任务声明范围过滤资产，是原版 `f48967a` 自带行为，删除会改变 ARTEX 语义。
- `ApproveTaskAssetsResolved` 等审批状态写入函数：仍被数据库层测试与归档流程引用，保留。

## 行为变化

- Agent 登记资产时不再额外查询活动表、不再写一段 JSON 出处；资产入库路径少一次查询。
- 任务资产列表不再返回 `origins` 字段（第 9 项删审批页后前端本就不再展示）。
- 审批/封禁接口仍无路由，与第 9 项一致。

## 验证

| 项目 | 命令 | 结果 |
| --- | --- | --- |
| 编译 | `go build ./...` | 通过 |
| 静态检查 | `go vet ./...` | 无输出 |
| Agent 单测 | `go test ./agent -count=1` | ok（2.9s） |
| 前端类型 | `cd web && npx tsc --noEmit` | 0 错误 |
| 前端单测 | `npm run test:ui` | 51/51 通过 |

## 与原版一致度

`db/asset_origins.go`、`db/asset_approval_groups.go` 在原版均不存在；删除后这两个文件不再出现在分支中。`source_origin` 列与 `task_asset_links` 其余审批字段仍留在库里，与原版表结构不完全一致，属于第 9 项“保留历史数据”的既有决定。
