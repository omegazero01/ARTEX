# 移除：资产审批与资产跳过（第 9 项）

日期：2026-11-05。目标与 ARTEX 原版一致：资产登记后即可测试，没有“待审批/已撤回/已封禁”闸门。

## 改法

审批状态散落在数据库函数、工具钩子和提示词三层，逐列删除会牵动核心资产表（`task_asset_links` 有 142 处引用）。
因此采用“函数级还原”：

1. **数据库函数恒通过**（`db/schema.sql` 末尾追加覆盖、`db/asset_templates.sql`）：
   - `task_asset_blocked` → 恒 `false`
   - `task_asset_owner_approval_state` / `task_asset_effective_approval_state` → 恒 `'approved'`
   - `task_host_approval_state` → 任务存在即 `'approved'`
   所有按审批过滤的查询、校验都走这些函数，改完即全部放行。
2. **工具钩子透传**（`guard/assets.go`）：`AssetPolicyHooks` / `AssetPolicyHooksWithGuard`
   不再按资产状态拦截目标，只保留原有拦截规则链（Guard）。
3. **`check_target_access` 不再注册给模型**（`agent/tools.go`、`agent/tools_insert.go`、
   `agent/tool_contract.go`），内部实现改为恒“可用”，仅保留函数避免改动调用点。
4. **提示词里的审批约束删除**（`agent/asset_context.go` 的审批模板/跳过清单/目标预检注入、
   `agent/planner.go`、`agent/main_execution.go` 中的 pending/审批措辞）。
5. **审批界面与接口删除**：
   - 任务详情“资产审批”页签（`_tabs/asset-approvals-tab.tsx`，646 行）与入口
   - 资产页签里的批准/撤回按钮与确认框、`components/asset-approval-template.tsx`
   - 新建任务里的“资产审批模板”下拉框
   - `api.ts` 的审批接口、`types.ts` 的审批类型
   - 后端路由：`PUT /api/tasks/{id}/asset-approval-template`、
     `GET /api/tasks/{id}/asset-approvals`、`POST .../asset-approvals/{approve,revoke,block}`
6. 删除对应前端测试 `mock/asset-approval-templates.test.ts`、`mock/asset-origins.test.ts`。

## 保留的东西

- 数据表与列（`task_asset_links.approval_state`、`task_asset_blocks`、`task_asset_skips`）仍在，
  历史数据不动；只是不再影响能否测试。
- 拦截规则（intercept/Guard）、任务隔离、删除墓碑（blocked 记录仍展示）保留。
- mock 数据里仍保留审批模拟路由，仅供前端离线开发；生产后端已无这些接口。

## 行为变化

- 新发现资产、待审批资产、已撤回资产都可以直接测试，不再需要人工点“批准”；
- 任务详情不再有“资产审批”页签；
- 模型侧不再有 `check_target_access` 工具和“跳过未授权资产”的提示；
- 删除/封禁的资产仍会在列表里显示为“已封禁”，但不再阻断执行（数据库函数恒放行）。

## 恢复方式

`git revert` 本提交可恢复审批界面、工具、钩子与提示词；数据库函数需要重新执行原定义（在 `schema.sql` 被覆盖前的版本里）。
