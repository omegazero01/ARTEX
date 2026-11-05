# 资产审批残留清理（并入审批总提交）

- 日期：2026-11-07
- 分支：`rebuild/from-series`
- 背景：`fd6bf94`（移除资产审批、封禁与出处登记）已在 db/schema.sql 第 1697 行后用
  `CREATE OR REPLACE FUNCTION` 把 `task_asset_blocked`、`task_asset_owner_approval_state`、
  `task_asset_effective_approval_state` 覆盖成“一律 approved / 未封禁”，Go 侧检查因此全部放行，
  但函数名和判断代码仍留着，容易误读为还在拦截。

## 改动

1. `db/task_assets.go`
   - `ValidateTaskAssetsApproved`：只校验资产编号存在，不再看 pending/approved/blocked/revoked；
   - `FilterApprovedAssets`：存在即保留，不再按审批状态过滤。
2. `db/worker_execution.go`
   - `ValidateWorkerAssets`：只校验资产编号存在，不再要求 approved/pending。
3. 调用点（intent_stop、节点读、evidence_id、intent_id、steer_work、worker 输出/轨迹、
   finding_workflow、execution_mode、guard、traffic）保持原签名，行为随之放行，不再逐处改。
4. SQL 覆盖保持 `fd6bf94` 原样，不重复添加。

## 删除的过期测试

审批/封禁语义已不存在，以下用例断言的“pending/revoked/blocked 必须被拒”不再成立，整函数删除：

- db 包 34 个：`TestTaskDSLAppliesApprovalBeforePagination`、`TestAssetApprovalTemplates`、
  `TestHostExecutionTemplatesFailClosedAndDNSEvidenceIsTaskLocal`、`TestApprovalGroupsAtomicAndDurable`、
  `TestMixedContextLiveAuthorizationAndInheritance`、`TestContextOutputUsesLiveAuthorization`、
  `TestDerivedApprovalControlsClaimAndTestedFromParent`、`TestDerivedManualParentAndAgentTransaction`、
  `TestBlackboardRevisionTracksTaskAssetContext`、`TestFindingCommitAuthorizationUsesTransactionAndScopeLock`、
  `TestFindingCommitRejectsRestrictedAncestor`、`TestReproductionTrafficUsesLiveTaskHostPolicy`、
  `TestContextLineageAssetIDsRemainPagedAndAuthorized`、`TestProjectedPayloadAssetLineageAndLegacyFallback`、
  `TestHostOnlyApprovalTemplates`、`TestPerHostAuthorizationDoesNotCascadeToSubdomains`、
  `TestPerHostIsolationAcrossSourceTasks`、`TestMainPendingIntentPermissionAndClaims`、
  `TestPlannerBootstrapScopedCountsAndEventPriority`、`TestTargetAccessLineageAndSources`、
  `TestTaskArchiveDatabaseRoundTrip`、`TestManualAssetBlockSurvivesGlobalRecreation`、
  `TestManualAssetBlockLifecycle`、`TestTaskAssetManagementView`、
  `TestTaskAssetApprovalInheritanceAndTombstone`、`TestAgentDiscoverySideEffectsStayPendingAndManualApprovalIsDurable`、
  `TestIntentAssetsIncludesDirectSourceProvenance`、`TestInheritedApprovalListSelectsOneEffectiveSource`、
  `TestTaskNotificationsVisibilityAndGrouping`、`TestTaskNotificationsPendingOnlyAndSources`、
  `TestToolNodePageAuthorizationBeforePaging`、`TestOpenIntentPageAuthorizedCountAndPriority`、
  `TestTaskAssetSkipLifecycle`、`TestTemplateArchiveAndInvalidQuarantine`。
- traffic 包 3 个：`TestWorkerProxyPendingAndManualLimits`、`TestTrafficReadRolesRetainApprovalPolicy`、
  `TestProxyBlockIsPerHost`（`worker_policy_test.go` 全文件移除）。
- 仅剩测试代码的文件整体删除：`asset_skips_test.go`、`derived_asset_approval_test.go`、
  `derived_asset_parents_test.go`、`finding_access_test.go`、`host_isolation_test.go`、
  `main_execution_test.go`、`target_access_test.go`、`task_asset_block_recreation_test.go`、
  `task_asset_view_test.go`。

## 验证

- 新建空库 `artex_verify`，`SERIES_PG_DSN` 指向它跑 `go test ./db`：**ok，91.2s**；
- `go test ./traffic`：ok；`go test ./agent ./guard`：ok；
- `go build ./...`、`go vet ./...`：通过；
- `go test ./server`：仍有 3 个环境相关失败（缺 python3、PTY 抓不到脚本输出、报告模板长度校验过短），
  与本次改动无关，属既有基线。
