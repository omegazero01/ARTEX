# 移除：联盟模块（画像 + 台账）

日期：2026-11-05。范围：整个联盟模块，含画像文档、台账系统、任务关联、brief 注入。

## 删除的内容

**后端 Go（约 2,500 行）**
- `allianceprofile/`：`document.go`（画像文档结构与校验）、`template.json`（画像模板）
- `db/alliance.go`（533 行）：台账 CRUD、端点、任务关联
- `db/alliance_profile.go`（826 行）：画像版本、快照、阅读记录、建议、提示词注入
- `db/alliance_profile_test.go`
- `server/alliance.go`（435 行）：台账接口、统计、brief 组装
- `server/alliance_profile.go`（585 行）：画像接口、反馈循环、阅读
- `server/alliance_brief_test.go`

**接线**
- `server/server.go`：15 条联盟路由；启动时旧台账画像回填；`allianceProfileFeedbackLoop` 后台协程；
  建任务时的联盟系统解析、brief 并入描述、画像快照落库、建后台账关联；请求字段 `alliance_system_ids`
- `server/goals.go`：建目标后触发画像阅读
- `db/tasks.go`：`TaskCreateOptions.AllianceSnapshots` 与快照写入
- `agent/{goals,planner,worker,mainagent}.go`：每轮把画像摘要拼进提示词的 4 处调用

**数据库表（schema.sql 末尾整段删除，共 6 张 + 3 个字段）**
- 表：`alliance_systems`、`alliance_endpoints`、`alliance_task_links`、
  `alliance_profile_revisions`、`alliance_task_snapshots`、`alliance_task_readings`、
  `alliance_profile_feedback_runs`、`alliance_profile_suggestions`
- `alliance_systems` 上为画像新增的列：`environment`、`document`、`version`

**前端**
- 页面 `web/src/app/(main)/function/alliance/page.tsx`（801 行，联盟画像页）
- 任务详情页签 `_tabs/alliance-profile-tab.tsx`（180 行）及详情页的注册
- 组件 `web/src/components/alliance-profile.tsx`（557 行）
- `web/src/lib/alliance-profile-stats.ts` 及其测试
- 侧边栏入口「联盟画像」
- 新建任务表单里的 `AlliancePicker`（选择联盟画像、预览 brief）
- `api.ts` 的 15 个联盟方法、`types.ts` 的 12 个联盟类型

## 保留

- `docs/proposals/alliance-*.md`、`docs/verification/alliance-*.md`：历史方案与验收记录，未删。

## 影响

- 新建任务不再有「联盟画像」选项；已有任务的画像页签消失（历史数据仍在库里，只是不再读取）。
- 老库里的联盟表不会自动删除（schema 只是不再创建）。要彻底清掉需手工 `DROP TABLE`，本仓库不包含这一步。
- 已有任务的描述里若曾并入过 brief，那段文字保留在描述中，不会被自动去除。

## 恢复方式

本提交之前的版本（`git log --grep 移除联盟` 的上一个提交）完整保留全部联盟代码；
需要恢复时 `git revert` 该提交即可。
