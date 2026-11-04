# SERIES vs ARTEX 核对原始记录

## 第一部分：230 个实质差异文件

# 语义核对记录（series vs artex）

进度：6 / 229（只记录读完完整 diff 的文件）

## agent/assembly.go
- 位置：DeferredInfo 结构、AugmentTools、guardedTool
- 差异：SERIES 有 ContextReads/AssetGuidance 字段与 WithTaskToolSet/TaskDomainToolsFrom，ARTEX 无；SERIES AugmentTools 返回 error 并对待注入工具去重，ARTEX 不返回 error 且直接 append（重名可重复）；ARTEX 末尾多 applyHostPromptOverrides（本次移植）。
- 风险：中（错误处理与工具去重行为不同）

## agent/cancelcause.go
- 位置：原因定义表
- 差异：SERIES 有 AssetAuthorizationChangedCause、AbortTaskFinishedByUser；ARTEX 无。ARTEX 新增 AbortCancelledByUser（本次移植）。AbortWorkCancelledByUser 语义不同：SERIES=取消保留意图，ARTEX=删除意图（软删/硬删）。
- 风险：中（缺少资产授权变更终止原因，取消语义不同）

## agent/capture.go
- 位置：captureRun、工具事件处理
- 差异：导入路径改名；SERIES captureRun 接受 settleGuard 并走 newSession，ARTEX 直接 agentcore.NewSession；SERIES 用 assetContextCall 归一化工具调用，ARTEX 直接用 ev.ToolUse。
- 风险：低

## agent/chat.go
- 位置：ChatAgent 字段、WithTaskPolicy、Chat 装配
- 差异：SERIES 有 ctxPolicyFn（自定义上下文清理/摘要线）、taskAssets/taskID（任务资产作用域）、WithTaskPolicy、SetContextPolicy；ARTEX 无。ARTEX 的 AugmentTools 不返回 error；RunInfo 只带 SessionID（SERIES 带 AgentKey/TaskID/Trigger）。Settlement 调用改为 wrapupSettlement(agentKey, nil)。ARTEX 无 AssetPolicyHooksWithGuard / withAssetContext / applyContextPolicy / newSession。
- 风险：中（聊天会话缺少任务资产作用域与上下文策略）

## agent/compaction.go
- 位置：Compactor 结构、OnPlannerRound、foldBlock、compress
- 差异：SERIES 有 assets 与 blockAuthorized（压缩块要过资产授权）、compactorRunContext（独立 RunInfo 归因）；ARTEX 无。foldBlock 语义不同：SERIES 在源签名/内容变化时整体丢弃重试；ARTEX 只剔除复活成员，剩余不足 K 才跳过，且用原 nodeByID/cvers 生成摘要。ARTEX 补了 exp<ID>-compactor 的 session id。compress 系统提示 SERIES 包 withChineseLanguage，ARTEX 未包。
- 风险：中（压缩范围授权校验缺失；提交时校验策略不同）

## agent/deferred.go
- 位置：deferredSystem、seedUnlockFromHistory
- 差异：SERIES 追加 toolAvailabilityRule 与 contextReadRule（依据 ContextReads）、AssetGuidance，并返回 boundary=1；ARTEX 全部没有，无块时 boundary=0。SERIES 有 plannerAssetGuidance 函数（按最终工具集生成资产检索指引），ARTEX 无。seedUnlockFromHistory：SERIES 只回放“调用+成功结果”配对的 Skill，ARTEX 只要看到 Skill 调用就解锁（未配对结果）。
- 风险：中（未成功/被拒绝的 Skill 调用也会在恢复时解锁，属行为放宽）

## agent/finding_workflow.go
- 位置：findingIDGuidance 常量、findingWorkflowTools 分支
- 差异：SERIES 的编号说明指向 context_search/context_get；ARTEX 改成 list_findings/list_task_findings/node_detail/get_task_node_detail（无上下文工具）。SERIES 在“未启用流量绑定”时直接返回“不要绑定”提示；ARTEX 删掉该分支。SERIES 的 guidance 里涉及 context_get/context_search 的取 ID 路径，ARTEX 换成任务详情/漏洞列表工具。
- 风险：中（提示词与工具面绑定，编号约定随工具改，模型可能按旧说明找 context_search）

## agent/goals.go
- 位置：goalsScopeTail、DecomposeGoalsWithProvider
- 差异：SERIES 用 register_user_target + registerDescriptionAsset 登记资产范围；ARTEX 改为 add_task_scope + goalsScopeTail。ARTEX 不再设置 RunInfo（TaskID/AgentKey/Trigger），只补 transcript session id。
- 风险：中（资产范围登记工具不同；缺少 RunInfo 归因）

## agent/host_lessons.go
- 位置：落盘缓冲、RecordHostLesson、FlushHostLessons
- 差异：SERIES 用 2 秒去抖缓冲批量落盘（hostLessonBuf/hostLessonFlushDelay/FlushHostLessons/mergeHostLessonLists）；ARTEX 每次错误直接读盘-合并-写盘。ARTEX 无 HostLessonNote 的 pending 覆盖。
- 风险：中（写盘频率高；并发下每次读改写，性能与竞争行为不同）

## agent/host_profile.go
- 位置：BashPromptOverride / bashPromptOverrideFor / windowsBashPrompt
- 差异：SERIES Windows 返回空串（整段清空 SDK 提示）；ARTEX 返回改写文本 windowsBashPrompt，保留“先写脚本/避免交互输入/超时”等建议（用户已确认保留）。hostClipRunes 位置调整。
- 风险：低（有意差异，用户已拍板）

## agent/intent_stop.go
- 位置：文件整体
- 差异：SERIES 是完整工具实现（cancelIntentTool/killWorkTool/intentStopTool，含角色校验、node 校验、planner 资产授权校验）；ARTEX 只剩 IntentStopResult/IntentStopper 类型与 SetCancelIntent，工具改在 server/orchestration.go 内联实现。
- 风险：中（工具层角色/资产校验未在 agent 侧统一，依赖 server 侧校验）

## agent/mainagent.go
- 位置：结构体字段、SetContextPolicy/SetSmartProxy、mainAgentDefaultTmpl、mainAgentSystem、Chat 装配
- 差异：SERIES 有 ctxPolicyFn、smart（智能代理）、SetContextPolicy/SetSmartProxy；ARTEX 无。提示词差异大：SERIES 用 context_search/context_get/asset_search，ARTEX 用 list_findings/list_facts/list_assets/get_worker_output；SERIES 有执行模式（手工/托管）与 cancel_intent 指引，ARTEX 无。Chat 签名少 guard/notifyNodes；ARTEX 无 withAssetContext/dispatchReceiptProvider/AssetPolicyHooksWithGuard/WorkerFeedbackContext/applyContextPolicy/newSession，代理地址直接用 m.proxyAddr（SERIES 按任务派生）。
- 风险：中（执行模式、智能代理、资产 hook、Worker 反馈上下文均缺失）

## agent/noa.go
- 位置：enableNoa
- 差异：SERIES 在启用前检查归档目录可写（MkdirAll+临时文件探测），失败回退内置压缩，并把 Compaction.ContextWindow 传给适配器；ARTEX 无这两项检查。
- 风险：中（不可写目录不会回退，可能启用后失败）

## agent/memory_tools.go
- 位置：memoryUsageHint、memoryCatalogTools、saveMemoryFile/renderMemoryFile
- 差异：SERIES 暴露 5 个工具（含 RecallMemory/RecordMemory，支持 pinned 常驻记忆）；ARTEX 只暴露 4 个（List/Read/Search/SaveMemory），无 pinned 字段与常驻注入，读取上限用常量 30000 而非 readOutputBudget(ctx)。
- 风险：中（缺少常驻记忆与按上下文预算裁剪）

## agent/planner.go
- 位置：Planner 字段与 Set*、TriggerEvent、renderTriggers/renderGraphOverview、plannerDefaultTmpl、Plan 装配
- 差异：SERIES 有 ctxPolicyFn、overviewProfileFn/overviewConfigFn（完整/精简态势）、smart（智能代理）、KillWork 为 IntentStopper、RunInfo 带 AgentKey；ARTEX 全无，KillWork 退化为 func(intentID) error。SERIES 的态势是“用户取消清单 + plannerBootstrap(强制 context_search) + 有界/精简 graph_overview（预算裁剪）”；ARTEX 是 renderTriggers + 完整 graph_overview JSON（无预算裁剪）。SERIES 有执行模式判断（手工模式不自动规划）、manualQueue、DeletedWorkers、requireRoleTools、AssetPolicyHooksWithGuard、withAssetContext、applyContextPolicy、StripCoverageParams、DisableBackgroundTasks；ARTEX 均无。SERIES 只有 Bash 暴露给 planner；ARTEX 用 actool.DefaultTools() 全量默认工具。
- 风险：中（规划者工具面扩大、缺少态势预算控制、缺少手工模式与资产 hook）

## agent/prompt.go
- 位置：chineseLanguagePrompt/withChineseLanguage、renderSystem、renderTmpl
- 差异：SERIES 强制给所有渲染后的系统提示追加“统一语言要求（简体中文）”，renderTmpl 对无模板变量文本直接返回；ARTEX 无语言策略，且总是走 template.Parse。
- 风险：中（ARTEX 缺少全局中文约束；含 {{ 字面量的文本可能解析失败）

## agent/promptcatalog.go
- 位置：autoDefaultTmpl、pentestDefaultTmpl、ReporterDefaultPrompt
- 差异：工具名随 ARTEX 工具面改写（context_search/graph_overview → list_findings/get_task_graph/get_task_worker_trace 等）；pentest 的资产检索/漏洞回看改 list_assets/list_findings；reporter 的报告步骤改成 get_task_node_detail/get_task_worker_trace；update_finding_report 参数名 finding_node_id→finding_id。
- 风险：中（提示与工具名强绑定，随工具面变化）

## agent/runinfo.go
- 位置：RunInfo 类型与 ctx 存取
- 差异：SERIES 直接复用 llmrec.RunInfo（含 Trigger/Phase/RetryOrdinal 等归因字段）；ARTEX 自定义精简 RunInfo（TaskID/ExplorationID/IntentID/SessionID），无 Trigger/Phase。
- 风险：中（llmrec 归因字段缺失，按 trigger/phase 的统计与提示分发不可用）

## agent/toolcatalog.go
- 位置：builtinToolsByAgent、defaultUnbound、ToolResolve、DecorateTool、injectDefaults、applyHostPromptOverrides
- 差异：SERIES 的 auto/pentest 绑定 context_search/context_get/graph_overview/asset_search 等；ARTEX 改为 list_assets/list_findings 等。SERIES 有 defaultUnbound 空表、ToolResolve 返回 error、DecorateTool 注入 factWriteContract、PrepareInput 做默认值注入+业务校验（injectDefaults 带 missingOnly 参数）；ARTEX ToolResolve 无 error、无校验、injectDefaults 不区分显式空串、无 normalizeToolSchema。ARTEX 新增 promptOverrideTool/applyHostPromptOverrides（Bash 说明改写）。
- 风险：中（参数校验与显式空值语义缺失；工具目录随工具面变化）

## agent/wrapup.go
- 位置：wrapupSettlement、wrapupSettlementForTask
- 差异：SERIES 签名无 disabledTools（注释明确要求传 nil，避免隐藏工具导致整段缓存重算），收尾拦截靠 newSession 的 settleGuard；ARTEX 保留 disabledTools 参数，本次移植调用处传 nil，并新增 SettlementGate 在请求观察后拒绝 Bash。
- 风险：低（调用处传 nil，缓存前缀不变；保留参数属兼容）

## agent/worker.go
- 位置：Worker 字段/Set*、marker、proxyEnv、TaskProxyAddr、worker 提示、Execute/execute
- 差异：SERIES 有 ErrWorkerAssetAuthorization、trafficRecording 开关、smart、ctxPolicyFn、IntentAssetIDs（含血缘）与 ValidateIntentAssets（启动前资产授权校验，失败零 token 退出）、TaskProxyAddr（带签名的任务代理凭据）、currentToolUAEnv、NO_PROXY 清空；ARTEX 全无，代理直连 m.proxyAddr。worker 提示差异大：SERIES 保留资产审批/待审批语义、流量工具说明与“最终总结返回下发者”；ARTEX 简化。notifyFinding 少 assetIDs 参数。ARTEX 新增 memory 工具注入与 defaultToolsExcept。
- 风险：中（缺少 worker 启动前资产授权、任务级代理凭据与流量录制开关）

## agent/blackboard_inheritance_test.go
- 位置：测试用例
- 差异：SERIES 有“派生节点继承资产授权”测试（依赖审批）、graph_overview 省略 scope；ARTEX 改为“展开公司范围”断言。GetWorkerTrace 超限行为不同：SERIES 报错，ARTEX 返回前 5 条 + omitted_step_ids + notice。
- 风险：低（测试差异，但反映 get_worker_trace 行为不同）

## agent/deferred_test.go
- 位置：断言
- 差异：SERIES 断言无全局块时 system 仍带 toolAvailabilityRule 且 boundary=1；ARTEX 断言纯文本 boundary=0。SERIES 有 TestSeedUnlockRequiresSuccessfulResult（只有成功结果才解锁），ARTEX 删掉且实现也放宽。
- 风险：低（测试差异，对应实现差异已在 deferred.go 记录）

## agent/finding_recorder_test.go
- 位置：notifyFinding stub
- 差异：SERIES 的 notifyFinding 带可变 assetIDs 参数；ARTEX 只有 (int64, string)。
- 风险：低

## agent/finding_workflow_test.go
- 位置：ToolResolve stub、AugmentTools 调用、禁用绑定断言
- 差异：SERIES ToolResolve 返回 error、断言禁用绑定时必须出现“当前未启用流量证据绑定”；ARTEX 无 error，且删掉该断言（实现也不再输出该提示）。
- 风险：低

## agent/host_lessons_test.go
- 差异：SERIES 有去抖不落盘测试与 asset_search 描述携带本机经验测试；ARTEX 删除，改为直接落盘（无 FlushHostLessons）。
- 风险：低

## agent/host_profile_test.go
- 差异：SERIES 断言 Windows 下 Bash 提示必须为空串；ARTEX 删除这些断言（改写版由 bash_prompt_test.go 覆盖）。
- 风险：低

## agent/memory_tools_test.go
- 差异：SERIES 断言 5 个记忆工具与 pinned 常驻注入；ARTEX 断言 4 个（List/Read/Search/SaveMemory），删除 pinned 相关用例。
- 风险：低

## agent/prompt_test.go
- 差异：SERIES 断言中文语言策略幂等且所有 agent 提示都追加；ARTEX 删除该策略及断言。
- 风险：低

## agent/tools.go
- 位置：ToolSet 字段、graphOverviewData、relatedTaskOverviews、list_findings/list_facts、node_detail、add_intent、prove_goal/goal_met、report_finding、record_fact、PlannerTools
- 差异（完整读完 2176 行 diff）：
  1) 资产授权：SERIES 有 visibleNodeIDs/restrictNodeRelations/restrictNodeOrigins、authorizedNodesForStore、nodeAuthorization、validatePlanningAssets、validateResultAssets、targetAccess 查询、report_finding 的 Revalidate 二次校验、markResultAssetsTested、digestAuthorizationBatch；ARTEX 全部没有，只有资产拦截（CheckAssetsIntercept）一项。
  2) 意图准入：SERIES 用 AddIntentDeduplicated + requireManaged 校验（planner RunInfo 必须匹配），返回 duplicates/dispatch_results/access_errors；ARTEX 直接 AddIntent，无去重、无 dispatch 结果。
  3) 工具面：SERIES PlannerTools 含 context_search/context_get/asset_search/asset_get/check_target_access/list_task_assets/mark_host_proxy/check_host_proxy/listFindingDeletionFeedback；ARTEX 改为 list_findings/list_facts/node_detail/get_worker_output/get_worker_trace/expand_digest 等，且默认带 goal_met（但 defaultUnbound 排除绑定）。
  4) 态势：SERIES 有有界窗口/精简投影/预算截断/授权过滤/related scope 展开；ARTEX 直接全量 JSON，无授权过滤、无预算。
  5) 其他：SERIES 的 node_detail 支持 field/index/offset/max_chars 续读；ARTEX 只按 id 返回整节点。SERIES 的 list_findings/list_facts 支持分页/筛选/预算截断；ARTEX list_findings 全量 500、list_facts 分页无 severity/asset 过滤。SERIES 有 factWriteContract 强制工具名/批量契约；ARTEX 无。
- 风险：中（缺失资产级读写授权与意图去重，属能力/安全边界差异）

## agent/tools_digest.go
- 位置：digest 折叠/展开授权与分页
- 差异：SERIES 的 cold digests 与 expand_digest 全部经过 digestAuthorizationBatch（任一成员未授权则整块隐藏，成员分页 ToolDigestMemberPage，正文 textWindow 偏移续读，响应超预算自动二分截断）；ARTEX 全无授权过滤、无分页/偏移/预算裁剪，直接返回全部成员与正文。
- 风险：中（未授权成员可随 digest 展开被读到）

## agent/tools_insert.go
- 位置：资产插入与审批状态
- 差异：SERIES 通过 RegisterAgentAssetWithOrigin 作用域写入，写后查询 TaskAssetApprovalStates，按 pending/blocked/revoked 统计并拒绝；有 related 参数控制是否纳入覆盖度；有 originWorker/toolUseID 归因。ARTEX 直接 Upsert*，只做资产拦截规则（EvaluateAssetGate）与 SetTaskAssetSource，无审批状态、无 related、无归因字段。
- 风险：中（资产审批与来源归因缺失）

## agent/proxyenv_test.go
- 差异：SERIES 有 TaskProxyAddr 测试（带 series-task-<id> 用户名与 64 位密码，直连代理不改写）；ARTEX 无该函数故无测试。
- 风险：低

## agent/worker_intervention_test.go
- 差异：SERIES 有“品牌变更后投递身份仍匹配”的标记解析测试（解析任意 *_WORKER_CHAT 注释，不依赖 SERIES 前缀）；ARTEX 只做精确前缀匹配，测试也删除。
- 风险：低（恢复旧品牌会话时可能重复投递）

## agent/settlement_gate_test.go
- 差异：SERIES 通过 dispatchHooks 测收尾门（覆盖：提示后拦截、落库放行、持续闭锁、PromptByReason、追加后缀、相似文本不误判、非 user 尾部不触发、无 guard/settlement 不拦截、真实 SDK 循环工具面不变）；ARTEX 用 NewSettlementGate 测（闭锁、前缀匹配、PromptByReason、惰性、WrapDeps 观察）。ARTEX 缺相似文本误判、非 user 尾部、工具面稳定性的对照测试。
- 风险：低

## db/asset_dsl.go
- 差异：SERIES 支持 approval_state（含 status==approved 别名）并要求任务上下文；有 CountDSLByTaskApproval/QueryDSLByTaskApproval/buildTaskDSLWhere（按任务关联与审批过滤、workerRead 投影）；ARTEX 无 approval_state、无任务级审批查询，DSL 只支持 =/==/!=/> 与裸词，不支持 contains/=~ 与运算符两侧空格；查询直接用 s.db 而 SERIES 用 queryRow/selectAssetColumns。
- 风险：中（审批维度 DSL 与任务级过滤缺失）

## db/companies.go
- 差异：SERIES 范围原始/单值上限 65536 rune；ARTEX 为 1024。ON CONFLICT 从 md5(value) 改为 value（icp/keyword 去重键不同）。注释提及 scopesentry sync。
- 风险：中（可登记范围长度骤降；去重键变化可能产生重复或冲突行为不同）

## db/company_scope.go
- 差异：SERIES 的 ScopeInput 有 Manual 字段，手工录入时 CIDR/URL/证书原样保留（不强制降级为 keyword）；ARTEX 无 Manual 分支。
- 风险：中（手工录入 URL/证书会被改写为关键词或拒绝）

## db/constants.go
- 差异：ARTEX 新增 StateIntentDeleted="deleted"（软删意图状态），SERIES 无此常量（其删除走 cancel_reason/state 组合）。
- 风险：低

## db/chat_mentions.go
- 差异：SERIES 分“精确 ID 命中（主键查询）”与“模糊候选（trigram/index）”两路 UNION，带 5s 上下文超时、exact 游标标记，模糊查询限定各表独立 ILIKE 列；ARTEX 单条 UNION 全表扫描式写法（id::text 比较、concat_ws ILIKE），无超时，exact 用 fmt.Sprint(id)==query 推断。
- 风险：中（大表上 @ 提及搜索性能与游标语义退化）

## db/config.go
- 差异：SERIES 有 ProfileByIDContext、保存时 TrimSpace(SessionHeaderKey)、MCP 使用统计（calls/tasks/usage_agents/last_used，MCPToolUsageStats/MCPServerUsageStats）、ImportMCP 批量导入（按 name upsert、清缓存）；ARTEX 全无（MCPServer 只多 Insecure 字段，MCPToolsDetailed 不再汇总调用次数）。
- 风险：中（MCP 导入/使用统计能力缺失）

## db/intercept.go
- 差异：SERIES 全部走 Context 版本并带 10s 超时，ListPendingInterceptsContext 支持按 task 过滤；ARTEX 只有非 Context 查询，不接受 task 过滤。
- 风险：中（长查询无超时、无法按任务过滤待审批）

## db/settings.go
- 差异：SERIES 有 SetSettingsAtomic（事务写入，失败全部回滚）；ARTEX 只有逐键 SetSetting。
- 风险：中（多键设置可能部分生效）

## db/tool_usage.go
- 差异：SERIES 内嵌 tool_usage_backfill.sql 做历史回填，并在汇总时用 CanonicalToolUsageKey 合并别名；ARTEX 无回填、无别名归一。
- 风险：低（统计口径不同）

## db/db.go
- 差异：SERIES 额外 embed asset_templates.sql、连接池 20/10、全新库检测后写入 traffic_retention_days(30/0)、ensurePerformanceIndexes、backfillToolUsage、把 browser MCP 改名 playwright 并 seed 3 个内置 MCP（playwright/web_search/chrome-devtools）、清理 ScopeSentry 占位；ARTEX 连接池 32/32，只 seed 一个 browser MCP，新增 seedDefaultAssetInterceptRules（.gov/.gov.cn/.edu/.edu.cn 模糊域名默认拦截）。DSN 环境变量名 SERIES_PG_DSN → ARTEX_PG_DSN。
- 风险：中（初始化差异：索引/回填/保留期默认值/内置 MCP 数量不同；ARTEX 多出默认资产拦截规则）

## db/finding_retests.go
- 差异：ARTEX 复测判 fixed 时改走 SetFindingStatusTx（带状态变更通知，注释说明此前裸 UPDATE 不发推送）；SERIES 仍是裸 UPDATE findings SET status。
- 风险：低（ARTEX 行为更完整）

## db/finding_traffic.go
- 差异：SERIES 的 RecordFindingInput 带 AccessRole/Revalidate/SuppressionCheck/SuppressionTargets，RecordFindingTx 内做 LockFindingRecordTx、ValidateFindingSuppressionTx、ValidateFindingAccessTx，并写禁报校验审计；ARTEX 去掉这些，改 LockTaskEvidenceTx，并在同事务内写 RecordNotificationEventTx(notify.EventFindingCreated) 推送事件。RecordFindingTx 增加 ctx 参数。
- 风险：中（SERIES 的禁报/访问双校验缺失，但 ARTEX 增加了推送事件原子性）

## db/findings.go
- 差异：SERIES 有 FindingFilter.SummaryOnly（列表省掉 evidence）、导出含 report_detail、FindingAccessMetadata、DeleteFinding(id, reason) 写 finding_deletion_feedback + finding_suppression_rules、PatchFinding 原子改多字段（含图投影 content_version）；ARTEX 无 SummaryOnly、导出不含 report_detail、删除不写反馈/禁报规则、改为 setFindingCol 单列更新 + SetFindingStatus 裸更新（另有 SetFindingStatusWithNotify 供生产用）。
- 风险：中（删除反馈与禁报联动、批量补丁、报告明细导出能力缺失）

## db/exploration.go（851 行 diff，已完整读完）
- 差异：
  1) 队列串行化：SERIES 所有写操作走 beginQueue/queueExec（advisory lock 串行），ARTEX 直接用 s.db.Exec/Begin。
  2) 意图去重：SERIES 有 IntentFingerprint（summary 归一 + 排序资产 SHA256）+ AddIntentDeduplicated/AddPlannerIntentDeduplicated（去重、抓指纹、校验 execution_mode=manual 拒绝 planner、写 fingerprint）；ARTEX 的 AddIntent 直接 AddNode，无去重。
  3) 资产血缘：SERIES 有 AnchorAssetIDs/LineageAnchorAssetIDs/LineageAnchorAssetIDsForNodes（递归 CTE 取上游资产），ClaimIntent 内用 intentAssetAllowedSQL 过滤未授权资产；ARTEX 无。
  4) 取消语义：SERIES StopIntentWithReason 把意图置 stopped + cancelled_by_user/cancel_reason/cancelled_at，并额外写一条 fact 记录取消原因；ARTEX 改为 SoftDeleteIntent（state='deleted' + delete_reason 字段），不写 fact。SERIES CancelIntent 只删直接 yields 且无其它父的 fact/finding；ARTEX CancelIntent 改为“独占子图级联删除”（沿 yields/derived_from 迭代，保护 goal/origin fact，共享节点保留）。
  5) 状态写：SERIES 的 SetIntentState/CompareAndSetIntentState/ResetRunningIntents 都带 cancelled_by_user 过滤；ARTEX 只在 ReopenIntent/ResetRunningIntents/ClaimIntent 过滤，SetNodeState/SetIntentState 允许改已取消意图（ReopenIntent 会清锁）。
  6) 查询能力：SERIES 有 HasRunningIntent/HasOpenIntent/WaitingForAssetApproval（资产审批等待）/ExecutionCountsContext/BlackboardRevision（对可见图+资产+范围做 md5）/FactsYieldedByIntents/FactsDerivedFromIntents/NodesPageContext(countOnly)/ActivityListContext/StatsContext；ARTEX 全无，只有 CountOpenIntents。
  7) 队列排序：SERIES Frontier 按 worker_queue_manual + queue_position 排序；ARTEX 只按 priority DESC,id。
  8) Node 结构：ARTEX 新增 DeleteReason 字段并进 nodeCols/scanNode；ARTEX NodesPage 搜索把 payload/origin ILIKE 与 id 精确匹配并列，SERIES 二选一且转义通配符。
- 风险：中（队列串行化、意图去重、资产授权过滤、审批等待、黑板版本号均缺失；取消/删除语义两边不同）

## db/intent_admission.go
- 差异：DiscardOpenIntent 从 beginQueue 改为 s.db.Begin，并新增 cancelled_by_user 过滤（不删已取消意图）。
- 风险：低

## db/intent_user_control.go
- 差异：SERIES 有 UserCancelledIntents、ReopenIntentByUser(id, expected)、RestoreUserReopen、ReopenBlockedIntentsByUser、ClaimPausedIntentByUser（均走队列锁并校验资产授权/任务状态）；ARTEX 只有 Node.UserCancelled 与 StopIntentWithReason（open/paused → stopped + cancelled_by_user/cancel_reason/cancelled_at + 清理副会话），没有“用户重新开启”的受控入口。
- 风险：中（用户侧重开/批量重开的受控路径缺失，只有通用 ReopenIntent 会清锁）

## db/llm_usage.go
- 差异：SERIES 的 LLMUsage 记录 intent_id、trigger、retry_ordinal、phase、system/message/tool_schema/request 字符数与估算输入 token，并有 idx_llm_usage_intent；ARTEX 精简为 task/exp/worker/model/latency/tokens，新增 JudgeUsageStats（worker='judge' 的累计与按日序列）。
- 风险：中（按意图/触发原因的计量与提示体积分析能力缺失）

## db/schema.sql（681 行 diff，已完整读完）
- SERIES 独有（ARTEX 无）：task_asset_links 的 tested/approval_state/approved_* 与 task_asset_blocks 墓碑表、authorization_host 生成列与 task_asset_* 授权函数族（host/identity/blocked/owner_approval/effective_approval/effectively_approved）、normalize_derived_task_asset_approvals、sync_task_asset_links 的 agent_discovery/user_registration 分支、first_test 回填、tasks.paused_at 与 execution_mode、worker 队列（worker_queue_manual/version、queue_position、deleted_workers、assign/bump 触发器）、worker_feedback(_delivery/_run) 与 notify_worker_feedback 触发器、finding_deletion_feedback/finding_suppression_rules、findings.report_detail、notification_xid 列、mcp_usage、tool_usage_migrations、tools.directory/usage_help/when_to_use、uq_sv3_value(md5)、uq_task_scope_v3(md5)、idx_act_worker_output、idx_findings_task_node、try_inet 快路径。
- ARTEX 独有（SERIES 无）：exploration_nodes.delete_reason 与 state='deleted'（约束与迁移）、asset_intercept_rules（7 种 kind）、task_intercept_rules（block/allow）、notification_channels/notification_events/notification_deliveries（实时+摘要、租约、batch）、task_templates.category_id 与 intercept_rules、schema 内 seed ScopeSentry MCP。
- 风险：中（资产审批授权模型整套缺失；ARTEX 新增拦截规则与通知队列属新增能力）

## db/task_assets.go（1035 行 diff）
- 差异：ARTEX 删除 23 个函数、无新增，全部属于资产审批/授权体系：ApprovalApproved/Pending/Revoked/Blocked 常量、TaskAssetApproval 结构、AssetKey/normalizeTaskAssetHost、ValidateTaskAssetsApproved、TaskAssetApprovalStates、FilterApprovedAssets、projectApprovedAssetHosts、TaskHostApprovalStates/ValidateTaskHostsApproved、RegisterAgentDiscoveredAsset/RegisterAgentAsset、rejectDerivedApprovalAssets、Approve/RevokeTaskAssets(Resolved)、lockTaskAssetApprovalRows、RunningIntentIDsForAssets、ListTaskAssetApprovals、taskAssetBlockInfo 等。attach 的 ON CONFLICT 也不再保护人工来源（manual/direct/company/api/task/legacy）。
- 风险：中（任务级资产审批、封禁、按主机投影、撤回时选中运行中 worker 的能力全部缺失）

## db/task_context.go
- 差异：SetIntentBlockedReason / ReopenIntentsByBlockedReason 从 queueExec 改为 s.db.Exec，并去掉 cancelled_by_user 过滤（ARTEX 可把已取消意图重新置 blocked/open）。
- 风险：中（用户取消锁可能被阻塞/重开路径覆盖）

## db/task_scope.go
- 差异：SERIES 手工（source=manual）登记范围时会在同一事务内写入 task_asset_grants、把命中资产挂到任务并置 approved、清除同名 task_asset_blocks 墓碑（含全局资产已删除的 host_key 场景）；ARTEX 只 upsert task_scope。覆盖度 SQL：SERIES 的 tested 统计同时含 task_asset_links.tested 与 fact/finding 锚点，且都过 task_asset_effectively_approved；ARTEX 只统计 fact 锚点、无审批过滤。
- 风险：中（手工范围不再自动授权资产；覆盖度口径变窄）

## db/task_templates.go
- 差异：ARTEX 新增 CategoryID 与 InterceptRules（模板保存任务级拦截/允许规则快照），TaskTemplatePatch 用 Set 标志区分“未传”与“清空”；SERIES 模板只有 name/description/goal。
- 风险：低（ARTEX 新增能力）

## db/tasks.go
- 差异：ARTEX 去掉 Task.ExecutionMode/AssetApprovalTemplate、TaskCreateOptions 的 AssetIDs 与 ApprovalTemplate 校验、insertTaskAssets（创建时直接挂资产并派生保守范围+授权）、seedUserAssetGrants；新增 TaskCreateOptions.InterceptRules 并在同事务写 task_intercept_rules。
- 风险：中（创建任务时选择资产/自动派生范围/资产审批模板能力缺失；新增任务级拦截规则）

## db/tools.go
- 差异：SERIES 的 Tool 有 Executable/Directory/UsageHelp/WhenToUse 与 shell 类型校验（ValidateShellCommand/ShellCommand/shellMetadata/shellExecutable），UpdateTool 在事务内取 advisory lock 并同步 context binding grants；ARTEX 删掉 shell 相关字段与校验，UpdateTool 单条 UPDATE，无授权同步。
- 风险：中（shell 自定义工具元数据/校验、工具授权同步缺失）

## db/tasks_test.go
- 差异：删除“用户提供资产必须 approved 且可执行”的断言。
- 风险：低

## db/tool_usage_test.go
- 差异：删除 TestToolUsageBackfill（历史回填）用例，只保留 ledger 用例。
- 风险：低

## db/task_scope_test.go
- 差异：删除 TestManualCompanyScopeApprovesExistingAssetsButKeywordDoesNot（手工企业范围应授权既有资产）。
- 风险：低

## db/assets.go
- 差异：SERIES 有 WithReadContext（请求级读上下文）、WithToolReadFields（模型侧只取必要列）、selectAssetColumns/query/queryRow 抽象、enableAgentDiscoveryMode（事务内把新关联标 pending）、taskAssetContextAssociationSQL/taskAssetSelectedLinkBoolSQL/taskAssetBlockedSQL（任务关联/审批/墓碑的规范 SQL）、QueryByTaskTested/QueryByTaskApproval；ARTEX 全无。ARTEX 的 linkHostAssets 从返回 error 改为 best-effort 无返回；各 Upsert* 不再写 agent_discovered 标记。
- 风险：中（模型侧字段裁剪、任务审批过滤、资产发现待审模式缺失）

## db/task_assets_context.go
- 差异：SERIES 有 CountHostsByTaskWithSources 与 taskHostContextCTE（含审批过滤/来源展开）；ARTEX 只保留 HostsByTaskWithSources，CTE 精简。
- 风险：中（主机计数与审批感知的上下文查询缺失）

## db/assets_test.go
- 差异：删除任务审批相关断言（QueryByTask 系列）。
- 风险：低

## db/activity_page_test.go
- 差异：ARTEX 用固定 summary "i" 造 25 条意图；SERIES 用带序号的 summary 并导入 fmt。属测试写法差异。
- 风险：低

## db/company_scope_test.go
- 差异：ARTEX 改用 t.Cleanup 注册 Close（注释解释 LIFO 顺序，避免清理落在已关闭连接上）并不吞清理错误；SERIES 是 defer d.Close() + `_,_=`。ICP 值从动态拆分改为固定串；IPv6 用例从随机 IPv6 改为固定 IPv4/24。无产品逻辑差异。
- 风险：低

## db/config_test.go
- 差异：SERIES 用时间戳唯一 profile 名；ARTEX 固定名（并发/残留时可能撞名）。
- 风险：低

## db/customtool_test.go
- 差异：删除 TestCustomToolShellMetadata 与 TestShellMetadataNormalization（对应 shell 元数据能力移除）。
- 风险：低

## db/exploration_sources_test.go
- 差异：删除“继承会话需终态意图”的 ActivitySearch/ActivityWindow/重开后拒绝读取断言（对应 ActivitySessionFilter.Inherited 移除）。
- 风险：低

## db/intent_user_control_test.go
- 差异：SERIES 的 TestUserCancellationLifecycle 覆盖：取消写锁、重复取消幂等、各种 SetIntentState/SetNodeState/CAS/Reopen/Claim/Reset/ReopenBlocked 均不能复活；ARTEX 换成 TestStopIntentWithReasonLocksClaim（需 ARTEX_PG_DSN，验证取消后 ClaimIntent=false、重开清锁）。
- 风险：低（ARTEX 用例覆盖面更窄）

## db/intercept_seed_test.go
- 差异：删除 TestDeleteEndpointSeedIsIdempotent（内置拦截规则 seed 幂等）。
- 风险：低

## db/llmretry_test.go
- 差异：SERIES 用 t.Cleanup(func(){d.Close()})；ARTEX 用 defer d.Close()（同上关闭顺序问题）。
- 风险：低

## db/settings_insert_test.go
- 差异：删除 TestInsertSettingIfAbsentConcurrent（并发首写竞争）。
- 风险：低

## db/task_archive_aggregate_stats.go
- 差异：单行差异（见 diff：无实质逻辑）。
- 风险：无

## db/task_archives.go
- 差异：删除 TaskArchiveBlocker/TaskArchiveBlockerContext（按源任务查活跃直接依赖）、remainingPausedTimeout（暂停时长顺延超时）、归档快照中与审批/队列相关字段的处理；其余为字段删减。
- 风险：中（暂停不烧超时的语义与按源任务的阻塞查询缺失）

## db/task_archives_restore.go
- 差异：删掉旧档案兼容填充（exploration_nodes.queue_position、task_asset_skips.observer_scopes、task_asset_blocks.block_kind、task_asset_links 的 tested/approval 字段）与 mcp_usage 序列推进；恢复时直接 json_populate_recordset。
- 风险：中（旧归档恢复可能因 NOT NULL 字段缺失失败；mcp_usage 序列未推进可能主键冲突）

## db/exploration_test.go
- 差异：删除 IntentFingerprint 归一化、AddIntent 只对活跃前沿去重、BlackboardRevision 随图/资产上下文变化四个用例。
- 风险：低

## db/task_list_performance_test.go
- 差异：删除 TestTaskListSeverityCountsAndDeletion。
- 风险：低

## db/task_assets_test.go
- 差异：删除 10 个审批相关用例（审批继承与墓碑、Agent 发现保持 pending、AssetKey 稳定性、重发现可重置关联、取消读上下文、分页、继承审批源选择、创建时直接资产独立测试态、缺资产回滚）。
- 风险：低

## db/task_archives_test.go
- 差异：删除 TestRemainingPausedTimeoutUsesPauseInstant 等归档/暂停用例。
- 风险：低

## server/assembly.go
- 差异：SERIES 有 mcpRuntimeCache（run 内复用连接/元数据）、meterMCPTool 带 RunInfo 归因、loadContextAccess + applyContextPolicy（按角色裁剪 context_search/get/graph_overview）、TaskDomainToolsFrom 任务域工具优先且缺任务域时注入 nil（不降级为全局资产）；ARTEX 每次 run 直连 MCP、closers 用 io.Closer、ToolResolve 返回无 error（读表失败放行代码默认）。wireTools 第 3 个可变参数 configRoot 被删。
- 风险：中（上下文能力裁剪、任务域工具隔离、MCP 归因缺失；读表失败时放行默认工具）

## server/assets.go
- 差异：SERIES 列表读加 20s 上下文、支持 tested/approval_state 过滤与任务级审批查询（CountDSLByTaskApproval/QueryByTaskApproval）、插入走 RegisterUserAssetWithSource；ARTEX 无超时/过滤，插入直接 Upsert + SetTaskAssetSource。
- 风险：中（资产页审批/测试筛选缺失）

## server/auth.go
- 差异：JWT subject SERIES→ARTEX；SERIES 校验 alg=HS256 且 subject=SERIES，token 提取要求单一 Authorization/X-Series-Token 头、cookie 名 series_token；ARTEX 只判断 "Bearer " 前缀、cookie 名 artex_token，无 subject/alg 约束。
- 风险：中（ARTEX 的 JWT 校验更宽松）

## server/broadcast.go
- 差异：SERIES 订阅者缓冲满时删除并 close(ch)，客户端按 last id 重连补洞；取消订阅有重复 close 保护。ARTEX 缓冲满时静默丢弃事件且不关闭 channel，删除订阅时统一 close(ch)。
- 风险：中（ARTEX 丢事件后客户端无法感知，可能永久缺一段活动流）

## server/chatupload.go
- 差异：SERIES 有 saveUpload/writeUpload（临时文件 + os.Link 原子落盘，避免截断/覆盖竞态）；ARTEX 无这两个函数。
- 风险：中（上传同名/中断时可能留下半截文件）

## server/conversations.go
- 差异：SERIES 在复测会话开跑前校验 FindingRetestForConversation / 原任务可用，并用 WithTaskPolicy 绑定任务资产策略、代理与 CA；ARTEX 无此段（复测会话不带任务资产策略）。
- 风险：中（复测会话缺少资产授权与代理上下文）

## server/customtool.go
- 差异：ARTEX 去掉 shell 类型（kind 只允许 command/script/http）、ValidateShellCommand 校验、testShellTool 分支、以及三处 customToolAssetPolicy（命令渲染前、script 代码、HTTP URL 与重定向）和 httpProxyTransport(ctx,...) 的 ctx 参数；参数解析错误被忽略。
- 风险：中（自定义工具不再做资产策略校验，HTTP 重定向可绕过；shell 自定义工具不可用）

## server/dto.go
- 差异：TaskDTO 去掉 ExecutionMode/AssetApprovalTemplate；FindingDTO 去掉 ReportDetail；TaskNodeDTO 新增 DeleteReason。
- 风险：低

## server/engine_timeout.go
- 差异：SERIES 的截止协调器在暂停时冻结倒计时（恢复时顺延 deadline）；ARTEX 按墙上时钟继续走。SERIES 收尾时持 workerControlMu 检查 ExecutionMode=manual（手工模式强制 met=false），终局规划前 beginPlannerRound 守卫并在 Plan 前注入 RunInfo{Trigger:"task_timeout"}、返回前检查 ctx 取消；ARTEX 全无。
- 风险：中（暂停会烧任务超时；手工模式仍会自动终局判定；缺 trigger 归因）

## server/finding_retests.go
- 差异：SERIES 的 get_finding_retest_context 支持 field/index/offset/max_chars 分段读取（ValidateToolDetailInput + ProjectToolDetail），并按 finding.TaskID 直接从 PG 读任务约束（任务未加载也可用），漏洞/任务不存在时报错；ARTEX 无参数、整包返回，约束只在 Manager 已加载任务时读取。
- 风险：中（大证据无法分段读；任务未加载时缺约束）

## server/finding_traffic.go
- 差异：finding_id/binding_id 参数从正整数校验（positiveIDParam）改为 strParam；提示语中 update_finding_report 的参数名说明不同。
- 风险：低

## server/finding_workflow.go
- 差异：SERIES 校验任务角色必须有 TaskID、用 FindingAccessMetadata 轻量读、并按 worker/mainagent/planner 分别 ValidateWorkerAssets / ValidateTaskAssetsApproved + TaskNodeAccess.CanRead；ARTEX 只校验任务归属与继承只读，无资产授权校验。提示语改为 list_task_findings/get_task_node_detail。
- 风险：中（流量证据读写不再做资产级授权）

## server/findings_groups.go
- 差异：findingFilterFromQuery 去掉 summary_only 参数。
- 风险：低

## server/goals_api.go
- 差异：SERIES 用 NotifyNodes("goal", ids)/NotifyNodes("goal_edited", ids) 唤醒 planner；ARTEX 改为 NotifyGoal(text)/NotifyGoalEdited(old,new)。
- 风险：低

## server/intent_stop.go
- 差异：SERIES 有 intentControlTask（校验 RunInfo 的 task/role/intentID）、KillWork（planner 终止等待中/运行中意图，含 CAS、controlWork kill、冲突重试、活动记录）；ARTEX 只剩 cancelIntentFromMain（任务操作闸门 + engine.StopIntent），文件 106 行内容被删。
- 风险：中（planner 终止意图的独立路径与角色校验缺失，需确认 killWorkTool 的实际落点）

## server/intercept.go
- 差异：ARTEX 新增 judgeWorkerLane 常量与 interceptJudgeUsage 接口（裁判模型消耗统计）；待审批列表/历史改为非 Context、非超时查询（ListPendingIntercepts/ListTaskInterceptsPage/ListAllInterceptsPage），不再支持按 task 过滤待审批。
- 风险：中（无请求取消/超时；待审批无法按任务过滤；新增裁判用量统计）

## server/platform_tools.go
- 差异：ARTEX 去掉 shell 工具的 executable/directory/usage_help/when_to_use 字段与说明，JSON 解析错误被忽略（_ = json.Unmarshal），create/update MCP 不再 normalizeMCPServer 校验；transport 描述改为 stdio | http / sse。
- 风险：中（参数错误静默；MCP 配置不再规范化校验）

## server/side_questions.go
- 差异：SERIES 在发问前检查 worker 意图是否 stopped（409 Worker 已取消），并在 runSide 中注入 RunInfo、FilterTaskAssetRequest、WithTaskAssetContext 做资产过滤；ARTEX 无发问前状态检查、无资产过滤，仅改提示语为“已删除”；环境变量 SERIES_BTW_MAX_OUTPUT_TOKENS → ARTEX_…
- 风险：中（旁问可绕过资产范围过滤，可能泄露未授权资产信息）

## server/manager.go（505 行 diff）
- SERIES 独有：Task.ExecutionMode/AssetApprovalTemplate、workerControlMu/plannerCancel/plannerGeneration/beginPlannerRound、workerWake 通道、trafficDiskMu 与 SetTrafficDiskSettings/TrafficSkipRules/TrafficRetentionDays/applyTrafficDiskSettings（静态资源跳过、保留天数）、smartProxyPool/smart（智能代理池）、ConfigureToolUA（curl/wget 默认 UA）、SetAssetPolicyStore、PlannerOverviewMode/Config、ContextPolicyConfig、TaskProxyAddr/TaskProxyCACert（任务级策略代理）、playwright MCP 名兼容回退。
- ARTEX：上述全部移除；syncPlaywrightMCPProxy → syncBrowserMCPProxy，MCP 名固定 "browser"；traffic 启动不再 SetRecordingEnabled/应用磁盘设置。
- 风险：中（执行模式、规划者取消、流量磁盘策略、智能代理、任务级代理凭据、态势/上下文策略配置全部缺失）

## server/mcpdiscover.go
- 差异：SERIES 有 mcpDiscoveryPhase/mcpDiscoveryError 阶段化错误、normalizeMCPServer（校验并规范化 transport 相关字段，清空不适用的 command/url）、discoverMCP（纯探测不落库，测试端点与缓存共用）；ARTEX 直接 discoverAndCacheMCP，无规范化与阶段错误。
- 风险：中（MCP 配置不规范/错误定位困难）

## server/task_control.go
- 差异：SERIES 的 applyTaskControl 支持 action=finish（走 finishTask）；暂停/恢复会 emitActivity；意图控制持 workerControlMu，cancel 走 StopIntentWithReason + 挂取消 fact + 等待注册/结算的循环，resume 支持 open 直接下发与 stopped(userCancelled) 重开（ReopenIntentByUser + admitTask + 失败回滚 RestoreUserReopen）。ARTEX 无 finish、无活动记录、无锁；cancel 改为 soft/hard 两种删除（soft=SoftDeleteIntent 置 deleted，hard=CancelIntent 独占子图物理删除），强制填删除原因，运行中先 ControlWork，并 NotifyCancelled(id, summary, reason)；resume 仅允许 paused→open 的 CAS。
- 风险：中（恢复语义收窄、用户取消后无法“重开”；删除语义整体改写，需关注前端对 mode/deleted 的适配）

## server/task_llm.go
- 差异：ARTEX 的 per-task agent 装配缺少：worker/planner/main 的 SetContextPolicy、planner 的 SetOverviewProfile/SetOverviewConfig、三者的 SetSmartProxy、worker 的 SetTrafficRecording、TaskProxyAddr/TaskProxyCACert（改用 m.ProxyAddr/ProxyCACert）、Compactor.SetAssetStore（冷节点压缩的资产授权校验）、以及 main.SetCancelIntent(s.cancelIntentFromMain)。
- 风险：中（context policy/态势投影/智能代理/任务级代理缺失；冷压缩不再做资产授权；**主 agent 的 cancel_intent 回调未装配，可能使该工具在 per-task 主 agent 上不可用**）

## server/task_assets.go
- 差异：删除资产审批模板接口（updateTaskAssetTemplate/SetAssetApprovalTemplate）、审批列表/批量批准/撤回/封禁接口（listTaskAssetApprovals/updateTaskAssetApprovals/mutateTaskAssetApprovals，含 group_by=host 分组、Resolved 接口、撤回后 CancelWorkersForAssets）、以及对应的错误映射（超时 504、blocked/not-approved 403）；taskAssetActor("SERIES") 一并删除。
- 风险：中（任务资产审批与撤回联动取消 worker 的能力整体缺失）

## server/tool_usage.go
- 差异：meteredTool 不再实现 PrepareInput（agent.PrepareToolInput 默认值注入/校验）与 AcceptsRawInput 透传。
- 风险：中（计量包装后的工具丢失入参默认值注入与原始输入能力）

## server/webui_embed.go
- 差异：ARTEX 新增 serveFileIfExists 辅助并用 http.ServeFileFS 提供 index.html；SERIES 直接调 serveFileIfExists（该函数在别处）。
- 风险：低

## server/orchestration.go（399 行 diff）
- 差异：
  1) 跨任务工具：SERIES 用 taskDelegatedTool 从基础工具动态生成 schema（参数/分页契约与任务内一致，只加 task_id）并做 graph_overview→get_task_graph 等改名；ARTEX 逐个手写 schema/描述，易与任务内工具漂移。
  2) update_finding_report：参数 finding_node_id → finding_id，且去掉了 FindingIDByNodeID 转换与 agentFindingTrafficAccess 写权限校验。
  3) 新增 cancel_intent 主机工具（仅 mainagent 绑定），走 cancelIntentFromMain。
  4) hostTools 不再把平台工具登记为 deferred。
  5) delegateToTask 不再 SetTaskID/SetFindingRecorder/SetNotifyNodes。
  6) refreshBuiltinToolSchemas 从 v9（逐工具 ResolveBuiltinTool + traffic.SeedToolMetas + 全量 domain 工具，失败重试）降为 v7（只刷 orchestration/platform + goal_met/insert_assets/list_facts），且不再处理失败重试。
  7) seed 迁移序列去掉 seedFindingSuppressionBindings，新增多项一次性 reseed（auto/planner 绑定、company scope、worker 读写解绑、goals/main/planner/worker 提示词）。
  8) reporter 触发消息：ARTEX 只迁移 V1→新文案，删掉 V2 冻结常量；reporter 工具绑定从 context_get/context_search/graph_overview 改为 get_task_node_detail/list_task_findings/get_task_worker_trace/list_task_worker_traces/search_task_worker_traces/get_task_graph。
- 风险：中（工具 schema 与任务内实现可能不一致；update_finding_report 少了写权限与节点→finding 解析；schema 刷新降级可能让旧库工具描述过期）

## server/engine.go（458 行 diff，已读完）
- 差异：
  1) 引擎持有全局 planner/worker（mu/snapshot/UseLLM/Mode）→ ARTEX 只保留任务解析器 + readiness 谓词，无全局对。
  2) 心跳跳过：SERIES 用 plannerRevision（BlackboardRevision）+ HasRunningIntent 判断“黑板无变化且无在跑 work”直接跳过 LLM；ARTEX 每轮照跑。
  3) 审批等待：SERIES plannerLoop 检查 WaitingForAssetApproval，有待审批且无已批资产时暂停规划；ARTEX 无。
  4) 手工模式：SERIES 在 plannerLoop/Run/终局规划多处检查 ExecutionMode=manual 并 beginPlannerRound 取消句柄；ARTEX 无执行模式概念。
  5) worker 领取：SERIES 用 ClaimNextWorker(predicate) + ValidateIntentAssets/ValidateIntentHosts（资产与主机授权）+ AssetPolicyHooksWithGuard + IntentAssetIDs；ARTEX 直接遍历 Frontier(20) 调 ClaimIntent，无授权校验。
  6) RunInfo：SERIES 注入 Trigger(claim/human_message/task_timeout)/RetryOrdinal；ARTEX 不注入。
  7) 完成判定：SERIES 用 completeAutomatically（admission 互斥 + guarded CAS）；ARTEX 直接 SetTaskStatus/SetTaskStatusGuarded。
  8) 取消：SERIES 有 CancelWorkersForAssets（撤回/删除资产时按 RunningIntentIDsForAssets 取消对应 worker 并写审计活动）、controlWork 支持 planner kill、AbortReason 记录；ARTEX 只有 ControlWork(pause/cancel) + KillWork(intentID) + StopIntent（主代理取消，15s 等待）。
  9) worker 循环：SERIES 用 workerWake 信号 + 空闲定时器（30s）精确唤醒；ARTEX 轮询（800ms/1000ms/1500ms）。
  10) 收尾门：ARTEX 在 steerHooks.PreToolUse 增加 SettlementBlocks 拦截 + PostToolUse 记录 host lesson；SERIES 无（其收尾门在 SDK settleGuard）。
- 风险：中（资产授权准入、审批等待、心跳跳过、手工模式、资产撤回联动取消 worker 均缺失；worker 改为轮询领取）

## server/server.go（973 行 diff，已读关键段落）
- 差异：
  1) 导出：ARTEX 保留 md-single（report.FindingsMarkdown，带流量证据）并新增 md-report/md-report-zip（提交报告），md-zip 走 buildFindingsEvidenceZip + evidenceStore().StageFindingsExport；SERIES 的 md-report 与 md-single 共用 FindingsReportMarkdown，md-zip 用 buildFindingsReportZip。临时目录前缀 series-→artex-。
  2) patchFinding：SERIES 用 db.FindingPatch 原子更新；ARTEX 拆成 SetFindingStatusWithNotify / SetFindingSeverity / SetFindingName / SetFindingVulnClass 四次独立写（状态走通知事务）。
  3) deleteFinding：SERIES 解析删除原因（严格 JSON、长度校验），DeleteFindingWithFeedback 写反馈+禁报规则，并 NotifyFindingDeleted 通知 planner；ARTEX 直接 DeleteFinding，无原因/无反馈/无通知。
  4) explorationNodes：SERIES 有 10s 超时与 count_only；ARTEX 去掉超时，新增 NodeAssets/EdgesTouching/NodesByIDs 组装 edges+refs+assets（播报板一次取全）。
  5) settingsPayload/putSettings：ARTEX 去掉 traffic_skip_static/types/retention_days、smart_proxy(_pool/_hosts)、planner_overview_mode/config、context_policy_config、shell_mode；新增 notify_enabled/notify_public_base_url/notify_digest_interval_min（含 URL 前缀与 1–1440 分钟校验）。
  6) 其他：buildPlannerWorker/profileConfig/providerForProfileContext/agentsForProfile/mainAgentRef 等按 profile 构建全局 planner/worker 的代码删除；getReport 新增；路由与 createTask 字段随执行模式/资产模板移除而变化。
- 风险：中（删除反馈与推送、若干设置项消失；导出语义与 ARTEX 新增通知配置需前端配合）

## server/server_mgmt.go
- 差异：删除 MCP 测试端点（pgTestMCP/mcpTestTool/decodeMCPBody）、MCP 批量导入（pgImportMCP/mcpImportServer.normalized/mcpImportRequest）、MCP 使用统计（pgMCPUsage）；skillRelPath/walkSkillFiles 的注释移动（ToSlash 逻辑保留）。
- 风险：中（MCP 测试/导入/用量能力缺失）

## server/assembly_test.go / core_test.go / engine_cancelcause_test.go / finding_traffic_test.go / finding_workflow_test.go
- 差异：模块路径改名；core_test 去掉 retestProvider 注入与 asset_counts 断言、stats 去掉 compact=1；engine_cancelcause 删除 controlWork("kill") 等结算等待与 AssetAuthorizationChangedCause 审计用例，改测 KillWork；finding_traffic_test 导出用例从 md-report-zip（提交报告）改为 md-zip（带证据），并删除 report/技术附录分割断言；finding_workflow_test 的 notifyFinding 去掉 assetIDs 变参与 list_task_findings 返回数组结构变化（不再是 {findings:[]} 分页包）。
- 风险：低（测试随实现变化；说明 list_task_findings 返回结构变了）

## server 其余测试与归档包
- auth_test.go：删除 TestJWTRequiresCurrentIdentityAndAlgorithm（JWT 身份/算法校验）与 TestAuthInitConcurrentCannotOverwrite（并发初始化防覆盖）。
- manager_delete_files_test.go / side_questions_test.go / chat_mentions_test.go / task_llm_test.go / task_metadata_test.go：模块路径、措辞与少量断言调整（side_questions 删除状态文案改为“已删除”）。
- skill_upload_test.go：TestWalkSkillFilesSlashPaths/TestSkillFileRoundTripNested 顺序调整，逻辑保留。
- task_archive_package.go(+test)：ARTEX 新增 TestTaskArchivePackageSkipsSymlink，打包时跳过符号链接（安全增强）。
- tool_usage_test.go：删除 TestToolCatalogUsageResponse（工具目录调用量统计）。
- tools_wire_test.go：删除 baseDescription/mustToolRows 辅助与 TaskDomainTools 注入相关断言（对应 wireTools 去掉 configRoot/任务域注入）。
- customtool_test.go：删除 TestPlatformShellQuoting（shell 类型移除），新增 TestHTTPToolTruncatesLargeBody（大响应截断）。
- 风险：低

## traffic/traffic.go
- 差异：SERIES 有 blob 内存缓存（8MiB）、SetAssetPolicyStore（代理出口做任务资产授权）、SetRecordingEnabled（抓包开关不关代理，保留授权兜底）、静态资源过滤（DefaultSkipStaticRules/SetSkipStatic/SetSkipRules/shouldSkipStatic）、保留期清扫（SetRetention/retentionLoop）；ARTEX 全无，traffic_search 描述也不同（host 必填 vs 主机检索）。
- 风险：中（抓包不再做资产授权兜底、无静态过滤与自动清理，磁盘会持续增长）

## traffic/evidence.go
- 差异：读取证据时 SERIES 对 req/resp 做 decompressInline 解压；ARTEX 直接返回原始（可能压缩的）字节。
- 风险：中（模型读到的响应体可能是压缩后的二进制）

## traffic/archive.go
- 差异：SERIES 归档 blob 为 zst 时解码、导入时按需压缩回 .zst、FTS 索引前对 inline body 解压并按 Content-Encoding 解码；ARTEX 一律 .bin 明文、不解压。
- 风险：中（压缩体进入 FTS 索引会搜不到内容）

## traffic 测试（traffic_test/content_encoding_test/reclaim_test/traffic_store_test）
- 差异：删除任务代理凭据用例（TestProxyTaskID）、readTrafficPart 解压断言、不可压缩随机体（压缩相关）、.zst 落盘断言改为 .bin 且大小等于原文。
- 风险：低

## evidence/store.go(+test)
- 差异：SERIES 在 Record 事务开头做 ValidateFindingSuppressionTx（禁报规则锁），把资产/任务锁推迟到 RecordFindingTx；ARTEX 改为 LockTaskEvidenceTx，并给 RecordFindingTx 传 ctx（用于推送事件写入）。测试为模块路径与调用签名调整。
- 风险：低（与禁报规则已回滚一致）

## llmrec/llmrec.go(+test)
- 差异：SERIES 有 RunInfo（trigger/retry/phase 等归因）与 WithRunInfo/WithIndependentRunInfo/RunInfoFrom，recordUsage 记录这些维度；ARTEX 删除 RunInfo，新增 WithWorker/workerFrom（显式 agent 泳道覆盖，用于 judge 计量）并在 recordUsage 中去掉 runInfo 参数；测试删除归因合并与请求维度用例。
- 风险：中（按 trigger/phase/retry 的归因统计缺失；新增裁判泳道）

## intercept/prompt.go
- 差异：EffectiveJudgePrompt 的默认提示文本有调整（未改逻辑分支）。
- 风险：低

## mcphttp/client.go
- 差异：初始化/关闭处少量清理（有 4 行删除与 1 行调整），无协议行为改动。
- 风险：低

## report/findings.go
- 差异：ARTEX 保留 FindingFilename/FindingReportTitle/FindingReportFilename/SingleFindingReportMarkdown/FindingsReportMarkdown（本次移植新增）；CSV 列不同：SERIES 导出「漏洞名称/等级/类型/状态/任务/时间/提交报告」，ARTEX 导出「名称/类别/严重等级/状态/任务/时间/概述/流量证据数量/流量证据ID」，且 ARTEX CSV 不含报告全文。
- 风险：低（导出列语义不同，使用方需注意）

## 根目录与脚本
- go.mod：模块路径 series→artex；norma 从 v0.4.5-0.20261017151310（伪版本）回退到 v0.4.3（**依赖版本回退**）；brotli 改直接依赖；去掉 mvdan.cc/sh/v3；新增 kr/pretty。go.sum 相应变化。
- docker-compose.yml：服务名 series→artex，POSTGRES_USER/DB 默认 series→artex，环境变量 SERIES_*→ARTEX_*，镜像从本地构建（series:local + pull_policy: never）改为拉取 autumn27/artex:${ARTEX_TAG:-latest}（**部署方式变化**）。
- Dockerfile：二进制名 series→artex，去掉 @zhafron/mcp-web-search 与 chrome-devtools-mcp 的全局安装，注释改为 browser MCP。
- start.sh：ARTEX 新增退出码 75（请求重启）分支并在注释说明校验/换装交给 Go 的 selfupdate；其余为重命名。
- update.sh：Docker 更新改为询问 tag + docker compose pull artex + up -d artex（不再本地构建、不再 stop 服务）；本地更新编译 ./cmd/artex。
- install.sh：重命名与变量名替换（SERIES_*→ARTEX_*）。
- reset-password.sh：重命名；**删除了密码长度校验（≥8 位且 ≤72 字节）与事务包裹**，只保留 pgcrypto upsert。
- stop/其他脚本：重命名。
- 风险：中（norma 版本回退、compose 改拉镜像、重置密码弱校验）

## web/src/lib/types.ts
- 差异：删除资产审批/执行模板/任务资产审批/资产来源/任务发现筛选/PlannerOverviewLimits 等类型；新增 WorkspaceEntry/Listing/File（工作目录文件管理）、TaskTemplate.category_id 与 intercept_rules、TaskNode.delete_reason、ExplorationNodesPage.assets（播报板节点锚定资产）、SessionStatus 增 "deleted"。
- 风险：中（前端不再有资产审批模型；新增工作区与任务拦截规则类型）

## web/src/lib/api.ts
- 差异：令牌机制 series_token + X-Series-Token + entryAwareFetch → artex_token + Authorization Bearer + 原生 fetch；删除 taskNotifications/toolCalls/toolCallDetail、资产审批（taskAssetApprovals/updateTaskAssetTemplate/…）、执行模式（setExecutionMode）、人工派发（dispatchIntents）、worker 队列（workerQueue/moveWorker）、MCP 测试/导入/用量、SmartProxyHost、HTTPAuthSettings、GlobalProxyProbeResult；新增 JudgeUsage、NotificationChannel/Delivery/Filter/Meta、SSProject/SSTask、UpdateCheck、Workspace*、AssetInterceptRule 与任务拦截规则 CRUD、taskAssets/searchTaskAssets 去掉 tested/approval 参数、controlIntent 增加 mode=soft|hard 且返回 state 可为 "deleted"/""、stats 去掉 compact=1、task 创建用 intercept_rules 取代 asset_approval_template/asset_ids、模板 CRUD 支持 category_id/intercept_rules。
- 风险：中（前端能力与后端同步削减/新增；Bearer 与 cookie 名变化，需与后端一致）

## web/src/components/ui/*（badge/button/calendar/chart/checkbox/dialog/input-group/native-select/select/sheet/table/toggle/toggle-group/tooltip）
- 差异：纯样式调整（圆角 md→[2px]/[4px]/calc(var(--radius)-n)、hover 透明度 90→80、遮罩 bg-black/55→/10 并支持 backdrop-blur、表格字号 13px→text-sm、移除行高 h-10、badge 去掉 notification 变体）。无逻辑变化。
- 风险：无

## web/src/app/(auth)/login/page.tsx
- 差异：ARTEX 新增《使用须知》弹窗（滚动到底才可同意、未勾选禁止登录）、左右分栏登录页（左 logo、右表单）、登录后跳 /function/tasks；用户名 SERIES→ARTEX；删除 HTTPEntryAuthError 分支，错误统一提示“用户名或密码错误/无法连接”。
- 风险：低

## web/src/app/(auth)/setup/page.tsx
- 差异：初始化密码页改为左右分栏布局与 /logo.png 展示，跳转改 /function/tasks，文案 SERIES→ARTEX；错误展示精简。
- 风险：低

## web/src/app/(external)/page.tsx
- 差异：根路径重定向 /function/tasks/new → /function/tasks。
- 风险：低

## web/src/app/(main)/_components/main-content.tsx
- 差异：ARTEX 新增全局头部（AccountSwitcher/ThemeSwitcher/LayoutControls/SearchDialog/UpdateBadge、版本号来自 /health），并按路径判断全出血页（/function/tasks/* 详情）；SERIES 只在 /function/tasks/detail 时去掉头部。
- 风险：低

## web/src/app/(main)/layout.tsx
- 差异：ARTEX 从 cookie 读 sidebar_variant/collapsible（静态导出可用的客户端方案），sidebar 宽度改为 calc(var(--spacing)*68)，移除 SettingsWorkspace 包裹与 NewTaskShortcut。
- 风险：低

## web sidebar（app-sidebar / layout-controls / nav-main / sidebar-items）
- 差异：app-sidebar 改用 zustand preferences store（sidebarVariant/collapsible/isSynced），新增 SidebarSupportCard 与 link，移除“设置/外观”菜单与 useSettings；layout-controls 从 settings 弹窗改为 Popover 内的主题/布局/字体选择（applyThemePreset/applyContentLayout/persistPreference 等）；nav-main 去掉 tasks/findings 详情页的特殊高亮、⌘J 提示与移动端关闭；sidebar-items 重排为 功能（dashboard/chat/traffic/commands/llm-records/sync/workspace）+ 系统（llm/agents/mcp/skills/tools/notify/intercept/asset-intercept/approvals/settings），新增 Ban/BellRing/Bot/Brain/FolderOpen/FolderSync 等图标。
- 风险：低（导航结构变化；新增通知推送/资产拦截/工作空间入口）

## web chat/page.tsx
- 差异：删除 SessionToolCalls 工具调用面板与 toolCallRevision；localStorage 键 series.chat.collapsed-agents→artex.*；输入区与容器样式调整（去掉 max-w-960 外框、圆角 xl→lg、高度按 4rem 头部计算）。
- 风险：低（前端不再有会话工具调用列表）

## web dashboard/page.tsx
- 差异：纯样式（筛选按钮选中态 border→bg-foreground/text-background、去掉 border 类）。
- 风险：无

## web function/findings/page.tsx
- 差异：重构列表页：移除 FindingWorkspaceDetail/FindingPagination/FindingLayoutToggle/FindingPresentation/FindingSelector/FindingsToolbar，改为 findings-table 组件 + TablePagination + Tabs/ToggleGroup/Select/InputGroup 内联筛选；默认排序 time→severity、每页 50→20；新增复测对话框 FindingRetestDialog 与 activeRetests 拉取；localStorage 键 series_finding_list_preferences→artex_*；删除 groupsSnapshot/groupsError 逻辑。
- 风险：中（页面行为与布局整体改写，依赖新组件 findings-table/table-pagination）

## web function/findings/detail/page.tsx
- 差异：详情页从复用 FindingDetailContent 改为自建布局：读 query id/context_task、内联严重度与状态下拉（setFindingSeverity/setFindingStatus）、右侧状态面板 FieldRow、Tabs（概览/流量/复测）、FindingTrafficPanel、FindingRetestPanel、Markdown 渲染、CopyButton、toast 提示；删除 router 与 revision 重载逻辑。
- 风险：低

## web function/tasks/page.tsx
- 差异：ARTEX 新增 CreateTaskSheet（内嵌建任务：上传附件并把绝对路径清单追加进描述、资产拦截规则编辑、任务模板、LLM 配置链、来源任务/企业 Combobox，MAX_SOURCE_TASKS=8），列表页新增“目标”列；移除任务未读通知（TaskUnreadBadge/useTaskNotifications）与 TaskPreviewSheet 预览抽屉；localStorage 键 series_task_list_*→artex_task_list_*；搜索占位文案去掉“名称”。
- 风险：中（新建任务流程整体改为抽屉；未读漏洞提醒移除）

## web tasks/detail/page.tsx
- 差异：Tabs 从 sessions/assets/asset-approvals/broadcast/findings/graph/coverage-graph/intercept/smart-proxy 等改为含 retests（复测）与 report（报告），删除 asset-approvals 与 smart-proxy；子页从 dynamic 懒加载改为静态导入；移除 useTaskNotifications/NotificationReadContext/未读 capture 与 mergeTaskStats；默认 tab 与 URL tab 解析逻辑调整。
- 风险：中（资产审批/智能代理页签移除；通知已读机制移除）

## web tasks/detail/_tabs/overview-tab.tsx
- 差异：新增 TaskInterceptRulesCard（任务级拦截/允许规则的列表、新增、行内编辑、删除、启用开关，7 种 kind：exact/fuzzy domain/ip/url 与 cidr），插在总览页。
- 风险：低

## web tasks/detail/_tabs/sessions-tab.tsx
- 差异：移除 WorkerQueue 队列视图、SessionToolCalls、批量选择 Checkbox 与 SourceTranscript、worker-feedback 的 lastMainActivity、router/searchParams 相关深链；新增 deleted 状态（CircleSlash/Trash2 图标、删除线、“已删除”徽章），删除动作改为弹窗填写原因并选择 soft/hard，恢复按钮从“已取消执行可重开”改为仅暂停可恢复；SessionsTab 入参去掉 pendingIntercepts，改用 useApprovalFocus({taskId})。
- 风险：中（队列/批量/来源记录能力移除；取消→删除语义变化需与后端一致）

## web broadcast-tab.tsx
- 差异：新增 AssetList（节点锚定资产：类型标签+可辨识文本）与 RelatedNodeCard（上下游节点悬停名片：类型/状态/来源/时间/摘要/资产），数据来自本页返回的 refs/assets，不再额外请求；移除 startPolling 轮询与 summaryOf 的 "summary" 回退字段。
- 风险：低

## web findings-tab.tsx（任务详情内）
- 差异：整体重写：移除 FindingsToolbar/FindingPagination/FindingSelector/FindingPresentation/FindingSuppressionRules/导出按钮/通知已读，改为轻量列表（排序仅 time、localStorage 键 artex_task_findings_sort）、状态徽章、卡片/行内展开、跳转详情；不再有批量选择与导出入口。
- 风险：中（任务内不再支持筛选/分页/批量导出/禁报规则入口）

## web assets-tab.tsx
- 差异：删除资产审批徽章逻辑（blocked/pending/revoked/derived 继承）、来源任务只读徽章、已测试/未测试徽章与提示；范围输入说明改为“域名和 IP 创建或复用全局资产；CIDR/ICP/关键词作为上下文”。
- 风险：中（任务资产页不再展示审批/测试状态）

## web system 七个页面（intercept/llm/mcp/agents/skills/tools/settings）
- 差异：SERIES 每个页面只有一行 `export { default } from "./_components/xxx-settings"`（配置界面在共享 SettingsWorkspace 组件里）；ARTEX 改为各自独立整页实现：
  - intercept：命令拦截（rules）与裁判模型（judge）两 Tab，含 JudgeCard/JudgeStat/JudgeSparkbars（裁判用量与按日条形）、按 DB 工具分组、动作徽章。
  - llm：profiles 与 retry 两 Tab，含 PoolSheet（池成员/健康/冷却）、ProfileSheet（编辑档案）。
  - mcp：config/tools 两 Tab，服务器卡片 + 工具列表。
  - agents：Agent 清单卡片网格 + 新建对话框。
  - skills：技能库总览 + 选中技能的 basic/mcp/visibility 三 Tab、文件树（buildTree）与上传。
  - tools：系统工具/自定义工具两 Tab、参数表编辑（toRows/applyRows/coerceDefault）、自定义工具对话框。
  - settings：整页系统配置（流量捕获、全局出口代理、搜索来源、操作约束注入、通知等分组）。
- 风险：中（配置 UI 从共享弹窗改为独立页面，交互与可用项随后端能力变化）

## web 组件与 lib（task-template-controls / approval-execution-focus / approval-records / finding-traffic-panel / scope-text-editor / company-scope / side-questions / status / task-assets）
- task-template-controls：模板草稿新增 categoryID 与 interceptRules（AssetInterceptRulesEditor、NativeSelect 选分类），seed 类型扩展，保存/编辑带入新字段。
- approval-execution-focus：去掉 useSearchParams 与 enabled 开关，改读 window.location.search（静态导出兼容）；replaceState 传 null state。
- approval-records：去掉通知已读（useNotificationRead）与 startPolling 轮询/AbortSignal，改为单次手动刷新 + 并行拉取待审批与历史（不再按 signal 取消）。
- finding-traffic-panel：文案“流量证据”→“关联流量”，样式类顺序调整；删除响应格式校验。
- scope-text-editor：keyword 文案“原文信息”→“关键词”，说明改为自动识别域名/IP/CIDR/ICP/关键词。
- lib/company-scope.ts：最大长度 65536→1024（与后端一致）；CIDR 前缀收紧为 IPv4 /16–/32、IPv6 /32–/128；URL/证书不再降级为 keyword 而是报错；无效 IP 直接报错；删除证书多行合并逻辑。
- lib/side-questions.ts：删除 parseSideHistory/parseSideExchange 的严格校验（响应格式错误不再抛错）。
- lib/status.ts：新增 deleted 状态（已删除/slate）与推送投递状态表（pending/sending/sent/failed/skipped）。
- lib/task-assets.ts：来源标签 api“资产 API”、company“企业关联”、manual“人工加入”、task“任务初始化”（原统一“用户提供”）。
- 风险：中（公司范围可登记长度与 CIDR 下限收紧、URL/证书不再接受；侧问响应校验移除；新增删除/投递状态）

## web transcript.tsx / agent-editor.tsx / fonts/registry.ts / globals.css
- transcript：工具调用配对 key 从 `${worker}:${intent}:${seg}:${tool_use_id}` 改为仅 `tool_use_id`（跨 worker 同名 id 可能错配）；import 顺序调整；其余为格式。
- agent-editor：移除 useSettingsDraft 依赖，改为独立编辑器（Tabs/Switch/Select/Textarea/Input 等重排），保留预览/对比/重置/删除；Markdown 预览引入。
- fonts/registry：SERIES 用 localFont 指向 node_modules/geist 的 Geist-Variable/GeistPixel-Square（六个字体别名共用同文件）；ARTEX 改用 `geist/font/pixel` 的 GeistPixelSquare（不再 preload 全部像素字重）。
- globals.css：圆角改为相对 `--radius`（0.625rem）推导（sm=radius-4 … 4xl=radius+16）；新增 brutalist/soft-pop/tangerine 三套主题预设 @import；调整 background/sidebar/sidebar-primary 色值；body 去掉固定 text-sm。
- 风险：低（transcript key 变化可能影响工具结果与调用的配对展示）

## web 其余（mock / preferences / proxy / logs / retry / 根 layout / package / logo / theme-boot）
- lib/mock/handler.ts：删除 LLM 重试默认值、finding 删除反馈与禁报规则 mock；新增 TrafficEvidenceRole 等类型适配。
- lib/mock/data.ts：intents 改为从 explorationGraph 派生；新增 fiRce/fiPriv/fiJenkins 资产映射、nodeAssetsFor、transcripts、workspaceList/workspaceRead。
- lib/preferences/preferences-config.ts：默认 sidebar_variant "sidebar"→"floating"；theme_preset/content_layout/navbar_style/sidebar_variant/sidebar_collapsible 的存储位置改为 "client-cookie"。
- proxy.ts：middleware matcher 从多行列表简化为单条正则（排除静态资源与 api）。
- hooks/use-side-questions.ts：SSE 数据不再经 parseSideExchange 校验，直接 as SideExchange。
- system/logs：entryAwareFetch→fetch，示例日志文本 SERIES→ARTEX，刷新图标尺寸调整。
- system/llm/_components/retry.tsx：去掉 useSettingsDraft 与 draftVersion/resetDraft。
- app/layout.tsx：html lang "zh-CN"→"en"。
- package.json：包名 series-web→artex-web；删除 test:ui 脚本。
- public/logo.svg：替换为新的 ARTEX 品牌矢量图（大文件）。
- scripts/theme-boot.tsx：主题/布局引导改为 readPreference 读取（内容/导航/侧栏变体/折叠）。
- 风险：低（lang 改 en 与中文界面不一致；mock 与偏好默认值变化）

## 补记：逐文件索引（供最终汇总按文件取用）
- .gitignore：改为忽略 data/*.db/*.db-wal/*.db-shm/config.json/scheduler_state.json/__pycache__/web/node_modules；移除 /dist/ 与 /server/webui/dist/ 忽略。
- CHANGELOG.md：品牌改写并加入 0.3.15 条目（拦截、漏洞推送、安全审计、测试基础设施、流量）。
- README.md：标题/截图/在线 Demo 链接改为 ARTEX；删除旧品牌截图说明。
- agent/context_prompt_migration.go：仅保留 reporter 旧提示词迁移，删除 PreviousPlannerProactivePromptHashes，补注释。
- agent/retester.go：复测提示词去掉分段续读说明，改为一次性读取关联漏洞/证据/约束。
- agent/terminalreason.go：模型错误说明从 context_search 改为 get_worker_trace。
- evidence/store_test.go：模块路径、SERIES_PG_DSN→ARTEX_PG_DSN、SetFindingReportVersionByNodeID 去掉 detail 参数。
- llmrec/llmrec_test.go：删除 RunInfo 合并与请求维度（system/message/tool chars）用例。
- server/finding_traffic_export.go：模块路径改名；buildFindingsReportZip 注释改为中文说明（逻辑未变）。
- server/task_archive_package_test.go：新增 TestTaskArchivePackageSkipsSymlink。
- start.bat：新增退出码 75（请求重启）分支与说明，二进制名 series.exe→artex.exe。
- traffic/content_encoding_test.go：删除 readTrafficPart 解压断言。
- traffic/reclaim_test.go：bulkRecord 用重复字符替代不可压缩随机体（因不再压缩）。
- traffic/traffic_store_test.go：blob 断言从 .zst 且更小改为 .bin 且等于原文。
- web/app-sidebar.tsx：改用 zustand 偏好 store、新增支持卡片、移除设置/外观菜单（见 sidebar 组）。
- web/layout-controls.tsx：设置弹窗改为 Popover 主题/布局/字体控制（见 sidebar 组）。
- web/nav-main.tsx：去掉详情页特殊高亮、⌘J 提示、移动端关闭（见 sidebar 组）。
- web/approval-execution-focus.tsx：改读 window.location.search，去掉 enabled 参数（见组件组）。
- web/approval-records.tsx：去掉通知已读与轮询/AbortSignal（见组件组）。
- web/finding-traffic-panel.tsx：文案与样式调整，删除响应格式校验（见组件组）。
- web/scope-text-editor.tsx：关键词与范围说明文案更新（见组件组）。
- web/task-template-controls.tsx：模板新增 categoryID/interceptRules（见组件组）。
- web/ui/badge.tsx、button.tsx、calendar.tsx、chart.tsx、checkbox.tsx、dialog.tsx、input-group.tsx、native-select.tsx、select.tsx、sheet.tsx、table.tsx、toggle-group.tsx、toggle.tsx、tooltip.tsx：纯样式（圆角/遮罩/字号/透明度）。
- web/navigation/sidebar/sidebar-items.ts：导航分组重排并新增工作空间/资产同步/通知推送/资产拦截入口（见 sidebar 组）。
- web/styles/presets/brutalist.css：--radius 5px→0。
- web/styles/presets/soft-pop.css：--radius 5px→1rem。
- web/styles/presets/tangerine.css：--radius 5px→0.625rem。

## web/src/components/exploration-graph.tsx（补记：此前被 git 当二进制而漏读，实际为逻辑差异）
- 位置：文件整体
- 差异：ARTEX 把 digest 折叠逻辑内联进本文件（新增 collapseDigestGraph：按 covers 边把成员折到 digest 超节点、重连边、去重、丢弃反向 derived_from），SERIES 则在 lib/exploration-graph.ts 中提供；同时 typeMeta 由多行对象压缩为单行、若干 forEach 改写为表达式、Prettier 风格重排。文件中含 2 个真实 NUL 字节（模板 key `${src}\0${rel}\0${dst}`），导致 git 把它识别为二进制、diff/blame 失效。
- 风险：中（逻辑差异未在 229 统计内；源码含 NUL 会影响 git diff/部分工具链）


## 第二部分：单边文件（566 + 94）

# 单边文件语义核对记录（SERIES 独有 566 + ARTEX 独有 94）

方法：逐文件读取（源码按文件头注释、导出符号与关键实现；测试按用例名与断言；文档按标题与首段）；小文件全文读取。结论按目录分组。

## agent/（105 个，SERIES 独有，14,036 行）

### 资产上下文与授权
- agent/asset_context.go（403）：资产上下文提供者——对每次模型请求重校验持久化的结构化工具结果；按任务授权过滤/重写资产字段与 scope；WithTaskAssetContext/FilterTaskAssetRequest 供聊天/侧问等路径使用；authorizedScope 过滤范围行。
- agent/asset_context_test.go（117）：验证 pending 资产不会出现在模型视图、也不会阻止同主机其它工作。
- agent/asset_management_context.go（79）：管理类工具结果（审批历史）每次请求重校验身份，保持历史快照语义。
- agent/asset_origin_test.go（72）：insert_assets 调用来源归因测试。
- agent/asset_skip_scope_test.go（80）：跳过提示的范围与内容稳定性测试。
- agent/asset_tools.go（177）：asset_search / asset_get 工具（选择器去重、总预算与续读）。
- agent/assets_approval_dsl_test.go（33）：list_assets 的 approval DSL 测试。
- agent/main_execution.go（69）：主代理直接执行工具的角色/资产校验（validatePlanningAssets/Hosts、mainRoleTool、PrepareInput）。
- agent/main_execution_test.go（108）：主代理 pending 执行的角色与上下文测试。
- agent/worker_execution.go（33）：worker 写回时的内部可见性掩码、validateResultAssets、markResultAssetsTested。
- agent/worker_execution_test.go（209）：worker pending 执行保持规划者边界。
- agent/worker_context.go（28）：worker 启动上下文构造。
- agent/target_access_context.go（243）：target access 快照与按请求刷新（check_target_access 的上下文侧）。
- agent/target_access_test.go（232）：target access 入参/历史/审批生命周期/上限。
- agent/tools_target_access.go（128）：targetAccessQuery/Asset/Node、视图与解码。
- agent/structured_capture.go（101）：延迟工具结果的结构化捕获与再校验。
- agent/structured_capture_test.go（140）：捕获有界、ID 全保留或整体丢弃预览。
- agent/structured_history.go（43）：SDK 压缩后无法校验的结构化历史做脱敏。
- agent/structured_history_test.go（48）：不可校验历史仍保留可用上下文。
- agent/tool_read_scope.go（59）：请求级 overview 读取范围与审批状态读取，不跨请求复用。
- agent/tools_asset_view.go（41）：task 资产视图查询解码与 listTaskAssets。
- agent/tools_asset_view_test.go（115）：任务资产管理工具边界测试。
- agent/worker_trace_tools_test.go（33）：worker 定向 trace 工具测试。
- agent/local_tools.go（13）：workerLocalTools = DefaultTools 去掉 Sleep。

### 上下文工程（检索/策略/预算/压缩）
- agent/context_budget.go（32）：按上下文预算裁剪输出（readOutputBudget/readContentBudget/contextJSON）。
- agent/context_compat_test.go（220）：旧记录不得进入已打开的检索；context_get 预算与续读；混合检索权限与游标。
- agent/context_match_read_test.go（122）：检索命中不等于详情授权。
- agent/context_mixed.go（42）：searchMixedContext 混合实体检索。
- agent/context_performance_test.go（111）：检索场景 SQL 计数与指标。
- agent/context_policy.go（347）：上下文两条线（工具清理线/摘要线）的可配置绝对值、默认值、归一与校验。
- agent/context_policy_test.go（198）：策略归一/校验、窗口推导、清理旧工具结果、仅超线清理。
- agent/context_prompt_measurement_test.go（21）：上下文提示体积指标。
- agent/context_review_test.go（253）：历史证据不被当作平台 ID；资产读取需可信 worker 上下文；worker 父上下文续读；预加载继承资产不拉整图。
- agent/context_task.go（57）：contextTaskID 双表示解析、跨任务 schema。
- agent/context_task_test.go（126）：模型提供的 ID/RunInfo 不能自建本地任务绑定。
- agent/context_tools.go（432）：ContextAccess 由宿主持有；context_search/context_get 的权限、目标、worker 读取边界。
- agent/context_tools_test.go（392）：读取授权/游标、能力边界、历史去重、热投影与压缩图一致、输出与 digest 视图、资产投影与流量刷新。
- agent/context_window_metrics_test.go（106）：上下文窗口检索指标。
- agent/retrieval_history.go（343）：检索历史窗口（copy-on-write 投影、去重、目录上限、查询分组）。
- agent/retrieval_history_test.go（244）：窗口与审计、超大/部分失败/延迟、目录有界且不含密钥、重读变最新、非文本与写结果不动。
- agent/retrieval_noa_test.go（108）：noa 与 provider 路径下的检索。
- agent/retrieval_access_test.go（104）：检索窗口刷新授权并可恢复。
- agent/tool_budget.go（263）：ProjectToolDetail 无损续读协议、文本窗口、输入解码与校验。
- agent/tool_budget_test.go（144）：Unicode/集合续读、历史配对、非法入参、工具定义与历史体积。
- agent/tool_history.go（198）：只对无副作用读取做去重与归一（cloneToolMessages/deduplicateToolHistory）。
- agent/tool_input.go（92）：业务契约边界（normalizeToolSchema/validateBusinessInput/PrepareToolInput）。
- agent/tool_input_test.go（23）：枚举错误展示允许值且不回显输入。
- agent/tool_defaults.go（76）：按 schema 注入标量默认值（missingOnly 语义）。
- agent/tool_detail_contract_test.go（108）：详情输入校验、无损文本预算、集合与延迟字段。
- agent/tool_contract.go（102）：ResolveBuiltinTool 保持结构由代码持有、合并持久化元数据、requireRoleTools。
- agent/tool_contract_test.go（75）：内置 schema 不可被库中版本替换、事实批量先解码后写、必需角色工具、资产查询契约。
- agent/tool_error_guidance_test.go（40）：资产查询错误指引与选择器透传。
- agent/tool_documentation_test.go（149）：内置提示文档不含已退役工具名、批量契约等。
- agent/tool_inventory_test.go（57）：内置工具清单契约。
- agent/tool_surface_metrics_test.go（55）：工具面与角色工具面指标。
- agent/compaction_review_test.go（61）：fold block 丢弃已变更来源。
- agent/lightformat.go（213）：轻量标签格式渲染提示上下文（稳定、有界、转义）。
- agent/lightformat_test.go（50）：渲染稳定性/优先级/嵌套去重。
- agent/planner_bootstrap.go（184）：规划者启动引导（SetTaskContextReadAccess/refreshPlannerBootstrap）。
- agent/planner_bootstrap_test.go（104）：引导预算与请求重校验、自定义提示保留使命、恢复部分引导不泄漏。
- agent/planner_context_metrics_test.go（90）：规划者按需上下文指标。
- agent/planner_proactive_test.go（247）：主动查询不作为准入证明、提示替换门禁、graph_overview 按需文档、无检索时权限保持。
- agent/planner_scenario_metrics_test.go（116）：规划场景指标（对照改前版本）。
- agent/planner_user_cancel_test.go（44）：规划者取消在 trigger drain 后仍生效。
- agent/planner_asset_prompt_test.go（50）：规划者资产提示按实际绑定，覆盖度关闭时不给不可用选择器。

### 态势投影（graph overview）
- agent/overview_config.go（183）：OverviewLimits/Config/Focus/LeanOptions/RenderOptions 与默认值。
- agent/overview_lean.go（484）：leanGraphOverviewData 有损精简投影（行裁剪、文本精简、目标/意图行）。
- agent/overview_lean_test.go（494）：精简投影、unavailable 保留、预算与分节、默认 profile。
- agent/overview_projection.go（287）：OverviewProfile 选择投影量、boundedGraphOverview、overviewRead/compactRecord/scope。
- agent/overview_projection_test.go（91）：分节公平、范围与极小预算、读取范围随请求任务。
- agent/overview_sync_test.go（171）：overview 输入体积与 SQL 调用对照、digest 排序与错误。

### SDK 会话与策略
- agent/sdk_session.go（417）：newSession 通过 SDK 扩展点接入策略（交互工具、审批钩子、PrepareInput）。
- agent/sdk_session_test.go（309）：延迟策略与来源、改写命令到禁止目标被拦截。
- agent/sdk_review_test.go（203）：真实 SDK 循环下收尾不能执行隐藏 Bash、空绑定不恢复默认工具、审批未变仍重校验资产。
- agent/official_provider_test.go（88）：真实 SDK 的 OpenAI 线格式与会话签名。
- agent/official_runtime_review_test.go（72）：真实运行时 HTTP 重试与 noa 溢出停止。
- agent/one_shot_session_test.go（114）：一次性会话头与归因、压缩器保留上下文并拒绝缺任务。
- agent/noa_test.go（137）：noa 开关与归档回退、视图刷新资产授权、使用配置窗口。
- agent/noa_asset_context.go（66）：noa 工具结果尾部标记过滤。

### 反馈/派发/删除
- agent/finding_feedback_context.go（192）：漏洞反馈（删报规则）历史投影与角色过滤。
- agent/finding_feedback_context_test.go（169）：规则动态刷新、取消、历史预算。
- agent/finding_deletion_feedback.go（167）：删除反馈分组/身份/分页/预算。
- agent/finding_deletion_feedback_test.go（81）：反馈预算与角色分页。
- agent/main_dispatch_feedback.go（128）：主代理派发回执（turn/record/errors）。
- agent/main_dispatch_feedback_test.go（170）：派发不轮询结束本轮、拒绝不结束、用户轮加载反馈。
- agent/intent_dispatch.go（83）：意图派发接口与宿主串行化。
- agent/intent_dispatch_test.go（104）：原子释放与重复、规划者/无任务不可用、手工模式不能建意图。
- agent/intent_stop_test.go（126）：意图停止工具的作用域与失败、目录与持久化文档。
- agent/findings_pagination_test.go（115）：漏洞分页规则与续读。
- agent/graph_context_test.go（19）：无任务的 graph_overview 返回工具错误。
- agent/hint_trigger_test.go（50）：提示触发批量与回退。

### 工具 UA / 其它
- agent/tool_ua.go（53）：curl/wget 默认浏览器 UA 的启动文件配置。
- agent/tool_ua_test.go（128）：启动文件生成、代理 env 携带 UA、真实 curl 生效。
- agent/taskmanager.go（12）：toolsNeedTaskManager 判定。
- agent/coverage_params_test.go（68）：stripCoverageParams 与 insert_assets related 参数。
- agent/chat_task_policy_test.go（16）：聊天任务策略不修改共享 agent。
- agent/assembly_failure_test.go（28）：工具装配失败覆盖所有角色。
- agent/builtin_review_test.go（168）：内置默认在预算内、asset_get 总预算与续读、overview worker 策略与读取一致。
- agent/digest_authorization_test.go（105）：跨任务 digest 授权。
- agent/fact_contract_test.go（36）：事实写入命名契约。
- agent/tools_contract_test.go（68）：规划者工具契约与按需提示一致、worker 本地工具无 Sleep、add_intent 批量>4 拒写、节点关系隐藏未授权端点。
- agent/tools_description_assets.go（27）：registerDescriptionAsset。
- agent/tools_smartproxy.go（109）：mark_host_proxy / check_host_proxy 工具与 SmartProxyView。
- agent/tools_smartproxy_test.go（162）：智能代理工具需注入才能用、标记记录。
- agent/context_prompt_measurement_test.go（21）：上下文提示指标（见上）。

## db/（80 个，SERIES 独有，11,600 行）

### 资产审批与授权
- db/asset_approval_groups.go（167）：按主机分组的审批视图、组键/等级与解析选择（与资产写锁同域）。
- db/asset_context.go（116）：TaskNodeAccess/TaskNodeApprovalStates/TaskTargetAssetStates——历史图结果的单查询重校验。
- db/asset_origins.go（99）：资产来源指针（会话/工具调用/活动）与审批来源。
- db/asset_origins_test.go（252）：旧触发器安装的命名空间刷新、注册来源、组来源与注册时间。
- db/asset_skips.go（105）：任务级资产拒绝记录（RememberTaskAssetDenials/ActiveTaskAssetSkips）。
- db/asset_skips_test.go（72）：拒绝记录生命周期。
- db/asset_templates.go（372）：资产审批模板（all/related/explicit）、DNS 元数据注册、RegisterAgentAsset。
- db/asset_templates.sql（193）：模板相关 SQL（审批模板/资产 upsert 辅助）。
- db/asset_templates_test.go（441）：模板归档与非法隔离、审批模板、按主机执行模板 fail-closed、Agent 资产校验、审批组原子性。
- db/derived_asset_parents.go（15）：用户选择的派生资产登记精确父资产。
- db/derived_asset_parents_test.go（68）：仅子关联不构成执行授权。
- db/derived_asset_approval_test.go（95）：派生审批由父资产决定 claim/tested。
- db/derived_asset_migration_test.go（161）：迁移保留撤回、旧归档恢复。
- db/finding_access.go（62）：finding 提交时的资产范围锁与授权校验（LockFindingRecordTx/ValidateFindingAccessTx）。
- db/finding_access_test.go（88）：跨会话锁顺序与受限祖先拒绝。
- db/main_execution.go（58）：主代理 pending 授权（ValidateIntentAssets/Hosts、intentAssetAllowedSQL、GrantMainIntentDispatch）。
- db/main_execution_test.go（110）：主代理 pending 意图许可与领取。
- db/host_approval_test.go（97）：仅主机审批模板、任务授权不消耗序列。
- db/host_isolation_test.go（259）：按主机授权不级联子域、跨来源任务隔离。
- db/task_asset_blocking.go（91）：封禁任务资产（保留关联与测试历史）、解析封禁。
- db/task_asset_blocking_test.go（264）：手工封禁生命周期与归档兼容。
- db/task_asset_block_recreation_test.go（56）：手工封禁在全局资产重建后仍生效。
- db/task_asset_view.go（169）：任务资产管理视图与组刷新（只重校验保留身份）。
- db/task_asset_view_test.go（145）：管理视图行为。
- db/task_company_assets.go（68）：任务企业预设物化为已批准资产（不扩展 CIDR）。
- db/task_company_assets_test.go（49）：企业预设置为 approved 资产。
- db/manual_scope_test.go（58）：手工范围接受任意输入、证书范围往返。
- db/asset_dsl_task_test.go（122）：任务 DSL 占位符顺序与授权、审批先于分页、审批字段。
- db/worker_execution.go（91）：WithWorkerRead 读取投影、ValidateWorkerAssets/Hosts、RememberWorkerAccess。
- db/read_performance_test.go（87）：读上下文取消、生成列随写更新、执行计数与 open intent。
- db/performance_indexes.go（47）：表达式索引 ensurePerformanceIndexes。
- db/performance_indexes_test.go（51）：索引与查询表达式匹配。

### 上下文检索与工具读取
- db/activity_search.go（51）：活动全文检索（不出全文，strpos 处理通配符）。
- db/activity_search_test.go（100）：全文快照与范围、结果锚点。
- db/activity_window.go（140）：会话连续窗口读取（锚点定位）。
- db/activity_window_test.go（92）：窗口配对与隔离。
- db/context_assets.go（65）：按组授权并投影上下文资产。
- db/context_mixed.go（80）：混合实体搜索游标（双 ID 边界）。
- db/context_mixed_test.go（129）：实时授权与继承、等时间与 ID 分页。
- db/context_reads.go（401）：ContextSearch 词法检索（先授权后检索）、命中与可见性。
- db/context_reads_test.go（147）：全文与窗口、活动与报告检索。
- db/context_review_test.go（220）：拒绝不可校验 digest、新目录显式绑定、大整数保真、删除任务拒绝、实时授权。
- db/tool_reads.go（380）：工具节点分页/摘要索引/worker trace 检索（授权快照内）。
- db/tool_reads_test.go（223）：授权先于分页、投影尺寸、千条活动后仍可取最新、open intent 分页计数。
- db/tool_calls.go（178）：工具调用配对/过滤/分页与详情窗口。
- db/tool_calls_test.go（250）：配对过滤分页、Unicode 隔离与只读、任务会话隔离、epoch/自定义类型。
- db/tool_context_catalog.go（329）：上下文工具目录迁移、别名仅用于历史归因、默认授权。
- db/tool_context_catalog_test.go（243）：迁移保留授权与自定义提示、用量映射不改账本。
- db/task_findings_page.go（143）：任务发现分页（仅活直接来源）、过滤先于分页。
- db/task_findings_page_test.go（215）：超 200 节点仍可翻页、过滤顺序。
- db/broadcast.go（95）：播报页节点/邻居/详情（不加载证据正文）。
- db/broadcast_test.go（100）：播报页有界。
- db/planner_context.go（111）：规划者上下文（计数、队列模式、引导）。
- db/planner_context_test.go（251）：引导计数/事件优先级/共享上限/心跳时间。
- db/planner_context_migration.go（66）：规划者按需/主动/图读取提示迁移。
- db/planner_context_migration_test.go（234）：迁移只升一次、前序迁移后再升级。
- db/planner_asset_search_test.go（148）：规划者资产/意图检索迁移与绑定。
- db/context_prompt_migration.go（111）：上下文提示默认值追加为新版本、契约迁移。

### 执行模式、队列与 worker 反馈
- db/execution_mode.go（92）：托管/手工模式、派发请求标记、用户结束。
- db/execution_mode_test.go（249）：手工切换后规划者意图被拒、模式切换竞态、等待队列锁时取消、claim。
- db/worker_queue.go（254）：write 前统一取 exploration advisory lock（beginQueue/queueExec）、队列快照。
- db/worker_queue_test.go（265）：队列顺序持久化与追加、删除保留产出与计量、删除/领取竞态。
- db/worker_feedback.go（126）：主代理派发回执、投递与未发布反馈、上下文。
- db/worker_feedback_test.go（179）：派发结算与恢复、暂停重跑失败隔离、并发去重与回滚。
- db/task_notifications.go（79）：任务通知计数/摘要（PG 可见性快照游标）。
- db/task_notifications_test.go（215）：可见性与分组、批量隔离、仅 pending 与来源。

### 发现删除/禁报/说明资产
- db/finding_deletion_feedback.go（105）：删除反馈状态、原因归一、带反馈删除。
- db/finding_deletion_feedback_test.go（191）：事务、原因校验、禁报目标快照与旧数据。
- db/finding_suppression.go（167）：禁报规则与目标身份（资产删除/重建后仍稳定）。
- db/findings_patch_test.go（74）：finding 补丁原子投影。
- db/description_assets.go（154）：描述型证据资产注册与校验。
- db/description_assets_test.go（36）：描述证据链接格式。

### MCP/用量/迁移
- db/mcp_usage.go（227）：MCP 调用账本与聚合、最近调用、服务器统计。
- db/mcp_usage_test.go（47）：聚合与最近调用。
- db/tool_usage_backfill.sql（42）：工具用量历史回填 SQL。
- db/config_session_header_test.go（64）：Profile 会话头键往返。
- db/review_migration_test.go（100）：流量保留期升级与 ScopeSentry 清理。
- db/prompt_contracts_test.go（117）：提示契约升级、报告参数不混淆 ID 命名空间。
- db/target_access_test.go（129）：本地批准不覆盖来源节点撤回。
- db/lineage_batch_test.go（58）：批量血缘保根、环终止。

## server/（76 个，SERIES 独有，10,199 行）

### 执行模式与队列
- server/execution_mode.go（384）：手工/托管模式切换、意图派发（setExecutionMode/dispatchIntents/dispatchTaskIntentsLocked），持久化终止先于取消。
- server/execution_mode_test.go（369）：手工模式只取消规划者并抑制唤醒、终局超时跳过规划者、无规划者也能起超时、删除屏障竞态。
- server/worker_queue.go（208）：worker 队列 HTTP、删除 worker、暂存 worker 文件（排空期间禁止新旁问）。
- server/worker_queue_test.go（94）：队列与删除 API。
- server/worker_signal_test.go（41）：worker 唤醒广播不消费规划者通知。
- server/intent_stop_test.go（355）：规划者停止排队意图及边界、持久化竞态。
- server/intent_user_control_test.go（126）：等待中 worker 取消与手工恢复、取消等待 claim 注册。
- server/main_execution_test.go（85）：主代理 pending 派发与准入回滚。
- server/task_pause_timeout_test.go（106）：暂停冻结任务超时预算、恢复顺延、协调器暂停/到期结算。
- server/worker_feedback.go（55）：把 worker 状态变化提交的反馈投递到主代理会话。
- server/worker_feedback_test.go（126）：投递不跑主代理/规划者、游标推进后重放。
- server/worker_trace_bindings.go（40）：worker trace 工具的绑定补齐。
- server/worker_trace_bindings_test.go（95）：绑定迁移。

### 上下文与工具契约
- server/context_tools.go（329）：上下文策略工具（默认访问、输入 schema、描述）。
- server/context_tools_test.go（178）：策略 schema 与运行时边界、worker context_search 拒绝不支持类型、任务选择器、目录无退役入口。
- server/context_surface_metrics_test.go（104）：默认上下文工具面指标。
- server/tool_calls.go（241）：工具调用查询/游标/会话隔离（只读构造，不调模型）。
- server/tool_calls_test.go（171）：查询校验、HTTP 隔离且不写模型。
- server/tool_contract_test.go（71）：任务包装 schema 与本地一致、宿主工具目录契约。
- server/tool_documentation_test.go（161）：上下文文档与角色视图一致、本地/跨域分离、无退役入口、schema 刷新保留配置。
- server/tool_execution_boundary_test.go（127）：工具目录读失败即关闭、复测详情先校验入参、分页与会话范围。
- server/tools_assembly_helpers_test.go（27）：resolveToolsForTest/augmentToolsForTest。
- server/tools_context_test.go（81）：注入的 graph_overview 使用 run store。
- server/retrieval_examples_test.go（221）：角色检索示例、窄授权、资产搜索模式必须显式。
- server/planner_asset_search_test.go（217）：规划者资产检索迁移与范围、资产搜索策略文档。
- server/planner_intent_search_test.go（148）：真实规划器装配下的意图检索。
- server/planner_on_demand_test.go（156）：按需检索真实运行、节点事件复制身份。
- server/planner_revision_test.go（21）：任务待处理触发。
- server/worker_context_search_test.go（161）：真实 worker 装配下的 context_search。
- server/task_findings_page.go（92）：任务发现分页 HTTP。
- server/task_findings_page_test.go（123）：分页来源与旧数据。
- server/activity_search.go（90）：活动检索 HTTP（游标）。
- server/activity_search_test.go（130）：校验与隔离、继承任务。
- server/activity_stream_test.go（92）：广播溢出断开客户端、先连后回放并取消池等待。
- server/activity_window_test.go（61）：历史锚点范围。
- server/broadcast_nodes.go（54）：播报详情节点。
- server/broadcast_nodes_test.go（64）：详情任务范围。

### 资产审批/封禁/模板
- server/task_asset_approval_types_test.go（108）：审批拒绝派生资产且原子。
- server/task_asset_blocking_test.go（198）：封禁按主机只取消当前任务 worker、不取消子域。
- server/asset_templates_test.go（103）：任务模板与分组 API。
- server/finding_suppression.go（450）：禁报规则服务（缓存、判定、基础设施错误处理）。
- server/finding_suppression_test.go（458）：无规则不调模型、判定、竞态隔离、最终事务竞态。
- server/finding_suppression_review_test.go（218）：规则读超时不 fail-open、profile 读期限、传输模型。
- server/finding_deletion_feedback_test.go（78）：删除反馈 HTTP 与仅规划者唤醒。
- server/findings_patch_test.go（33）：补丁先校验后写。

### 认证/代理/网络
- server/http_auth.go（210）：HTTP 入口认证（两层、失败窗口、缓存命中仍读实时设置）。
- server/http_auth_settings.go（209）：入口认证设置（保存者浏览器授权、配置加载、用户名校验）。
- server/http_auth_settings_test.go（211）：运行时设置、浏览器授权范围、校验。
- server/http_auth_test.go（281）：全路由两层认证、实时哈希缓存与失败。
- server/global_proxy.go（148）：CIPCC 出口 IP 探测与解析。
- server/global_proxy_test.go（65）：解析与探测 URL。
- server/smartproxy.go（207）：智能代理开关/池/主机名单设置与持久化。

### 自定义工具/Shell
- server/shell_execution.go（32）：平台默认 shell 模式与检测信息。
- server/shell_settings_test.go（42）：shell 设置只读、旧值拒绝。
- server/shell_test_api.go（105）：shell 检测命令与试运行端点。
- server/shell_test_api_test.go（120）：检测不执行命令、试运行与取消。
- server/shell_usage.go（211）：shell 命令请求计量与归属（字面词、目录）。
- server/shell_usage_test.go（210）：命令统计、目录归属、计量与失败。
- server/customtool_capture_test.go（79）：HTTP/脚本工具传递捕获上下文。
- server/customtool_metadata_test.go（68）：shell 工具说明与 agent schema。
- server/customtool_windows_test.go（53）：Windows 命令引号。
- server/chatupload_cleanup_test.go（77）：上传不覆盖且清理暂存、失败不留文件。

### MCP/通知/其它
- server/mcp_runtime.go（142）：MCP 元数据缓存（不缓存客户端/凭据/任务绑定）。
- server/mcp_runtime_test.go（127）：缓存隔离与失效。
- server/mcp_usage.go（46）：MCP 调用计量包装。
- server/mcp_usage_test.go（67）：成功前记录、失败计数且隐藏账本错误。
- server/mcpdiscover_test.go（71）：MCP 规范化与 transport 错误、SSE 导入。
- server/task_notifications.go（93）：任务通知 HTTP 与快照校验。
- server/task_notifications_test.go（79）：快照校验、HTTP 校验与只读。
- server/task_severity_test.go（50）：任务列表严重度统计。
- server/hint_trigger_test.go（21）：任务提示触发。
- server/builtin_review_test.go（168）：报告工具范围与证据 ID、reporter 默认提示升级、ID 解析不饱和。
- server/removed_update_routes_test.go（46）：在线更新路由已移除。
- server/webui_files.go（33）：静态资源缓存策略与 serveFileIfExists。
- server/webui_files_test.go（34）：缓存策略。
- server/task_compactor_test.go（44）：任务规划轮推进冷压缩。

## web/（79 个，SERIES 独有，14,370 行）

### 页面与大型组件
- web/scripts/test-register.cjs（8）：让 node 测试复用 Next 的 `@/` 别名。
- app/(main)/_components/new-task-shortcut.tsx（29）：⌘J 新建任务快捷键。
- app/(main)/function/tasks/detail/_tabs/asset-approvals-tab.tsx（646）：资产审批页签（按主机分组、批准/撤回/封禁、来源与墓碑、可带导航状态）。
- app/(main)/function/tasks/detail/_tabs/smart-proxy-tab.tsx（113）：智能代理页签（开关/池/已标记主机）。
- app/(main)/function/tasks/detail/_tabs/worker-queue.tsx（310）：worker 队列视图（手工排序、删除、批量选择）。
- app/(main)/function/tasks/detail/execution-control.tsx（120）：任务执行控制（暂停/恢复/结束/模式切换）。
- app/(main)/function/tasks/new/page.tsx（4）：/function/tasks/new 入口页。
- app/(main)/system/agents/_components/agents-settings.tsx（266）：Agent 设置工作区（卡片打开编辑器）。
- app/(main)/system/intercept/_components/intercept-settings.tsx（906）：命令拦截设置（规则表、工具范围、裁判模型配置）。
- app/(main)/system/llm/_components/llm-settings.tsx（1018）：LLM 设置（profile 链、思考开关/强度、池与重试）。
- app/(main)/system/mcp/_components/mcp-settings.tsx（807）：MCP 设置（服务器、工具缓存、导入/测试、用量）。
- app/(main)/system/settings/_components/general-settings.tsx（1477）：通用设置（含上下文压缩表单、流量保留、智能代理等）。
- app/(main)/system/settings/_components/http-auth-card.tsx（135）：HTTP 入口认证卡片。
- app/(main)/system/skills/_components/skills-settings.tsx（1312）：技能库设置（文件树、编辑器、可见性）。
- app/(main)/system/tools/_components/tools-settings.tsx（1143）：工具设置（系统/自定义、参数表编辑、traffic host 工具）。

### 复用组件
- components/asset-approval-template.tsx（53）：审批模板标签与选择字段。
- components/finding-delete-dialog.tsx（110）：删除漏洞弹窗（必填原因）。
- components/finding-detail-content.tsx（399）：漏洞详情/工作区（按来源与身份 key 加载）。
- components/finding-pagination.tsx（85）：漏洞分页。
- components/finding-presentation.tsx（197）：列表/详情布局切换与折叠状态。
- components/finding-selector.tsx（163）：漏洞选择器（批量导出）。
- components/finding-suppression-rules.tsx（207）：禁报规则展示/撤销。
- components/findings-toolbar.tsx（148）：漏洞筛选工具栏。
- components/session-search.tsx（213）：会话内检索。
- components/session-tool-calls.tsx（338）：会话工具调用面板（保持 transcript 挂载）。
- components/settings/appearance-settings.tsx（64）：外观设置。
- components/settings/settings-context.tsx（37）：设置工作区的上下文/draft/脏标记。
- components/settings/settings-workspace.tsx（220）：设置工作区外壳与分区。
- components/source-transcript.tsx（263）：来源会话独立历史窗口。
- components/task-notifications.tsx（165）：任务未读通知（并发首访收敛基线、已读上下文）。
- components/tasks/task-creation-form.tsx（842）：新建任务表单（附件上传清单、资产/企业/来源选择、模板、LLM 链）。

### lib
- lib/activity-highlight.ts（74）：CSS Highlight 风格的活动文本高亮。
- lib/exploration-graph.ts（40）：collapseDigestGraph 折叠 digest。
- lib/findings.ts（29）：严重度/状态常量、行键与比较。
- lib/http-auth.ts（17）：HTTP 入口 Basic 认证与 entryAwareFetch。
- lib/mcp-import.ts（84）：把常见 MCP 客户端 JSON 规范化为导入项。
- lib/polling.ts（60）：单代轮询（慢响应不叠加）。
- lib/task-notifications.ts（34）：通知分类/游标/计数类型与存储修复。
- lib/tool-calls.ts（85）：工具调用类型/状态/分页与 revision。
- lib/worker-feedback.ts（7）：lastMainActivity（worker 投递归属）。
- lib/workspace-navigation.ts（46）：新建任务路径、设置页签与 URL 规则。
- styles/control-radius.css（8）：控件圆角归一（非圆形胶囊改普通控件）。

### mock 与测试
- lib/exploration-graph.test.mjs、findings.test.ts、http-auth.test.ts、mcp-import.test.ts、polling.test.ts、side-questions.test.ts、task-notifications.test.ts、tool-calls.test.ts、worker-feedback.test.mjs、workspace-navigation.test.ts：对应 lib 的单测。
- lib/mock/*.test.ts（21 个）：活动检索、审批执行、资产审批模板、资产来源、播报、上下文工具、执行模式、删除反馈、筛选、禁报、HTTP 认证、shell 工具、任务严重度、流量、上游同步、worker 取消、队列、工作空间等 mock 行为测试。
- lib/mock/context-tools.json（378）：上下文工具 mock 数据。
- lib/mock/finding-export.ts（123）：demo 用未压缩 ZIP 生成器。
- lib/mock/finding-traffic.ts（111）：流量证据 mock。
- lib/mock/task-notifications.ts（33）：通知账本 mock（创建身份与状态分离）。
- lib/mock/tool-calls.ts（142）：工具调用 mock（代码侧当前工具名）。
- lib/mock/traffic.ts（50）：流量 mock（变更与夹具分离）。
- web/tests/asset-source-retry.cjs（56）：真实后端浏览器回归（隔离实例）。

### 资源与二进制
- web/public/series-icon.png（29 行二进制）：旧品牌图标。
- web/public/series-icon.svg（1）：旧品牌矢量图标。

## traffic / guard / cmd / report / mcphttp / third_party（SERIES 独有，25 个）

### traffic
- traffic/compress.go（131）：内联体与 blob 的 zstd 压缩/解压、压缩门槛与编码选择。
- traffic/disk_optimization_test.go（344）：内联压缩往返、小体不压、blob 压缩、旧原始体可读、静态过滤。
- traffic/smartproxy.go（217）：智能代理（按任务标记主机、代理池、持久化回调、按请求路由）。
- traffic/smartproxy_test.go（363）：禁用即无操作、任务范围、主机归一、取消标记/清空、池清空回退直连。
- traffic/smartproxy_e2e_test.go（165）：真实 HTTP/HTTPS 流量的端到端代理验证。
- traffic/smartproxy_routing_test.go（136）：带凭据标记的路由决定性。
- traffic/context_review_test.go（55）：旧 blob 与缺失文件读取。
- traffic/sync_regression_test.go（65）：同步过滤与回收延迟回归。
- traffic/tool_reads.go（469）：traffic_search/traffic_get 工具（预算、游标、二进制续读、角色授权）。
- traffic/tool_reads_test.go（136）：分段与稳定游标、二进制续读。
- traffic/tool_review_test.go（79）：默认端口与游标、任务角色不得回退全局。
- traffic/worker_policy_test.go（197）：worker 代理 pending/手工限制、读角色审批策略、封禁按主机。

### guard
- guard/assets.go（359）：任务资产策略（显式 ID 预检、结构化目标、skip scope、worker scope、审批钩子）。
- guard/assets_test.go（101）：命令目标不把本地文件当主机、结构化资产 ID、审计主体、URL/IP 收集、需正向目标识别。
- guard/proxy_tag.go（85）：任务级代理签名凭据与解析。
- guard/proxy_scope_test.go（24）：代理跳过范围签名与兼容。
- guard/compress_test.go（106）：压缩路径跳过资产检查但保留钩子。

### cmd / report / mcphttp / third_party
- cmd/series/main.go（131）：SERIES 入口（双 SQLite 图 → PG、banner、启动/优雅关闭）。
- cmd/series/main_test.go（19）：关闭保留命名取消原因。
- report/helpers.go（12）：nz 辅助。
- report/findings_test.go（65）：提交报告 Markdown 不含摘要/证据、CSV 只导出提交报告、标题优先提交名。
- mcphttp/capture_test.go（53）：远程工具捕获保留错误与完整输出。
- third_party/.DS_Store（1）：macOS 元数据（非代码）。

## docs/（126 个，SERIES 独有，16,171 行）

### 设计与契约文档
- docs/builtin-tools-query-contract.md（55）：内置查询工具的模型视图契约。
- docs/context-retrieval-window-20261019.md（95）：连续检索上下文控制与命中直读。
- docs/http-entry-auth.md（71）：HTTP 入口认证设计。
- docs/main-pending-execution-verification-20260922.md（28）：资产审批只限制 Planner 的实现与验证。
- docs/mcp-shell-verification-20260922.md（42）：MCP 界面、Shell 检测与调用统计验证。
- docs/norma-official-20261018.md（125）：官方 norma SDK 接入与验证。
- docs/planner-asset-search-20261019.md（28）：Planner 资产搜索权限修复。
- docs/series-rename-20261009.md（65）：SERIES 更名交付与切换。
- docs/session-search-verification-20260921.md（58）：会话关键词搜索与来源定位验收。
- docs/tool-contract-review-20260913.md（32）：内置工具与调用契约审查。
- docs/tool-execution-boundary-fixes-20260913.md（47）：工具执行边界修复与验收。
- docs/worker-feedback-verification-20260925.md（43）：主 Agent 异步下发与 Worker 回传验证。
- docs/worker-queue.md（15）：Worker 删除与等待队列。
- docs/upstream-sync-20260913/20260929/20260930/20261008.md（52/80/95/91）：官方增量同步记录（四份）。

### proposals（方案，5 个）
- docs/proposals/agent-view-slimming.md（691）：规划者态势窗口瘦身方案。
- docs/proposals/planner-fact-selection.md（752）：Planner 事实相关性选择实施方案。
- docs/proposals/planner-fact-selection-v2.md（428）：上述方案 v2。
- docs/proposals/tool-result-compaction.md（456）：工具返回内容压缩方案（文档标注**已否决**）。
- docs/proposals/windows-skill-path.md（153）：Windows 技能路径兼容方案。

### verification（验证记录与产物）
- agent-view-slimming-20261022.md（60）：态势精简实现与验证。
- context-policy-20261026.md（87）：上下文压缩阈值可自定义（清理线/摘要线）验证。
- export-unified-20261024.md（36）：导出统一为提交版报告。
- host-adaptation-20261025.md（212）：本机环境适配（阶段 1+2）验证。
- planner-fact-selection-20261023.md（57）：精简态势事实选择。
- planner-overview-config-20261022.md（73）：态势窗口数值配置。
- report-length-budget-20261026.md（48）：提交报告字数预算。
- report-template-20261023.md（50）：报告模板改造。
- revert-05c8e90-prompt-20261030.md（70）：回滚 05c8e90 规划行为（提示词层）。
- settlement-prefix-20261028.md（73）：收尾相位不隐藏 Bash、改 hook 拒执行。
- upstream-335bd2e-20261029.md / upstream-b76c684-20261030.md（50/64）：上游移植与补全。
- windows-skill-path-20261021.md（65）：Windows 技能路径修复。
- context-tools-20261018/、planner-on-demand-20261023/、planner-proactive-20261026/、planner-proactive-review-20261026/、worker-context-search-20261026/、finding-suppression-20261019/：对应功能的测试输出（*.txt）、结果 JSON、截图与度量数据。

### ui-neo 与截图
- docs/ui-neo/verification.md（105）：Neo 风格 UI 改版记录；同目录含 build-metrics.json、request-metrics.json 与登录/会话/任务/设置等截图。
- docs/mcp-shell-review/*.png：MCP/Shell 界面截图。
- docs/screenshots/findings-ui-20261019/、task-controls-20261019/：发现页与任务控制截图（暗/亮/移动端）。
- docs/verification-20260929/*.png：资产来源、探索图、任务严重度截图。

### 渗透测试报告（客户交付物，未展开内容）
- docs/review/报告-155-影像系统水平越权-20261023.md、报告-156-订单接管、积分商城退款金额篡改、权益重复领取、银行卡号越权查询、报告-158-验证码绕过、报告-158-卫星公交未授权信息泄露、报告-积分商城SQL注入-20261023-v2、docs/资源计量报告.md：均为 SRC/客户报告与计量说明（含目标与证据，属敏感业务数据，本记录只登记存在与用途）。
- docs/review/series-code-audit.md、series-commits-inventory.md：SERIES 自身代码审计与提交清单（此前审计已引用）。

## artifacts/（60 个，SERIES 独有，1,925 行）

- 说明性文档：asset-approval-performance-review.md（41，资产审批/入库/页面性能复核）、asset-approval-refactor.md（33，审批重构交付）、asset-approval-templates.md（38，审批模板）、builtin-system-prompts.md（514，内置系统提示词整理）、review-20260908-20260911.md（95，提交区间审查）、task-detail-loading-review.md（57，详情页性能）、upstream-sync-20260916.md（96，官方提交同步）、task-detail-first-entry-waterfall.json（509，性能瀑布数据）。
- 核验目录（README + 测试/服务/浏览器/竞态日志）：asset-source-review-20260916、cold-compactor-sync-20260919、mcp-optimization-20260919、performance-fixes-20260919、task-execution-mode-20260920、upstream-a590b41-20260920、upstream-sync-20260916、upstream-sync-20260919。
- 这些是 SERIES 开发过程的验证证据（日志/度量），非运行时代码。

## 根目录与杂项（SERIES 独有）
- screenshots/series-http-auth-settings.jpg、series-login.jpg、series-session.jpg、series-settings-dark.jpg、series-tasks.jpg（5 个旧品牌界面截图）。
- .DS_Store（63 行二进制）、third_party/.DS_Store（1）：macOS 元数据。
- .zcode/plans/plan-sess_354ed7fe-...md（34）：curl/wget 默认浏览器 UA 的实现方案笔记。

## ARTEX 独有（94 个，42,564 行）——series 不存在的能力

### notify/（26 个，5,055 行）：漏洞推送渠道适配层（全新能力）
- channel.go（194）：渠道适配器接口（无状态、Kind 注册、PermanentError 永久失败）。
- notify.go（95）：包入口、严重度排序/标签、状态标签。
- event.go（69）：notification_events.snapshot 的 JSON 契约与标题/正文生成。
- filter.go（119）：channel.filter JSON 契约（最低严重度、任务/资产范围、类别关键词、状态变更开关）。
- render.go（160）：UTF-8 安全截断（字节/符文）、单行化、按整条打包。
- markdown.go（203）：钉钉/企微共用 Markdown 渲染与转义。
- html.go（140）：邮件 HTML 正文（内联样式、表格布局）。
- http.go（255）：HTTP 投递与 SSRF 防护（环回/链路本地/内网地址拦截、错误分级、响应片段）。
- mask.go（252）：凭据掩码回显（保留尾部提示、目标变更需重填）。
- dingtalk.go（160）/ feishu.go（206）/ wecom.go（78）/ telegram.go（191）/ webhook.go（224）/ email.go（215）：六种渠道实现（签名、限长、卡片/文本、SMTP）。
- channel_test/channels_test（506）、filter_test（143）、http_test（119）、mask_test（403）、pack_test（208）、redact_test（236）、render_test（150）、sign_test（94）、ssrf_test（193）、webhook_template_test（106）、email_test（336）：协议级与安全不变量测试（错误不泄漏凭据、签名对齐 OpenSSL、模板能力边界、SSRF、UTF-8 截断、整条打包）。

### db/（8 个，2,622 行）
- notification.go（551）：推送渠道/事件存储与查询（ValidNotifyMode、channel CRUD、事件快照）。
- notification_delivery.go（446）：投递任务领取与状态（realtime/digest、租约、重试、batch）。
- notification_test.go（874）：真库下的渠道/事件/投递/资产名解析/事务回滚用例。
- asset_intercept.go（80）：资产拦截规则存储。
- asset_intercept_match.go（250）：规则匹配与执行（7 种 kind、候选域名/IP/URL、原因描述）。
- asset_intercept_match_test.go（129）：命中拒绝优先、gate 评估、候选提取。
- task_intercept.go（122）：任务级 block/allow 规则（只对本任务生效、拆分全局/任务）。
- intent_delete_test.go（170）：软删除意图（state=deleted、保留产出/血缘）测试。

### server/（10 个，3,731 行）
- notifier.go（555）：推送调度器（事件→渠道 fan-out、realtime/digest、限速、失败重试、启动恢复）。
- notify_api.go（515）：推送 HTTP API（渠道 CRUD、掩码回显、投递历史、元数据）。
- notify_api_test.go（951）：端到端（漏洞落库→事件→分派→真实 HTTP 投递）。
- asset_intercept.go（156）：资产拦截规则 CRUD HTTP。
- task_intercept.go（189）：任务级拦截/允许规则 CRUD 与校验。
- sync_scopesentry.go（513）：资产同步（ScopeSentry MCP 数据源，状态/数据源同步）。
- update.go（373）：页面一键更新 HTTP（GitHub release、缓存、重启请求）。
- update_test.go（155）：release 缓存 TTL/强制刷新/错误短 TTL。
- workspace.go（290）：工作目录文件管理（列出/读/写/下载/删除，路径逃逸防护）。
- finding_report_zip_test.go（34）：提交报告 ZIP 构建测试。

### selfupdate/（5 个，1,411 行）：在线更新
- selfupdate.go（175）：入口与路径解析、标记文件。
- bootstrap.go（299）：启动时换装/回滚/冒烟确认。
- stage.go（339）：下载、校验和清单、分阶段写盘。
- github.go（139）：固定仓库的 release 查询与资产选择。
- selfupdate_test.go（459）：隔离目录下的升级/回滚/校验测试。

### agent/（3 个，256 行）
- settlement_gate.go（142）：收尾门（工具清单不变、收尾提示出现后拒执行 Bash、前缀匹配、惰性）。
- bash_prompt_test.go（87）：Bash 提示改写保持能力建议、注册表 schema 去掉 SDK 提示、只改 Bash。
- context_prompt_migration_test.go（27）：仅精确命中旧哈希才升级 reporter 提示。

### cmd/、report/、根目录、其它
- cmd/artex/main.go（160）+ main_test.go（19）：ARTEX 入口与优雅关闭（与 SERIES 的 cmd/series 对应）。
- report/report.go（99）：从发现渲染交付 Markdown 报告（Input/findingView/parseFinding）。
- report/findings_report_test.go（51）：报告标题/文件名/跳过空报告。
- PORTING.md（68）、PORTING-AUDIT.md（53）、PORTING-CODE-COMPARE.md（1753）、PORTING-RULES.md（33）：11 条上游移植的记录/审计/代码对照/规则。
- LICENSE（661）、.env.example（13）、.gitattributes（2）、.dockerignore（16）：许可证与工程配置。
- .github/workflows/release.yml（135）：多架构二进制构建与 release 校验和。
- skills/scopesentry/SKILL.md（366）：ScopeSentry 数据源技能说明。
- screenshots/*.png（13 个，22,370 行二进制）：ARTEX README 界面截图（dashboard/tasks/findings/graph/agents/assets/chat/intercept/llm/logs/sessions/traffic/assets_test）。

### web/（15 个，3,690 行）
- app/(main)/_components/sidebar/search-dialog.tsx（157）：全局搜索对话框。
- app/(main)/_components/update-badge.tsx（58）：新版本提示徽章。
- function/findings/_components/findings-table.tsx（454）：发现表格（新列表实现）。
- function/sync/page.tsx（486）：资产同步页（ScopeSentry 状态/数据源）。
- function/tasks/detail/_tabs/report-tab.tsx（90）：任务报告页签。
- function/tasks/detail/_tabs/retests-tab.tsx（168）：复测页签。
- function/workspace/page.tsx（364）：工作空间文件管理页。
- system/intercept/assets/page.tsx（352）：全局资产拦截规则页。
- system/notify/page.tsx（605）+ _components/channel-form.tsx（153）、channel-fields.ts（181）、delivery-list.tsx（184）、stat-tile.tsx（21）：通知推送配置页与渠道表单/投递历史/统计。
- system/settings/_components/update-card.tsx（314）：一键更新卡片。
- components/asset-intercept-rules-editor.tsx（103）：拦截/允许规则编辑器（复用组件）。

### 补充：遗漏的单文件
- agent/tool_assets_budget_test.go（133）：工具资产摘要与显式凭据、overview 读取范围不跨调用存活。

## 附录：单边文件逐条索引（SERIES 566 + ARTEX 94）

说明：正文已按目录给出语义结论；本附录保证每个路径都被登记。标记：代码=源码/测试；文档=md/txt/json；资源=图片/元数据；日志=验证日志。


### SERIES 独有（566）

- [代码] `"docs/\346\274\217\346\264\236\346\265\201\351\207\217\350\257\201\346\215\256.md"`
- [代码] `"docs/review/\344\270\215\347\272\263\345\205\245/\346\212\245\345\221\212-158-\351\233\205\346\230\237\345\205\254\344\272\244\346\234\252\346\216\210\346\235\203\344\277\241\346\201\257\346\263\204\351\234\262-20261023.md"`
- [代码] `"docs/review/\346\212\245\345\221\212-155-\345\275\261\345\203\217\347\263\273\347\273\237\346\260\264\345\271\263\350\266\212\346\235\203-20261023.md"`
- [代码] `"docs/review/\346\212\245\345\221\212-156-\346\235\203\347\233\212\351\207\215\345\244\215\351\242\206\345\217\226-20261023.md"`
- [代码] `"docs/review/\346\212\245\345\221\212-156-\347\247\257\345\210\206\345\225\206\345\237\216\351\200\200\346\254\276\351\207\221\351\242\235\347\257\241\346\224\271-20261023.md"`
- [代码] `"docs/review/\346\212\245\345\221\212-156-\350\256\242\345\215\225\346\216\245\347\256\241-20261023.md"`
- [代码] `"docs/review/\346\212\245\345\221\212-156-\351\223\266\350\241\214\345\215\241\345\217\267\350\266\212\346\235\203\346\237\245\350\257\242-20261023.md"`
- [代码] `"docs/review/\346\212\245\345\221\212-158-\351\252\214\350\257\201\347\240\201\347\273\225\350\277\207-20261023.md"`
- [代码] `"docs/review/\346\212\245\345\221\212-\347\247\257\345\210\206\345\225\206\345\237\216SQL\346\263\250\345\205\245-20261023-v2.md"`
- [资源] `.DS_Store`
- [文档] `.zcode/plans/plan-sess_354ed7fe-47b6-4853-8dfe-718620b00145.md`
- [代码] `agent/assembly_failure_test.go`
- [代码] `agent/asset_context.go`
- [代码] `agent/asset_context_test.go`
- [代码] `agent/asset_management_context.go`
- [代码] `agent/asset_origin_test.go`
- [代码] `agent/asset_skip_scope_test.go`
- [代码] `agent/asset_tools.go`
- [代码] `agent/assets_approval_dsl_test.go`
- [代码] `agent/builtin_review_test.go`
- [代码] `agent/chat_task_policy_test.go`
- [代码] `agent/compaction_review_test.go`
- [代码] `agent/context_budget.go`
- [代码] `agent/context_compat_test.go`
- [代码] `agent/context_match_read_test.go`
- [代码] `agent/context_mixed.go`
- [代码] `agent/context_performance_test.go`
- [代码] `agent/context_policy.go`
- [代码] `agent/context_policy_test.go`
- [代码] `agent/context_prompt_measurement_test.go`
- [代码] `agent/context_review_test.go`
- [代码] `agent/context_task.go`
- [代码] `agent/context_task_test.go`
- [代码] `agent/context_tools.go`
- [代码] `agent/context_tools_test.go`
- [代码] `agent/context_window_metrics_test.go`
- [代码] `agent/coverage_params_test.go`
- [代码] `agent/digest_authorization_test.go`
- [代码] `agent/fact_contract_test.go`
- [代码] `agent/finding_deletion_feedback.go`
- [代码] `agent/finding_deletion_feedback_test.go`
- [代码] `agent/finding_feedback_context.go`
- [代码] `agent/finding_feedback_context_test.go`
- [代码] `agent/findings_pagination_test.go`
- [代码] `agent/graph_context_test.go`
- [代码] `agent/hint_trigger_test.go`
- [代码] `agent/intent_dispatch.go`
- [代码] `agent/intent_dispatch_test.go`
- [代码] `agent/intent_stop_test.go`
- [代码] `agent/lightformat.go`
- [代码] `agent/lightformat_test.go`
- [代码] `agent/local_tools.go`
- [代码] `agent/main_dispatch_feedback.go`
- [代码] `agent/main_dispatch_feedback_test.go`
- [代码] `agent/main_execution.go`
- [代码] `agent/main_execution_test.go`
- [代码] `agent/noa_asset_context.go`
- [代码] `agent/noa_test.go`
- [代码] `agent/official_provider_test.go`
- [代码] `agent/official_runtime_review_test.go`
- [代码] `agent/one_shot_session_test.go`
- [代码] `agent/overview_config.go`
- [代码] `agent/overview_lean.go`
- [代码] `agent/overview_lean_test.go`
- [代码] `agent/overview_projection.go`
- [代码] `agent/overview_projection_test.go`
- [代码] `agent/overview_sync_test.go`
- [代码] `agent/planner_asset_prompt_test.go`
- [代码] `agent/planner_bootstrap.go`
- [代码] `agent/planner_bootstrap_test.go`
- [代码] `agent/planner_context_metrics_test.go`
- [代码] `agent/planner_proactive_test.go`
- [代码] `agent/planner_scenario_metrics_test.go`
- [代码] `agent/planner_user_cancel_test.go`
- [代码] `agent/retrieval_access_test.go`
- [代码] `agent/retrieval_history.go`
- [代码] `agent/retrieval_history_test.go`
- [代码] `agent/retrieval_noa_test.go`
- [代码] `agent/sdk_review_test.go`
- [代码] `agent/sdk_session.go`
- [代码] `agent/sdk_session_test.go`
- [代码] `agent/structured_capture.go`
- [代码] `agent/structured_capture_test.go`
- [代码] `agent/structured_history.go`
- [代码] `agent/structured_history_test.go`
- [代码] `agent/target_access_context.go`
- [代码] `agent/target_access_test.go`
- [代码] `agent/taskmanager.go`
- [代码] `agent/tool_assets_budget_test.go`
- [代码] `agent/tool_budget.go`
- [代码] `agent/tool_budget_test.go`
- [代码] `agent/tool_contract.go`
- [代码] `agent/tool_contract_test.go`
- [代码] `agent/tool_defaults.go`
- [代码] `agent/tool_detail_contract_test.go`
- [代码] `agent/tool_documentation_test.go`
- [代码] `agent/tool_error_guidance_test.go`
- [代码] `agent/tool_history.go`
- [代码] `agent/tool_input.go`
- [代码] `agent/tool_input_test.go`
- [代码] `agent/tool_inventory_test.go`
- [代码] `agent/tool_read_scope.go`
- [代码] `agent/tool_surface_metrics_test.go`
- [代码] `agent/tool_ua.go`
- [代码] `agent/tool_ua_test.go`
- [代码] `agent/tools_asset_view.go`
- [代码] `agent/tools_asset_view_test.go`
- [代码] `agent/tools_contract_test.go`
- [代码] `agent/tools_description_assets.go`
- [代码] `agent/tools_smartproxy.go`
- [代码] `agent/tools_smartproxy_test.go`
- [代码] `agent/tools_target_access.go`
- [代码] `agent/worker_context.go`
- [代码] `agent/worker_execution.go`
- [代码] `agent/worker_execution_test.go`
- [代码] `agent/worker_trace_tools_test.go`
- [文档] `artifacts/asset-approval-performance-review.md`
- [文档] `artifacts/asset-approval-refactor.md`
- [文档] `artifacts/asset-approval-templates.md`
- [文档] `artifacts/asset-source-review-20260916/README.md`
- [日志] `artifacts/asset-source-review-20260916/agent.log`
- [日志] `artifacts/asset-source-review-20260916/db-final.log`
- [日志] `artifacts/asset-source-review-20260916/error-after.log`
- [日志] `artifacts/asset-source-review-20260916/error-before.log`
- [日志] `artifacts/asset-source-review-20260916/multiple-verified.log`
- [日志] `artifacts/asset-source-review-20260916/pagination-verified.log`
- [日志] `artifacts/asset-source-review-20260916/retry-verified.log`
- [日志] `artifacts/asset-source-review-20260916/scenarios-verified.log`
- [日志] `artifacts/asset-source-review-20260916/server-final.log`
- [日志] `artifacts/asset-source-review-20260916/types-final.log`
- [文档] `artifacts/builtin-system-prompts.md`
- [文档] `artifacts/cold-compactor-sync-20260919/README.md`
- [日志] `artifacts/cold-compactor-sync-20260919/agent.log`
- [日志] `artifacts/cold-compactor-sync-20260919/before.log`
- [日志] `artifacts/cold-compactor-sync-20260919/server.log`
- [文档] `artifacts/mcp-optimization-20260919/README.md`
- [文档] `artifacts/performance-fixes-20260919/README.md`
- [日志] `artifacts/performance-fixes-20260919/agent.log`
- [日志] `artifacts/performance-fixes-20260919/browser.log`
- [日志] `artifacts/performance-fixes-20260919/db-verified.log`
- [日志] `artifacts/performance-fixes-20260919/server-verified.log`
- [日志] `artifacts/performance-fixes-20260919/types-final.log`
- [日志] `artifacts/performance-fixes-20260919/web-verified.log`
- [文档] `artifacts/review-20260908-20260911.md`
- [文档] `artifacts/task-detail-first-entry-waterfall.json`
- [文档] `artifacts/task-detail-loading-review.md`
- [文档] `artifacts/task-execution-mode-20260920/README.md`
- [日志] `artifacts/task-execution-mode-20260920/agent.log`
- [日志] `artifacts/task-execution-mode-20260920/browser.log`
- [日志] `artifacts/task-execution-mode-20260920/db-race.log`
- [日志] `artifacts/task-execution-mode-20260920/db.log`
- [日志] `artifacts/task-execution-mode-20260920/mock.log`
- [日志] `artifacts/task-execution-mode-20260920/server-race.log`
- [日志] `artifacts/task-execution-mode-20260920/server.log`
- [文档] `artifacts/upstream-a590b41-20260920/README.md`
- [日志] `artifacts/upstream-a590b41-20260920/agent.log`
- [日志] `artifacts/upstream-a590b41-20260920/noa-race.log`
- [日志] `artifacts/upstream-a590b41-20260920/norma.log`
- [文档] `artifacts/upstream-sync-20260916.md`
- [日志] `artifacts/upstream-sync-20260916/agent-fixed.log`
- [日志] `artifacts/upstream-sync-20260916/browser.log`
- [日志] `artifacts/upstream-sync-20260916/noa-race.log`
- [日志] `artifacts/upstream-sync-20260916/norma-fixed.log`
- [日志] `artifacts/upstream-sync-20260916/norma-local.log`
- [日志] `artifacts/upstream-sync-20260916/norma-upstream.log`
- [日志] `artifacts/upstream-sync-20260916/server-fixed.log`
- [文档] `artifacts/upstream-sync-20260919/README.md`
- [日志] `artifacts/upstream-sync-20260919/agent.log`
- [日志] `artifacts/upstream-sync-20260919/build.log`
- [日志] `artifacts/upstream-sync-20260919/db.log`
- [日志] `artifacts/upstream-sync-20260919/features.log`
- [日志] `artifacts/upstream-sync-20260919/local-browser.log`
- [日志] `artifacts/upstream-sync-20260919/server-final.log`
- [日志] `artifacts/upstream-sync-20260919/source-browser.log`
- [日志] `artifacts/upstream-sync-20260919/types.log`
- [日志] `artifacts/upstream-sync-20260919/web-tests.log`
- [代码] `build-docker.sh`
- [代码] `cmd/series/main.go`
- [代码] `cmd/series/main_test.go`
- [代码] `db/activity_search.go`
- [代码] `db/activity_search_test.go`
- [代码] `db/activity_window.go`
- [代码] `db/activity_window_test.go`
- [代码] `db/asset_approval_groups.go`
- [代码] `db/asset_context.go`
- [代码] `db/asset_dsl_task_test.go`
- [代码] `db/asset_origins.go`
- [代码] `db/asset_origins_test.go`
- [代码] `db/asset_skips.go`
- [代码] `db/asset_skips_test.go`
- [代码] `db/asset_templates.go`
- [文档] `db/asset_templates.sql`
- [代码] `db/asset_templates_test.go`
- [代码] `db/broadcast.go`
- [代码] `db/broadcast_test.go`
- [代码] `db/config_session_header_test.go`
- [代码] `db/context_assets.go`
- [代码] `db/context_mixed.go`
- [代码] `db/context_mixed_test.go`
- [代码] `db/context_prompt_migration.go`
- [代码] `db/context_reads.go`
- [代码] `db/context_reads_test.go`
- [代码] `db/context_review_test.go`
- [代码] `db/derived_asset_approval_test.go`
- [代码] `db/derived_asset_migration_test.go`
- [代码] `db/derived_asset_parents.go`
- [代码] `db/derived_asset_parents_test.go`
- [代码] `db/description_assets.go`
- [代码] `db/description_assets_test.go`
- [代码] `db/execution_mode.go`
- [代码] `db/execution_mode_test.go`
- [代码] `db/finding_access.go`
- [代码] `db/finding_access_test.go`
- [代码] `db/finding_deletion_feedback.go`
- [代码] `db/finding_deletion_feedback_test.go`
- [代码] `db/finding_suppression.go`
- [代码] `db/findings_patch_test.go`
- [代码] `db/host_approval_test.go`
- [代码] `db/host_isolation_test.go`
- [代码] `db/lineage_batch_test.go`
- [代码] `db/main_execution.go`
- [代码] `db/main_execution_test.go`
- [代码] `db/manual_scope_test.go`
- [代码] `db/mcp_usage.go`
- [代码] `db/mcp_usage_test.go`
- [代码] `db/performance_indexes.go`
- [代码] `db/performance_indexes_test.go`
- [代码] `db/planner_asset_search_test.go`
- [代码] `db/planner_context.go`
- [代码] `db/planner_context_migration.go`
- [代码] `db/planner_context_migration_test.go`
- [代码] `db/planner_context_test.go`
- [代码] `db/prompt_contracts_test.go`
- [代码] `db/read_performance_test.go`
- [代码] `db/review_migration_test.go`
- [代码] `db/target_access_test.go`
- [代码] `db/task_asset_block_recreation_test.go`
- [代码] `db/task_asset_blocking.go`
- [代码] `db/task_asset_blocking_test.go`
- [代码] `db/task_asset_view.go`
- [代码] `db/task_asset_view_test.go`
- [代码] `db/task_company_assets.go`
- [代码] `db/task_company_assets_test.go`
- [代码] `db/task_findings_page.go`
- [代码] `db/task_findings_page_test.go`
- [代码] `db/task_notifications.go`
- [代码] `db/task_notifications_test.go`
- [代码] `db/tool_calls.go`
- [代码] `db/tool_calls_test.go`
- [代码] `db/tool_context_catalog.go`
- [代码] `db/tool_context_catalog_test.go`
- [代码] `db/tool_reads.go`
- [代码] `db/tool_reads_test.go`
- [文档] `db/tool_usage_backfill.sql`
- [代码] `db/worker_execution.go`
- [代码] `db/worker_feedback.go`
- [代码] `db/worker_feedback_test.go`
- [代码] `db/worker_queue.go`
- [代码] `db/worker_queue_test.go`
- [文档] `docs/builtin-tools-query-contract.md`
- [文档] `docs/context-retrieval-window-20261019.md`
- [文档] `docs/http-entry-auth.md`
- [文档] `docs/main-pending-execution-verification-20260922.md`
- [资源] `docs/mcp-shell-review/mcp-desktop-light.png`
- [资源] `docs/mcp-shell-review/mcp-mobile-dark.png`
- [资源] `docs/mcp-shell-review/shell-test-dark.png`
- [文档] `docs/mcp-shell-verification-20260922.md`
- [文档] `docs/norma-official-20261018.md`
- [文档] `docs/planner-asset-search-20261019.md`
- [文档] `docs/proposals/agent-view-slimming.md`
- [文档] `docs/proposals/planner-fact-selection-v2.md`
- [文档] `docs/proposals/planner-fact-selection.md`
- [文档] `docs/proposals/tool-result-compaction.md`
- [文档] `docs/proposals/windows-skill-path.md`
- [资源] `docs/screenshots/findings-ui-20261019/global-dark.png`
- [资源] `docs/screenshots/findings-ui-20261019/global-light.png`
- [资源] `docs/screenshots/findings-ui-20261019/task-light.png`
- [资源] `docs/screenshots/findings-ui-20261019/task-mobile.png`
- [资源] `docs/screenshots/task-controls-20261019/desktop-dark.png`
- [资源] `docs/screenshots/task-controls-20261019/desktop-light.png`
- [资源] `docs/screenshots/task-controls-20261019/mobile-dark.png`
- [资源] `docs/screenshots/task-controls-20261019/settings.png`
- [文档] `docs/series-rename-20261009.md`
- [文档] `docs/session-search-verification-20260921.md`
- [文档] `docs/tool-contract-review-20260913.md`
- [文档] `docs/tool-execution-boundary-fixes-20260913.md`
- [文档] `docs/ui-neo/build-metrics.json`
- [资源] `docs/ui-neo/conversation-dark-source.png`
- [资源] `docs/ui-neo/conversation-source.png`
- [资源] `docs/ui-neo/login.png`
- [资源] `docs/ui-neo/new-task-desktop.png`
- [资源] `docs/ui-neo/new-task-mobile.png`
- [文档] `docs/ui-neo/request-metrics.json`
- [资源] `docs/ui-neo/settings-desktop.png`
- [资源] `docs/ui-neo/settings-mobile.png`
- [资源] `docs/ui-neo/setup.png`
- [资源] `docs/ui-neo/sidebar-without-search.png`
- [资源] `docs/ui-neo/tasks-desktop.png`
- [文档] `docs/ui-neo/verification.md`
- [文档] `docs/upstream-sync-20260913.md`
- [文档] `docs/upstream-sync-20260929.md`
- [文档] `docs/upstream-sync-20260930.md`
- [文档] `docs/upstream-sync-20261008.md`
- [资源] `docs/verification-20260929/asset-origin.png`
- [资源] `docs/verification-20260929/exploration-graph.png`
- [资源] `docs/verification-20260929/task-severity.png`
- [文档] `docs/verification/agent-view-slimming-20261022.md`
- [文档] `docs/verification/context-policy-20261026.md`
- [资源] `docs/verification/context-tools-20261018/context-get-editor.png`
- [文档] `docs/verification/context-tools-20261018/measurements.txt`
- [文档] `docs/verification/context-tools-20261018/results.json`
- [文档] `docs/verification/context-tools-20261018/review-compile-final.txt`
- [文档] `docs/verification/context-tools-20261018/review-final.txt`
- [文档] `docs/verification/context-tools-20261018/review-metrics.txt`
- [文档] `docs/verification/context-tools-20261018/review-norma.txt`
- [文档] `docs/verification/context-tools-20261018/review-server-final.txt`
- [文档] `docs/verification/context-tools-20261018/review-surface.txt`
- [文档] `docs/verification/context-tools-20261018/review.json`
- [资源] `docs/verification/context-tools-20261018/source-location.png`
- [文档] `docs/verification/context-tools-20261018/test-summary.txt`
- [文档] `docs/verification/export-unified-20261024.md`
- [文档] `docs/verification/finding-suppression-20261019/results.json`
- [文档] `docs/verification/finding-suppression-20261019/review-results.json`
- [资源] `docs/verification/finding-suppression-20261019/rules-active.png`
- [资源] `docs/verification/finding-suppression-20261019/rules-revoked.png`
- [文档] `docs/verification/host-adaptation-20261025.md`
- [文档] `docs/verification/planner-fact-selection-20261023.md`
- [文档] `docs/verification/planner-on-demand-20261023/agent-tests.txt`
- [文档] `docs/verification/planner-on-demand-20261023/compile.txt`
- [文档] `docs/verification/planner-on-demand-20261023/db-tests.txt`
- [文档] `docs/verification/planner-on-demand-20261023/guard-tests.txt`
- [文档] `docs/verification/planner-on-demand-20261023/jit-diagnostic.txt`
- [文档] `docs/verification/planner-on-demand-20261023/overview-metrics.txt`
- [文档] `docs/verification/planner-on-demand-20261023/planner-runs.txt`
- [文档] `docs/verification/planner-on-demand-20261023/race-tests.txt`
- [文档] `docs/verification/planner-on-demand-20261023/results.json`
- [文档] `docs/verification/planner-on-demand-20261023/scenario-after.txt`
- [文档] `docs/verification/planner-on-demand-20261023/scenario-before.txt`
- [文档] `docs/verification/planner-on-demand-20261023/server-focused.txt`
- [文档] `docs/verification/planner-on-demand-20261023/server-tests.txt`
- [文档] `docs/verification/planner-on-demand-20261023/vet.txt`
- [文档] `docs/verification/planner-overview-config-20261022.md`
- [文档] `docs/verification/planner-proactive-20261026/agent-full.txt`
- [文档] `docs/verification/planner-proactive-20261026/compile.txt`
- [文档] `docs/verification/planner-proactive-20261026/db-full.txt`
- [文档] `docs/verification/planner-proactive-20261026/examples.txt`
- [文档] `docs/verification/planner-proactive-20261026/guard-full.txt`
- [文档] `docs/verification/planner-proactive-20261026/metrics-after.txt`
- [文档] `docs/verification/planner-proactive-20261026/metrics-before.txt`
- [文档] `docs/verification/planner-proactive-20261026/race.txt`
- [文档] `docs/verification/planner-proactive-20261026/results.json`
- [文档] `docs/verification/planner-proactive-20261026/runs-before.txt`
- [文档] `docs/verification/planner-proactive-20261026/server-final.txt`
- [文档] `docs/verification/planner-proactive-20261026/server-full.txt`
- [文档] `docs/verification/planner-proactive-20261026/vet.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/agent-full.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/agent-race.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/build.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/compile.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/db-full.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/guard-full.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/results.json`
- [文档] `docs/verification/planner-proactive-review-20261026/server-full.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/server-race.txt`
- [文档] `docs/verification/planner-proactive-review-20261026/vet.txt`
- [文档] `docs/verification/report-length-budget-20261026.md`
- [文档] `docs/verification/report-template-20261023.md`
- [文档] `docs/verification/revert-05c8e90-prompt-20261030.md`
- [文档] `docs/verification/settlement-prefix-20261028.md`
- [文档] `docs/verification/upstream-335bd2e-20261029.md`
- [文档] `docs/verification/upstream-b76c684-20261030.md`
- [文档] `docs/verification/windows-skill-path-20261021.md`
- [文档] `docs/verification/worker-context-search-20261026/agent-full.txt`
- [文档] `docs/verification/worker-context-search-20261026/agent-race.txt`
- [文档] `docs/verification/worker-context-search-20261026/build.txt`
- [文档] `docs/verification/worker-context-search-20261026/db-full.txt`
- [文档] `docs/verification/worker-context-search-20261026/guard-full.txt`
- [文档] `docs/verification/worker-context-search-20261026/results.json`
- [文档] `docs/verification/worker-context-search-20261026/sdk-focused.txt`
- [文档] `docs/verification/worker-context-search-20261026/server-full-final.txt`
- [文档] `docs/verification/worker-context-search-20261026/server-full.txt`
- [文档] `docs/verification/worker-context-search-20261026/server-race.txt`
- [文档] `docs/verification/worker-context-search-20261026/vet.txt`
- [文档] `docs/worker-feedback-verification-20260925.md`
- [文档] `docs/worker-queue.md`
- [代码] `guard/assets.go`
- [代码] `guard/assets_test.go`
- [代码] `guard/compress_test.go`
- [代码] `guard/proxy_scope_test.go`
- [代码] `guard/proxy_tag.go`
- [代码] `mcphttp/capture_test.go`
- [代码] `report/findings_test.go`
- [代码] `report/helpers.go`
- [资源] `screenshots/series-http-auth-settings.jpg`
- [资源] `screenshots/series-login.jpg`
- [资源] `screenshots/series-session.jpg`
- [资源] `screenshots/series-settings-dark.jpg`
- [资源] `screenshots/series-tasks.jpg`
- [代码] `server/activity_search.go`
- [代码] `server/activity_search_test.go`
- [代码] `server/activity_stream_test.go`
- [代码] `server/activity_window_test.go`
- [代码] `server/asset_templates_test.go`
- [代码] `server/broadcast_nodes.go`
- [代码] `server/broadcast_nodes_test.go`
- [代码] `server/builtin_review_test.go`
- [代码] `server/chatupload_cleanup_test.go`
- [代码] `server/context_surface_metrics_test.go`
- [代码] `server/context_tools.go`
- [代码] `server/context_tools_test.go`
- [代码] `server/customtool_capture_test.go`
- [代码] `server/customtool_metadata_test.go`
- [代码] `server/customtool_windows_test.go`
- [代码] `server/execution_mode.go`
- [代码] `server/execution_mode_test.go`
- [代码] `server/finding_deletion_feedback_test.go`
- [代码] `server/finding_suppression.go`
- [代码] `server/finding_suppression_review_test.go`
- [代码] `server/finding_suppression_test.go`
- [代码] `server/findings_patch_test.go`
- [代码] `server/global_proxy.go`
- [代码] `server/global_proxy_test.go`
- [代码] `server/hint_trigger_test.go`
- [代码] `server/http_auth.go`
- [代码] `server/http_auth_settings.go`
- [代码] `server/http_auth_settings_test.go`
- [代码] `server/http_auth_test.go`
- [代码] `server/intent_stop_test.go`
- [代码] `server/intent_user_control_test.go`
- [代码] `server/main_execution_test.go`
- [代码] `server/mcp_runtime.go`
- [代码] `server/mcp_runtime_test.go`
- [代码] `server/mcp_usage.go`
- [代码] `server/mcp_usage_test.go`
- [代码] `server/mcpdiscover_test.go`
- [代码] `server/planner_asset_search_test.go`
- [代码] `server/planner_intent_search_test.go`
- [代码] `server/planner_on_demand_test.go`
- [代码] `server/planner_revision_test.go`
- [代码] `server/removed_update_routes_test.go`
- [代码] `server/retrieval_examples_test.go`
- [代码] `server/shell_execution.go`
- [代码] `server/shell_settings_test.go`
- [代码] `server/shell_test_api.go`
- [代码] `server/shell_test_api_test.go`
- [代码] `server/shell_usage.go`
- [代码] `server/shell_usage_test.go`
- [代码] `server/smartproxy.go`
- [代码] `server/task_asset_approval_types_test.go`
- [代码] `server/task_asset_blocking_test.go`
- [代码] `server/task_compactor_test.go`
- [代码] `server/task_findings_page.go`
- [代码] `server/task_findings_page_test.go`
- [代码] `server/task_notifications.go`
- [代码] `server/task_notifications_test.go`
- [代码] `server/task_pause_timeout_test.go`
- [代码] `server/task_severity_test.go`
- [代码] `server/tool_calls.go`
- [代码] `server/tool_calls_test.go`
- [代码] `server/tool_contract_test.go`
- [代码] `server/tool_documentation_test.go`
- [代码] `server/tool_execution_boundary_test.go`
- [代码] `server/tools_assembly_helpers_test.go`
- [代码] `server/tools_context_test.go`
- [代码] `server/webui_files.go`
- [代码] `server/webui_files_test.go`
- [代码] `server/worker_context_search_test.go`
- [代码] `server/worker_feedback.go`
- [代码] `server/worker_feedback_test.go`
- [代码] `server/worker_queue.go`
- [代码] `server/worker_queue_test.go`
- [代码] `server/worker_signal_test.go`
- [代码] `server/worker_trace_bindings.go`
- [代码] `server/worker_trace_bindings_test.go`
- [资源] `third_party/.DS_Store`
- [代码] `traffic/compress.go`
- [代码] `traffic/context_review_test.go`
- [代码] `traffic/disk_optimization_test.go`
- [代码] `traffic/smartproxy.go`
- [代码] `traffic/smartproxy_e2e_test.go`
- [代码] `traffic/smartproxy_routing_test.go`
- [代码] `traffic/smartproxy_test.go`
- [代码] `traffic/sync_regression_test.go`
- [代码] `traffic/tool_reads.go`
- [代码] `traffic/tool_reads_test.go`
- [代码] `traffic/tool_review_test.go`
- [代码] `traffic/worker_policy_test.go`
- [资源] `web/public/series-icon.png`
- [资源] `web/public/series-icon.svg`
- [代码] `web/scripts/test-register.cjs`
- [代码] `web/src/app/(main)/_components/new-task-shortcut.tsx`
- [代码] `web/src/app/(main)/function/tasks/detail/_tabs/asset-approvals-tab.tsx`
- [代码] `web/src/app/(main)/function/tasks/detail/_tabs/smart-proxy-tab.tsx`
- [代码] `web/src/app/(main)/function/tasks/detail/_tabs/worker-queue.tsx`
- [代码] `web/src/app/(main)/function/tasks/detail/execution-control.tsx`
- [代码] `web/src/app/(main)/function/tasks/new/page.tsx`
- [代码] `web/src/app/(main)/system/agents/_components/agents-settings.tsx`
- [代码] `web/src/app/(main)/system/intercept/_components/intercept-settings.tsx`
- [代码] `web/src/app/(main)/system/llm/_components/llm-settings.tsx`
- [代码] `web/src/app/(main)/system/mcp/_components/mcp-settings.tsx`
- [代码] `web/src/app/(main)/system/settings/_components/general-settings.tsx`
- [代码] `web/src/app/(main)/system/settings/_components/http-auth-card.tsx`
- [代码] `web/src/app/(main)/system/skills/_components/skills-settings.tsx`
- [代码] `web/src/app/(main)/system/tools/_components/tools-settings.tsx`
- [代码] `web/src/components/asset-approval-template.tsx`
- [代码] `web/src/components/finding-delete-dialog.tsx`
- [代码] `web/src/components/finding-detail-content.tsx`
- [代码] `web/src/components/finding-pagination.tsx`
- [代码] `web/src/components/finding-presentation.tsx`
- [代码] `web/src/components/finding-selector.tsx`
- [代码] `web/src/components/finding-suppression-rules.tsx`
- [代码] `web/src/components/findings-toolbar.tsx`
- [代码] `web/src/components/session-search.tsx`
- [代码] `web/src/components/session-tool-calls.tsx`
- [代码] `web/src/components/settings/appearance-settings.tsx`
- [代码] `web/src/components/settings/settings-context.tsx`
- [代码] `web/src/components/settings/settings-workspace.tsx`
- [代码] `web/src/components/source-transcript.tsx`
- [代码] `web/src/components/task-notifications.tsx`
- [代码] `web/src/components/tasks/task-creation-form.tsx`
- [代码] `web/src/lib/activity-highlight.ts`
- [代码] `web/src/lib/exploration-graph.test.mjs`
- [代码] `web/src/lib/exploration-graph.ts`
- [代码] `web/src/lib/findings.test.ts`
- [代码] `web/src/lib/findings.ts`
- [代码] `web/src/lib/http-auth.test.ts`
- [代码] `web/src/lib/http-auth.ts`
- [代码] `web/src/lib/mcp-import.test.ts`
- [代码] `web/src/lib/mcp-import.ts`
- [代码] `web/src/lib/mock/activity-search.test.ts`
- [代码] `web/src/lib/mock/approval-execution.test.ts`
- [代码] `web/src/lib/mock/asset-approval-templates.test.ts`
- [代码] `web/src/lib/mock/asset-origins.test.ts`
- [代码] `web/src/lib/mock/broadcast.test.ts`
- [文档] `web/src/lib/mock/context-tools.json`
- [代码] `web/src/lib/mock/context-tools.test.ts`
- [代码] `web/src/lib/mock/execution-mode.test.ts`
- [代码] `web/src/lib/mock/finding-deletion.test.ts`
- [代码] `web/src/lib/mock/finding-export.ts`
- [代码] `web/src/lib/mock/finding-filters.test.ts`
- [代码] `web/src/lib/mock/finding-suppression.test.ts`
- [代码] `web/src/lib/mock/finding-traffic.ts`
- [代码] `web/src/lib/mock/http-auth.test.ts`
- [代码] `web/src/lib/mock/shell-tools.test.ts`
- [代码] `web/src/lib/mock/task-notifications.ts`
- [代码] `web/src/lib/mock/task-severity.test.ts`
- [代码] `web/src/lib/mock/tool-calls.ts`
- [代码] `web/src/lib/mock/traffic.test.ts`
- [代码] `web/src/lib/mock/traffic.ts`
- [代码] `web/src/lib/mock/upstream-sync.test.ts`
- [代码] `web/src/lib/mock/worker-cancel.test.ts`
- [代码] `web/src/lib/mock/worker-queue.test.ts`
- [代码] `web/src/lib/mock/workspace.test.ts`
- [代码] `web/src/lib/polling.test.ts`
- [代码] `web/src/lib/polling.ts`
- [代码] `web/src/lib/side-questions.test.ts`
- [代码] `web/src/lib/task-notifications.test.ts`
- [代码] `web/src/lib/task-notifications.ts`
- [代码] `web/src/lib/tool-calls.test.ts`
- [代码] `web/src/lib/tool-calls.ts`
- [代码] `web/src/lib/worker-feedback.test.mjs`
- [代码] `web/src/lib/worker-feedback.ts`
- [代码] `web/src/lib/workspace-navigation.test.ts`
- [代码] `web/src/lib/workspace-navigation.ts`
- [代码] `web/src/styles/control-radius.css`
- [代码] `web/tests/asset-source-retry.cjs`

### ARTEX 独有（94）

- [文档] `.dockerignore`
- [文档] `.env.example`
- [文档] `.gitattributes`
- [文档] `.github/workflows/release.yml`
- [文档] `LICENSE`
- [文档] `PORTING-AUDIT.md`
- [文档] `PORTING-CODE-COMPARE.md`
- [文档] `PORTING-RULES.md`
- [文档] `PORTING.md`
- [代码] `agent/bash_prompt_test.go`
- [代码] `agent/context_prompt_migration_test.go`
- [代码] `agent/settlement_gate.go`
- [代码] `cmd/artex/main.go`
- [代码] `cmd/artex/main_test.go`
- [代码] `db/asset_intercept.go`
- [代码] `db/asset_intercept_match.go`
- [代码] `db/asset_intercept_match_test.go`
- [代码] `db/intent_delete_test.go`
- [代码] `db/notification.go`
- [代码] `db/notification_delivery.go`
- [代码] `db/notification_test.go`
- [代码] `db/task_intercept.go`
- [代码] `notify/channel.go`
- [代码] `notify/channels_test.go`
- [代码] `notify/dingtalk.go`
- [代码] `notify/email.go`
- [代码] `notify/email_test.go`
- [代码] `notify/event.go`
- [代码] `notify/feishu.go`
- [代码] `notify/filter.go`
- [代码] `notify/filter_test.go`
- [代码] `notify/html.go`
- [代码] `notify/http.go`
- [代码] `notify/http_test.go`
- [代码] `notify/markdown.go`
- [代码] `notify/mask.go`
- [代码] `notify/mask_test.go`
- [代码] `notify/notify.go`
- [代码] `notify/pack_test.go`
- [代码] `notify/redact_test.go`
- [代码] `notify/render.go`
- [代码] `notify/render_test.go`
- [代码] `notify/sign_test.go`
- [代码] `notify/ssrf_test.go`
- [代码] `notify/telegram.go`
- [代码] `notify/webhook.go`
- [代码] `notify/webhook_template_test.go`
- [代码] `notify/wecom.go`
- [代码] `report/findings_report_test.go`
- [代码] `report/report.go`
- [资源] `screenshots/agents.png`
- [资源] `screenshots/assets.png`
- [资源] `screenshots/assets_test.png`
- [资源] `screenshots/chat.png`
- [资源] `screenshots/dashboard.png`
- [资源] `screenshots/findings.png`
- [资源] `screenshots/graph.png`
- [资源] `screenshots/intercept.png`
- [资源] `screenshots/llm.png`
- [资源] `screenshots/logs.png`
- [资源] `screenshots/sessions.png`
- [资源] `screenshots/tasks.png`
- [资源] `screenshots/traffic.png`
- [代码] `selfupdate/bootstrap.go`
- [代码] `selfupdate/github.go`
- [代码] `selfupdate/selfupdate.go`
- [代码] `selfupdate/selfupdate_test.go`
- [代码] `selfupdate/stage.go`
- [代码] `server/asset_intercept.go`
- [代码] `server/finding_report_zip_test.go`
- [代码] `server/notifier.go`
- [代码] `server/notify_api.go`
- [代码] `server/notify_api_test.go`
- [代码] `server/sync_scopesentry.go`
- [代码] `server/task_intercept.go`
- [代码] `server/update.go`
- [代码] `server/update_test.go`
- [代码] `server/workspace.go`
- [文档] `skills/scopesentry/SKILL.md`
- [代码] `web/src/app/(main)/_components/sidebar/search-dialog.tsx`
- [代码] `web/src/app/(main)/_components/update-badge.tsx`
- [代码] `web/src/app/(main)/function/findings/_components/findings-table.tsx`
- [代码] `web/src/app/(main)/function/sync/page.tsx`
- [代码] `web/src/app/(main)/function/tasks/detail/_tabs/report-tab.tsx`
- [代码] `web/src/app/(main)/function/tasks/detail/_tabs/retests-tab.tsx`
- [代码] `web/src/app/(main)/function/workspace/page.tsx`
- [代码] `web/src/app/(main)/system/intercept/assets/page.tsx`
- [代码] `web/src/app/(main)/system/notify/_components/channel-fields.ts`
- [代码] `web/src/app/(main)/system/notify/_components/channel-form.tsx`
- [代码] `web/src/app/(main)/system/notify/_components/delivery-list.tsx`
- [代码] `web/src/app/(main)/system/notify/_components/stat-tile.tsx`
- [代码] `web/src/app/(main)/system/notify/page.tsx`
- [代码] `web/src/app/(main)/system/settings/_components/update-card.tsx`
- [代码] `web/src/components/asset-intercept-rules-editor.tsx`

## 第三部分：单边文件正文级核对（10 路并行，2026-11-02）

方法：10 个子代理各自独立上下文，逐文件打开正文（不是只看符号），按「文件 | 行数 | 职责 | 关键实现 | 风险」输出；原始结果存 `docs/audit/subagent-summary-*.txt`（5 份被截断的长结果），其余已并入本节。数量已与目录实测核对：db 非测试 79、db 测试 90、server 非测试 71、server 测试 102、web app+components 166、web lib/mock/tests 69、agent 测试 94、traffic/guard/cmd/report/mcphttp 39、ARTEX 独有代码 57。

### 覆盖结论
- SERIES `db/`：79 个非测试 + 90 个测试全部读完。核心是资产审批/封禁/模板、上下文检索（context_reads/tool_reads）、探索图状态机（exploration.go 2465 行）、队列（worker_queue）、worker 反馈、禁报规则、MCP 用量、性能索引。
- SERIES `server/`：71 个非测试 + 102 个测试全部读完。核心是 engine（1461 行）/manager（2285 行）/server.go（4126 行）三层运行时、执行模式与派发、HTTP 入口认证、拦截、发现工作流、归档、LLM 路由与故障转移。
- SERIES `web/`：166 个页面/组件 + 69 个 lib/mock/测试全部读完。核心是任务详情（sessions 2638 行、tasks 3002 行）、发现中心、资产库、设置工作区、审批页签。
- SERIES `agent/`：94 个测试全部读完（此前已读完 25 个非测试核心），覆盖授权、上下文预算、工具契约、收尾门、主机经验。
- `traffic/guard/cmd/report/mcphttp`：39 个全部读完（含真实 mitmproxy 端到端用例）。
- ARTEX 独有：57 个代码文件全部读完；前端 15 个文件、根目录文档、docs 126 个、artifacts 60 个已登记。

### 本轮新发现的实质风险（按严重度）
1. **ARTEX `server/workspace.go` 无 `EvalSymlinks`**（已核对源码）：`wsResolve` 只做 `Clean` + 前缀比较，工作区内指向外部的符号链接可越界读写；上传/MCP 都可能写入此类链接。
2. **SERIES `server/triggers.go` 更新/删除只按 id**，不校验触发器是否属于当前 agent → 可跨 agent 改/删触发器。
3. **SERIES `server/intercept.go` 的 interceptGetOne/Detail/Execution 只按 id 取**，无任务级授权；`server/goals_api.go` 的 edit/delete 只校验 kind，不校验 gid 归属任务。
4. **SERIES `server/task_assets.go` 的写操作没有 `beginTaskOperation` 保护**，与任务删除存在竞态；actor 固定为 SERIES，缺操作者审计。
5. **SERIES `server/skill_zip.go` 解压前不限总解压体积/条目数**（zip bomb 面），依赖上层限制。
6. **SERIES `guard/guard.go` 在未装配 interceptor 时完全不拦截**：破坏性命令与反弹 shell 均放行；安全性全押在 DB 规则是否加载。
7. **SERIES 资产审批谓词散落 4 段 SQL**（`assets.go` 的 QueryByTask*/CountByTaskTested/taskAssetBlockedSQL 等），改一处易不一致 → 越权或漏封。
8. **ARTEX `notify/` SSRF 默认放行 RFC1918**（内网自建渠道是合法用法），逃生口 `ARTEX_NOTIFY_ALLOW_LOCAL` 是进程级全局开关；邮件渠道在未配账号时明文投递。
9. **ARTEX `report/findings.go` 文件名按字节截断 120**，可能切断汉字产生非法 UTF-8；同包 `FindingReportFilename` 已改 rune 截断，两处口径不一致。
10. **SERIES `performance_indexes.go` 用 `CREATE INDEX CONCURRENTLY`**，必须在迁移锁的专用连接上执行；失败会拖慢或阻塞启动。`activity_search` 无 trigram 索引，长任务检索是区间全扫。

### 与 ARTEX 的总体差异（正文级确认）
- SERIES 多出的是「授权/审批/上下文预算」三条主线：资产级读写授权、任务执行模式与 Worker 队列、上下文检索与态势投影；这些在 ARTEX 中整体不存在。
- ARTEX 多出的是「交付/运维」四条线：通知推送（6 渠道 + SSRF/掩码/签名/整条打包）、在线更新（selfupdate + 退出码 75 + 回滚）、资产/任务拦截规则、ScopeSentry 同步与工作区文件管理。
- 两边共有的核心（探索图、worker、发现工作流）逻辑一致，差异集中在授权与投影层，而不是算法层。
