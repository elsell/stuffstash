#!/usr/bin/env python3
"""Exercise concurrent publication against actual local Git repositories."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).with_name('push-pages.sh').resolve()


class PagesPushTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        self.remote = self.root / 'pages.git'
        self.git(self.root, 'init', '--bare', str(self.remote))
        self.first = self.clone('first')
        self.write(self.first, 'index.html', 'production')
        self.write(self.first, 'CNAME', 'docs.example.test')
        self.write(self.first, '.nojekyll', '')
        self.write(self.first, 'pr-10/index.html', 'preview ten')
        self.commit(self.first, 'Initial site')
        self.git(self.first, 'push', 'origin', 'HEAD:gh-pages')
        self.git(self.remote, 'symbolic-ref', 'HEAD', 'refs/heads/gh-pages')
        self.second = self.clone('second')

    def git(self, cwd, *args, check=True):
        return subprocess.run(['git', *args], cwd=cwd, text=True, capture_output=True,
                              check=check, env={**os.environ, 'GIT_CONFIG_NOSYSTEM': '1'})

    def clone(self, name):
        path = self.root / name
        self.git(self.root, 'clone', str(self.remote), str(path))
        self.git(path, 'config', 'user.name', 'Pages Test')
        self.git(path, 'config', 'user.email', 'pages@example.test')
        self.git(path, 'config', 'commit.gpgsign', 'false')
        return path

    def write(self, repo, path, value):
        file = repo / path
        file.parent.mkdir(parents=True, exist_ok=True)
        file.write_text(value)

    def commit(self, repo, message):
        self.git(repo, 'add', '-A')
        self.git(repo, 'commit', '-m', message)

    def push(self, repo, succeeds=True):
        result = subprocess.run(['bash', str(SCRIPT)], cwd=repo, text=True, capture_output=True)
        self.assertEqual(result.returncode == 0, succeeds, result.stdout + result.stderr)
        return result

    def published(self, path):
        return self.git(self.remote, 'show', f'gh-pages:{path}').stdout

    def test_disjoint_preview_survives_concurrent_production_publish(self):
        self.write(self.first, 'index.html', 'new production')
        self.commit(self.first, 'Production')
        self.write(self.second, 'pr-20/index.html', 'preview twenty')
        self.commit(self.second, 'Preview')
        self.push(self.second)
        self.push(self.first)
        self.assertEqual(self.published('index.html'), 'new production')
        self.assertEqual(self.published('pr-20/index.html'), 'preview twenty')
        self.assertEqual(self.published('pr-10/index.html'), 'preview ten')
        self.assertEqual(self.published('CNAME'), 'docs.example.test')

    def test_cleanup_keeps_concurrent_preview_and_production(self):
        (self.first / 'pr-10/index.html').unlink()
        self.commit(self.first, 'Remove closed preview')
        self.write(self.second, 'pr-20/index.html', 'preview twenty')
        self.write(self.second, 'index.html', 'new production')
        self.commit(self.second, 'Other publication')
        self.push(self.second)
        self.push(self.first)
        self.assertNotEqual(self.git(self.remote, 'cat-file', '-e', 'gh-pages:pr-10/index.html', check=False).returncode, 0)
        self.assertEqual(self.published('pr-20/index.html'), 'preview twenty')
        self.assertEqual(self.published('index.html'), 'new production')

    def test_conflicting_content_fails_without_overwriting_remote(self):
        for repo, value in [(self.first, 'first'), (self.second, 'second')]:
            self.write(repo, 'index.html', value)
            self.commit(repo, value)
        self.push(self.second)
        before = self.git(self.remote, 'rev-parse', 'gh-pages').stdout
        self.push(self.first, succeeds=False)
        self.assertEqual(self.git(self.remote, 'rev-parse', 'gh-pages').stdout, before)
        self.assertEqual(self.published('index.html'), 'second')
        self.assertEqual(self.git(self.first, 'status', '--porcelain').stdout, '')

    def test_identical_concurrent_cleanup_succeeds(self):
        for repo in [self.first, self.second]:
            (repo / 'pr-10/index.html').unlink()
            self.commit(repo, 'Remove preview')
        self.push(self.second)
        self.push(self.first)
        self.assertEqual(self.published('index.html'), 'production')

    def test_continuing_contention_stops_after_five_push_attempts(self):
        self.write(self.first, 'pr-20/index.html', 'preview twenty')
        self.commit(self.first, 'Preview twenty')
        self.write(self.second, 'counter', '0')
        self.commit(self.second, 'Concurrent publication')
        self.push(self.second)
        hook = self.first / '.git/hooks/post-rewrite'
        hook.write_text(f'''#!/bin/sh
cd "{self.second}" || exit 1
value="$(cat counter)"
printf '%s' "$((value + 1))" > counter
git add counter
git commit -m "Concurrent publication" >/dev/null
git push origin HEAD:gh-pages >/dev/null 2>&1
''')
        hook.chmod(0o755)
        result = self.push(self.first, succeeds=False)
        self.assertIn('after five attempts', result.stderr)
        self.assertEqual(self.published('counter'), '4')
        self.assertNotEqual(self.git(self.remote, 'cat-file', '-e', 'gh-pages:pr-20/index.html', check=False).returncode, 0)
        self.assertEqual(self.published('index.html'), 'production')

    def test_receive_hook_rejection_is_not_retried(self):
        counter = self.root / 'received'
        hook = self.remote / 'hooks/pre-receive'
        hook.write_text(f'#!/bin/sh\necho rejected >> "{counter}"\necho publication-disabled >&2\nexit 1\n')
        hook.chmod(0o755)
        self.write(self.first, 'index.html', 'new production')
        self.commit(self.first, 'Production')
        result = self.push(self.first, succeeds=False)
        self.assertEqual(counter.read_text(), 'rejected\n')
        self.assertIn('publication-disabled', result.stderr)
        self.assertEqual(self.published('index.html'), 'production')

    def test_concurrent_initialization_preserves_both_previews(self):
        initial = self.root / 'initial'
        self.git(self.root, 'init', str(initial))
        self.git(initial, 'remote', 'add', 'origin', str(self.remote))
        self.git(initial, 'config', 'user.name', 'Pages Test')
        self.git(initial, 'config', 'user.email', 'pages@example.test')
        self.git(initial, 'config', 'commit.gpgsign', 'false')
        self.write(initial, 'pr-20/index.html', 'preview twenty')
        self.write(initial, '.nojekyll', '')
        self.commit(initial, 'Initialize Pages preview')
        self.push(initial)
        self.assertEqual(self.published('pr-20/index.html'), 'preview twenty')
        self.assertEqual(self.published('index.html'), 'production')


if __name__ == '__main__':
    unittest.main()
