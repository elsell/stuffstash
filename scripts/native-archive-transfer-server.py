#!/usr/bin/env python3
"""Loopback-only native transport fixture; never handles production credentials."""
import argparse
import hashlib
import json
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

LIMIT = 16 * 1024 * 1024
release = threading.Event()

class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_args):
        pass

    def do_GET(self):
        if self.path == '/release':
            release.set()
        self.send_response(200 if self.path in ('/health', '/release') else 404)
        self.send_header('Content-Length', '0')
        self.end_headers()

    def do_POST(self):
        if self.headers.get('Authorization') != 'Bearer native-archive-fixture' or self.headers.get('Idempotency-Key') != 'native-fixture':
            self.send_error(403)
            return
        content = bytearray()
        if self.headers.get('Transfer-Encoding', '').lower() == 'chunked':
            while True:
                size = int(self.rfile.readline(128).split(b';')[0], 16)
                if size == 0:
                    self.rfile.readline(128)
                    break
                if size < 0 or len(content) + size > LIMIT:
                    self.send_error(413)
                    return
                content.extend(self.rfile.read(size))
                if self.rfile.read(2) != b'\r\n':
                    self.send_error(400)
                    return
        else:
            size = int(self.headers.get('Content-Length', '-1'))
            if size < 0 or size > LIMIT:
                self.send_error(413)
                return
            content.extend(self.rfile.read(size))
        if self.path == '/stall':
            release.clear()
            release.wait(15)
        if self.path == '/redirect':
            self.send_response(302)
            self.send_header('Location', '/forbidden')
            self.send_header('Content-Length', '0')
            self.end_headers()
            return
        body = b'x' * 65537 if self.path == '/oversized' else json.dumps({'size': len(content), 'sha256': hashlib.sha256(content).hexdigest()}).encode()
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        try:
            self.wfile.write(body)
        except (BrokenPipeError, ConnectionResetError):
            pass

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--port', required=True, type=int)
    args = parser.parse_args()
    ThreadingHTTPServer(('127.0.0.1', args.port), Handler).serve_forever()
