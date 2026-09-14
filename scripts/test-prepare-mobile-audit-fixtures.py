import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


class FixtureRouteIsolationTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        (self.root / "scripts").mkdir()
        self.script = self.root / "scripts/prepare-mobile-audit-fixtures.py"
        shutil.copyfile(Path(__file__).with_name("prepare-mobile-audit-fixtures.py"), self.script)
        self.routes = self.root / "apps/mobile/src/app"
        self.routes.mkdir(parents=True)
        (self.routes / "index.tsx").write_text("production route\n")
        self.runner = self.root / "runner"
        self.runner.mkdir()

    def run_script(self, **settings):
        env = {**os.environ, "GITHUB_ACTIONS": "true", "RUNNER_TEMP": str(self.runner),
               "AUDIT_SUITE": "fixtures", **settings}
        return subprocess.run(["python3", str(self.script)], env=env, capture_output=True, text=True)

    def test_refuses_nonrunner_or_nonfixture_calls_without_modifying_production(self):
        for settings in [{"GITHUB_ACTIONS": "false"}, {"AUDIT_SUITE": "onboarding"}]:
            self.assertNotEqual(self.run_script(**settings).returncode, 0)
            self.assertEqual((self.routes / "index.tsx").read_text(), "production route\n")

    def test_retains_production_routes_and_installs_only_fixture_exports(self):
        self.assertEqual(self.run_script().returncode, 0)
        self.assertEqual((self.runner / "production-mobile-routes/index.tsx").read_text(), "production route\n")
        self.assertEqual({p.name for p in self.routes.iterdir()},
                         {"_layout.tsx", "index.tsx", "audit-add.tsx", "audit-browse.tsx", "audit-expiration.tsx", "audit-expiration-medium.tsx"})
        self.assertIn("FixtureMenu as default", (self.routes / "index.tsx").read_text())
        self.assertNotEqual(self.run_script().returncode, 0)
        self.assertEqual((self.runner / "production-mobile-routes/index.tsx").read_text(), "production route\n")


if __name__ == "__main__":
    unittest.main()
