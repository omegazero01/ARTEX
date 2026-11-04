# SERIES（二开） vs ARTEX（基准）全量代码核对报告

> **状态声明（以此为准）**
> - 已读并核对：229 个「两边都有且逻辑有差异」的文件（其中 20 余个大文件按功能块读取）。
> - 单边文件：SERIES 独有 566 个（73,957 行）与 ARTEX 独有 94 个（42,564 行）已逐个打开并读正文（2026-11-02 用 10 路并行子代理完成，逐文件读实现与断言，不是只看符号）；文档读标题与章节，验证日志按用途登记，截图未做视觉分析。
> - 仍未逐行：74 个纯重命名与 216 个纯空白差异（仅按哈希/行数分类）；约 20 个超大差异文件按功能块读取（清单见下）。
> - 因此结论口径是「229 个差异文件 + 660 个单边文件已完成文件级语义核对」，不是「每一行都读过」。

- 基准仓库：E:\AI\ARTEX-mian（artex）
- 二开仓库：E:\AI\series（series）
- 文件盘点：SERIES 1097 / ARTEX 625，交集 531

## 文件盘点状态：已完成

- SERIES 文件总数：1097
- ARTEX 文件总数：625
- 交集：531
- SERIES 独有：566
- ARTEX 独有：94
- 实质逻辑差异（已逐个读完整 diff）：230（含 1 个被 git 当二进制的文件）
- 纯重命名：74
- 纯空白/格式：216

## 语义核对进度：230 / 230 差异文件 + 660 / 660 单边文件


## 核对方法与边界（必须先看）

- 229 个实质逻辑差异文件的 diff 全部打开并做了语义判断；下表按“读取粒度”区分：
  - **逐行读取**：其余全部文件（含 db/exploration.go 851 行、db/schema.sql 681 行、agent/tools.go 2176 行、server/engine.go 458 行、server/orchestration.go 399 行等）。
  - **按功能块/hunk 读取（非逐行）**：`server/server.go`（973 行，读了导出/补丁/删除/播报/设置/路由等功能块）、`server/manager.go`（505 行，前 330 行逐行 + 其余按函数签名与 hunk）、`db/task_assets.go`（1035 行，前 330 行逐行 + 其余 23 个被删函数逐个确认）、`db/task_assets_test.go`（634 行，按被删用例清单确认）；前端 `system/{intercept,llm,mcp,agents,skills,tools,settings}/page.tsx`（均为 SERIES 一行 re-export → ARTEX 整页新增，按组件结构与函数清单读取）、`function/tasks/page.tsx`、`function/findings/page.tsx`、`findings/detail/page.tsx`、`tasks/detail/page.tsx`、`_tabs/{sessions,broadcast,findings,overview,assets}-tab.tsx`、`components/{agent-editor,transcript}.tsx`、`lib/{api,types}.ts`、`lib/mock/handler.ts`（均为大段 UI/类型改写，按新增/删除的功能符号与关键 hunk 读取）。
- 未逐行读取的部分均为：重复的样式类名、整页新增 JSX 的结构性代码、或成批删除的审批测试；其**功能增删与语义**已逐项确认，但若要求“每个字符都读过”，上述文件不满足，需再补。

## 一、series 独有（artex 不存在）文件 & 能力清单

共 566 个文件。按目录：

| 目录 | 文件数 |
|---|---|
| docs | 126 |
| agent | 105 |
| db | 80 |
| web | 79 |
| server | 76 |
| artifacts | 60 |
| traffic | 12 |
| "docs | 9 |
| guard | 5 |
| screenshots | 5 |
| (根目录) | 2 |
| cmd | 2 |
| report | 2 |
| .zcode | 1 |
| mcphttp | 1 |
| third_party | 1 |

### 能力清单（按目录归纳）

- **agent/**：SERIES 独有的上下文工程与资产授权工具层：context_policy/context_tools、asset_tools、sdk_session、overview_lean/config/projection、planner_bootstrap、tool_budget/defaults/input/contract、retrieval_history、finding_feedback_context、finding_deletion_feedback、worker 资产授权与任务域工具等（约 105 个 Go 文件）。
- **db/**：SERIES 独有的数据库层：asset 授权（task_asset_*）、finding_suppression_rules、worker_feedback、mcp_usage、tool_usage_backfill.sql、asset_templates.sql 等。
- **docs/**：SERIES 独有：系列设计/审计文档（docs/review 等）。
- **server/**：SERIES 独有的服务端文件：finding_suppression（禁报）、asset approval 相关 handler、smart proxy、planner overview/context policy 接口、profile 级 agent 构建等。
- **traffic/**：SERIES 独有：smartproxy.go、压缩/保留期相关实现与测试。
- **web/**：SERIES 独有的前端组件与页面：settings 工作区组件（settings/*）、finding-detail-content、findings-toolbar/pagination/presentation/selector、task-notifications、worker-queue、source-transcript、smart-proxy 页签、asset-approvals 页签、session-tool-calls、tool-calls、http-auth 等。
- **"docs/**：9 个独有文件（256.md", 262-20261023.md", 203-20261023.md", 226-20261023.md", 271-20261023.md", 241-20261023.md" …）
- **(根目录)/**：2 个独有文件（.DS_Store, build-docker.sh）
- **.zcode/**：1 个独有文件（plan-sess_354ed7fe-47b6-4853-8dfe-718620b00145.md）
- **artifacts/**：60 个独有文件（asset-approval-performance-review.md, asset-approval-refactor.md, asset-approval-templates.md, README.md, agent.log, db-final.log …）
- **cmd/**：2 个独有文件（main.go, main_test.go）
- **guard/**：5 个独有文件（assets.go, assets_test.go, compress_test.go, proxy_scope_test.go, proxy_tag.go）
- **mcphttp/**：1 个独有文件（capture_test.go）
- **report/**：2 个独有文件（findings_test.go, helpers.go）
- **screenshots/**：5 个独有文件（series-http-auth-settings.jpg, series-login.jpg, series-session.jpg, series-settings-dark.jpg, series-tasks.jpg）
- **third_party/**：1 个独有文件（.DS_Store）

<details><summary>展开 SERIES 独有文件完整清单</summary>


**"docs/**（9）

- `"docs/\346\274\217\346\264\236\346\265\201\351\207\217\350\257\201\346\215\256.md"`
- `"docs/review/\344\270\215\347\272\263\345\205\245/\346\212\245\345\221\212-158-\351\233\205\346\230\237\345\205\254\344\272\244\346\234\252\346\216\210\346\235\203\344\277\241\346\201\257\346\263\204\351\234\262-20261023.md"`
- `"docs/review/\346\212\245\345\221\212-155-\345\275\261\345\203\217\347\263\273\347\273\237\346\260\264\345\271\263\350\266\212\346\235\203-20261023.md"`
- `"docs/review/\346\212\245\345\221\212-156-\346\235\203\347\233\212\351\207\215\345\244\215\351\242\206\345\217\226-20261023.md"`
- `"docs/review/\346\212\245\345\221\212-156-\347\247\257\345\210\206\345\225\206\345\237\216\351\200\200\346\254\276\351\207\221\351\242\235\347\257\241\346\224\271-20261023.md"`
- `"docs/review/\346\212\245\345\221\212-156-\350\256\242\345\215\225\346\216\245\347\256\241-20261023.md"`
- `"docs/review/\346\212\245\345\221\212-156-\351\223\266\350\241\214\345\215\241\345\217\267\350\266\212\346\235\203\346\237\245\350\257\242-20261023.md"`
- `"docs/review/\346\212\245\345\221\212-158-\351\252\214\350\257\201\347\240\201\347\273\225\350\277\207-20261023.md"`
- `"docs/review/\346\212\245\345\221\212-\347\247\257\345\210\206\345\225\206\345\237\216SQL\346\263\250\345\205\245-20261023-v2.md"`

**(根目录)/**（2）

- `.DS_Store`
- `build-docker.sh`

**.zcode/**（1）

- `.zcode/plans/plan-sess_354ed7fe-47b6-4853-8dfe-718620b00145.md`

**agent/**（105）

- `agent/assembly_failure_test.go`
- `agent/asset_context.go`
- `agent/asset_context_test.go`
- `agent/asset_management_context.go`
- `agent/asset_origin_test.go`
- `agent/asset_skip_scope_test.go`
- `agent/asset_tools.go`
- `agent/assets_approval_dsl_test.go`
- `agent/builtin_review_test.go`
- `agent/chat_task_policy_test.go`
- `agent/compaction_review_test.go`
- `agent/context_budget.go`
- `agent/context_compat_test.go`
- `agent/context_match_read_test.go`
- `agent/context_mixed.go`
- `agent/context_performance_test.go`
- `agent/context_policy.go`
- `agent/context_policy_test.go`
- `agent/context_prompt_measurement_test.go`
- `agent/context_review_test.go`
- `agent/context_task.go`
- `agent/context_task_test.go`
- `agent/context_tools.go`
- `agent/context_tools_test.go`
- `agent/context_window_metrics_test.go`
- `agent/coverage_params_test.go`
- `agent/digest_authorization_test.go`
- `agent/fact_contract_test.go`
- `agent/finding_deletion_feedback.go`
- `agent/finding_deletion_feedback_test.go`
- `agent/finding_feedback_context.go`
- `agent/finding_feedback_context_test.go`
- `agent/findings_pagination_test.go`
- `agent/graph_context_test.go`
- `agent/hint_trigger_test.go`
- `agent/intent_dispatch.go`
- `agent/intent_dispatch_test.go`
- `agent/intent_stop_test.go`
- `agent/lightformat.go`
- `agent/lightformat_test.go`
- `agent/local_tools.go`
- `agent/main_dispatch_feedback.go`
- `agent/main_dispatch_feedback_test.go`
- `agent/main_execution.go`
- `agent/main_execution_test.go`
- `agent/noa_asset_context.go`
- `agent/noa_test.go`
- `agent/official_provider_test.go`
- `agent/official_runtime_review_test.go`
- `agent/one_shot_session_test.go`
- `agent/overview_config.go`
- `agent/overview_lean.go`
- `agent/overview_lean_test.go`
- `agent/overview_projection.go`
- `agent/overview_projection_test.go`
- `agent/overview_sync_test.go`
- `agent/planner_asset_prompt_test.go`
- `agent/planner_bootstrap.go`
- `agent/planner_bootstrap_test.go`
- `agent/planner_context_metrics_test.go`
- `agent/planner_proactive_test.go`
- `agent/planner_scenario_metrics_test.go`
- `agent/planner_user_cancel_test.go`
- `agent/retrieval_access_test.go`
- `agent/retrieval_history.go`
- `agent/retrieval_history_test.go`
- `agent/retrieval_noa_test.go`
- `agent/sdk_review_test.go`
- `agent/sdk_session.go`
- `agent/sdk_session_test.go`
- `agent/structured_capture.go`
- `agent/structured_capture_test.go`
- `agent/structured_history.go`
- `agent/structured_history_test.go`
- `agent/target_access_context.go`
- `agent/target_access_test.go`
- `agent/taskmanager.go`
- `agent/tool_assets_budget_test.go`
- `agent/tool_budget.go`
- `agent/tool_budget_test.go`
- `agent/tool_contract.go`
- `agent/tool_contract_test.go`
- `agent/tool_defaults.go`
- `agent/tool_detail_contract_test.go`
- `agent/tool_documentation_test.go`
- `agent/tool_error_guidance_test.go`
- `agent/tool_history.go`
- `agent/tool_input.go`
- `agent/tool_input_test.go`
- `agent/tool_inventory_test.go`
- `agent/tool_read_scope.go`
- `agent/tool_surface_metrics_test.go`
- `agent/tool_ua.go`
- `agent/tool_ua_test.go`
- `agent/tools_asset_view.go`
- `agent/tools_asset_view_test.go`
- `agent/tools_contract_test.go`
- `agent/tools_description_assets.go`
- `agent/tools_smartproxy.go`
- `agent/tools_smartproxy_test.go`
- `agent/tools_target_access.go`
- `agent/worker_context.go`
- `agent/worker_execution.go`
- `agent/worker_execution_test.go`
- `agent/worker_trace_tools_test.go`

**artifacts/**（60）

- `artifacts/asset-approval-performance-review.md`
- `artifacts/asset-approval-refactor.md`
- `artifacts/asset-approval-templates.md`
- `artifacts/asset-source-review-20260916/README.md`
- `artifacts/asset-source-review-20260916/agent.log`
- `artifacts/asset-source-review-20260916/db-final.log`
- `artifacts/asset-source-review-20260916/error-after.log`
- `artifacts/asset-source-review-20260916/error-before.log`
- `artifacts/asset-source-review-20260916/multiple-verified.log`
- `artifacts/asset-source-review-20260916/pagination-verified.log`
- `artifacts/asset-source-review-20260916/retry-verified.log`
- `artifacts/asset-source-review-20260916/scenarios-verified.log`
- `artifacts/asset-source-review-20260916/server-final.log`
- `artifacts/asset-source-review-20260916/types-final.log`
- `artifacts/builtin-system-prompts.md`
- `artifacts/cold-compactor-sync-20260919/README.md`
- `artifacts/cold-compactor-sync-20260919/agent.log`
- `artifacts/cold-compactor-sync-20260919/before.log`
- `artifacts/cold-compactor-sync-20260919/server.log`
- `artifacts/mcp-optimization-20260919/README.md`
- `artifacts/performance-fixes-20260919/README.md`
- `artifacts/performance-fixes-20260919/agent.log`
- `artifacts/performance-fixes-20260919/browser.log`
- `artifacts/performance-fixes-20260919/db-verified.log`
- `artifacts/performance-fixes-20260919/server-verified.log`
- `artifacts/performance-fixes-20260919/types-final.log`
- `artifacts/performance-fixes-20260919/web-verified.log`
- `artifacts/review-20260908-20260911.md`
- `artifacts/task-detail-first-entry-waterfall.json`
- `artifacts/task-detail-loading-review.md`
- `artifacts/task-execution-mode-20260920/README.md`
- `artifacts/task-execution-mode-20260920/agent.log`
- `artifacts/task-execution-mode-20260920/browser.log`
- `artifacts/task-execution-mode-20260920/db-race.log`
- `artifacts/task-execution-mode-20260920/db.log`
- `artifacts/task-execution-mode-20260920/mock.log`
- `artifacts/task-execution-mode-20260920/server-race.log`
- `artifacts/task-execution-mode-20260920/server.log`
- `artifacts/upstream-a590b41-20260920/README.md`
- `artifacts/upstream-a590b41-20260920/agent.log`
- `artifacts/upstream-a590b41-20260920/noa-race.log`
- `artifacts/upstream-a590b41-20260920/norma.log`
- `artifacts/upstream-sync-20260916.md`
- `artifacts/upstream-sync-20260916/agent-fixed.log`
- `artifacts/upstream-sync-20260916/browser.log`
- `artifacts/upstream-sync-20260916/noa-race.log`
- `artifacts/upstream-sync-20260916/norma-fixed.log`
- `artifacts/upstream-sync-20260916/norma-local.log`
- `artifacts/upstream-sync-20260916/norma-upstream.log`
- `artifacts/upstream-sync-20260916/server-fixed.log`
- `artifacts/upstream-sync-20260919/README.md`
- `artifacts/upstream-sync-20260919/agent.log`
- `artifacts/upstream-sync-20260919/build.log`
- `artifacts/upstream-sync-20260919/db.log`
- `artifacts/upstream-sync-20260919/features.log`
- `artifacts/upstream-sync-20260919/local-browser.log`
- `artifacts/upstream-sync-20260919/server-final.log`
- `artifacts/upstream-sync-20260919/source-browser.log`
- `artifacts/upstream-sync-20260919/types.log`
- `artifacts/upstream-sync-20260919/web-tests.log`

**cmd/**（2）

- `cmd/series/main.go`
- `cmd/series/main_test.go`

**db/**（80）

- `db/activity_search.go`
- `db/activity_search_test.go`
- `db/activity_window.go`
- `db/activity_window_test.go`
- `db/asset_approval_groups.go`
- `db/asset_context.go`
- `db/asset_dsl_task_test.go`
- `db/asset_origins.go`
- `db/asset_origins_test.go`
- `db/asset_skips.go`
- `db/asset_skips_test.go`
- `db/asset_templates.go`
- `db/asset_templates.sql`
- `db/asset_templates_test.go`
- `db/broadcast.go`
- `db/broadcast_test.go`
- `db/config_session_header_test.go`
- `db/context_assets.go`
- `db/context_mixed.go`
- `db/context_mixed_test.go`
- `db/context_prompt_migration.go`
- `db/context_reads.go`
- `db/context_reads_test.go`
- `db/context_review_test.go`
- `db/derived_asset_approval_test.go`
- `db/derived_asset_migration_test.go`
- `db/derived_asset_parents.go`
- `db/derived_asset_parents_test.go`
- `db/description_assets.go`
- `db/description_assets_test.go`
- `db/execution_mode.go`
- `db/execution_mode_test.go`
- `db/finding_access.go`
- `db/finding_access_test.go`
- `db/finding_deletion_feedback.go`
- `db/finding_deletion_feedback_test.go`
- `db/finding_suppression.go`
- `db/findings_patch_test.go`
- `db/host_approval_test.go`
- `db/host_isolation_test.go`
- `db/lineage_batch_test.go`
- `db/main_execution.go`
- `db/main_execution_test.go`
- `db/manual_scope_test.go`
- `db/mcp_usage.go`
- `db/mcp_usage_test.go`
- `db/performance_indexes.go`
- `db/performance_indexes_test.go`
- `db/planner_asset_search_test.go`
- `db/planner_context.go`
- `db/planner_context_migration.go`
- `db/planner_context_migration_test.go`
- `db/planner_context_test.go`
- `db/prompt_contracts_test.go`
- `db/read_performance_test.go`
- `db/review_migration_test.go`
- `db/target_access_test.go`
- `db/task_asset_block_recreation_test.go`
- `db/task_asset_blocking.go`
- `db/task_asset_blocking_test.go`
- `db/task_asset_view.go`
- `db/task_asset_view_test.go`
- `db/task_company_assets.go`
- `db/task_company_assets_test.go`
- `db/task_findings_page.go`
- `db/task_findings_page_test.go`
- `db/task_notifications.go`
- `db/task_notifications_test.go`
- `db/tool_calls.go`
- `db/tool_calls_test.go`
- `db/tool_context_catalog.go`
- `db/tool_context_catalog_test.go`
- `db/tool_reads.go`
- `db/tool_reads_test.go`
- `db/tool_usage_backfill.sql`
- `db/worker_execution.go`
- `db/worker_feedback.go`
- `db/worker_feedback_test.go`
- `db/worker_queue.go`
- `db/worker_queue_test.go`

**docs/**（126）

- `docs/builtin-tools-query-contract.md`
- `docs/context-retrieval-window-20261019.md`
- `docs/http-entry-auth.md`
- `docs/main-pending-execution-verification-20260922.md`
- `docs/mcp-shell-review/mcp-desktop-light.png`
- `docs/mcp-shell-review/mcp-mobile-dark.png`
- `docs/mcp-shell-review/shell-test-dark.png`
- `docs/mcp-shell-verification-20260922.md`
- `docs/norma-official-20261018.md`
- `docs/planner-asset-search-20261019.md`
- `docs/proposals/agent-view-slimming.md`
- `docs/proposals/planner-fact-selection-v2.md`
- `docs/proposals/planner-fact-selection.md`
- `docs/proposals/tool-result-compaction.md`
- `docs/proposals/windows-skill-path.md`
- `docs/screenshots/findings-ui-20261019/global-dark.png`
- `docs/screenshots/findings-ui-20261019/global-light.png`
- `docs/screenshots/findings-ui-20261019/task-light.png`
- `docs/screenshots/findings-ui-20261019/task-mobile.png`
- `docs/screenshots/task-controls-20261019/desktop-dark.png`
- `docs/screenshots/task-controls-20261019/desktop-light.png`
- `docs/screenshots/task-controls-20261019/mobile-dark.png`
- `docs/screenshots/task-controls-20261019/settings.png`
- `docs/series-rename-20261009.md`
- `docs/session-search-verification-20260921.md`
- `docs/tool-contract-review-20260913.md`
- `docs/tool-execution-boundary-fixes-20260913.md`
- `docs/ui-neo/build-metrics.json`
- `docs/ui-neo/conversation-dark-source.png`
- `docs/ui-neo/conversation-source.png`
- `docs/ui-neo/login.png`
- `docs/ui-neo/new-task-desktop.png`
- `docs/ui-neo/new-task-mobile.png`
- `docs/ui-neo/request-metrics.json`
- `docs/ui-neo/settings-desktop.png`
- `docs/ui-neo/settings-mobile.png`
- `docs/ui-neo/setup.png`
- `docs/ui-neo/sidebar-without-search.png`
- `docs/ui-neo/tasks-desktop.png`
- `docs/ui-neo/verification.md`
- `docs/upstream-sync-20260913.md`
- `docs/upstream-sync-20260929.md`
- `docs/upstream-sync-20260930.md`
- `docs/upstream-sync-20261008.md`
- `docs/verification-20260929/asset-origin.png`
- `docs/verification-20260929/exploration-graph.png`
- `docs/verification-20260929/task-severity.png`
- `docs/verification/agent-view-slimming-20261022.md`
- `docs/verification/context-policy-20261026.md`
- `docs/verification/context-tools-20261018/context-get-editor.png`
- `docs/verification/context-tools-20261018/measurements.txt`
- `docs/verification/context-tools-20261018/results.json`
- `docs/verification/context-tools-20261018/review-compile-final.txt`
- `docs/verification/context-tools-20261018/review-final.txt`
- `docs/verification/context-tools-20261018/review-metrics.txt`
- `docs/verification/context-tools-20261018/review-norma.txt`
- `docs/verification/context-tools-20261018/review-server-final.txt`
- `docs/verification/context-tools-20261018/review-surface.txt`
- `docs/verification/context-tools-20261018/review.json`
- `docs/verification/context-tools-20261018/source-location.png`
- `docs/verification/context-tools-20261018/test-summary.txt`
- `docs/verification/export-unified-20261024.md`
- `docs/verification/finding-suppression-20261019/results.json`
- `docs/verification/finding-suppression-20261019/review-results.json`
- `docs/verification/finding-suppression-20261019/rules-active.png`
- `docs/verification/finding-suppression-20261019/rules-revoked.png`
- `docs/verification/host-adaptation-20261025.md`
- `docs/verification/planner-fact-selection-20261023.md`
- `docs/verification/planner-on-demand-20261023/agent-tests.txt`
- `docs/verification/planner-on-demand-20261023/compile.txt`
- `docs/verification/planner-on-demand-20261023/db-tests.txt`
- `docs/verification/planner-on-demand-20261023/guard-tests.txt`
- `docs/verification/planner-on-demand-20261023/jit-diagnostic.txt`
- `docs/verification/planner-on-demand-20261023/overview-metrics.txt`
- `docs/verification/planner-on-demand-20261023/planner-runs.txt`
- `docs/verification/planner-on-demand-20261023/race-tests.txt`
- `docs/verification/planner-on-demand-20261023/results.json`
- `docs/verification/planner-on-demand-20261023/scenario-after.txt`
- `docs/verification/planner-on-demand-20261023/scenario-before.txt`
- `docs/verification/planner-on-demand-20261023/server-focused.txt`
- `docs/verification/planner-on-demand-20261023/server-tests.txt`
- `docs/verification/planner-on-demand-20261023/vet.txt`
- `docs/verification/planner-overview-config-20261022.md`
- `docs/verification/planner-proactive-20261026/agent-full.txt`
- `docs/verification/planner-proactive-20261026/compile.txt`
- `docs/verification/planner-proactive-20261026/db-full.txt`
- `docs/verification/planner-proactive-20261026/examples.txt`
- `docs/verification/planner-proactive-20261026/guard-full.txt`
- `docs/verification/planner-proactive-20261026/metrics-after.txt`
- `docs/verification/planner-proactive-20261026/metrics-before.txt`
- `docs/verification/planner-proactive-20261026/race.txt`
- `docs/verification/planner-proactive-20261026/results.json`
- `docs/verification/planner-proactive-20261026/runs-before.txt`
- `docs/verification/planner-proactive-20261026/server-final.txt`
- `docs/verification/planner-proactive-20261026/server-full.txt`
- `docs/verification/planner-proactive-20261026/vet.txt`
- `docs/verification/planner-proactive-review-20261026/agent-full.txt`
- `docs/verification/planner-proactive-review-20261026/agent-race.txt`
- `docs/verification/planner-proactive-review-20261026/build.txt`
- `docs/verification/planner-proactive-review-20261026/compile.txt`
- `docs/verification/planner-proactive-review-20261026/db-full.txt`
- `docs/verification/planner-proactive-review-20261026/guard-full.txt`
- `docs/verification/planner-proactive-review-20261026/results.json`
- `docs/verification/planner-proactive-review-20261026/server-full.txt`
- `docs/verification/planner-proactive-review-20261026/server-race.txt`
- `docs/verification/planner-proactive-review-20261026/vet.txt`
- `docs/verification/report-length-budget-20261026.md`
- `docs/verification/report-template-20261023.md`
- `docs/verification/revert-05c8e90-prompt-20261030.md`
- `docs/verification/settlement-prefix-20261028.md`
- `docs/verification/upstream-335bd2e-20261029.md`
- `docs/verification/upstream-b76c684-20261030.md`
- `docs/verification/windows-skill-path-20261021.md`
- `docs/verification/worker-context-search-20261026/agent-full.txt`
- `docs/verification/worker-context-search-20261026/agent-race.txt`
- `docs/verification/worker-context-search-20261026/build.txt`
- `docs/verification/worker-context-search-20261026/db-full.txt`
- `docs/verification/worker-context-search-20261026/guard-full.txt`
- `docs/verification/worker-context-search-20261026/results.json`
- `docs/verification/worker-context-search-20261026/sdk-focused.txt`
- `docs/verification/worker-context-search-20261026/server-full-final.txt`
- `docs/verification/worker-context-search-20261026/server-full.txt`
- `docs/verification/worker-context-search-20261026/server-race.txt`
- `docs/verification/worker-context-search-20261026/vet.txt`
- `docs/worker-feedback-verification-20260925.md`
- `docs/worker-queue.md`

**guard/**（5）

- `guard/assets.go`
- `guard/assets_test.go`
- `guard/compress_test.go`
- `guard/proxy_scope_test.go`
- `guard/proxy_tag.go`

**mcphttp/**（1）

- `mcphttp/capture_test.go`

**report/**（2）

- `report/findings_test.go`
- `report/helpers.go`

**screenshots/**（5）

- `screenshots/series-http-auth-settings.jpg`
- `screenshots/series-login.jpg`
- `screenshots/series-session.jpg`
- `screenshots/series-settings-dark.jpg`
- `screenshots/series-tasks.jpg`

**server/**（76）

- `server/activity_search.go`
- `server/activity_search_test.go`
- `server/activity_stream_test.go`
- `server/activity_window_test.go`
- `server/asset_templates_test.go`
- `server/broadcast_nodes.go`
- `server/broadcast_nodes_test.go`
- `server/builtin_review_test.go`
- `server/chatupload_cleanup_test.go`
- `server/context_surface_metrics_test.go`
- `server/context_tools.go`
- `server/context_tools_test.go`
- `server/customtool_capture_test.go`
- `server/customtool_metadata_test.go`
- `server/customtool_windows_test.go`
- `server/execution_mode.go`
- `server/execution_mode_test.go`
- `server/finding_deletion_feedback_test.go`
- `server/finding_suppression.go`
- `server/finding_suppression_review_test.go`
- `server/finding_suppression_test.go`
- `server/findings_patch_test.go`
- `server/global_proxy.go`
- `server/global_proxy_test.go`
- `server/hint_trigger_test.go`
- `server/http_auth.go`
- `server/http_auth_settings.go`
- `server/http_auth_settings_test.go`
- `server/http_auth_test.go`
- `server/intent_stop_test.go`
- `server/intent_user_control_test.go`
- `server/main_execution_test.go`
- `server/mcp_runtime.go`
- `server/mcp_runtime_test.go`
- `server/mcp_usage.go`
- `server/mcp_usage_test.go`
- `server/mcpdiscover_test.go`
- `server/planner_asset_search_test.go`
- `server/planner_intent_search_test.go`
- `server/planner_on_demand_test.go`
- `server/planner_revision_test.go`
- `server/removed_update_routes_test.go`
- `server/retrieval_examples_test.go`
- `server/shell_execution.go`
- `server/shell_settings_test.go`
- `server/shell_test_api.go`
- `server/shell_test_api_test.go`
- `server/shell_usage.go`
- `server/shell_usage_test.go`
- `server/smartproxy.go`
- `server/task_asset_approval_types_test.go`
- `server/task_asset_blocking_test.go`
- `server/task_compactor_test.go`
- `server/task_findings_page.go`
- `server/task_findings_page_test.go`
- `server/task_notifications.go`
- `server/task_notifications_test.go`
- `server/task_pause_timeout_test.go`
- `server/task_severity_test.go`
- `server/tool_calls.go`
- `server/tool_calls_test.go`
- `server/tool_contract_test.go`
- `server/tool_documentation_test.go`
- `server/tool_execution_boundary_test.go`
- `server/tools_assembly_helpers_test.go`
- `server/tools_context_test.go`
- `server/webui_files.go`
- `server/webui_files_test.go`
- `server/worker_context_search_test.go`
- `server/worker_feedback.go`
- `server/worker_feedback_test.go`
- `server/worker_queue.go`
- `server/worker_queue_test.go`
- `server/worker_signal_test.go`
- `server/worker_trace_bindings.go`
- `server/worker_trace_bindings_test.go`

**third_party/**（1）

- `third_party/.DS_Store`

**traffic/**（12）

- `traffic/compress.go`
- `traffic/context_review_test.go`
- `traffic/disk_optimization_test.go`
- `traffic/smartproxy.go`
- `traffic/smartproxy_e2e_test.go`
- `traffic/smartproxy_routing_test.go`
- `traffic/smartproxy_test.go`
- `traffic/sync_regression_test.go`
- `traffic/tool_reads.go`
- `traffic/tool_reads_test.go`
- `traffic/tool_review_test.go`
- `traffic/worker_policy_test.go`

**web/**（79）

- `web/public/series-icon.png`
- `web/public/series-icon.svg`
- `web/scripts/test-register.cjs`
- `web/src/app/(main)/_components/new-task-shortcut.tsx`
- `web/src/app/(main)/function/tasks/detail/_tabs/asset-approvals-tab.tsx`
- `web/src/app/(main)/function/tasks/detail/_tabs/smart-proxy-tab.tsx`
- `web/src/app/(main)/function/tasks/detail/_tabs/worker-queue.tsx`
- `web/src/app/(main)/function/tasks/detail/execution-control.tsx`
- `web/src/app/(main)/function/tasks/new/page.tsx`
- `web/src/app/(main)/system/agents/_components/agents-settings.tsx`
- `web/src/app/(main)/system/intercept/_components/intercept-settings.tsx`
- `web/src/app/(main)/system/llm/_components/llm-settings.tsx`
- `web/src/app/(main)/system/mcp/_components/mcp-settings.tsx`
- `web/src/app/(main)/system/settings/_components/general-settings.tsx`
- `web/src/app/(main)/system/settings/_components/http-auth-card.tsx`
- `web/src/app/(main)/system/skills/_components/skills-settings.tsx`
- `web/src/app/(main)/system/tools/_components/tools-settings.tsx`
- `web/src/components/asset-approval-template.tsx`
- `web/src/components/finding-delete-dialog.tsx`
- `web/src/components/finding-detail-content.tsx`
- `web/src/components/finding-pagination.tsx`
- `web/src/components/finding-presentation.tsx`
- `web/src/components/finding-selector.tsx`
- `web/src/components/finding-suppression-rules.tsx`
- `web/src/components/findings-toolbar.tsx`
- `web/src/components/session-search.tsx`
- `web/src/components/session-tool-calls.tsx`
- `web/src/components/settings/appearance-settings.tsx`
- `web/src/components/settings/settings-context.tsx`
- `web/src/components/settings/settings-workspace.tsx`
- `web/src/components/source-transcript.tsx`
- `web/src/components/task-notifications.tsx`
- `web/src/components/tasks/task-creation-form.tsx`
- `web/src/lib/activity-highlight.ts`
- `web/src/lib/exploration-graph.test.mjs`
- `web/src/lib/exploration-graph.ts`
- `web/src/lib/findings.test.ts`
- `web/src/lib/findings.ts`
- `web/src/lib/http-auth.test.ts`
- `web/src/lib/http-auth.ts`
- `web/src/lib/mcp-import.test.ts`
- `web/src/lib/mcp-import.ts`
- `web/src/lib/mock/activity-search.test.ts`
- `web/src/lib/mock/approval-execution.test.ts`
- `web/src/lib/mock/asset-approval-templates.test.ts`
- `web/src/lib/mock/asset-origins.test.ts`
- `web/src/lib/mock/broadcast.test.ts`
- `web/src/lib/mock/context-tools.json`
- `web/src/lib/mock/context-tools.test.ts`
- `web/src/lib/mock/execution-mode.test.ts`
- `web/src/lib/mock/finding-deletion.test.ts`
- `web/src/lib/mock/finding-export.ts`
- `web/src/lib/mock/finding-filters.test.ts`
- `web/src/lib/mock/finding-suppression.test.ts`
- `web/src/lib/mock/finding-traffic.ts`
- `web/src/lib/mock/http-auth.test.ts`
- `web/src/lib/mock/shell-tools.test.ts`
- `web/src/lib/mock/task-notifications.ts`
- `web/src/lib/mock/task-severity.test.ts`
- `web/src/lib/mock/tool-calls.ts`
- `web/src/lib/mock/traffic.test.ts`
- `web/src/lib/mock/traffic.ts`
- `web/src/lib/mock/upstream-sync.test.ts`
- `web/src/lib/mock/worker-cancel.test.ts`
- `web/src/lib/mock/worker-queue.test.ts`
- `web/src/lib/mock/workspace.test.ts`
- `web/src/lib/polling.test.ts`
- `web/src/lib/polling.ts`
- `web/src/lib/side-questions.test.ts`
- `web/src/lib/task-notifications.test.ts`
- `web/src/lib/task-notifications.ts`
- `web/src/lib/tool-calls.test.ts`
- `web/src/lib/tool-calls.ts`
- `web/src/lib/worker-feedback.test.mjs`
- `web/src/lib/worker-feedback.ts`
- `web/src/lib/workspace-navigation.test.ts`
- `web/src/lib/workspace-navigation.ts`
- `web/src/styles/control-radius.css`
- `web/tests/asset-source-retry.cjs`

</details>

## 二、artex 原有、series 移除的文件 & 能力清单

共 94 个文件。


**(根目录)/**（8）

- `.dockerignore`
- `.env.example`
- `.gitattributes`
- `LICENSE`
- `PORTING-AUDIT.md`
- `PORTING-CODE-COMPARE.md`
- `PORTING-RULES.md`
- `PORTING.md`

**.github/**（1）

- `.github/workflows/release.yml`

**agent/**（3）

- `agent/bash_prompt_test.go`
- `agent/context_prompt_migration_test.go`
- `agent/settlement_gate.go`

**cmd/**（2）

- `cmd/artex/main.go`
- `cmd/artex/main_test.go`

**db/**（8）

- `db/asset_intercept.go`
- `db/asset_intercept_match.go`
- `db/asset_intercept_match_test.go`
- `db/intent_delete_test.go`
- `db/notification.go`
- `db/notification_delivery.go`
- `db/notification_test.go`
- `db/task_intercept.go`

**notify/**（26）

- `notify/channel.go`
- `notify/channels_test.go`
- `notify/dingtalk.go`
- `notify/email.go`
- `notify/email_test.go`
- `notify/event.go`
- `notify/feishu.go`
- `notify/filter.go`
- `notify/filter_test.go`
- `notify/html.go`
- `notify/http.go`
- `notify/http_test.go`
- `notify/markdown.go`
- `notify/mask.go`
- `notify/mask_test.go`
- `notify/notify.go`
- `notify/pack_test.go`
- `notify/redact_test.go`
- `notify/render.go`
- `notify/render_test.go`
- `notify/sign_test.go`
- `notify/ssrf_test.go`
- `notify/telegram.go`
- `notify/webhook.go`
- `notify/webhook_template_test.go`
- `notify/wecom.go`

**report/**（2）

- `report/findings_report_test.go`
- `report/report.go`

**screenshots/**（13）

- `screenshots/agents.png`
- `screenshots/assets.png`
- `screenshots/assets_test.png`
- `screenshots/chat.png`
- `screenshots/dashboard.png`
- `screenshots/findings.png`
- `screenshots/graph.png`
- `screenshots/intercept.png`
- `screenshots/llm.png`
- `screenshots/logs.png`
- `screenshots/sessions.png`
- `screenshots/tasks.png`
- `screenshots/traffic.png`

**selfupdate/**（5）

- `selfupdate/bootstrap.go`
- `selfupdate/github.go`
- `selfupdate/selfupdate.go`
- `selfupdate/selfupdate_test.go`
- `selfupdate/stage.go`

**server/**（10）

- `server/asset_intercept.go`
- `server/finding_report_zip_test.go`
- `server/notifier.go`
- `server/notify_api.go`
- `server/notify_api_test.go`
- `server/sync_scopesentry.go`
- `server/task_intercept.go`
- `server/update.go`
- `server/update_test.go`
- `server/workspace.go`

**skills/**（1）

- `skills/scopesentry/SKILL.md`

**web/**（15）

- `web/src/app/(main)/_components/sidebar/search-dialog.tsx`
- `web/src/app/(main)/_components/update-badge.tsx`
- `web/src/app/(main)/function/findings/_components/findings-table.tsx`
- `web/src/app/(main)/function/sync/page.tsx`
- `web/src/app/(main)/function/tasks/detail/_tabs/report-tab.tsx`
- `web/src/app/(main)/function/tasks/detail/_tabs/retests-tab.tsx`
- `web/src/app/(main)/function/workspace/page.tsx`
- `web/src/app/(main)/system/intercept/assets/page.tsx`
- `web/src/app/(main)/system/notify/_components/channel-fields.ts`
- `web/src/app/(main)/system/notify/_components/channel-form.tsx`
- `web/src/app/(main)/system/notify/_components/delivery-list.tsx`
- `web/src/app/(main)/system/notify/_components/stat-tile.tsx`
- `web/src/app/(main)/system/notify/page.tsx`
- `web/src/app/(main)/system/settings/_components/update-card.tsx`
- `web/src/components/asset-intercept-rules-editor.tsx`

### 能力层面的移除（来自 229 个差异文件的核对结论）

- **资产审批/授权体系**：ARTEX 有而 SERIES 缺失的那一半不成立——实际是 SERIES 有、ARTEX 无。相对 ARTEX 基线，SERIES 移除了：任务资产审批状态（approved/pending/revoked/blocked）、删除墓碑（task_asset_blocks）、按主机的授权继承与投影、审批接口与页签、撤回资产时取消对应 Worker、Worker 启动前 ValidateIntentAssets/ValidateIntentHosts、finding 流量的资产级读写校验。
- **执行模式与队列**：手工/托管模式（execution_mode）、规划者取消句柄（beginPlannerRound）、worker 唤醒通道、人工队列（worker_queue_manual/queue_position）、批量派发（dispatch_intents）。
- **上下文与态势配置**：planner overview full/lean 与数值配置、context policy 配置、BlackboardRevision 心跳跳过、任务域工具注入（TaskDomainToolsFrom）。
- **流量治理**：抓包静态资源过滤、保留期清扫、blob 压缩（zst）与解压、任务级代理凭据（TaskProxyAddr）、智能代理池（smart proxy）、curl/wget 默认 UA。
- **禁报与删除反馈**：finding_suppression_rules、finding_deletion_feedback、删除漏洞需填原因并通知 planner。
- **通知已读**：任务未读漏洞/审批提醒（task-notifications）。
- **平台能力**：MCP 测试/导入/用量统计、shell 类型自定义工具与其元数据、per-profile 全局 planner/worker 构建、会话工具调用列表、来源会话/队列视图。
- **其它**：公司范围手工 URL/证书保留与 65536 长度上限、侧问响应严格校验、任务归档暂停顺延超时、旧归档兼容填充。

## 五、单边文件语义核对结论（566 + 94）

完整逐文件记录见 `docs/CODE-COMPARE-NOTES.md`。要点：

### SERIES 独有能力的真实构成（566 个）
- **agent/ 105 个（14,036 行）**：资产上下文与授权（asset_context/target_access/structured_capture）、上下文工程（context_policy/context_tools/retrieval_history/tool_budget/tool_contract/tool_defaults）、态势投影（overview_config/lean/projection）、SDK 会话策略（sdk_session/settlement 相关）、派发与反馈（intent_dispatch/main_dispatch_feedback/finding_deletion_feedback）、智能代理工具、工具 UA。
- **db/ 80 个（11,600 行）**：资产审批/封禁/模板/来源/跳过、任务资产视图、finding 访问与禁报/删除反馈、worker 执行与反馈、队列与执行模式、上下文检索与工具读取、任务通知、MCP 用量、性能索引与迁移。
- **server/ 76 个（10,199 行）**：执行模式与队列、上下文/工具契约、资产审批与封禁 API、禁报服务、HTTP 入口认证、智能代理、shell 工具与计量、MCP 运行时与用量、任务通知。
- **web/ 79 个（14,370 行）**：设置工作区组件（agents/intercept/llm/mcp/skills/tools/general）、资产审批页签、智能代理页签、worker 队列、来源会话、会话工具调用、漏洞选择/分页/筛选/禁报、任务通知、任务创建表单、mock 与单测。
- **traffic/guard 等 25 个**：blob 压缩、智能代理端到端、traffic 读取工具、任务资产策略钩子、代理签名凭据。
- **docs 126 + artifacts 60 + 截图/杂项**：设计/方案/验证记录、性能瀑布与验证日志、SRC 报告（敏感，仅登记）。

### ARTEX 独有能力（94 个，series 不存在）
- **notify/ 26 个（5,055 行）**：六种渠道（钉钉/飞书/企微/Telegram/Webhook/邮件）+ 过滤、渲染、掩码、SSRF 防护、签名、整条打包、错误不泄凭据。
- **db/ 8 个 + server/ 10 个**：notification_channels/events/deliveries 与调度器/API、资产与任务拦截规则、ScopeSentry 同步、页面一键更新、工作区文件管理。
- **selfupdate/ 5 个（1,411 行）**：GitHub release 拉取、校验和、分阶段换装、冒烟确认与回滚、重启（退出码 75）。
- **agent/ 3 个**：settlement_gate（收尾门保持工具清单不变）、Bash 提示改写测试、reporter 提示迁移测试。
- **web/ 15 个**：通知推送页、资产同步页、工作空间页、资产拦截页、报告/复测页签、findings-table、更新徽章与卡片、搜索对话框。
- **根目录**：PORTING 四件套、LICENSE、.env.example、.github/release.yml、ScopeSentry 技能、README 截图 13 张。

## 三、修改文件汇总（230 个实质逻辑差异）


### .gitignore/

- `.gitignore`（+52/-22）— **低风险**：改为忽略 data/*.db/*.db-wal/*.db-shm/config.json/scheduler_state.json/__pycache__/web/node_modules；移除 /dist/ 与 /server/webui/dist/ 忽略。

### CHANGELOG.md/

- `CHANGELOG.md`（+282/-73）— **低风险**：品牌改写并加入 0.3.15 条目（拦截、漏洞推送、安全审计、测试基础设施、流量）。

### Dockerfile/

- `Dockerfile`（+12/-13）— **低风险**：已读完整 diff。

### README.md/

- `README.md`（+116/-74）— **低风险**：标题/截图/在线 Demo 链接改为 ARTEX；删除旧品牌截图说明。

### agent/

- `agent/assembly.go`（+10/-59）— **低风险**：SERIES 有 ContextReads/AssetGuidance 字段与 WithTaskToolSet/TaskDomainToolsFrom，ARTEX 无；SERIES AugmentTools 返回 error 并对待注入工具去重，ARTEX 不返回 error 且直接 append（重名可重复）；ARTEX 末尾多 applyHostPromptOverrides（本次移植）。
- `agent/blackboard_inheritance_test.go`（+49/-81）— **低风险**：SERIES 有“派生节点继承资产授权”测试（依赖审批）、graph_overview 省略 scope；ARTEX 改为“展开公司范围”断言。GetWorkerTrace 超限行为不同：SERIES 报错，ARTEX 返回前 5 条 + omitted_step_ids + notice。
- `agent/cancelcause.go`（+4/-11）— **低风险**：SERIES 有 AssetAuthorizationChangedCause、AbortTaskFinishedByUser；ARTEX 无。ARTEX 新增 AbortCancelledByUser（本次移植）。AbortWorkCancelledByUser 语义不同：SERIES=取消保留意图，ARTEX=删除意图（软删/硬删）。
- `agent/capture.go`（+12/-17）— **低风险**：导入路径改名；SERIES captureRun 接受 settleGuard 并走 newSession，ARTEX 直接 agentcore.NewSession；SERIES 用 assetContextCall 归一化工具调用，ARTEX 直接用 ev.ToolUse。
- `agent/chat.go`（+13/-41）— **低风险**：SERIES 有 ctxPolicyFn（自定义上下文清理/摘要线）、taskAssets/taskID（任务资产作用域）、WithTaskPolicy、SetContextPolicy；ARTEX 无。ARTEX 的 AugmentTools 不返回 error；RunInfo 只带 SessionID（SERIES 带 AgentKey/TaskID/Trigger）。Settlement 调用改为 wrapupSettlement(agentKey, nil)。ARTEX 无 AssetPolicyHooksWithGuard / withAssetContext / applyCont
- `agent/compaction.go`（+24/-72）— **低风险**：SERIES 有 assets 与 blockAuthorized（压缩块要过资产授权）、compactorRunContext（独立 RunInfo 归因）；ARTEX 无。foldBlock 语义不同：SERIES 在源签名/内容变化时整体丢弃重试；ARTEX 只剔除复活成员，剩余不足 K 才跳过，且用原 nodeByID/cvers 生成摘要。ARTEX 补了 exp<ID>-compactor 的 session id。compress 系统提示 SERIES 包 withChineseLanguage，ARTEX 未包。
- `agent/context_prompt_migration.go`（+3/-11）— **低风险**：仅保留 reporter 旧提示词迁移，删除 PreviousPlannerProactivePromptHashes，补注释。
- `agent/deferred.go`（+8/-68）— **低风险**：SERIES 追加 toolAvailabilityRule 与 contextReadRule（依据 ContextReads）、AssetGuidance，并返回 boundary=1；ARTEX 全部没有，无块时 boundary=0。SERIES 有 plannerAssetGuidance 函数（按最终工具集生成资产检索指引），ARTEX 无。seedUnlockFromHistory：SERIES 只回放“调用+成功结果”配对的 Skill，ARTEX 只要看到 Skill 调用就解锁（未配对结果）。
- `agent/deferred_test.go`（+5/-20）— **低风险**：SERIES 断言无全局块时 system 仍带 toolAvailabilityRule 且 boundary=1；ARTEX 断言纯文本 boundary=0。SERIES 有 TestSeedUnlockRequiresSuccessfulResult（只有成功结果才解锁），ARTEX 删掉且实现也放宽。
- `agent/finding_recorder_test.go`（+4/-4）— **低风险**：SERIES 的 notifyFinding 带可变 assetIDs 参数；ARTEX 只有 (int64, string)。
- `agent/finding_workflow.go`（+5/-10）— **低风险**：SERIES 的编号说明指向 context_search/context_get；ARTEX 改成 list_findings/list_task_findings/node_detail/get_task_node_detail（无上下文工具）。SERIES 在“未启用流量绑定”时直接返回“不要绑定”提示；ARTEX 删掉该分支。SERIES 的 guidance 里涉及 context_get/context_search 的取 ID 路径，ARTEX 换成任务详情/漏洞列表工具。
- `agent/finding_workflow_test.go`（+5/-14）— **低风险**：SERIES ToolResolve 返回 error、断言禁用绑定时必须出现“当前未启用流量证据绑定”；ARTEX 无 error，且删掉该断言（实现也不再输出该提示）。
- `agent/goals.go`（+32/-6）— **低风险**：SERIES 用 register_user_target + registerDescriptionAsset 登记资产范围；ARTEX 改为 add_task_scope + goalsScopeTail。ARTEX 不再设置 RunInfo（TaskID/AgentKey/Trigger），只补 transcript session id。
- `agent/host_lessons.go`（+22/-106）— **低风险**：SERIES 用 2 秒去抖缓冲批量落盘（hostLessonBuf/hostLessonFlushDelay/FlushHostLessons/mergeHostLessonLists）；ARTEX 每次错误直接读盘-合并-写盘。ARTEX 无 HostLessonNote 的 pending 覆盖。
- `agent/host_lessons_test.go`（+0/-29）— **低风险**：SERIES 有去抖不落盘测试与 asset_search 描述携带本机经验测试；ARTEX 删除，改为直接落盘（无 FlushHostLessons）。
- `agent/host_profile.go`（+27/-20）— **低风险**：SERIES Windows 返回空串（整段清空 SDK 提示）；ARTEX 返回改写文本 windowsBashPrompt，保留“先写脚本/避免交互输入/超时”等建议（用户已确认保留）。hostClipRunes 位置调整。
- `agent/host_profile_test.go`（+0/-47）— **低风险**：SERIES 断言 Windows 下 Bash 提示必须为空串；ARTEX 删除这些断言（改写版由 bash_prompt_test.go 覆盖）。
- `agent/intent_stop.go`（+6/-64）— **低风险**：SERIES 是完整工具实现（cancelIntentTool/killWorkTool/intentStopTool，含角色校验、node 校验、planner 资产授权校验）；ARTEX 只剩 IntentStopResult/IntentStopper 类型与 SetCancelIntent，工具改在 server/orchestration.go 内联实现。
- `agent/mainagent.go`（+24/-77）— **低风险**：SERIES 有 ctxPolicyFn、smart（智能代理）、SetContextPolicy/SetSmartProxy；ARTEX 无。提示词差异大：SERIES 用 context_search/context_get/asset_search，ARTEX 用 list_findings/list_facts/list_assets/get_worker_output；SERIES 有执行模式（手工/托管）与 cancel_intent 指引，ARTEX 无。Chat 签名少 guard/notifyNodes；ARTEX 无 withAssetContext/dispatchRec
- `agent/memory_tools.go`（+25/-167）— **低风险**：SERIES 暴露 5 个工具（含 RecallMemory/RecordMemory，支持 pinned 常驻记忆）；ARTEX 只暴露 4 个（List/Read/Search/SaveMemory），无 pinned 字段与常驻注入，读取上限用常量 30000 而非 readOutputBudget(ctx)。
- `agent/memory_tools_test.go`（+6/-56）— **低风险**：SERIES 断言 5 个记忆工具与 pinned 常驻注入；ARTEX 断言 4 个（List/Read/Search/SaveMemory），删除 pinned 相关用例。
- `agent/noa.go`（+0/-26）— **低风险**：SERIES 在启用前检查归档目录可写（MkdirAll+临时文件探测），失败回退内置压缩，并把 Compaction.ContextWindow 传给适配器；ARTEX 无这两项检查。
- `agent/planner.go`（+164/-264）— **低风险**：SERIES 有 ctxPolicyFn、overviewProfileFn/overviewConfigFn（完整/精简态势）、smart（智能代理）、KillWork 为 IntentStopper、RunInfo 带 AgentKey；ARTEX 全无，KillWork 退化为 func(intentID) error。SERIES 的态势是“用户取消清单 + plannerBootstrap(强制 context_search) + 有界/精简 graph_overview（预算裁剪）”；ARTEX 是 renderTriggers + 完整 graph_overview JSON（无
- `agent/prompt.go`（+1/-18）— **低风险**：SERIES 强制给所有渲染后的系统提示追加“统一语言要求（简体中文）”，renderTmpl 对无模板变量文本直接返回；ARTEX 无语言策略，且总是走 template.Parse。
- `agent/prompt_test.go`（+4/-38）— **低风险**：SERIES 断言中文语言策略幂等且所有 agent 提示都追加；ARTEX 删除该策略及断言。
- `agent/promptcatalog.go`（+11/-11）— **低风险**：工具名随 ARTEX 工具面改写（context_search/graph_overview → list_findings/get_task_graph/get_task_worker_trace 等）；pentest 的资产检索/漏洞回看改 list_assets/list_findings；reporter 的报告步骤改成 get_task_node_detail/get_task_worker_trace；update_finding_report 参数名 finding_node_id→finding_id。
- `agent/proxyenv_test.go`（+0/-18）— **低风险**：SERIES 有 TaskProxyAddr 测试（带 series-task-<id> 用户名与 64 位密码，直连代理不改写）；ARTEX 无该函数故无测试。
- `agent/retester.go`（+1/-1）— **低风险**：复测提示词去掉分段续读说明，改为一次性读取关联漏洞/证据/约束。
- `agent/runinfo.go`（+14/-5）— **低风险**：SERIES 直接复用 llmrec.RunInfo（含 Trigger/Phase/RetryOrdinal 等归因字段）；ARTEX 自定义精简 RunInfo（TaskID/ExplorationID/IntentID/SessionID），无 Trigger/Phase。
- `agent/settlement_gate_test.go`（+78/-138）— **低风险**：SERIES 通过 dispatchHooks 测收尾门（覆盖：提示后拦截、落库放行、持续闭锁、PromptByReason、追加后缀、相似文本不误判、非 user 尾部不触发、无 guard/settlement 不拦截、真实 SDK 循环工具面不变）；ARTEX 用 NewSettlementGate 测（闭锁、前缀匹配、PromptByReason、惰性、WrapDeps 观察）。ARTEX 缺相似文本误判、非 user 尾部、工具面稳定性的对照测试。
- `agent/terminalreason.go`（+1/-1）— **低风险**：模型错误说明从 context_search 改为 get_worker_trace。
- `agent/toolcatalog.go`（+96/-29）— **低风险**：SERIES 的 auto/pentest 绑定 context_search/context_get/graph_overview/asset_search 等；ARTEX 改为 list_assets/list_findings 等。SERIES 有 defaultUnbound 空表、ToolResolve 返回 error、DecorateTool 注入 factWriteContract、PrepareInput 做默认值注入+业务校验（injectDefaults 带 missingOnly 参数）；ARTEX ToolResolve 无 error、无校验、injectDefau
- `agent/tools.go`（+568/-1608）— **低风险**：（完整读完 2176 行 diff）：
- `agent/tools_digest.go`（+72/-218）— **低风险**：SERIES 的 cold digests 与 expand_digest 全部经过 digestAuthorizationBatch（任一成员未授权则整块隐藏，成员分页 ToolDigestMemberPage，正文 textWindow 偏移续读，响应超预算自动二分截断）；ARTEX 全无授权过滤、无分页/偏移/预算裁剪，直接返回全部成员与正文。
- `agent/tools_insert.go`（+174/-312）— **低风险**：SERIES 通过 RegisterAgentAssetWithOrigin 作用域写入，写后查询 TaskAssetApprovalStates，按 pending/blocked/revoked 统计并拒绝；有 related 参数控制是否纳入覆盖度；有 originWorker/toolUseID 归因。ARTEX 直接 Upsert*，只做资产拦截规则（EvaluateAssetGate）与 SetTaskAssetSource，无审批状态、无 related、无归因字段。
- `agent/worker.go`（+73/-229）— **低风险**：SERIES 有 ErrWorkerAssetAuthorization、trafficRecording 开关、smart、ctxPolicyFn、IntentAssetIDs（含血缘）与 ValidateIntentAssets（启动前资产授权校验，失败零 token 退出）、TaskProxyAddr（带签名的任务代理凭据）、currentToolUAEnv、NO_PROXY 清空；ARTEX 全无，代理直连 m.proxyAddr。worker 提示差异大：SERIES 保留资产审批/待审批语义、流量工具说明与“最终总结返回下发者”；ARTEX 简化。notifyFinding 少 a
- `agent/worker_intervention_test.go`（+0/-24）— **低风险**：SERIES 有“品牌变更后投递身份仍匹配”的标记解析测试（解析任意 *_WORKER_CHAT 注释，不依赖 SERIES 前缀）；ARTEX 只做精确前缀匹配，测试也删除。
- `agent/wrapup.go`（+4/-33）— **低风险**：SERIES 签名无 disabledTools（注释明确要求传 nil，避免隐藏工具导致整段缓存重算），收尾拦截靠 newSession 的 settleGuard；ARTEX 保留 disabledTools 参数，本次移植调用处传 nil，并新增 SettlementGate 在请求观察后拒绝 Bash。

### db/

- `db/activity_page_test.go`（+3/-6）— **低风险**：ARTEX 用固定 summary "i" 造 25 条意图；SERIES 用带序号的 summary 并导入 fmt。属测试写法差异。
- `db/asset_dsl.go`（+10/-118）— **低风险**：SERIES 支持 approval_state（含 status==approved 别名）并要求任务上下文；有 CountDSLByTaskApproval/QueryDSLByTaskApproval/buildTaskDSLWhere（按任务关联与审批过滤、workerRead 投影）；ARTEX 无 approval_state、无任务级审批查询，DSL 只支持 =/==/!=/> 与裸词，不支持 contains/=~ 与运算符两侧空格；查询直接用 s.db 而 SERIES 用 queryRow/selectAssetColumns。
- `db/assets.go`（+31/-390）— **低风险**：SERIES 有 WithReadContext（请求级读上下文）、WithToolReadFields（模型侧只取必要列）、selectAssetColumns/query/queryRow 抽象、enableAgentDiscoveryMode（事务内把新关联标 pending）、taskAssetContextAssociationSQL/taskAssetSelectedLinkBoolSQL/taskAssetBlockedSQL（任务关联/审批/墓碑的规范 SQL）、QueryByTaskTested/QueryByTaskApproval；ARTEX 全无。ARTEX 的 lin
- `db/assets_test.go`（+4/-25）— **低风险**：删除任务审批相关断言（QueryByTask 系列）。
- `db/chat_mentions.go`（+24/-46）— **低风险**：SERIES 分“精确 ID 命中（主键查询）”与“模糊候选（trigram/index）”两路 UNION，带 5s 上下文超时、exact 游标标记，模糊查询限定各表独立 ILIKE 列；ARTEX 单条 UNION 全表扫描式写法（id::text 比较、concat_ws ILIKE），无超时，exact 用 fmt.Sprint(id)==query 推断。
- `db/companies.go`（+4/-4）— **低风险**：SERIES 范围原始/单值上限 65536 rune；ARTEX 为 1024。ON CONFLICT 从 md5(value) 改为 value（icp/keyword 去重键不同）。注释提及 scopesentry sync。
- `db/company_scope.go`（+0/-19）— **低风险**：SERIES 的 ScopeInput 有 Manual 字段，手工录入时 CIDR/URL/证书原样保留（不强制降级为 keyword）；ARTEX 无 Manual 分支。
- `db/company_scope_test.go`（+22/-13）— **低风险**：ARTEX 改用 t.Cleanup 注册 Close（注释解释 LIFO 顺序，避免清理落在已关闭连接上）并不吞清理错误；SERIES 是 defer d.Close() + `_,_=`。ICP 值从动态拆分改为固定串；IPv6 用例从随机 IPv6 改为固定 IPv4/24。无产品逻辑差异。
- `db/config.go`（+4/-99）— **低风险**：SERIES 有 ProfileByIDContext、保存时 TrimSpace(SessionHeaderKey)、MCP 使用统计（calls/tasks/usage_agents/last_used，MCPToolUsageStats/MCPServerUsageStats）、ImportMCP 批量导入（按 name upsert、清缓存）；ARTEX 全无（MCPServer 只多 Insecure 字段，MCPToolsDetailed 不再汇总调用次数）。
- `db/config_test.go`（+7/-12）— **低风险**：SERIES 用时间戳唯一 profile 名；ARTEX 固定名（并发/残留时可能撞名）。
- `db/constants.go`（+5/-0）— **低风险**：ARTEX 新增 StateIntentDeleted="deleted"（软删意图状态），SERIES 无此常量（其删除走 cancel_reason/state 组合）。
- `db/customtool_test.go`（+0/-66）— **低风险**：删除 TestCustomToolShellMetadata 与 TestShellMetadataNormalization（对应 shell 元数据能力移除）。
- `db/db.go`（+67/-74）— **低风险**：SERIES 额外 embed asset_templates.sql、连接池 20/10、全新库检测后写入 traffic_retention_days(30/0)、ensurePerformanceIndexes、backfillToolUsage、把 browser MCP 改名 playwright 并 seed 3 个内置 MCP（playwright/web_search/chrome-devtools）、清理 ScopeSentry 占位；ARTEX 连接池 32/32，只 seed 一个 browser MCP，新增 seedDefaultAssetInterceptRules
- `db/exploration.go`（+207/-644）— **低风险**：：
- `db/exploration_sources_test.go`（+0/-22）— **低风险**：删除“继承会话需终态意图”的 ActivitySearch/ActivityWindow/重开后拒绝读取断言（对应 ActivitySessionFilter.Inherited 移除）。
- `db/exploration_test.go`（+0/-222）— **低风险**：删除 IntentFingerprint 归一化、AddIntent 只对活跃前沿去重、BlackboardRevision 随图/资产上下文变化四个用例。
- `db/finding_retests.go`（+10/-1）— **低风险**：ARTEX 复测判 fixed 时改走 SetFindingStatusTx（带状态变更通知，注释说明此前裸 UPDATE 不发推送）；SERIES 仍是裸 UPDATE findings SET status。
- `db/finding_traffic.go`（+21/-32）— **低风险**：SERIES 的 RecordFindingInput 带 AccessRole/Revalidate/SuppressionCheck/SuppressionTargets，RecordFindingTx 内做 LockFindingRecordTx、ValidateFindingSuppressionTx、ValidateFindingAccessTx，并写禁报校验审计；ARTEX 去掉这些，改 LockTaskEvidenceTx，并在同事务内写 RecordNotificationEventTx(notify.EventFindingCreated) 推送事件。RecordFindin
- `db/findings.go`（+33/-117）— **低风险**：SERIES 有 FindingFilter.SummaryOnly（列表省掉 evidence）、导出含 report_detail、FindingAccessMetadata、DeleteFinding(id, reason) 写 finding_deletion_feedback + finding_suppression_rules、PatchFinding 原子改多字段（含图投影 content_version）；ARTEX 无 SummaryOnly、导出不含 report_detail、删除不写反馈/禁报规则、改为 setFindingCol 单列更新 + SetFindingS
- `db/intent_admission.go`（+3/-2）— **低风险**：DiscardOpenIntent 从 beginQueue 改为 s.db.Begin，并新增 cancelled_by_user 过滤（不删已取消意图）。
- `db/intent_user_control.go`（+44/-59）— **低风险**：SERIES 有 UserCancelledIntents、ReopenIntentByUser(id, expected)、RestoreUserReopen、ReopenBlockedIntentsByUser、ClaimPausedIntentByUser（均走队列锁并校验资产授权/任务状态）；ARTEX 只有 Node.UserCancelled 与 StopIntentWithReason（open/paused → stopped + cancelled_by_user/cancel_reason/cancelled_at + 清理副会话），没有“用户重新开启”的受控入口。
- `db/intent_user_control_test.go`（+34/-123）— **低风险**：SERIES 的 TestUserCancellationLifecycle 覆盖：取消写锁、重复取消幂等、各种 SetIntentState/SetNodeState/CAS/Reopen/Claim/Reset/ReopenBlocked 均不能复活；ARTEX 换成 TestStopIntentWithReasonLocksClaim（需 ARTEX_PG_DSN，验证取消后 ClaimIntent=false、重开清锁）。
- `db/intercept.go`（+6/-14）— **低风险**：SERIES 全部走 Context 版本并带 10s 超时，ListPendingInterceptsContext 支持按 task 过滤；ARTEX 只有非 Context 查询，不接受 task 过滤。
- `db/intercept_seed_test.go`（+0/-19）— **低风险**：删除 TestDeleteEndpointSeedIsIdempotent（内置拦截规则 seed 幂等）。
- `db/llm_usage.go`（+68/-51）— **低风险**：SERIES 的 LLMUsage 记录 intent_id、trigger、retry_ordinal、phase、system/message/tool_schema/request 字符数与估算输入 token，并有 idx_llm_usage_intent；ARTEX 精简为 task/exp/worker/model/latency/tokens，新增 JudgeUsageStats（worker='judge' 的累计与按日序列）。
- `db/llmretry_test.go`（+3/-3）— **低风险**：SERIES 用 t.Cleanup(func(){d.Close()})；ARTEX 用 defer d.Close()（同上关闭顺序问题）。
- `db/schema.sql`（+175/-506）— **低风险**：已读完整 diff。
- `db/settings.go`（+0/-20）— **低风险**：SERIES 有 SetSettingsAtomic（事务写入，失败全部回滚）；ARTEX 只有逐键 SetSetting。
- `db/settings_insert_test.go`（+0/-31）— **低风险**：删除 TestInsertSettingIfAbsentConcurrent（并发首写竞争）。
- `db/task_archive_aggregate_stats.go`（+0/-1）— **低风险**：单行差异（见 diff：无实质逻辑）。
- `db/task_archives.go`（+8/-64）— **低风险**：删除 TaskArchiveBlocker/TaskArchiveBlockerContext（按源任务查活跃直接依赖）、remainingPausedTimeout（暂停时长顺延超时）、归档快照中与审批/队列相关字段的处理；其余为字段删减。
- `db/task_archives_restore.go`（+11/-181）— **低风险**：删掉旧档案兼容填充（exploration_nodes.queue_position、task_asset_skips.observer_scopes、task_asset_blocks.block_kind、task_asset_links 的 tested/approval 字段）与 mcp_usage 序列推进；恢复时直接 json_populate_recordset。
- `db/task_archives_test.go`（+3/-132）— **低风险**：删除 TestRemainingPausedTimeoutUsesPauseInstant 等归档/暂停用例。
- `db/task_assets.go`（+33/-1002）— **低风险**：ARTEX 删除 23 个函数、无新增，全部属于资产审批/授权体系：ApprovalApproved/Pending/Revoked/Blocked 常量、TaskAssetApproval 结构、AssetKey/normalizeTaskAssetHost、ValidateTaskAssetsApproved、TaskAssetApprovalStates、FilterApprovedAssets、projectApprovedAssetHosts、TaskHostApprovalStates/ValidateTaskHostsApproved、RegisterAgentDiscovere
- `db/task_assets_context.go`（+4/-40）— **低风险**：SERIES 有 CountHostsByTaskWithSources 与 taskHostContextCTE（含审批过滤/来源展开）；ARTEX 只保留 HostsByTaskWithSources，CTE 精简。
- `db/task_assets_test.go`（+10/-624）— **低风险**：删除 10 个审批相关用例（审批继承与墓碑、Agent 发现保持 pending、AssetKey 稳定性、重发现可重置关联、取消读上下文、分页、继承审批源选择、创建时直接资产独立测试态、缺资产回滚）。
- `db/task_context.go`（+4/-4）— **低风险**：SetIntentBlockedReason / ReopenIntentsByBlockedReason 从 queueExec 改为 s.db.Exec，并去掉 cancelled_by_user 过滤（ARTEX 可把已取消意图重新置 blocked/open）。
- `db/task_list_performance_test.go`（+0/-41）— **低风险**：删除 TestTaskListSeverityCountsAndDeletion。
- `db/task_scope.go`（+2/-123）— **低风险**：SERIES 手工（source=manual）登记范围时会在同一事务内写入 task_asset_grants、把命中资产挂到任务并置 approved、清除同名 task_asset_blocks 墓碑（含全局资产已删除的 host_key 场景）；ARTEX 只 upsert task_scope。覆盖度 SQL：SERIES 的 tested 统计同时含 task_asset_links.tested 与 fact/finding 锚点，且都过 task_asset_effectively_approved；ARTEX 只统计 fact 锚点、无审批过滤。
- `db/task_scope_test.go`（+1/-57）— **低风险**：删除 TestManualCompanyScopeApprovesExistingAssetsButKeywordDoesNot（手工企业范围应授权既有资产）。
- `db/task_templates.go`（+54/-9）— **低风险**：ARTEX 新增 CategoryID 与 InterceptRules（模板保存任务级拦截/允许规则快照），TaskTemplatePatch 用 Set 标志区分“未传”与“清空”；SERIES 模板只有 name/description/goal。
- `db/tasks.go`（+13/-149）— **低风险**：ARTEX 去掉 Task.ExecutionMode/AssetApprovalTemplate、TaskCreateOptions 的 AssetIDs 与 ApprovalTemplate 校验、insertTaskAssets（创建时直接挂资产并派生保守范围+授权）、seedUserAssetGrants；新增 TaskCreateOptions.InterceptRules 并在同事务写 task_intercept_rules。
- `db/tasks_test.go`（+0/-6）— **低风险**：删除“用户提供资产必须 approved 且可执行”的断言。
- `db/tool_usage.go`（+1/-19）— **低风险**：SERIES 内嵌 tool_usage_backfill.sql 做历史回填，并在汇总时用 CanonicalToolUsageKey 合并别名；ARTEX 无回填、无别名归一。
- `db/tool_usage_test.go`（+1/-72）— **低风险**：删除 TestToolUsageBackfill（历史回填）用例，只保留 ledger 用例。
- `db/tools.go`（+14/-82）— **低风险**：SERIES 的 Tool 有 Executable/Directory/UsageHelp/WhenToUse 与 shell 类型校验（ValidateShellCommand/ShellCommand/shellMetadata/shellExecutable），UpdateTool 在事务内取 advisory lock 并同步 context binding grants；ARTEX 删掉 shell 相关字段与校验，UpdateTool 单条 UPDATE，无授权同步。

### docker-compose.yml/

- `docker-compose.yml`（+10/-13）— **低风险**：已读完整 diff。

### evidence/

- `evidence/store.go`（+4/-7）— **低风险**：SERIES 在 Record 事务开头做 ValidateFindingSuppressionTx（禁报规则锁），把资产/任务锁推迟到 RecordFindingTx；ARTEX 改为 LockTaskEvidenceTx，并给 RecordFindingTx 传 ctx（用于推送事件写入）。测试为模块路径与调用签名调整。
- `evidence/store_test.go`（+6/-6）— **低风险**：模块路径、SERIES_PG_DSN→ARTEX_PG_DSN、SetFindingReportVersionByNodeID 去掉 detail 参数。

### go.mod/

- `go.mod`（+5/-5）— **低风险**：已读完整 diff。

### go.sum/

- `go.sum`（+9/-10）— **低风险**：已读完整 diff。

### install.sh/

- `install.sh`（+16/-16）— **低风险**：已读完整 diff。

### intercept/

- `intercept/prompt.go`（+6/-7）— **低风险**：EffectiveJudgePrompt 的默认提示文本有调整（未改逻辑分支）。

### llmrec/

- `llmrec/llmrec.go`（+33/-124）— **低风险**：SERIES 有 RunInfo（trigger/retry/phase 等归因）与 WithRunInfo/WithIndependentRunInfo/RunInfoFrom，recordUsage 记录这些维度；ARTEX 删除 RunInfo，新增 WithWorker/workerFrom（显式 agent 泳道覆盖，用于 judge 计量）并在 recordUsage 中去掉 runInfo 参数；测试删除归因合并与请求维度用例。
- `llmrec/llmrec_test.go`（+3/-27）— **低风险**：删除 RunInfo 合并与请求维度（system/message/tool chars）用例。

### mcphttp/

- `mcphttp/client.go`（+2/-8）— **低风险**：初始化/关闭处少量清理（有 4 行删除与 1 行调整），无协议行为改动。

### report/

- `report/findings.go`（+43/-36）— **低风险**：ARTEX 保留 FindingFilename/FindingReportTitle/FindingReportFilename/SingleFindingReportMarkdown/FindingsReportMarkdown（本次移植新增）；CSV 列不同：SERIES 导出「漏洞名称/等级/类型/状态/任务/时间/提交报告」，ARTEX 导出「名称/类别/严重等级/状态/任务/时间/概述/流量证据数量/流量证据ID」，且 ARTEX CSV 不含报告全文。

### reset-password.sh/

- `reset-password.sh`（+21/-35）— **低风险**：已读完整 diff。

### server/

- `server/assembly.go`（+38/-122）— **低风险**：SERIES 有 mcpRuntimeCache（run 内复用连接/元数据）、meterMCPTool 带 RunInfo 归因、loadContextAccess + applyContextPolicy（按角色裁剪 context_search/get/graph_overview）、TaskDomainToolsFrom 任务域工具优先且缺任务域时注入 nil（不降级为全局资产）；ARTEX 每次 run 直连 MCP、closers 用 io.Closer、ToolResolve 返回无 error（读表失败放行代码默认）。wireTools 第 3 个可变参数 configRoot
- `server/assembly_test.go`（+3/-3）— **低风险**：模块路径改名；core_test 去掉 retestProvider 注入与 asset_counts 断言、stats 去掉 compact=1；engine_cancelcause 删除 controlWork("kill") 等结算等待与 AssetAuthorizationChangedCause 审计用例，改测 KillWork；finding_traffic_test 导出用例从 md-report-zip（提交报告）改为 md-zip（带证据），并删除 report/技术附录分割断言；finding_workflow_test 的 notifyFinding 去掉 assetID
- `server/assets.go`（+15/-37）— **低风险**：SERIES 列表读加 20s 上下文、支持 tested/approval_state 过滤与任务级审批查询（CountDSLByTaskApproval/QueryByTaskApproval）、插入走 RegisterUserAssetWithSource；ARTEX 无超时/过滤，插入直接 Upsert + SetTaskAssetSource。
- `server/auth.go`（+9/-18）— **低风险**：JWT subject SERIES→ARTEX；SERIES 校验 alg=HS256 且 subject=SERIES，token 提取要求单一 Authorization/X-Series-Token 头、cookie 名 series_token；ARTEX 只判断 "Bearer " 前缀、cookie 名 artex_token，无 subject/alg 约束。
- `server/auth_test.go`（+0/-80）— **低风险**：已读完整 diff。
- `server/broadcast.go`（+3/-12）— **低风险**：SERIES 订阅者缓冲满时删除并 close(ch)，客户端按 last id 重连补洞；取消订阅有重复 close 保护。ARTEX 缓冲满时静默丢弃事件且不关闭 channel，删除订阅时统一 close(ch)。
- `server/chat_mentions_test.go`（+5/-8）— **低风险**：已读完整 diff。
- `server/chatupload.go`（+1/-36）— **低风险**：SERIES 有 saveUpload/writeUpload（临时文件 + os.Link 原子落盘，避免截断/覆盖竞态）；ARTEX 无这两个函数。
- `server/conversations.go`（+4/-24）— **低风险**：SERIES 在复测会话开跑前校验 FindingRetestForConversation / 原任务可用，并用 WithTaskPolicy 绑定任务资产策略、代理与 CA；ARTEX 无此段（复测会话不带任务资产策略）。
- `server/core_test.go`（+3/-28）— **低风险**：模块路径改名；core_test 去掉 retestProvider 注入与 asset_counts 断言、stats 去掉 compact=1；engine_cancelcause 删除 controlWork("kill") 等结算等待与 AssetAuthorizationChangedCause 审计用例，改测 KillWork；finding_traffic_test 导出用例从 md-report-zip（提交报告）改为 md-zip（带证据），并删除 report/技术附录分割断言；finding_workflow_test 的 notifyFinding 去掉 assetID
- `server/customtool.go`（+9/-94）— **低风险**：ARTEX 去掉 shell 类型（kind 只允许 command/script/http）、ValidateShellCommand 校验、testShellTool 分支、以及三处 customToolAssetPolicy（命令渲染前、script 代码、HTTP URL 与重定向）和 httpProxyTransport(ctx,...) 的 ctx 参数；参数解析错误被忽略。
- `server/customtool_test.go`（+21/-13）— **低风险**：删除 TestCustomToolShellMetadata 与 TestShellMetadataNormalization（对应 shell 元数据能力移除）。
- `server/dto.go`（+8/-12）— **低风险**：TaskDTO 去掉 ExecutionMode/AssetApprovalTemplate；FindingDTO 去掉 ReportDetail；TaskNodeDTO 新增 DeleteReason。
- `server/engine.go`（+160/-298）— **低风险**：：
- `server/engine_cancelcause_test.go`（+6/-30）— **低风险**：模块路径改名；core_test 去掉 retestProvider 注入与 asset_counts 断言、stats 去掉 compact=1；engine_cancelcause 删除 controlWork("kill") 等结算等待与 AssetAuthorizationChangedCause 审计用例，改测 KillWork；finding_traffic_test 导出用例从 md-report-zip（提交报告）改为 md-zip（带证据），并删除 report/技术附录分割断言；finding_workflow_test 的 notifyFinding 去掉 assetID
- `server/engine_timeout.go`（+6/-28）— **低风险**：SERIES 的截止协调器在暂停时冻结倒计时（恢复时顺延 deadline）；ARTEX 按墙上时钟继续走。SERIES 收尾时持 workerControlMu 检查 ExecutionMode=manual（手工模式强制 met=false），终局规划前 beginPlannerRound 守卫并在 Plan 前注入 RunInfo{Trigger:"task_timeout"}、返回前检查 ctx 取消；ARTEX 全无。
- `server/finding_retests.go`（+10/-32）— **低风险**：SERIES 的 get_finding_retest_context 支持 field/index/offset/max_chars 分段读取（ValidateToolDetailInput + ProjectToolDetail），并按 finding.TaskID 直接从 PG 读任务约束（任务未加载也可用），漏洞/任务不存在时报错；ARTEX 无参数、整包返回，约束只在 Manager 已加载任务时读取。
- `server/finding_traffic.go`（+5/-5）— **低风险**：finding_id/binding_id 参数从正整数校验（positiveIDParam）改为 strParam；提示语中 update_finding_report 的参数名说明不同。
- `server/finding_traffic_export.go`（+5/-6）— **低风险**：模块路径改名；buildFindingsReportZip 注释改为中文说明（逻辑未变）。
- `server/finding_traffic_test.go`（+19/-62）— **低风险**：模块路径改名；core_test 去掉 retestProvider 注入与 asset_counts 断言、stats 去掉 compact=1；engine_cancelcause 删除 controlWork("kill") 等结算等待与 AssetAuthorizationChangedCause 审计用例，改测 KillWork；finding_traffic_test 导出用例从 md-report-zip（提交报告）改为 md-zip（带证据），并删除 report/技术附录分割断言；finding_workflow_test 的 notifyFinding 去掉 assetID
- `server/finding_workflow.go`（+8/-35）— **低风险**：SERIES 校验任务角色必须有 TaskID、用 FindingAccessMetadata 轻量读、并按 worker/mainagent/planner 分别 ValidateWorkerAssets / ValidateTaskAssetsApproved + TaskNodeAccess.CanRead；ARTEX 只校验任务归属与继承只读，无资产授权校验。提示语改为 list_task_findings/get_task_node_detail。
- `server/finding_workflow_test.go`（+20/-28）— **低风险**：SERIES ToolResolve 返回 error、断言禁用绑定时必须出现“当前未启用流量证据绑定”；ARTEX 无 error，且删掉该断言（实现也不再输出该提示）。
- `server/findings_groups.go`（+1/-2）— **低风险**：findingFilterFromQuery 去掉 summary_only 参数。
- `server/goals_api.go`（+4/-3）— **低风险**：SERIES 用 NotifyNodes("goal", ids)/NotifyNodes("goal_edited", ids) 唤醒 planner；ARTEX 改为 NotifyGoal(text)/NotifyGoalEdited(old,new)。
- `server/intent_stop.go`（+7/-106）— **低风险**：SERIES 有 intentControlTask（校验 RunInfo 的 task/role/intentID）、KillWork（planner 终止等待中/运行中意图，含 CAS、controlWork kill、冲突重试、活动记录）；ARTEX 只剩 cancelIntentFromMain（任务操作闸门 + engine.StopIntent），文件 106 行内容被删。
- `server/intercept.go`（+31/-6）— **低风险**：ARTEX 新增 judgeWorkerLane 常量与 interceptJudgeUsage 接口（裁判模型消耗统计）；待审批列表/历史改为非 Context、非超时查询（ListPendingIntercepts/ListTaskInterceptsPage/ListAllInterceptsPage），不再支持按 task 过滤待审批。
- `server/manager.go`（+42/-463）— **低风险**：已读完整 diff。
- `server/manager_delete_files_test.go`（+4/-4）— **低风险**：已读完整 diff。
- `server/mcpdiscover.go`（+20/-164）— **低风险**：SERIES 有 mcpDiscoveryPhase/mcpDiscoveryError 阶段化错误、normalizeMCPServer（校验并规范化 transport 相关字段，清空不适用的 command/url）、discoverMCP（纯探测不落库，测试端点与缓存共用）；ARTEX 直接 discoverAndCacheMCP，无规范化与阶段错误。
- `server/orchestration.go`（+203/-196）— **低风险**：：
- `server/platform_tools.go`（+13/-44）— **低风险**：ARTEX 去掉 shell 工具的 executable/directory/usage_help/when_to_use 字段与说明，JSON 解析错误被忽略（_ = json.Unmarshal），create/update MCP 不再 normalizeMCPServer 校验；transport 描述改为 stdio | http / sse。
- `server/server.go`（+386/-587）— **低风险**：：
- `server/server_mgmt.go`（+12/-190）— **低风险**：删除 MCP 测试端点（pgTestMCP/mcpTestTool/decodeMCPBody）、MCP 批量导入（pgImportMCP/mcpImportServer.normalized/mcpImportRequest）、MCP 使用统计（pgMCPUsage）；skillRelPath/walkSkillFiles 的注释移动（ToSlash 逻辑保留）。
- `server/side_questions.go`（+6/-23）— **低风险**：SERIES 在发问前检查 worker 意图是否 stopped（409 Worker 已取消），并在 runSide 中注入 RunInfo、FilterTaskAssetRequest、WithTaskAssetContext 做资产过滤；ARTEX 无发问前状态检查、无资产过滤，仅改提示语为“已删除”；环境变量 SERIES_BTW_MAX_OUTPUT_TOKENS → ARTEX_…
- `server/side_questions_test.go`（+4/-4）— **低风险**：已读完整 diff。
- `server/skill_upload_test.go`（+48/-120）— **低风险**：已读完整 diff。
- `server/task_archive_package.go`（+7/-2）— **低风险**：已读完整 diff。
- `server/task_archive_package_test.go`（+34/-1）— **低风险**：新增 TestTaskArchivePackageSkipsSymlink。
- `server/task_assets.go`（+2/-150）— **低风险**：删除资产审批模板接口（updateTaskAssetTemplate/SetAssetApprovalTemplate）、审批列表/批量批准/撤回/封禁接口（listTaskAssetApprovals/updateTaskAssetApprovals/mutateTaskAssetApprovals，含 group_by=host 分组、Resolved 接口、撤回后 CancelWorkersForAssets）、以及对应的错误映射（超时 504、blocked/not-approved 403）；taskAssetActor("SERIES") 一并删除。
- `server/task_control.go`（+46/-84）— **低风险**：SERIES 的 applyTaskControl 支持 action=finish（走 finishTask）；暂停/恢复会 emitActivity；意图控制持 workerControlMu，cancel 走 StopIntentWithReason + 挂取消 fact + 等待注册/结算的循环，resume 支持 open 直接下发与 stopped(userCancelled) 重开（ReopenIntentByUser + admitTask + 失败回滚 RestoreUserReopen）。ARTEX 无 finish、无活动记录、无锁；cancel 改为 soft/hard
- `server/task_llm.go`（+14/-27）— **低风险**：ARTEX 的 per-task agent 装配缺少：worker/planner/main 的 SetContextPolicy、planner 的 SetOverviewProfile/SetOverviewConfig、三者的 SetSmartProxy、worker 的 SetTrafficRecording、TaskProxyAddr/TaskProxyCACert（改用 m.ProxyAddr/ProxyCACert）、Compactor.SetAssetStore（冷节点压缩的资产授权校验）、以及 main.SetCancelIntent(s.cancelIntentFromM
- `server/task_llm_test.go`（+6/-4）— **低风险**：已读完整 diff。
- `server/task_metadata_test.go`（+1/-9）— **低风险**：已读完整 diff。
- `server/task_templates.go`（+56/-12）— **低风险**：ARTEX 新增 CategoryID 与 InterceptRules（模板保存任务级拦截/允许规则快照），TaskTemplatePatch 用 Set 标志区分“未传”与“清空”；SERIES 模板只有 name/description/goal。
- `server/tool_usage.go`（+2/-11）— **低风险**：meteredTool 不再实现 PrepareInput（agent.PrepareToolInput 默认值注入/校验）与 AcceptsRawInput 透传。
- `server/tool_usage_test.go`（+2/-58）— **低风险**：删除 TestToolUsageBackfill（历史回填）用例，只保留 ledger 用例。
- `server/tools_wire_test.go`（+11/-36）— **低风险**：已读完整 diff。
- `server/webui_embed.go`（+16/-1）— **低风险**：ARTEX 新增 serveFileIfExists 辅助并用 http.ServeFileFS 提供 index.html；SERIES 直接调 serveFileIfExists（该函数在别处）。

### start.bat/

- `start.bat`（+19/-7）— **低风险**：新增退出码 75（请求重启）分支与说明，二进制名 series.exe→artex.exe。

### start.sh/

- `start.sh`（+22/-11）— **低风险**：已读完整 diff。

### traffic/

- `traffic/archive.go`（+6/-28）— **低风险**：SERIES 归档 blob 为 zst 时解码、导入时按需压缩回 .zst、FTS 索引前对 inline body 解压并按 Content-Encoding 解码；ARTEX 一律 .bin 明文、不解压。
- `traffic/content_encoding_test.go`（+0/-8）— **低风险**：删除 readTrafficPart 解压断言。
- `traffic/evidence.go`（+0/-1）— **低风险**：读取证据时 SERIES 对 req/resp 做 decompressInline 解压；ARTEX 直接返回原始（可能压缩的）字节。
- `traffic/reclaim_test.go`（+1/-17）— **低风险**：bulkRecord 用重复字符替代不可压缩随机体（因不再压缩）。
- `traffic/traffic.go`（+156/-566）— **低风险**：SERIES 有 blob 内存缓存（8MiB）、SetAssetPolicyStore（代理出口做任务资产授权）、SetRecordingEnabled（抓包开关不关代理，保留授权兜底）、静态资源过滤（DefaultSkipStaticRules/SetSkipStatic/SetSkipRules/shouldSkipStatic）、保留期清扫（SetRetention/retentionLoop）；ARTEX 全无，traffic_search 描述也不同（host 必填 vs 主机检索）。
- `traffic/traffic_store_test.go`（+6/-8）— **低风险**：blob 断言从 .zst 且更小改为 .bin 且等于原文。
- `traffic/traffic_test.go`（+0/-19）— **低风险**：模块路径改名；core_test 去掉 retestProvider 注入与 asset_counts 断言、stats 去掉 compact=1；engine_cancelcause 删除 controlWork("kill") 等结算等待与 AssetAuthorizationChangedCause 审计用例，改测 KillWork；finding_traffic_test 导出用例从 md-report-zip（提交报告）改为 md-zip（带证据），并删除 report/技术附录分割断言；finding_workflow_test 的 notifyFinding 去掉 assetID

### update.sh/

- `update.sh`（+27/-15）— **低风险**：已读完整 diff。

### web/

- `web/package.json`（+2/-3）— **低风险**：已读完整 diff。
- `web/public/logo.svg`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/app/(auth)/login/page.tsx`（+231/-19）— **低风险**：ARTEX 新增《使用须知》弹窗（滚动到底才可同意、未勾选禁止登录）、左右分栏登录页（左 logo、右表单）、登录后跳 /function/tasks；用户名 SERIES→ARTEX；删除 HTTPEntryAuthError 分支，错误统一提示“用户名或密码错误/无法连接”。
- `web/src/app/(auth)/setup/page.tsx`（+18/-10）— **低风险**：初始化密码页改为左右分栏布局与 /logo.png 展示，跳转改 /function/tasks，文案 SERIES→ARTEX；错误展示精简。
- `web/src/app/(external)/page.tsx`（+1/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/_components/main-content.tsx`（+72/-7）— **低风险**：ARTEX 新增全局头部（AccountSwitcher/ThemeSwitcher/LayoutControls/SearchDialog/UpdateBadge、版本号来自 /health），并按路径判断全出血页（/function/tasks/* 详情）；SERIES 只在 /function/tasks/detail 时去掉头部。
- `web/src/app/(main)/_components/sidebar/app-sidebar.tsx`（+36/-36）— **低风险**：已读完整 diff。
- `web/src/app/(main)/_components/sidebar/layout-controls.tsx`（+257/-4）— **低风险**：已读完整 diff。
- `web/src/app/(main)/_components/sidebar/nav-main.tsx`（+8/-21）— **低风险**：已读完整 diff。
- `web/src/app/(main)/chat/page.tsx`（+14/-43）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/dashboard/page.tsx`（+7/-7）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/function/findings/detail/page.tsx`（+310/-67）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/function/findings/page.tsx`（+390/-249）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/function/tasks/detail/_tabs/assets-tab.tsx`（+9/-230）— **低风险**：删除资产审批徽章逻辑（blocked/pending/revoked/derived 继承）、来源任务只读徽章、已测试/未测试徽章与提示；范围输入说明改为“域名和 IP 创建或复用全局资产；CIDR/ICP/关键词作为上下文”。
- `web/src/app/(main)/function/tasks/detail/_tabs/broadcast-tab.tsx`（+142/-99）— **低风险**：新增 AssetList（节点锚定资产：类型标签+可辨识文本）与 RelatedNodeCard（上下游节点悬停名片：类型/状态/来源/时间/摘要/资产），数据来自本页返回的 refs/assets，不再额外请求；移除 startPolling 轮询与 summaryOf 的 "summary" 回退字段。
- `web/src/app/(main)/function/tasks/detail/_tabs/findings-tab.tsx`（+195/-222）— **低风险**：整体重写：移除 FindingsToolbar/FindingPagination/FindingSelector/FindingPresentation/FindingSuppressionRules/导出按钮/通知已读，改为轻量列表（排序仅 time、localStorage 键 artex_task_findings_sort）、状态徽章、卡片/行内展开、跳转详情；不再有批量选择与导出入口。
- `web/src/app/(main)/function/tasks/detail/_tabs/overview-tab.tsx`（+275/-0）— **低风险**：新增 TaskInterceptRulesCard（任务级拦截/允许规则的列表、新增、行内编辑、删除、启用开关，7 种 kind：exact/fuzzy domain/ip/url 与 cidr），插在总览页。
- `web/src/app/(main)/function/tasks/detail/_tabs/sessions-tab.tsx`（+205/-638）— **低风险**：移除 WorkerQueue 队列视图、SessionToolCalls、批量选择 Checkbox 与 SourceTranscript、worker-feedback 的 lastMainActivity、router/searchParams 相关深链；新增 deleted 状态（CircleSlash/Trash2 图标、删除线、“已删除”徽章），删除动作改为弹窗填写原因并选择 soft/hard，恢复按钮从“已取消执行可重开”改为仅暂停可恢复；SessionsTab 入参去掉 pendingIntercepts，改用 useApprovalFocus({taskId})。
- `web/src/app/(main)/function/tasks/detail/page.tsx`（+61/-152）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/function/tasks/page.tsx`（+709/-294）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/function/traffic/page.tsx`（+0/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/layout.tsx`（+49/-10）— **低风险**：ARTEX 从 cookie 读 sidebar_variant/collapsible（静态导出可用的客户端方案），sidebar 宽度改为 calc(var(--spacing)*68)，移除 SettingsWorkspace 包裹与 NewTaskShortcut。
- `web/src/app/(main)/system/agents/page.tsx`（+283/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/system/intercept/page.tsx`（+985/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/system/llm/_components/retry.tsx`（+0/-7）— **低风险**：已读完整 diff。
- `web/src/app/(main)/system/llm/page.tsx`（+990/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/system/logs/page.tsx`（+4/-10）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/system/mcp/page.tsx`（+482/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/system/settings/page.tsx`（+713/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/system/skills/page.tsx`（+1127/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/(main)/system/tools/page.tsx`（+843/-1）— **低风险**：根路径重定向 /function/tasks/new → /function/tasks。
- `web/src/app/globals.css`（+17/-105）— **低风险**：已读完整 diff。
- `web/src/app/layout.tsx`（+1/-1）— **低风险**：ARTEX 从 cookie 读 sidebar_variant/collapsible（静态导出可用的客户端方案），sidebar 宽度改为 calc(var(--spacing)*68)，移除 SettingsWorkspace 包裹与 NewTaskShortcut。
- `web/src/components/agent-editor.tsx`（+103/-346）— **低风险**：已读完整 diff。
- `web/src/components/approval-execution-focus.tsx`（+5/-16）— **低风险**：已读完整 diff。
- `web/src/components/approval-records.tsx`（+35/-53）— **低风险**：已读完整 diff。
- `web/src/components/finding-traffic-panel.tsx`（+5/-6）— **低风险**：已读完整 diff。
- `web/src/components/scope-text-editor.tsx`（+2/-2）— **低风险**：已读完整 diff。
- `web/src/components/task-template-controls.tsx`（+88/-47）— **低风险**：已读完整 diff。
- `web/src/components/exploration-graph.tsx`（+约30/-约60）— **中风险**：ARTEX 把 digest 折叠逻辑内联到本文件（collapseDigestGraph：按 covers 折叠成员、重连去重、删除反向 derived_from），SERIES 在 lib/exploration-graph.ts 中提供；文件含 2 个真实 NUL 字节（模板 key 的分隔符），导致 git 将其识别为二进制、此前的 229 统计与 diff 工具都漏掉了它。
- `web/src/components/transcript.tsx`（+103/-256）— **低风险**：已读完整 diff。
- `web/src/components/ui/badge.tsx`（+0/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/button.tsx`（+7/-7）— **低风险**：已读完整 diff。
- `web/src/components/ui/calendar.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/chart.tsx`（+2/-2）— **低风险**：已读完整 diff。
- `web/src/components/ui/checkbox.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/dialog.tsx`（+2/-2）— **低风险**：已读完整 diff。
- `web/src/components/ui/input-group.tsx`（+3/-3）— **低风险**：已读完整 diff。
- `web/src/components/ui/native-select.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/select.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/sheet.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/table.tsx`（+2/-2）— **低风险**：已读完整 diff。
- `web/src/components/ui/toggle-group.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/toggle.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/components/ui/tooltip.tsx`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/hooks/use-side-questions.ts`（+2/-2）— **低风险**：已读完整 diff。
- `web/src/lib/api.ts`（+255/-261）— **低风险**：令牌机制 series_token + X-Series-Token + entryAwareFetch → artex_token + Authorization Bearer + 原生 fetch；删除 taskNotifications/toolCalls/toolCallDetail、资产审批（taskAssetApprovals/updateTaskAssetTemplate/…）、执行模式（setExecutionMode）、人工派发（dispatchIntents）、worker 队列（workerQueue/moveWorker）、MCP 测试/导入/用量、SmartProxy
- `web/src/lib/company-scope.ts`（+12/-31）— **低风险**：已读完整 diff。
- `web/src/lib/fonts/registry.ts`（+22/-83）— **低风险**：已读完整 diff。
- `web/src/lib/mock/data.ts`（+107/-161）— **低风险**：已读完整 diff。
- `web/src/lib/mock/handler.ts`（+303/-1533）— **低风险**：已读完整 diff。
- `web/src/lib/preferences/preferences-config.ts`（+6/-6）— **低风险**：已读完整 diff。
- `web/src/lib/side-questions.ts`（+13/-40）— **低风险**：已读完整 diff。
- `web/src/lib/status.ts`（+14/-1）— **低风险**：已读完整 diff。
- `web/src/lib/task-assets.ts`（+4/-5）— **低风险**：已读完整 diff。
- `web/src/lib/types.ts`（+211/-208）— **低风险**：删除资产审批/执行模板/任务资产审批/资产来源/任务发现筛选/PlannerOverviewLimits 等类型；新增 WorkspaceEntry/Listing/File（工作目录文件管理）、TaskTemplate.category_id 与 intercept_rules、TaskNode.delete_reason、ExplorationNodesPage.assets（播报板节点锚定资产）、SessionStatus 增 "deleted"。
- `web/src/navigation/sidebar/sidebar-items.ts`（+30/-9）— **低风险**：已读完整 diff。
- `web/src/proxy.ts`（+4/-6）— **低风险**：已读完整 diff。
- `web/src/scripts/theme-boot.tsx`（+5/-5）— **低风险**：已读完整 diff。
- `web/src/styles/presets/brutalist.css`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/styles/presets/soft-pop.css`（+1/-1）— **低风险**：已读完整 diff。
- `web/src/styles/presets/tangerine.css`（+1/-1）— **低风险**：已读完整 diff。

## 四、待补充文件

无。229 个实质差异文件均已读取完整 diff 并记录结论；未出现缺少一方文件信息的情况。


## ✅ 229 个实质差异文件核对完毕（单边文件未读，整体未完成）

230 / 230 个实质逻辑差异文件均已打开 diff 并完成语义判断（其中 20 余个大文件按功能块读取，见“核对方法与边界”）；74 个纯重命名与 216 个纯空白差异不计入语义核对，但已在盘点中列出。
