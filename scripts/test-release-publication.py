#!/usr/bin/env python3
import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest


def load(name):
    spec = importlib.util.spec_from_file_location(name, Path(__file__).with_name(name + '.py'))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


publication = load('publish-release-assets')
docs = load('generate-cli-downloads')


class ReleaseStore:
    """Faithful mutable remote release/asset store, including lost upload replies."""
    repo = 'example/stuffstash'

    def __init__(self, commit):
        self.commit = commit
        self.release = None
        self.bytes = {}
        self.disconnect_after_upload = False
        self.has_newer_stable = False
        self.hide_new_draft = False

    def tag_commit(self, _tag):
        return self.commit

    def find(self, _tag):
        return None if self.hide_new_draft else copy.deepcopy(self.release)

    def create(self, plan, _root):
        self.release = dict(id=123, tag_name=plan['tag'], draft=True, prerelease=False, assets=[], published_at=None)
        return copy.deepcopy(self.release)

    def refresh(self, release):
        if release['id'] != self.release['id']: raise ValueError('Unknown release')
        return copy.deepcopy(self.release)

    def upload(self, _tag, path):
        if path.name in self.bytes:
            raise ValueError('Duplicate asset upload')
        self.bytes[path.name] = path.read_bytes()
        self.release['assets'].append(dict(id=len(self.bytes), name=path.name))
        if self.disconnect_after_upload:
            self.disconnect_after_upload = False
            raise ConnectionError('Upload completed but response lost')

    def download(self, asset, target):
        target.write_bytes(self.bytes[asset['name']])

    def newer_stable(self, _tag):
        return self.has_newer_stable

    def publish(self, _tag):
        self.release.update(draft=False, published_at='2026-10-03T00:00:00Z')


def fixture(root):
    tag, commit = 'v1.2.3', 'a' * 40
    assets = []
    for system, arch in [('linux','amd64'),('linux','arm64'),('darwin','amd64'),('darwin','arm64'),('windows','amd64')]:
        name = f'stuffstash_{tag}_{system}_{arch}' + ('.zip' if system == 'windows' else '.tar.gz')
        path = root / name
        path.write_bytes((system + arch + commit).encode())
        sha = publication.digest(path)
        (root / (name + '.sha256')).write_text(f'{sha}  {name}\n')
        assets.append(dict(os=system, architecture=arch, name=name, sha256=sha, usbPrinting=False))
    (root / 'stuffstash-cli-release.json').write_text(json.dumps(dict(version=tag, commit=commit, assets=assets)))
    path = root / 'stuffstash-selfhost.tar.gz'; path.write_bytes(b'selfhost bytes')
    (root / (path.name + '.sha256')).write_text(f'{publication.digest(path)}  {path.name}\n')
    plan = dict(tag=tag, commit=commit, assets=[dict(name=p.name, sha256=publication.digest(p)) for p in root.iterdir()])
    (root / 'publication.json').write_text(json.dumps(plan))
    (root / 'release-notes.md').write_text('Exact release notes')
    return plan


class PublicationTests(unittest.TestCase):
    def test_new_draft_need_not_be_immediately_visible_in_collection(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp); plan = fixture(root); remote = ReleaseStore(plan['commit'])
            remote.hide_new_draft = True
            metadata = publication.publish(root, remote)
            self.assertEqual(metadata['version'], plan['tag'])
            self.assertFalse(remote.release['draft'])
            self.assertEqual(len(remote.bytes), len(plan['assets']))

    def test_lost_upload_response_recovers_original_bytes_before_stable_docs(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp); plan = fixture(root); remote = ReleaseStore(plan['commit'])
            remote.disconnect_after_upload = True
            with self.assertRaises(ConnectionError):
                publication.publish(root, remote)
            self.assertTrue(remote.release['draft'])
            metadata = publication.publish(root, remote)
            self.assertFalse(remote.release['draft'])
            self.assertEqual(len(remote.bytes), len(plan['assets']))
            self.assertEqual(publication.publish(root, remote), metadata)
            rendered = docs.render(metadata)
            self.assertIn('/releases/download/v1.2.3/stuffstash_v1.2.3_linux_amd64.tar.gz', rendered)
            self.assertIn(metadata['assets'][0]['sha256'], rendered)
            self.assertNotIn('/latest/', rendered)

    def test_tampering_or_different_tag_target_never_publishes_or_overwrites(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp); plan = fixture(root); remote = ReleaseStore('b' * 40)
            with self.assertRaisesRegex(ValueError, 'tag'):
                publication.publish(root, remote)
            self.assertIsNone(remote.release)
            remote.commit = plan['commit']; remote.disconnect_after_upload = True
            with self.assertRaises(ConnectionError):
                publication.publish(root, remote)
            first = next(iter(remote.bytes)); remote.bytes[first] = b'other bytes'
            with self.assertRaisesRegex(ValueError, 'refusing overwrite'):
                publication.publish(root, remote)
            self.assertTrue(remote.release['draft'])
            self.assertEqual(remote.bytes[first], b'other bytes')
            (root / first).write_bytes(b'local tampering')
            with self.assertRaisesRegex(ValueError, 'checksum mismatch'):
                publication.publish(root, remote)

    def test_old_repair_finishes_assets_without_advertising_an_older_download(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp); plan = fixture(root); remote = ReleaseStore(plan['commit'])
            remote.has_newer_stable = True
            self.assertIsNone(publication.publish(root, remote))
            self.assertFalse(remote.release['draft'])
            self.assertEqual(len(remote.bytes), len(plan['assets']))

    def test_incomplete_archive_matrix_cannot_be_advertised(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp); plan = fixture(root)
            manifest = root / 'stuffstash-cli-release.json'; data = json.loads(manifest.read_text())
            data['assets'].pop(); manifest.write_text(json.dumps(data))
            for asset in plan['assets']:
                if asset['name'] == manifest.name: asset['sha256'] = publication.digest(manifest)
            (root / 'publication.json').write_text(json.dumps(plan))
            remote = ReleaseStore(plan['commit'])
            with self.assertRaisesRegex(ValueError, 'matrix'):
                publication.publish(root, remote)
            self.assertIsNone(remote.release)


if __name__ == '__main__':
    unittest.main()
