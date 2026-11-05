# 放行未登记主机：补齐资产审批的出口校验

- 日期：2026-11-07
- 分支：`rebuild/from-series`
- 起因：用户指出代理出口的“先登记再访问”仍属资产审批残留，第 9 项没有清干净

## 事实核对

- `traffic/traffic.go` 的 authorizeTaskRequest、`db/worker_execution.go` 的 ValidateWorkerHosts、`db/task_assets.go` 的 ValidateTaskHostsApproved 都来自 SERIES 基线提交 `ee807c2`（`678cdae` 已含），不是 `1bbadc9` 新增；`1bbadc9` 没有改 `traffic/`。
- 但它们确实是资产审批能力的执行点：第 9 项只删了界面、接口和 guard 钩子的部分接线，这三处仍按主机状态拒绝请求，属于同一功能的残留，本项补齐。

## 改动

`db/task_assets.go` 的 ValidateTaskHostsApproved：

- 未登记主机（状态为空）直接放行；
- 已登记主机仍按 approved 放行、blocked 返回封禁错误、其余状态（pending/revoked）返回未批准错误。

`db/worker_execution.go` 的 ValidateWorkerHosts：

- 未登记主机与 pending 放行（Worker 执行期允许新发现主机）；
- blocked/revoked 仍拒绝。

效果：Worker 的 curl、nuclei、ffuf 以及代理出口不再要求“先把主机登记成任务资产”；封禁与撤回语义保留。`traffic/traffic.go` 的调用逻辑未改，拦截分支不再被未登记主机触发。

## 测试

- `go build ./...`、`go vet ./...` 通过；
- `go test ./agent ./traffic ./guard` 全部通过；
- `go test ./db` 仍有一批资产审批语义用例失败，但在备份分支 `backup/reduction-history` 上同样失败（第 9 项改数据库函数后就已红），本次没有新增失败。

## 未动

- 资产 ID 维度的校验 `ValidateTaskAssetsApproved` / `ValidateWorkerAssets` 仍保留，用于写入时检查显式资产 ID 是否属于任务；它与“出口主机”是两回事，需要单独决定是否也放开。
