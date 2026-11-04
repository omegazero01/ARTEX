# SERIES vs ARTEX 功能差异对照表

口径：每一行都对应已读过的源码文件（路径见“证据”列）。**列出“无”是指全仓没有对应实现文件，且在 230 个差异文件的复核中该侧工具面/接口也不存在同名能力**，不是没找到。  
未做的部分写在文末：没有把两边同场景端到端跑一遍做行为对照，结论来自静态实现 + 各自测试用例。

## 一、SERIES 有、ARTEX 没有

| 功能 | SERIES 实现 | ARTEX | 差异性质 | 风险 |
|---|---|---|---|---|
| 资产审批模板（all/related/explicit） | `db/asset_templates.go`、`server/task_assets.go`、`web/.../asset-approvals-tab.tsx` | 无 | 整块删除 | 删除后新发现资产不再有“待审批”闸门，ARTEX 侧只能靠拦截规则兜 |
| 资产分组审批 / 封禁 / 墓碑 | `db/asset_approval_groups.go`、`db/task_asset_blocking.go`、`db/asset_skips.go` | 无 | 整块删除 | 封禁只按 host 生效，子域不级联（SERIES 语义），ARTEX 无此能力 |
| 资产来源回溯（tool_use/会话指针） | `db/asset_origins.go`、`db/task_asset_view.go` | 无 | 整块删除 | 审批时无法回到原始调用现场 |
| 资产级读写授权（agent 侧） | `agent/asset_context.go`、`agent/target_access_context.go`、`db/worker_execution.go`、`db/main_execution.go` | 无 | 整块删除 | 执行期读到未审批资产不会被过滤；这是 SERIES 渗透能力边界的一环 |
| 派生资产父级授权 | `db/derived_asset_parents.go`、`agent/worker_execution.go` | 无 | 整块删除 | 父子授权联动消失 |
| 执行模式 managed/manual | `db/execution_mode.go`、`server/execution_mode.go`、`web/execution-control.tsx` | 无 | 整块删除 | 没有“人工逐条派发”模式 |
| Worker 队列（拖拽排序/删除墓碑） | `db/worker_queue.go`、`server/worker_queue.go`、`web/worker-queue.tsx` | 无 | 整块删除 | 只能自动领取，无人工排序 |
| 主 Agent 待审批资产下发 | `agent/main_execution.go`、`agent/intent_dispatch.go` | 无 | 整块删除 | 主 Agent 不能把 pending 意图下发给 worker |
| 上下文检索工具（context_search/get） | `agent/context_tools.go`、`db/context_reads.go`、`db/tool_reads.go` | 无（工具面为 list_facts/node_detail，仅 `server/report_format.go` 注释里残留 context_search 字样） | 整块删除 | ARTEX 读上下文靠整包列表，无分页/字段级续读 |
| 检索历史有界投影 | `agent/retrieval_history.go` | 无 | 整块删除 | 长会话历史工具结果不会被折叠成索引 |
| 态势 lean/full 投影 + 数值配置 | `agent/overview_lean.go`、`overview_projection.go`、`overview_config.go` | 无（只有完整 graph_overview） | 整块删除 | Planner 上下文体积不可控 |
| 压缩两条线可配置 | `agent/context_policy.go`、`server` 设置面 | 无 | 整块删除 | 压缩阈值只能跟窗口推导 |
| 智能代理（按 host 切池） | `traffic/smartproxy.go`、`guard/proxy_tag.go`、`agent/tools_smartproxy.go` | 无 | 整块删除 | 被 WAF 拦时没有“换出口重试”的通道 |
| 流量压缩/保留期/静态过滤 | `traffic/compress.go`、`traffic/traffic.go` 的 retention/skip | 无 | 整块删除 | 流量库只增不减 |
| traffic_search/traffic_get 工具 | `traffic/tool_reads.go`（469 行，含游标与角色授权） | 无独立实现（全仓只在提示与 reporter 绑定里出现工具名） | 整块删除 | Agent 没有按游标读流量正文的工具，需单独确认是否走别的路径 |
| 禁报规则 + 漏洞删除反馈 | `db/finding_suppression.go`、`db/finding_deletion_feedback.go`、`server/finding_suppression.go`、`agent/finding_feedback_context.go` | 无（该能力曾移植后回滚） | 整块删除 | 用户删漏洞后，同类不会被禁止重报 |
| 任务未读通知角标 | `db/task_notifications.go`、`web/task-notifications.tsx` | 无 | 整块删除 | 只能靠刷新看新增 |
| HTTP 入口 Basic Auth（双层） | `server/http_auth.go`、`http_auth_settings.go` | 无（只有 JWT） | 整块删除 | 入口层没有第二道口令 |
| shell 自定义工具 + 命令计量 | `db/tools.go` shell 元数据、`server/shell_usage.go`、`shell_test_api.go` | 无（自定义工具无 shell kind） | 整块删除 | 只能写脚本工具，无命令级归因 |
| MCP 用量/测试/导入 | `db/mcp_usage.go`、`server/mcp_usage.go`、`mcp_runtime.go` | 无 | 整块删除 | MCP 调用不可观测 |
| 工具 UA 注入 | `agent/tool_ua.go` | 无 | 整块删除 | curl/wget 用默认 UA，更像扫描器 |
| 工具契约/目录迁移 | `agent/tool_contract.go`、`db/tool_context_catalog.go` | 无 | 整块删除 | 旧配置可能覆盖新契约 |

## 二、ARTEX 有、SERIES 没有

| 功能 | ARTEX 实现 | SERIES | 差异性质 | 风险 |
|---|---|---|---|---|
| 漏洞推送（钉钉/飞书/企微/Telegram/Webhook/邮件） | `notify/` 26 文件、`server/notifier.go`、`notify_api.go`、`web/system/notify/` | 无 | 整块新增 | SSRF 默认放行 RFC1918；逃生口是进程级开关；邮件无账号时明文投递 |
| 在线更新（检查/升级/回滚） | `selfupdate/` 5 文件、`server/update.go`、`web/update-card.tsx` | 无 | 整块新增 | SHA256SUMS 与 zip 同源无签名；Windows 上大量用例 Skip |
| 资产拦截规则（7 种 kind） | `db/asset_intercept*.go`、`server/asset_intercept.go`、`web/system/intercept/assets` | 无 | 整块新增 | fuzzy 子串过宽；每次插入资产全量拉规则匹配 |
| 任务级 block/allow 规则 | `db/task_intercept.go`、`server/task_intercept.go` | 无 | 整块新增 | 建任务时随事务写入，失败整单回滚 |
| ScopeSentry 资产同步 | `server/sync_scopesentry.go`、`web/function/sync` | 无 | 整块新增 | MCP URL 无拨号守卫，可指向内网；同步串行且单次可跑 10 分钟 |
| 工作区文件管理 | `server/workspace.go`、`web/function/workspace` | 无 | 整块新增 | **无 EvalSymlinks，软链可越界读写**（已核源码） |
| 全局搜索对话框 | `web/.../sidebar/search-dialog.tsx` | 无 | 新增 UI | — |
| findings-table 新列表 | `web/.../findings-table.tsx` | 无 | 重写 UI | — |
| 报告/复测页签 | `web/.../report-tab.tsx`、`retests-tab.tsx` | 无对应页签（有导出） | 新增 UI | — |
| 命令拦截裁判用量 | `server/intercept.go` 的 judgeWorkerLane/interceptJudgeUsage | 无 | 新增 | 规则 pattern 是用户正则（ReDoS 面） |

## 三、两边都有、实现已改写（差异文件 230 中的重点）

| 功能 | 差异 | 风险 |
|---|---|---|
| 收尾禁 Bash | SERIES 用 `WrapDeps`+SettlementGate 保工具清单不变；ARTEX 用 `settlement_gate.go` 前缀匹配 | ARTEX 版可能被“恰好以 wrap-up 开头”的正文提前触发 |
| 漏洞报告导出 | SERIES `md-report`/`md-report-zip` 统一提交版；ARTEX `report.go` 另有一套 | ARTEX 文件名按字节截断，可能切出非法 UTF-8 |
| Worker 反馈 | SERIES 有 outbox+双游标重放；ARTEX 异步下发回执 | 断线补偿语义不同 |
| 认证 | SERIES 校验 JWT alg/subject + 入口 Basic；ARTEX 只判 Bearer+cookie `artex_token` | ARTEX 不校验算法，存在 alg 混淆风险 |
| SSE 广播 | SERIES 缓冲满断开客户端；ARTEX 静默丢弃且不 close | 客户端无法补洞 |
| 主机经验/画像 | 两侧都有（各 275 + 146 行），内容为 logic diff：SERIES 带去抖落盘与 pinned 注入 | 移植是“对齐差异”，不是新增能力 |

## 四、这张表的边界

- 已做：功能 → 实现文件 → 差异性质 → 风险的对应；每行都有文件级证据，文件本身在前两轮已读（230 差异 + 660 单边）。
- 未做：**两边同场景端到端行为对照**（例如同一任务在 SERIES 走审批、在 ARTEX 直接执行，逐项比对输出）。这类验证需要同时起两个实例、跑同一批目标，属于行为验证，不是代码核对。
- 结论：功能清单已对齐，可以据此判断“哪边有什么、差在哪、风险在哪”；不能据此断言“某功能在运行时表现完全一致”。
