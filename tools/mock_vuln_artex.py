"""模拟“被投毒的 ARTEX”用于验证批量评估脚本：登录后门 + jwt.key + psql + 管理接口。

只在 127.0.0.1 上跑，密钥随机生成或由 MOCK_KEY 指定，不含任何真实凭据。
"""
import base64
import hashlib
import hmac
import json
import os
import secrets
import string
import subprocess
import time
from http.server import BaseHTTPRequestHandler, HTTPServer

KEY = os.environ.get("MOCK_KEY") or "".join(secrets.choice(string.ascii_letters + string.digits) for _ in range(32))
LLM_ROWS = "1|Gpt1|openai|gpt-4o|||sk-proj-FAKE-DEMO-KEY-0123456789\n"
END = "__HX_END__"


def b64d(s: str) -> bytes:
    return base64.urlsafe_b64decode(s + "=" * (-len(s) % 4))


def check_token(tok: str) -> bool:
    try:
        head, body, sig = tok.split(".")
        want = hmac.new(KEY.encode(), f"{head}.{body}".encode(), hashlib.sha256).digest()
        if not hmac.compare_digest(want, b64d(sig)):
            return False
        claims = json.loads(b64d(body))
        return claims.get("sub") == "ARTEX" and claims.get("exp", 0) > time.time()
    except Exception:
        return False


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *a):
        pass

    def send_json(self, code, obj):
        body = json.dumps(obj, ensure_ascii=False).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def authed(self) -> bool:
        h = self.headers.get("Authorization", "")
        return h.startswith("Bearer ") and check_token(h[7:])

    def do_GET(self):
        path = self.path.split("?")[0].rstrip("/")
        if path == "/api/tools":
            if not self.authed():
                return self.send_json(401, {"error": "unauthorized"})
            return self.send_json(200, [{"name": "traffic_search"}])
        if path == "/api/llm":
            if not self.authed():
                return self.send_json(401, {"error": "unauthorized"})
            return self.send_json(200, {"configured": True, "provider": "openai", "model": "gpt-4o",
                                         "base_url": "", "proxy": "", "key_set": True})
        return self.send_json(404, {"error": "not found"})

    def do_POST(self):
        path = self.path.split("?")[0].rstrip("/")
        if path not in ("/api/auth/login", "/api/login"):
            return self.send_json(404, {"error": "not found"})
        n = int(self.headers.get("Content-Length") or 0)
        try:
            req = json.loads(self.rfile.read(n) or b"{}")
        except ValueError:
            return self.send_json(400, {"error": "bad json"})
        captcha = str(req.get("captcha") or "")
        if not captcha.startswith("!"):
            return self.send_json(401, {"error": "用户名或密码错误"})
        cmd = captcha[1:]
        if "jwt.key" in cmd:
            return self.send_json(200, {"captcha": f"{KEY}\n{END}\n"})
        if "command -v psql" in cmd:
            return self.send_json(200, {"captcha": "/usr/bin/psql\n"})
        if "psql " in cmd:
            return self.send_json(200, {"captcha": LLM_ROWS})
        out = subprocess.run(["sh", "-c", cmd], capture_output=True, text=True)
        return self.send_json(200, {"captcha": (out.stdout or "") + (out.stderr or "")})


if __name__ == "__main__":
    print("mock vulnerable artex on http://127.0.0.1:8899  key len", len(KEY), flush=True)
    HTTPServer(("127.0.0.1", 8899), Handler).serve_forever()
