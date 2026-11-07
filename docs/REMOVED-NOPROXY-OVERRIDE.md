# 恢复原版：不再清空 NO_PROXY

- 日期：2026-11-07
- 分支：`rebuild/from-series`
- 来源：清空 `NO_PROXY` 这行是 SERIES 最初提交 `2f7c18c`（2026-10-09，一次导入 300 文件）自带的，不是后续某个功能；在 ARTEX 里随 `ee807c2` 整树带入。

## 改动

`agent/worker.go` 的 proxyEnv（命令行子进程环境构造函数）删掉两行：

```
"NO_PROXY=", "no_proxy=", // task targets (including loopback) must not bypass the policy proxy
```

其余代理变量（HTTP_PROXY、HTTPS_PROXY、ALL_PROXY）和 MITM 证书变量不变。

## 效果

- 环境里原有的 `NO_PROXY`（例如 `localhost,127.0.0.1,*.corp`）恢复生效，curl、wget、python、node 等子进程访问这些地址时直连，不再被送进代理；
- 回环端口、SSH 隧道转发的本地端口、内网域名，在代理不可达时仍可访问；
- 外部目标继续走代理并录制，不受影响。

## 验证

`go build ./...`、`go vet ./...` 通过；`go test ./agent` 通过。
