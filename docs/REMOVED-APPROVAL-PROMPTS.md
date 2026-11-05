# 清理审批/封禁残留提示词

- 日期：2026-11-07
- 分支：`rebuild/from-series`
- 对照原版：`f48967a`（原版一处“审批/封禁”提示都没有）

## 结论

第 9 项已删掉审批模块的界面、接口和数据库强制，但提示词里还留着“未审批/封禁须跳过”“只能下发 approved 资产”等句子，模型会据此主动放弃可测目标。本项按用户选择 A 清理这些文案，只保留真实生效的约束：任务范围、任务隔离、删除、用户约束、并发与取消。

## 改了哪些文案

| 文件 | 改动 |
| --- | --- |
| `agent/deferred.go` | “未审批、封禁、撤回或越界” → “超出任务范围” |
| `agent/intent_dispatch.go` | 删除“可下发待审批资产意图；封禁、撤回” |
| `agent/mainagent.go` | 执行约束改为“任务范围、任务隔离及并发准入” |
| `agent/main_execution.go` | 删除“封禁、撤回”，保留删除与任务隔离 |
| `agent/planner.go` | 去掉“满足审批”与“绕过审批”；资产视图说明改为“可见不等于可测，只对任务范围内资产下发” |
| `agent/planner_bootstrap.go` | 约束列举去掉“资产审批、禁报”；读数失败文案改为“资产范围版本” |
| `agent/tools.go` | 工具描述去掉“审批过滤”“封禁/撤回”；9 处“未获授权或已被封禁”改为“不在当前任务范围内或不可用”；“审批查询失败”改为“资产查询失败” |
| `agent/tools_asset_view.go` | `list_task_assets` 描述由审批视图改为任务资产只读视图 |
| `agent/tools_insert.go` | `insert_assets` / `list_assets` 描述去掉 approved/pending/blocked/revoked 说法；删除 `approval_state` 筛选说明 |
| `agent/tools_target_access.go` | `check_target_access` 描述与规则改为“可用状态/任务范围” |
| `agent/worker.go` | 删除“待审批 pending 只限制 Planner…”整段中的审批说法，保留删除与隔离限制 |
| `agent/intent_stop.go` | 报错文案改为“不在当前任务范围内或不可用” |

## 测试调整

`agent/planner_asset_prompt_test.go` 原来断言提示词包含“未审批、封禁、撤回或越界”，改为断言包含“超出任务范围”。

## 保留未动

- 数据库里的 `approval_state` 列与 `revoked` 等状态值保留（历史数据与表结构不动）；
- `agent/cancelcause.go` 的“用户撤回或删除了当前意图关联的任务资产”保留，因为取消原因链路仍在使用；
- 数据库函数 `task_asset_blocked` / `task_asset_effective_approval_state` 仍恒返回“未封禁/通过”（第 9 项改动）。

## 行为变化

模型不再因为“未审批/封禁”而跳过资产；提示词描述的约束与当前代码一致。扫描与利用路径本身不改，只去掉误导性的自我限制。

## 验证

```
go build ./...                                   通过
go vet ./...                                     无输出
go test ./agent                                  通过
go test ./server（4 个相关用例，DSN 指向副本库）  Auth/Reporter/ContextTaskSelector 通过；FindingReport 用例因固定模板“报告过短”失败（既有基线）
```
