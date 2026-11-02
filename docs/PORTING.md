# ARTEX 移植进度与跳过清单

本文件记录「把 SERIES 二开提交搬到 ARTEX」的进度、每条的依据与验证方式，以及**明确跳过**的条目。
更新方式：每完成或跳过一条，追加一行；不要只改结论，要带上证据命令或产物路径。

- 工作区：`E:\AI\ARTEX-mian`（`omegazero01/ARTEX` 的干净 clone，基线 `f48967a` / v0.3.15）
- 来源仓库：`E:\AI\series`（`omegazero01/series`，`local-dev` 分支）
- 本地端口：ARTEX 5437（PG）/ 8787（API）/ 8788（代理）；SERIES 5435 / 8797 / 8798，两套可并存
- 纪律：一个能力块一个分支，先小后大；每条必须有可测指标；不按时间重放；只信代码不信提交标题

## 一、已完成移植

| 条目 | 分支 | ARTEX 提交 | 来源提交 | 内容 | 验证证据 |
|---|---|---|---|---|---|
| 流量正文解码 | `port/01-traffic-decode` | `79f2fed` | `95f9f05` | 流量正文按 Content-Encoding 解码（gzip/deflate/zlib/br/zstd，多重编码），原始字节仍存盘，索引/预览用解码视图 | `go test ./traffic/` ok；经 8788 抓 gzip 响应，`_index` 中响应体前两字节为 `1f 8b`，FTS 命中 `abc123`、`内网测` |
| 主机环境说明 + DSL 放宽 | `port/02-host-profile-dsl` | `bdf8d24` | `9de2728` | 启动时探测本机环境（PowerShell 版本、curl 别名、curl.exe 路径、临时目录）拼进 Bash 工具说明；资产 DSL 允许运算符两侧空格、`contains`/`=~` 模糊匹配、引号取值、报错带示例 | `go build ./...`=0；4 条 DSL 测试通过；真实探测输出 `len=792`，识别 PowerShell 5.1、curl 为 Invoke-WebRequest 别名 |
| 长期记忆工具 | `port/03-memory-tools` | `8643b01` | `542be7f` | worker 记忆四工具（列目录/搜索/读取/保存），中文双字滑窗检索，纯中文名用哈希当文件名防覆盖，写完重建 MEMORY.md | 6 条记忆测试全过；`go build`=0；新 exe 启动 `health=200` |
| 报错沉淀本机经验 | `port/04-host-lessons` | `86a840d` | `3b393cd` | 已知工具报错按签名沉淀（去重计数），只存签名与提示、不存原始命令/输出，把命中最高的 6 条追加到 Bash 描述 | 5 条测试通过；真实调用记录两次同一 `head` 报错 → `count=2`、经验提示生成、`lessons.json` 落盘 |
| 统一提交报告模板 | `port/05-report` | `f956f46` → `d29b10c` / 前端 `cf1681a` + 任务页导出 | `67dc00c` + `8b6ef32`（**已补齐**） | 统一提交报告模板（漏洞名称/等级/类型/URL + 详细描述/复现步骤/修复方案）；字数预算（350/200/400 字）；技术附录用 `<!-- SERIES:TECHNICAL -->` 切分，不导出；新增 `findings.report_detail` | 报告格式测试（含 11 条预算子用例）全过；新 exe 启动后 `findings` 表出现 `report_detail`（启动时自动迁移）；`GetFinding` 读回新列正常；任务发现页批量导出（勾选/全选 → `md-report-zip`）已补，`npx tsc --noEmit` 与 `NEXT_EXPORT=1 next build` 通过，产物 chunk 命中「导出提交报告」 |
| 收尾闸门改用 hook | `port/06-settlement-hook` | `2a971ee` | `77d45b0` | 收尾相位不再把 Bash 从工具清单删掉（会改写请求前缀、缓存全失效）；改为 gate 观察每次请求，命中收尾提示词后由 `steerHooks.PreToolUse` 拒绝执行，工具清单全程不变 | 6 条单测（置位/前缀/PromptByReason/惰性/包装透传/ctx）；`go build ./...`=0；`go test ./agent/ ./db/` 全绿 |
| 主代理取消意图 | `port/07-intent-cancel` | `caa4a13` → `e5286c0` | `14a567b`（**已补齐**） | 新增 `cancel_intent` 工具（默认只绑 mainagent）；`Engine.StopIntent`：运行中 worker 用 `AbortCancelledByUser` 取消并轮询等待退出注册表（15s 上限），等待/暂停/阻塞意图直接置 `stopped`，产出保留，写 mainagent 活动记录 | build + vet 干净；`go test ./agent/ ./db/` 全绿；未搬 SERIES 的 355 行 server 级测试 |
| 禁报规则 | `port/08-finding-suppression` | `a01fd25` | `78ab23f`（**部分**） | 新增 `finding_suppressions` 表（任务级：标题/类型/摘要/资产/原因/撤销时间）；删除漏洞时快照并写规则；`RecordFindingTx` 插入前拦截，命中返回 `ErrFindingSuppressed` 与解除指引；新增 `GET /api/finding-suppressions` 与 `DELETE /api/finding-suppressions/{id}` | 新 exe 启动后 5437 库中 `finding_suppressions` 表存在、`health=200`；3 条匹配单测（同标题命中/类型不同不命中/撤销不命中/共享资产命中/无交集不命中）通过 |
| Windows Bash 提示改写 | `port/09-bash-prompt` | `baf9ef2` | `f5a1b6e` + `092a081`（**用户已确认保留改写版**） | Windows 下把 Bash 工具的扩展说明换成改写版：删掉 `&&` 串联和 heredoc 两条 PowerShell 5.1 会报错的写法，保留“复杂任务先写脚本再执行、避免需要交互输入的命令”；`AugmentTools` 末尾统一套用，覆盖 worker/planner/mainagent/chat/reporter 全部入口 | 3 条测试通过（按系统判断、渲染后的 schema 不含 Unix 写法且基础描述/替换建议都在、只动 Bash 不动其它工具） |
| 技能路径统一斜杠 | `port/10-skill-path` | `7b9d5f5` | `b4e19b7` | 技能子目录路径统一 `/`：`skillRelPath` 改用 `path.Clean`/`path.IsAbs`，`walkSkillFiles` 返回前做 `filepath.ToSlash` | 修复前 `go test ./server -run TestSkillRelPath` 在本机 FAIL（子目录用例返回 `scripts\un.py` 这类反斜杠）；修复后 PASS；新增列表回归测试 `TestWalkSkillFilesSlashPaths` |
| 界面手动刷新按钮 | `port/11-ui-refresh` | `1dd98d5` | `2d8e8a5` + `b79fb75` | 前端四处手动刷新按钮：流量页（等流量+host 两个请求都结束再恢复）、工具执行、LLM 录制、系统日志（拉 /api/logs 快照按 seq 合并，上限 5000 行） | `tsc --noEmit` 通过；`NEXT_EXPORT=1 next build` 通过；dist 内 chunk 含「刷新流量记录」；exe 52,151,296 字节；起服 health=200 且线上 chunk 命中该按钮，验证后已停服 |
| MCP 工具调用台账 | main（直接提交） | `5579798` → `d5993ed` → `df0ffc4` → `20442b5` | SERIES `db/mcp_usage.go`（227 行）+ `server/mcp_usage.go`（46 行） | 新增 `mcp_usage` 表与三个索引，记录服务器/工具/agent/任务/探索/意图/会话；`meterMCPTool` 在 MCP 工具调用前写一行台账再原样转发，写入失败只打日志；`ListMCP` 附带 calls/tasks/usage_agents/last_used；工具列表 `MCPToolsDetailed` 合并每工具调用次数；新增 `GET /api/mcp/{id}/usage`；前端 MCP 页显示服务器/工具调用次数、工具统计与最近调用；任务归档快照采集 `mcp_usage`，恢复时按“包内是否存在该表”决定是否清空（旧归档包兼容）。不记参数与返回内容 | 真实 yakit MCP（`yak.exe mcp --transport stdio`）发现 21 个工具；经台账包装调用 `get_mitm_filter` 成功返回 MITM 过滤配置；台账校验 get_mitm_filter×1、agent=worker、task=990002；运行中实例 `GET /api/mcp` 显示 calls=1，`GET /api/mcp/122/usage` 返回同一行。`go test ./db/ -run TestMCPUsageLedger` 与 `./server -run TestMCPUsageLedgerLive` 通过；`TestMCPToolsDetailedCarriesCalls` 校验每工具次数；`TestTaskArchiveDatabaseRoundTrip` 增加断言：快照 DataCounts[mcp_usage]=1 → 热归档后清 0 → 恢复回 1；`NEXT_EXPORT=1 npx next build` 通过；浏览器实测 `/system/mcp → yakit → 工具列表` 可见“调用 N 次”徽标与“调用统计”面板（`get_mitm_filter` 1 次 · 1 个任务 · worker）；`20442b5` 修复用例清理顺序（`t.Cleanup` 在 `defer d.Close()` 之后跑导致测试残留），复跑后 `zz-%` 服务器与台账行均为 0 |

## 二、明确跳过（用户拍板，暂不做）

| 条目 | 来源提交 | 用户原话 | 跳过的理由 | 将来要做需要什么 |
|---|---|---|---|---|
| 智能代理（被拦截主机自动切代理池） | `36213bd` | “智能代理先不加，下一个” | 不是纯搬运：需要 guard 侧补代理标记、`server/manager.go` 接线，先造底座再搬功能 | 代理池探测与健康检查、guard 标记注入、按主机切换出口的开关与回退 |
| 事实加权（规划者事实打分） | `45d7fde` | “事实加权先不做” | 与 saspper 省钱分析绑定，权重参数需按真实任务标定，没有实测支撑不建议先上 | 真实任务的多轮规划数据（每轮事实条数、命中率），再定权重与档位 |
| 一跳上下文（`derived_from` 一跳事实补入概览） | `45d7fde` | “跳过这个一跳上下文” | ARTEX 没有 `agent/overview_lean.go`、也没有 lean 精简概览这一层，单独搬“一跳”没有落点；属于造子系统而非移植 | 先建 lean 事实选择层（按当前 open/running 意图召回），再叠加一跳与打分；可用开关灰度 |
| 记忆注入三件套（pinned 常驻注入 / 按意图召回 3 条 / 经验写入去抖） | `5252215` | “这三先跳过” | 与已移植的记忆工具同源，但会拉长每轮系统提示（pinned ≤900 字 + 3 条相关记忆），且中文短意图的召回质量未实测；先去抖/注入都不影响现有四个记忆工具可用 | 真实任务的 token 与召回命中数据；确认后可先只开 pinned、把 relevant 留开关 |
| 压缩阈值可配置（清理线/摘要线改绝对值） | `3b4bb93` | “压缩阈值可配跳过” | ARTEX 没有 `agent/context_policy.go` 这层压缩接管代码，属造底座；且它改的是“何时丢历史”，有损摘要可能把关键线索压没，必须用真实长会话标定 | 先有上下文压缩接管层（`agentcore.Options.Compactor`），再用真实长会话的 token 曲线定清理线/摘要线与保留条数 |
| 发现页导出弹窗去掉重复 Markdown 选项 | `5fd06b7` | “其他先不迁移” | ARTEX 后端没有 `md-report` 导出格式，前端照搬会让导出直接报错；要做必须先补后端 | 先补后端导出格式，再改前端选项 |
| 分栏折叠 | `452d6c7` | “其他先不迁移” | 要新造 `finding-presentation` 组件（ARTEX 没有），收益是界面舒适度，不影响探测 | 想要再做，先比对发现页现有渲染结构 |
| 漏洞归并（finding cases） | `3bc602e`（后端）+ `9a52f5d`、`e19ea54`、`5426444`（前端） | “这个先跳过” | 造子系统：后端约 3000 行（含 637 行 db、558 行 server、3 张新表），前端约 1000 行；只影响报告整理（同一漏洞多条合成一案），不影响探测能力 | 先出差异清单（`report/findings.go`、`server/finding_traffic_export.go` 在 ARTEX 的对应物），再决定 |
| 联盟系统（台账/画像/反哺） | `46903b4`、`3ceae46`、`977909c`、`4c89e3a` 等约 10 条 | “其他先不迁移” | 约 4900 行、33 文件，是城商行联盟定制业务，与通用探测无关 | 按客户定制需求单独立项 |

补充：第 5 条的前端已搬（`cf1681a`：`api.ts`、发现页导出弹窗、详情页「提交报告」+ 技术附录折叠）；任务发现页的批量导出入口也已补齐（`findings-tab.tsx`：勾选/全选 → `md-report-zip`）。SERIES 的 `mock/finding-export.ts` 是测试替身，ARTEX mock 层没有对应实现，未搬，不影响真实环境导出。

补充：第 6～8 条的移植差距（别误以为与 SERIES 完全等价）：

- **第 6 条（收尾禁 Bash）**：ARTEX 没有 `agent/sdk_session.go`，收尾检测改成 worker 包装 `QueryDeps`，再用 ctx 把 gate 传给 hooks；planner 的 settlement 本来就不删工具，未改。
- **第 7 条（意图取消）**：ARTEX 的 `RunInfo` 没有 AgentKey，角色限制靠工具绑定（只绑 mainagent）+ 任务/意图上下文校验；未搬 SERIES 的 `workerControlMu`、显式的用户取消锁字段和 355 行 server 级测试。`stopped` 在 ARTEX 不会被自动领取（claim 只取 `open`），但没有“用户取消”专用的锁字段。
- **第 8 条（禁报规则）**：只搬核心（表 + 删除即禁报 + 上报前拦截 + 查看/解除接口）。**未搬**：上报前的模型自审、资产失效/worker 停止的事务级权限校验（SERIES 的 `finding_access.go` 一层）、禁报清单注入 planner/worker 系统提示、前端管理页。模型会先尝试上报、被拒绝后再学习。

## 三、尚未决定（未开工）

暂无：本轮可选的 SERIES 提交已全部处理完（完成 11 条、明确跳过 8 条）。新的来源提交先按第一节的格式补记，再决定做/跳过。

## 四、验证方式与遗留问题

- 每条移植后固定跑：`go build ./...`、`go test ./agent/ ./db/`、`go build -tags embedui -o artex.exe ./cmd/artex`，并起服务确认 `/api/health=200`。
- ARTEX 未配 LLM，所以**没有跑过真实任务**：工具挂载、报告校验、经验注入目前只有单元测试与库级验证。
- Windows 下 `TestFindingWorkflowReporterBindsBeforeWritingReport` 因临时目录证据文件同步报 `Access is denied`，与报告改动无关。
- `server/webui/dist/index.html` 被替换为跳转 `/login/` 的静态页（静态导出不支持服务端 redirect），属产物覆盖，不是源码改动。

## 五、移植缺口（未完成，别当等价）

| # | 条目 | 缺什么 | 影响 | 补齐代价 |
|---|---|---|---|---|
| 5 | 报告统一 | 无（前端、任务发现页批量导出均已搬；SERIES 的 `mock/finding-export.ts` 是测试替身，未搬） | — | — |
| 7 | 意图取消 | ~~用户取消锁、领取过滤、通知规划者~~ 已补（`e5286c0`）；仅缺 SERIES 的 355 行 server 级测试 | 无功能缺口 | 可选：补等价测试 |
| 8 | 禁报规则 | 上报前模型自审、资产/worker 事务级权限校验、禁报清单注入 planner/worker 提示、前端规则管理页 | 模型会先尝试上报一次再被拒；无越权兜底 | 约 500–700 行 |
| 6 | 收尾禁 Bash | 与 SERIES 实现位置不同（`sdk_session.go` → worker 包装 + hook），行为等价但无原版测试 | 无功能缺口 | 可选：补等价测试 |
