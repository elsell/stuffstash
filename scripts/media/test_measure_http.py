import http.client
import http.server
import threading
import unittest

from measure_http import measure, summarize, validate_manifest


class Handler(http.server.BaseHTTPRequestHandler):
    requests = []

    def do_GET(self):
        self.requests.append(dict(self.headers))
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"image fixture")

    def log_message(self, *_):
        pass


class MeasurementTests(unittest.TestCase):
    def test_real_http_preserves_auth_and_trace_without_exporting_secrets(self):
        server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            result = measure("https://example.test", "/private", "small", 1, 1, "private-token",
                             connection_factory=lambda *_args, **_kwargs: http.client.HTTPConnection(*server.server_address, timeout=2))
            self.assertEqual(result["status"], 200)
            self.assertEqual(result["bytes"], 13)
            self.assertGreaterEqual(result["total_ms"], result["headers_ms"])
            self.assertNotIn("private", str(result))
            self.assertEqual(Handler.requests[-1]["Authorization"], "Bearer private-token")
            self.assertIn(result["trace_id"], Handler.requests[-1]["traceparent"])
        finally:
            server.shutdown()
            server.server_close()
            thread.join()

    def test_rejects_unsafe_or_ambiguous_destinations(self):
        path = "/tenants/T/inventories/I/assets/A/attachments/P/thumbnail?variant=small"
        self.assertEqual(validate_manifest({"base_url": "https://example.test", "paths": [path]})[1], [(path, "small")])
        for origin in ["http://example.test", "https://secret@example.test", "https://example.test/path", "https://example.test?token=secret"]:
            with self.assertRaises(ValueError):
                validate_manifest({"base_url": origin, "paths": [path]})
        for bad in ["https://elsewhere.test/", "//elsewhere.test/", path + "&token=secret", path.replace("small", "private"), path.replace("/T/", "/../")]:
            with self.assertRaises(ValueError):
                validate_manifest({"base_url": "https://example.test", "paths": [bad]})

    def test_percentiles_separate_first_from_repeat_without_claiming_cache_state(self):
        values = [{"variant": "small", "attempt": attempt, "total_ms": duration, "status": 200, "outcome": "success"}
                  for attempt, duration in [(1, 100), (1, 200), (2, 10), (2, 20)]]
        summary = summarize(values)
        self.assertEqual(summary["small:first"]["p95_ms"], 200)
        self.assertEqual(summary["small:repeat"]["p50_ms"], 10)


if __name__ == "__main__":
    unittest.main()
