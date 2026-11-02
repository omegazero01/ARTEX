# 11 条移植的重新检查（符号级 + 行为级）

检查方法：
1. 从每个原提交的 diff 里抽出新增的顶层符号（func / type / const / export function），得到“该提交新增了什么能力”；
2. 在 ARTEX 全仓按词边界搜索同名符号，统计缺失；
3. 对缺失项逐个人工判断：是“ARTEX 没有宿主所以不适用”，还是“有落点但没搬”；
4. 对换了实现的条目，直接对比两边的判定逻辑。

判定口径：**完整** = 原提交能力全部落地；**等价** = 换实现位置但行为一致；**部分** = 有落点却没搬。

| # | 来源 → ARTEX | 原提交新增符号 | ARTEX 缺失 | 结论 |
|---|---|---|---|---|
| 1 | `95f9f05` → `79f2fed` | 9 | 0 | 完整 |
| 2 | `9de2728` → `bdf8d24` | 12 | 0 | 完整（`asset_search` 工具 ARTEX 没有，DSL 解析器已改） |
| 3 | `542be7f` → `8643b01` | 21 | 0 | 完整（读取上限用常量 30000 替代 `readOutputBudget`） |
| 4 | `3b393cd` → `86a840d` | 11 | 0 | 完整（记录点从 `dispatchHooks` 换到 `steerHooks`） |
| 5 | `67dc00c`+`8b6ef32` → `f956f46` | 24 | 0（后端） | **后端完整**：导出助手、report-zip、reporter 提示迁移已补；前端 6 文件未搬 |
| 6 | `77d45b0` → `2a971ee` | 14 | 10（仅命名不同） | 等价，见下方核对 |
| 7 | `14a567b` → `e5286c0` | 14 | 4（仅测试辅助与 workerControlMu 命名不同） | **已补齐**：取消锁、领取过滤、通知规划者、ControlWork 等待均落地 |
| 8 | `78ab23f` → `a01fd25` | 53 | 43 | **部分**：缺自审、权限校验、注入、归档、前端 |
| 9 | `f5a1b6e`+`092a081` → `baf9ef2` | 5 | 0 | 等价（说明由清空改为改写，保留了仍适用的建议） |
| 10 | `b4e19b7` → `7b9d5f5` | 0（改既有函数） | 0 | 完整，改动前后测试均在本机跑过 |
| 11 | `2d8e8a5`+`b79fb75` → `1dd98d5` | 0（纯前端） | 0 | 完整，构建产物含新按钮 |

## 第 5 条缺的具体符号

`FindingReportTitle`、`FindingReportFilename`、`SingleFindingReportMarkdown`、`FindingsReportMarkdown`、`buildFindingsReportZip`、`IsLegacyReporterPrompt`、`reseedReporterPrompt`。

影响：导出仍是 ARTEX 旧格式（外面套状态/证据），不是提交模板；文件名不取“漏洞名称”；reporter 的旧提示词不会被识别/替换。

## 第 7 条缺的具体符号

`IntentStopper`、`controlWork`、`SetCancelIntent`、`cancelIntentTool`、`intentStopTool`、`IntentStopResult`、`intentControlTask` 等 11 个。

影响：没有用户取消锁与领取过滤、没有等执行真正停止的 `controlWork`、取消后不通知规划者。

## 第 8 条缺的具体符号

`ValidateFindingSuppressionTx`、`ValidateFindingAccessTx`、`LockFindingRecordTx`、`FindingSuppressionCheck`、`FindingDeletionFeedbackContext`、`FindingFeedbackStates`、`RulesPage` 等 43 个。

影响：没有事务级权限校验、没有模型自审与反馈注入、没有禁报规则前端。

## 第 6 条（收尾禁 Bash）为什么判等价

- 原实现：`watchSettlement` 收集提示词；`observeSettlement` 只在“最后一条是 user 消息、文本等于提示词或以提示词前 32 字符开头”时置位；置位后拒绝 `Bash`；工具清单不动。
- ARTEX 实现：`NewSettlementGate` 收集 `Prompt`+`PromptByReason`；`observe` 同样看最后一条 user 消息，判定为“完全相等或以提示词开头”；置位后由 `PreToolUse` 拒绝 `Bash`；工具清单不动。
- 两侧调用点一致：原仓库只有 worker 传 `"Bash"`，chat/mainagent 不传；ARTEX 只有 worker 设 guard，其余入口 `nil`。
- 差异：变量命名与文件位置不同；匹配用“提示词全文前缀”而不是“前 32 字符”，对 SDK 追加后缀的场景同样命中。

## 结论

- 8 条完整或等价：1、2、3、4、6、9、10、11。
- 3 条部分：5（报告导出层）、7（意图取消）、8（禁报规则）。
