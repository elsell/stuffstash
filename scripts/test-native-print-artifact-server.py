import importlib.util
import json
from pathlib import Path
import threading
import unittest
from urllib.error import HTTPError
from urllib.request import Request, urlopen

spec = importlib.util.spec_from_file_location('peer', Path(__file__).with_name('native-print-artifact-server.py'))
peer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(peer)

class PrintArtifactPeerTests(unittest.TestCase):
    def setUp(self):
        self.server = peer.make_server(0)
        threading.Thread(target=self.server.serve_forever, daemon=True).start()
        self.addCleanup(self.server.server_close)
        self.addCleanup(self.server.shutdown)
        self.base = f'http://127.0.0.1:{self.server.server_port}'

    def test_scoped_render_content_is_real_png_with_matching_geometry(self):
        with urlopen(Request(self.base + peer.asset + '/label-renders', data=b'{}', headers={'Content-Type': 'application/json'})) as response:
            render = json.load(response)['data']
        with urlopen(self.base + peer.content) as response:
            data = response.read()
            self.assertEqual(response.headers['Content-Type'], 'image/png')
        self.assertEqual(data[:8], b'\x89PNG\r\n\x1a\n')
        self.assertEqual(int.from_bytes(data[16:20], 'big'), render['widthPixels'])
        self.assertEqual(int.from_bytes(data[20:24], 'big'), render['heightPixels'])

    def test_unknown_scope_and_output_requests_are_rejected(self):
        for path in ['/tenants/other/inventories/inventory/label-renders/native-preview/content', peer.asset + '/print-jobs']:
            with self.assertRaises(HTTPError) as error:
                urlopen(Request(self.base + path, data=b'{}'))
            self.assertEqual(error.exception.code, 404)

if __name__ == '__main__': unittest.main()
