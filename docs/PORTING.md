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
| 禁报规则（**已回滚**） | `port/08-finding-suppression` | `a01fd25` → `e798b64` 撤销 | `78ab23f`（**部分**） | 新增 `finding_suppressions` 表（任务级：标题/类型/摘要/资产/原因/撤销时间）；删除漏洞时快照并写规则；`RecordFindingTx` 插入前拦截，命中返回 `ErrFindingSuppressed` 与解除指引；新增 `GET /api/finding-suppressions` 与 `DELETE /api/finding-suppressions/{id}` | 新 exe 启动后 5437 库中 `finding_suppressions` 表存在、`health=200`；3 条匹配单测（同标题命中/类型不同不命中/撤销不命中/共享资产命中/无交集不命中）通过 |
| Windows Bash 提示改写 | `port/09-bash-prompt` | `baf9ef2` | `f5a1b6e` + `092a081`（**用户已确认保留改写版**） | Windows 下把 Bash 工具的扩展说明换成改写版：删掉 `&&` 串联和 heredoc 两条 PowerShell 5.1 会报错的写法，保留“复杂任务先写脚本再执行、避免需要交互输入的命令”；`AugmentTools` 末尾统一套用，覆盖 worker/planner/mainagent/chat/reporter 全部入口 | 3 条测试通过（按系统判断、渲染后的 schema 不含 Unix 写法且基础描述/替换建议都在、只动 Bash 不动其它工具） |
| 技能路径统一斜杠 | `port/10-skill-path` | `7b9d5f5` | `b4e19b7` | 技能子目录路径统一 `/`：`skillRelPath` 改用 `path.Clean`/`path.IsAbs`，`walkSkillFiles` 返回前做 `filepath.ToSlash` | 修复前 `go test ./server -run TestSkillRelPath` 在本机 FAIL（子目录用例返回 `scripts\un.py` 这类反斜杠）；修复后 PASS；新增列表回归测试 `TestWalkSkillFilesSlashPaths` |
| 界面手动刷新按钮 | `port/11-ui-refresh` | `1dd98d5` | `2d8e8a5` + `b79fb75` | 前端四处手动刷新按钮：流量页（等流量+host 两个请求都结束再恢复）、工具执行、LLM 录制、系统日志（拉 /api/logs 快照按 seq 合并，上限 5000 行） | `tsc --noEmit` 通过；`NEXT_EXPORT=1 next build` 通过；dist 内 chunk 含「刷新流量记录」；exe 52,151,296 字节；起服 health=200 且线上 chunk 命中该按钮，验证后已停服 |
| MCP 工具调用台账 | main（直接提交） | `5579798` → `d5993ed` → `df0ffc4` → `20442b5` | SERIES `db/mcp_usage.go`（227 行）+ `server/mcp_usage.go`（46 行） | 新增 `mcp_usage` 表与三个索引，记录服务器/工具/agent/任务/探索/意图/会话；`meterMCPTool` 在 MCP 工具调用前写一行台账再原样转发，写入失败只打日志；`ListMCP` 附带 calls/tasks/usage_agents/last_used；工具列表 `MCPToolsDetailed` 合并每工具调用次数；新增 `GET /api/mcp/{id}/usage`；前端 MCP 页显示服务器/工具调用次数、工具统计与最近调用；任务归档快照采集 `mcp_usage`，恢复时按“包内是否存在该表”决定是否清空（旧归档包兼容）。不记参数与返回内容 | 真实 yakit MCP（`yak.exe mcp --transport stdio`）发现 21 个工具；经台账包装调用 `get_mitm_filter` 成功返回 MITM 过滤配置；台账校验 get_mitm_filter×1、agent=worker、task=990002；运行中实例 `GET /api/mcp` 显示 calls=1，`GET /api/mcp/122/usage` 返回同一行。`go test ./db/ -run TestMCPUsageLedger` 与 `./server -run TestMCPUsageLedgerLive` 通过；`TestMCPToolsDetailedCarriesCalls` 校验每工具次数；`TestTaskArchiveDatabaseRoundTrip` 增加断言：快照 DataCounts[mcp_usage]=1 → 热归档后清 0 → 恢复回 1；`NEXT_EXPORT=1 npx next build` 通过；浏览器实测 `/system/mcp → yakit → 工具列表` 可见“调用 N 次”徽标与“调用统计”面板（`get_mitm_filter` 1 次 · 1 个任务 · worker）；`20442b5` 修复用例清理顺序（`t.Cleanup` 在 `defer d.Close()` 之后跑导致测试残留），复跑后 `zz-%` 服务器与台账行均为 0 |
| SSE 广播缓冲满断开订阅者 | main（直接提交） | `8b42163` | `c6b6121` | `Publish` 缓冲满时摘除订阅者并 `close(ch)`（原来空 `default` 静默丢事件）、清理空 task 项；取消订阅只在“仍在集合里”时 close，避免 Publish 已关通道后 handler 的 `defer unsub()` 重复关闭 panic。SSE handler 原本就支持“通道关闭 → return → 浏览器带 Last-Event-ID 重连 → 从数据库回放缺口”，此前因通道永不关闭而触发不到 | `go test ./server/ -run TestBroadcast` 3 用例通过；RED 证据：回退 `broadcast.go` 后同一用例 FAIL“channel still open after 256 buffered events”（9.053s），恢复后 PASS（0.213s）；`go build ./...` 与 `-tags embedui` 通过；实例已重启（8787 健康）。未做端到端：开发库无任务、无 LLM，无法制造 256+ 条实时活动 |
| 工具 UA 注入（curl/wget 默认浏览器 UA） | main（直接提交） | `7915a9f` | SERIES `agent/tool_ua.go` | 启动时写 `<data>/tool-http/{.curlrc,wgetrc}`，已存在不覆盖（手改能留住）；`CURL_HOME`/`WGETRC` 由 `proxyEnv` 追加进四个 BashEnv 构造点（chat/mainagent/planner/worker），无代理时也下发（原实现无代理直接返回 nil，会把 UA 注入静默丢掉）；显式 `-A` 仍优先；配置失败只记日志不阻断启动 | `go test ./agent/ -run 'ToolUA|ProxyEnv'` 5 用例 PASS（含代理/CA 既有行为未破坏）；`TestRealCurlPicksUpUserAgent` PASS：真 curl 8.18.0 仅给 `CURL_HOME` → 服务端收到 Chrome UA，`-A pentest-override/1.0` 覆盖成功；实例重启后自动生成 `data/tool-http`，用该真实目录跑 curl 实测默认 UA 为 Chrome 串、`-A` 覆盖成功；wget 本机未安装，只做文件层校验 |
| 任务未读漏洞角标 | main（直接提交） | `247034b` | SERIES `db/task_notifications.go` + `server/task_notifications.go` + `web/src/components/task-notifications.tsx` | findings 加 `notification_xid xid8`（插入时 `pg_current_xact_id()`，旧行 NULL=已读）；`GET /api/tasks/notifications?queries=[{task_id,findings}]` 一次 SQL 批量返回未读数 + 新快照（只读，不写库）；前端逐函数移植 SERIES `components/task-notifications.tsx` + `lib/task-notifications.ts`（TaskUnreadBadge、notificationLabel、单调合并游标、基线取更早、请求串行化；无定时器，随列表刷新/可见性/游标事件触发）；点开任务推进 `artex:task-unread:<id>:findings`；**只搬漏洞一支**：新资产与待审批项依赖资产审批体系（ARTEX 无宿主，见第五节） | `go test ./db/ -run TestTaskNotifications` 2 用例 PASS（首查 0、NULL 旧行不计、新增 1、推进后 0、软删任务不出现）；`TestValidNotificationSnapshot` 12 组 PASS；接口实测：基线 `2288:2288:` findings=0 → 插入 1 条 → findings=1；浏览器实测：首次无角标 → 插 2 条后轮询出现红色角标「2」 → 点开任务写入游标 `2290:2290:` → 返回列表角标消失；测试任务与漏洞已删除（计数 0） |
| 流量正文压缩 + 保留期 | main（直接提交） | `448f8e8` | SERIES `traffic/compress.go`（131 行）+ `traffic.go` 的 retentionLoop/DeleteOlderThan + `disk_optimization_test.go` | inline 正文 >=1KB zstd 压缩，前缀 `SERIESZ1` 兼容旧行；blob 用 `.zst`/`.bin` 扩展名区分，文件名仍是明文 sha256；归档导出解压成明文、导入按需重压；保留期设置键 `traffic_retention_days`（0=永久），每小时按 500 条一批清理并回收空间；设置页新增保留期卡片。**未搬**：静态资源过滤（会丢弃 HTML/JS/CSS/图片，属减少可读信息，见第五节） | `go test ./traffic/ -count=1` 全包通过，新增 8 个用例（inline/blob 往返、旧数据直通、批删、越界拒绝、归档往返+FTS、不可压缩样本下的索引回收）；真机：36,800 字节响应 inline 421 字节、读回完整；462,900 字节响应 blob `.zst` 4,099 字节（省 99.1%），blob 接口返回 462,900 字节明文；设置 0→30→0 往返正确，非法值 400；测试流量已清理（0 行）；构建/前端构建通过 |
| shell 声明四字段 + Bash 命令计量 | main（直接提交） | `7a02662` | SERIES `db/tools.go`（shell 字段/校验）、`server/shell_usage.go`（211 行）、`server/assembly.go` 的 `shellToolNote`、`server/shell_usage_test.go`、`tools-settings.tsx` 的四个输入框 | `tools` 表加 `executable/directory/usage_help/when_to_use`（`ALTER ... IF NOT EXISTS`，旧行空=按 key 匹配）；自定义工具 REST 与 agent 侧 `create_custom_tool`/`update_custom_tool` 接受并落库四字段，可执行命令必须是命令名或绝对路径（带参数直接 400）；Bash 描述改为 SERIES 的 `shellToolNote`：名称/描述/可执行命令/所在目录/用法帮助/何时调用逐行展开；Bash 工具外面套 `shellMeteredTool`，按 `mvdan.cc/sh/v3` 解析出的字面量命令匹配声明（剥 `env/sudo/command`、支持绝对路径+目录、同命令多声明只记日志），命中写 `tool_usage`（工具/agent/任务/探索/意图/会话），台账失败不阻断执行。前端同名命令冲突提示已按原版补齐（`e94537a`）；Windows 下命令名按系统规则大小写折叠（`5153eef`） | `go test ./server/ -run TestShellUsage`：解析/目录归属/计量+失败透传/台账失败不中断 4 项 PASS；`ARTEX_PG_DSN` 下 `TestShellUsagePersistedAttribution` PASS（建声明→`GetTool` 四字段读回→真实 `Bash.Call` 写 `tool_usage` 各 1 行→`wireTools` 后 ToolResolve 的 Bash 描述含四行提示）；接口实测：`POST /api/tools/custom`（curl+/usr/bin/…）200，`GET /api/tools` 四字段读回一致，`executable="curl -sI"` 返回 400「可执行命令需为命令名或绝对路径，不含参数」，测试声明已删除（`zz_shell%` 计数 0）；库里 `tools` 四列存在；`tsc --noEmit`、`NEXT_EXPORT=1 next build`、`go build -tags embedui` 通过，实例已重启 |
| 流量读取层（search/get 分页） | main（直接提交） | 见下条提交 | SERIES `traffic/tool_reads.go`（469 行）+ `traffic/tool_reads_test.go` + `agent/worker.go` 的流量提示 | `traffic_search` 换成游标分页（首次锚定 `max(rowid)`，之后按 `(ts,id)` keyset，翻页期间新增流量不重复不漏），新增 method/status/since/until 过滤、URL 截断标记、80/443 裸域名互认；`traffic_get` 换成 `part`(请求头/请求正文/响应头/响应正文) + `offset` + `max_bytes`(≤8KB) 分段读取，返回 `total_bytes/truncated/next_offset/next`，二进制转十六进制，blob 由服务端解析；两个工具按 `ToolContext.MaxOutputChars` 控预算；`traffic_blob` 工具退役（功能并入 `traffic_get`）；worker 提示词换成 SERIES 原文 | `go test ./traffic/ -count=1` 全包通过（新增 3 用例：parts 稳定游标、二进制续读、400KB blob 分段读完整，后者 0.17s 远低于 5s 超时）；`go build ./...`、`-tags embedui` 通过；实例重启后 `GET /api/tools` 只剩 `traffic_search`/`traffic_get`，schema 为 `host/contains/body_contains/method/status/since/until/limit/cursor` 与 `id/part/offset/max_bytes`，绑定保留 `["worker","reporter"]`；`TestFindingWorkflowMigrationPreservesUserConfiguration` PASS（描述断言更新为新文案）；已知 Windows 临时目录用例仍按原状失败 |
| 工具契约（内置说明以代码为准） | main（直接提交） | 见下条提交 | SERIES `agent/tool_contract.go`（102 行）+ 其默认值校验 | 协议类内置工具的 description/schema 结构改为代码优先，库里旧文本不再覆盖；合法标量默认值保留并提示"以 Schema 为准"；非协议工具仍可编辑；默认值注入前做类型/枚举/边界校验。`server/assembly.go` 的 ToolResolve 改用它 | 4 个单测 PASS（旧 schema 不能替换契约、合法默认保留/越界丢弃、非协议工具用户描述保留、list_assets 元数据被忽略）；PG 实测：库内 traffic_get 写成旧文案+page 参数，经 ToolResolve 后仍为 part/offset，测试后自动恢复；`go build ./...`、embedui 通过，实例 health ok |
| 结构化现场复现 | main（直接提交） | `7120ba5` | SERIES `db/finding_reproduction.go`（170 行）+ `report/reproduction.go`（60）+ `agent/finding_reproduction.go`（27） | `findings.reproduction` JSONB（可空）；`report_finding.reproduction` 收 entry_point/prerequisites/steps，每步 title/action/result 必填，`traffic_refs` 在服务端解析成绑定 id（模型传 binding_ids 直接拒）；写库同事务同步探索节点 payload；技术导出无提交报告时回退分步复现；详情接口返回、前端新增复现面板。**未搬** 原版 `validateReproductionTrafficTx` 的资产范围校验 | db 5 用例（模型传 binding_ids 被拒 / Resolve 生成绑定 id 且不改输入 / 未绑定流量整笔失败 / 真实库写入校验）+ report 5 用例 + agent schema 1 用例全 PASS；`next build`、`tsc --noEmit` 通过；`GET /api/exploration/findings/99` 实测返回 reproduction；浏览器实测详情页出现「分步复现」面板（入口/前置/步骤/请求响应/实际结果/证明范围） |
| 工具调用明细 | main（直接提交） | `a2bcf0b` | SERIES `db/tool_calls.go`（178）+ `server/tool_calls.go`（241）+ `web/src/lib/tool-calls.ts`（85）+ `web/src/components/session-tool-calls.tsx`（338）+ `db/tool_calls_test.go`（250） | 新增 `GET /api/exploration/tool-calls[/{seq}]` 与 `GET /api/conversations/{id}/tool-calls[/{seq}]`：把 `activity` / `conversation_activities` 里的 tool_use 与 tool_result 按 `tool_use_id` 配成一条，列表只取元数据（工具名/类型/状态/时间），详情按 offset/limit 在 PG 里切 Unicode 字符，不把大正文读进服务端内存；会话范围由 HTTP 层解析（task+session 或 conversation id），游标是 base64 编码的 scope+q+type+status+before+snapshot；前端在任务「会话」页与「聊天」页各多一个「工具调用」页签。**宿主差异**：ARTEX 没有 `SessionSearch`（去掉对话页签里的会话检索包装，直接渲染 children）；没有 `db.RetiredToolReplacements`（builtin 名单少这一来源）；引擎原本不记录 planner 是否在跑 → 新增 `Engine.plannerActive`，在 `planner.Plan` 前后置位，用于把调用标为「执行中」 | `go build ./...` 通过；`go test ./db -run TestToolCalls -count=1` PASS（配对/筛选/翻页/详情范围 4 类）；`npx tsc --noEmit` 通过；`NEXT_EXPORT=1 npx next build` 通过。**未做端到端**：运行中的实例仍是旧二进制，需用户重启后才会出现「工具调用」页签 |

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
| 资产审批体系（模板/分组审批/封禁墓碑/来源回溯/agent 侧读写授权/派生父级授权） | SERIES `db/asset_templates.go`、`asset_approval_groups.go`、`task_asset_blocking.go`、`asset_origins.go`、`derived_asset_parents.go`、`agent/asset_context.go`（约 1159 行加前端） | “这四组先都跳过” | 这是 SERIES 的“未审批资产不可读”闸门；ARTEX 用拦截规则替代，搬过来等于给 ARTEX 加读取限制，与“渗透能力优先”冲突 | 先确认是否愿意接受读取闸门，再按块立项 |
| 人工控制三件（managed/manual 执行模式、worker 队列拖拽排序、主 Agent 下发待审批意图） | SERIES `db/execution_mode.go`、`db/worker_queue.go`、`agent/main_execution.go`（约 600 行加前端） | “这四组先都跳过” | 都是给人插手的入口，与既定“不引入人工审批阻断”冲突 | 需要先改这条原则 |
| 禁报规则 + 漏洞删除反馈 | SERIES `db/finding_suppression.go`、`finding_deletion_feedback.go`、`server/finding_suppression.go`、`agent/finding_feedback_context.go`（约 464 行） | “这四组先都跳过” | 曾移植（`a01fd25`）后被 `e798b64` 整体撤销，代码中已不存在；回滚原因未写进提交信息，重做前必须先查清 | 先查 `e798b64` 回滚原因，再决定是否按原版重搬 |

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
- **安全待修（非移植）**：`FEATURE-DIFF.md` 第 55 行记录——ARTEX 校验 JWT 时只判 Bearer 与 `artex_token` cookie，不校验签名算法（`alg`），存在算法混淆风险；SERIES 会校验 `alg` 与 subject。修它不属于搬功能，但优先级高于待搬项。

## 五、移植缺口（未完成，别当等价）

| # | 条目 | 缺什么 | 影响 | 补齐代价 |
|---|---|---|---|---|
| 5 | 报告统一 | 无（前端、任务发现页批量导出均已搬；SERIES 的 `mock/finding-export.ts` 是测试替身，未搬） | — | — |
| 7 | 意图取消 | ~~用户取消锁、领取过滤、通知规划者~~ 已补（`e5286c0`）；仅缺 SERIES 的 355 行 server 级测试 | 无功能缺口 | 可选：补等价测试 |
| 8 | 禁报规则（已回滚） | 整块能力不存在：`a01fd25` 加入的表、删除即禁报、上报前拦截、查看/解除接口，被 `e798b64` 全部撤销 | 用户删掉漏洞后，agent 会重复报同类 | 见第二节“禁报规则 + 漏洞删除反馈” |
| 13 | 任务未读角标 | SERIES 还统计“新资产”“待审批项”；ARTEX 只统计漏洞 | 资产/审批两类提示缺失 | 需先移植资产审批体系（红区，已定不搬） |
| 14 | 流量压缩/保留期 | SERIES 还有静态资源过滤（默认不录 JS/CSS/图片/字体） | 静态资源仍会入流量库 | 按“不减少可读信息”不搬；需要时再评估按 content-type 跳过 |
| 15 | shell 声明+命令计量 | Bash 解析只认字面量（`$VAR`/`$(...)`/`bash -c` 字符串不计数，SERIES 同样如此）；同名冲突提示已补（`e94537a`，浏览器实测正/负向） | 复杂拼接命令不会被归因，属原版已知限制 | 无（如需支持需自定义解析，会偏离原版） |
| 16 | shell 声明+命令计量 | `db.Tool.ValidateShellCommand` 与 `shellCommandMatches` 在 Windows 上按 bash/Windows 两种绝对路径风格归一化，并按系统规则做大小写折叠（`5153eef`；SERIES 只跑 Linux 语义） | 行为等价；差异仅在 Windows 下 `/opt/x` 可识别、`CURL.EXE` 能命中 `curl.exe` | 无需补；回到逐字一致会让 Windows 下 bash 路径与大小写差异静默漏记 |
| 17 | 流量读取层 | SERIES 的 `allowToolHost`（按任务资产审批限定可读 host）未搬：ARTEX 没有 `task_assets`/审批层，硬搬属造子系统 | 任何已抓过的 host 都能读，与移植前行为一致 | 需先移植任务资产审批体系（红区，已定不搬） |
| 18 | 流量读取层/工具契约 | 描述与 schema 已改为解析时以代码为准（ResolveBuiltinTool），库里的旧文本不再影响模型；但**工具页仍显示数据库原文**，且"代码里已删除的工具"（如 `traffic_blob`）在库里仍留行，需要手工删或等目录迁移 | 模型侧已不会按旧参数调用；界面可能显示过期说明，已删工具会出现在清单里 | 搬 #5 的目录迁移（SERIES `db/tool_context_catalog.go`）才能自动清理已删工具行 |
| 6 | 收尾禁 Bash | 与 SERIES 实现位置不同（`sdk_session.go` → worker 包装 + hook），行为等价但无原版测试 | 无功能缺口 | 可选：补等价测试 |
| 19 | 结构化现场复现 | 原版校验「复现步骤引用的流量，其主机必须在本任务资产范围内且已审批」（`validateReproductionTrafficTx`，依赖 `task_asset_links` 与审批状态）未搬 | 只能保证引用的流量已绑定到本漏洞；不能保证其目标主机在本任务授权范围内（与缺口 17 同源） | 需先有资产审批体系（红区，已定不搬） |
| 20 | 工具调用明细 | ARTEX 没有 SERIES 的 `SessionSearch`（会话内关键词检索并定位到活动）；「对话」页签里少了这条检索栏 | 只影响“在对话里搜关键词定位”，不影响工具调用列表、筛选与详情 |
| 21 | 工具调用明细 | SERIES 的 `TestToolCallsHTTPIsolationAndZeroModelWrites`（HTTP 层越权隔离 + 断言不写模型/会话表）未搬；ARTEX 只搬了参数/游标校验用例，加上 db 层 4 类用例 | 会话隔离只由解析器与 SQL 条件保证，没有 HTTP 端到端用例；重启后需手工验一次“另一会话的 id 取不到” |
