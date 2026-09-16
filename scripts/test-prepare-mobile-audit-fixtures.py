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
        self.tab_layouts = ("(tabs)/_layout.tsx", "(tabs)/(home)/_layout.tsx", "(tabs)/search/_layout.tsx")
        for layout in self.tab_layouts:
            target = self.routes / layout
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(f"production layout {layout}\n")
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
        self.assertEqual({p.name for p in self.routes.iterdir()} - {"audit-contents-search-preconfigured.tsx", "audit-tabs", "audit-invitation.tsx", "audit-customization-editor.tsx", "voice.tsx", "voice-plan-location.tsx", "audit-native-search-placement.tsx", "add.tsx", "settings.tsx"},
                         {"audit-customization.tsx", "_layout.tsx", "index.tsx", "audit-add.tsx", "audit-add-push.tsx", "audit-add-header.tsx", "audit-inventory-query.tsx", "audit-inventory-switcher.tsx", "audit-home-return.tsx", "audit-home-header.tsx", "home-return-details.tsx", "audit-checkout-history.tsx", "audit-edit-recovery.tsx", "audit-edit-tags.tsx", "audit-move-here-recovery.tsx", "audit-command-height.tsx", "audit-footer-appearance.tsx", "audit-sharing.tsx", "audit-account.tsx", "audit-connection.tsx", "audit-provider-editor.tsx", "audit-notice.tsx", "audit-notice-sheet.tsx", "audit-region-recovery.tsx", "audit-contents-search.tsx", "audit-detail-commands.tsx", "audit-sheet-diagnostic.tsx", "audit-browse.tsx", "audit-expiration.tsx", "audit-expiration-medium.tsx"})
        self.assertIn("AssetContentsSearchFixture as default", (self.routes / "audit-contents-search-preconfigured.tsx").read_text())
        self.assertIn("FixtureMenu as default", (self.routes / "index.tsx").read_text())
        self.assertIn("CustomizationEditorFixture as default", (self.routes / "audit-customization-editor.tsx").read_text())
        self.assertIn("VoiceProposalFixture as default", (self.routes / "voice.tsx").read_text())
        self.assertIn("VoicePlanLocationFixture as default", (self.routes / "voice-plan-location.tsx").read_text())
        self.assertIn("NativeSearchPlacementFixture as default", (self.routes / "audit-native-search-placement.tsx").read_text())
        self.assertIn("HomeAddProbeDestination as default", (self.routes / "add.tsx").read_text())
        self.assertIn("HomeProfileProbeDestination as default", (self.routes / "settings.tsx").read_text())
        self.assertIn("InvitationAcceptanceFixture as default", (self.routes / "audit-invitation.tsx").read_text())
        for layout in self.tab_layouts:
            self.assertEqual((self.routes / layout.replace("(tabs)/", "audit-tabs/", 1)).read_text(), f"production layout {layout}\n")
            self.assertEqual((self.runner / "production-mobile-routes" / layout).read_text(), f"production layout {layout}\n")
        self.assertEqual({str(p.relative_to(self.routes)) for p in (self.routes / "audit-tabs").rglob("*.tsx")},
                         {*[p.replace("(tabs)/", "audit-tabs/", 1) for p in self.tab_layouts], "audit-tabs/(home)/index.tsx", "audit-tabs/search/index.tsx"})
        self.assertIn("HomeTabShellFixture as default", (self.routes / "audit-tabs/(home)/index.tsx").read_text())
        self.assertIn("TabShellBrowsePlaceholder as default", (self.routes / "audit-tabs/search/index.tsx").read_text())
        self.assertNotEqual(self.run_script().returncode, 0)
        self.assertEqual((self.runner / "production-mobile-routes/index.tsx").read_text(), "production route\n")


if __name__ == "__main__":
    unittest.main()
