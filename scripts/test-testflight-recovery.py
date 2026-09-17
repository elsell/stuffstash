import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).with_name('validate-testflight-recovery.sh').resolve()

class RecoveryValidation(unittest.TestCase):
    def test_git_boundaries(self):
        with tempfile.TemporaryDirectory() as folder:
            def git(*args):
                return subprocess.run(['git', *args], cwd=folder, check=True, capture_output=True, text=True).stdout.strip()
            git('init', '-b', 'main')
            git('config', 'user.name', 'Release test')
            git('config', 'user.email', 'release@example.invalid')
            git('commit', '--allow-empty', '-m', 'initial')
            git('tag', 'v1.2.3')
            git('update-ref', 'refs/remotes/origin/main', 'HEAD')
            git('checkout', '-b', 'outside')
            git('commit', '--allow-empty', '-m', 'outside')
            git('tag', 'v1.2.4')
            cases = [
                ('v1.2.3', '116.2', 'refs/heads/main', True),
                ('v1.2.3', '116.2', 'refs/heads/other', False),
                ('main', '116.2', 'refs/heads/main', False),
                ('v01.2.3', '116.2', 'refs/heads/main', False),
                ('v1.2.3;echo bad', '116.2', 'refs/heads/main', False),
                ('v1.2.3', '0.1', 'refs/heads/main', False),
                ('v1.2.3', '116.0', 'refs/heads/main', False),
                ('v1.2.3', '116.100', 'refs/heads/main', False),
                ('v9.9.9', '116.2', 'refs/heads/main', False),
                ('v1.2.4', '116.2', 'refs/heads/main', False),
            ]
            for tag, build, ref, valid in cases:
                with self.subTest(tag=tag, build=build, ref=ref):
                    result = subprocess.run(['bash', str(SCRIPT), tag, build], cwd=folder,
                        env={**os.environ, 'GITHUB_REF': ref}, capture_output=True, text=True)
                    self.assertEqual(result.returncode == 0, valid, result.stderr)

if __name__ == '__main__':
    unittest.main()
