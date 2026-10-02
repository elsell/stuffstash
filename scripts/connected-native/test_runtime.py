import importlib.util
from pathlib import Path
import subprocess
import sys
import threading
from http.server import BaseHTTPRequestHandler, HTTPServer
import unittest
import tempfile
import json

spec = importlib.util.spec_from_file_location('connected_runtime', Path(__file__).with_name('runtime.py'))
runtime = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runtime)

class RuntimeTests(unittest.TestCase):
    def test_evidence_excludes_auth_screenshots_and_runtime_logs(self):
        spec = importlib.util.spec_from_file_location('native_evidence', Path(__file__).with_name('export-evidence.py'))
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            exported = root / 'private'
            evidence = root / 'public'
            exported.mkdir(); evidence.mkdir()
            names = ['connected-safe-persisted-detail_0.png', 'connected-stage_0.txt', 'connected-failure_0.txt',
                     'FailureScreenshot_0.png', 'auth-token_0.txt', 'runtime-log_0.txt']
            attachments = []
            for index, name in enumerate(names):
                filename = str(index) + Path(name).suffix
                (exported / filename).write_text(name)
                attachments.append({'suggestedHumanReadableName': name, 'exportedFileName': filename})
            (exported / 'manifest.json').write_text(json.dumps([{'attachments': attachments}]))
            module.retain_evidence(exported, evidence)
            self.assertEqual(sorted(p.name for p in evidence.iterdir()),
                             ['connected-failure.txt', 'connected-safe-persisted-detail.png', 'connected-stage.txt'])

    def test_cleanup_reaps_running_and_already_exited_services(self):
        live = subprocess.Popen([sys.executable, '-c', 'import time; time.sleep(60)'], start_new_session=True)
        exited = subprocess.Popen([sys.executable, '-c', 'pass'], start_new_session=True)
        exited.wait()
        try:
            runtime.stop([live, exited])
            self.assertIsNotNone(live.poll())
            self.assertIsNotNone(exited.poll())
        finally:
            runtime.stop([live, exited])

    def test_anonymous_discovery_requires_401_on_existing_route(self):
        class Discovery(BaseHTTPRequestHandler):
            denial = 401
            def do_GET(self):
                self.send_response(self.denial if self.path == '/me/tenants' else 405)
                self.end_headers()
            def log_message(self, *args):
                pass
        server = HTTPServer(('127.0.0.1', 0), Discovery)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            base = 'http://127.0.0.1:' + str(server.server_port)
            self.assertEqual(runtime.verify_anonymous_discovery(base), 401)
            for status in [200, 403, 404, 500]:
                Discovery.denial = status
                with self.assertRaisesRegex(RuntimeError, str(status)):
                    runtime.verify_anonymous_discovery(base)
        finally:
            server.shutdown()
            server.server_close()
            thread.join()

    def test_readiness_requires_live_process_and_successful_http(self):
        class Ready(BaseHTTPRequestHandler):
            def do_GET(self):
                self.send_response(200)
                self.end_headers()
            def log_message(self, *args):
                pass
        server = HTTPServer(('127.0.0.1', 0), Ready)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        process = subprocess.Popen([sys.executable, '-c', 'import time; time.sleep(60)'], start_new_session=True)
        url = 'http://127.0.0.1:' + str(server.server_port)
        try:
            runtime.wait_http(url, [process], timeout=2)
            runtime.stop([process])
            with self.assertRaisesRegex(RuntimeError, 'exited before readiness'):
                runtime.wait_http(url, [process], timeout=2)
        finally:
            runtime.stop([process])
            server.shutdown()
            server.server_close()
            thread.join()

if __name__ == '__main__':
    unittest.main()
