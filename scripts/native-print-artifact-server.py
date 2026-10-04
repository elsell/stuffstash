#!/usr/bin/env python3
"""Loopback-only PNG peer for the isolated native printing audit; never prints."""
import argparse
import base64
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from pathlib import Path
import re

root = Path(__file__).resolve().parents[1]
sample = (root / 'apps/mobile/native-audit/PrintingLabelSample.ts').read_text()
png = base64.b64decode(re.search(r"printingLabelPNG = '([^']+)'", sample)[1], validate=True)
scope = '/tenants/tenant/inventories/inventory'
asset = scope + '/assets/printing-item'
content = scope + '/label-renders/native-preview/content'

class ArtifactHandler(BaseHTTPRequestHandler):
    def log_message(self, *_): pass
    def send(self, status, body, media='application/json'):
        self.send_response(status)
        self.send_header('Content-Type', media)
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def json(self, data): self.send(200, json.dumps({'data': data, 'meta': {}}).encode())
    def do_GET(self):
        if self.path == '/health': self.json({'ready': True})
        elif self.path == content: self.send(200, png, 'image/png')
        else: self.send(404, b'{}')
    def do_POST(self):
        length = int(self.headers.get('Content-Length', '0'))
        if length > 65536: self.send(413, b'{}'); return
        self.rfile.read(length)
        if self.path == asset + '/label': self.json({'id': 'native-label'})
        elif self.path == asset + '/label-renders':
            self.json({'id': 'native-preview', 'selectionFingerprint': 'native-selection', 'widthPixels': 306, 'heightPixels': 991, 'displayRotation': 270})
        else: self.send(404, b'{}')

def make_server(port): return ThreadingHTTPServer(('127.0.0.1', port), ArtifactHandler)
if __name__ == '__main__':
    parser = argparse.ArgumentParser(); parser.add_argument('--port', type=int, required=True)
    make_server(parser.parse_args().port).serve_forever()
