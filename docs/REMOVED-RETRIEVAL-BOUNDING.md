# 删除读历史去重与检索历史占位

- 日期：2026-11-06
- 分支：`rebuild/from-series`
- 对照原版：`f48967a`（ARTEX v0.3.15 官方快照）

## 结论

原版没有这两套机制：历史工具结果一直保留全文。SERIES 加的“同查询去重”和“检索历史目录”会把旧证据正文换成占位或索引，模型跨轮看不到原文，只能重读。本项按用户选择 A 整块删除。

## 删了什么

| 文件 | 改动 |
| --- | --- |
| `agent/tool_history.go` | 删除 `normalizedReadQuery`（查询归一化）、`deduplicateToolHistory`（同查询去重）、`partialReadFailure` / `overviewHasFailure`（只服务于上述两处）；只保留 `cloneToolMessages` |
| `agent/retrieval_history.go` | 整个删除：`boundRetrievalHistory`、`clearRetrievalDirectory`、检索历史目录渲染与 `retrievalContextRule` 提示词 |
| `agent/noa_asset_context.go` | 去掉 `clearRetrievalDirectory(req)` 与 `boundRetrievalHistory(deduplicateToolHistory(out))` 两处调用 |
| `agent/tools.go` | 保留 `retrievalTool`（只用于 context-read 耗时日志） |

## 测试调整

- 删除 `agent/retrieval_history_test.go`（9 个用例，全部针对有界化/去重）；
- 删除三个只验证去重的用例：`TestContextReadHistoryDedup`、`TestPartialContextFailureDoesNotEraseEarlierEvidence`、`TestNormalizedToolHistoryKeepsPairingAndOriginal`；
- `agent/retrieval_access_test.go`、`agent/retrieval_noa_test.go`：把“旧页应被移除”的断言改成“旧页不应再出现 history_omitted 占位”，保留授权刷新与 NOA 路径验证；
- `agent/structured_history_test.go`：不再串去重函数，直接验证结构化历史脱敏；
- `agent/retrieval_fixture_test.go`：保留测试夹具 `retrievalFixture`（分页读取请求构造器）。

## 行为变化

- 模型请求里的历史工具结果保持全文，不再出现“同一查询页已有后续结果”或“检索历史目录（仅索引）”；
- 上下文变长，长任务会更早触发 SDK 自带的窗口压缩（原版同样行为）；
- 数据库、会话记录、归档不受影响，只改发给模型的请求投影。

## 原版一致度

`agent/retrieval_history.go` 在原版不存在；`agent/tool_history.go` 在原版也没有这两个函数。删除后读取投影回到原版“保留全文”的行为。

## 验证

```
go build ./...      通过
go vet ./...        无输出
go test ./agent     通过（2.4 秒）
```
