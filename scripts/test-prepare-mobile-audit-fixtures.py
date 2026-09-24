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
        self.root = Path(self.temporary.name) / "checkout"
        self.root.mkdir()
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
               "AUDIT_SUITE": "fixtures", "AUDIT_TEST_CASE": "all", **settings}
        return subprocess.run(["python3", str(self.script)], env=env, capture_output=True, text=True)

    def test_refuses_nonrunner_or_nonfixture_calls_without_modifying_production(self):
        for settings in [{"GITHUB_ACTIONS": "false"}, {"AUDIT_SUITE": "onboarding"}]:
            self.assertNotEqual(self.run_script(**settings).returncode, 0)
            self.assertEqual((self.routes / "index.tsx").read_text(), "production route\n")

    def test_retains_production_routes_and_installs_only_fixture_exports(self):
        self.assertEqual(self.run_script().returncode, 0)
        self.assertEqual((self.runner / "production-mobile-routes/index.tsx").read_text(), "production route\n")

        self.assertEqual({p.name for p in self.routes.iterdir()} - {"audit-browse-journey.tsx", "assets", "audit-edit-journey.tsx", "audit-android-header-composition.tsx", "audit-notifications.tsx", "audit-notification-target.tsx", "audit-managed-search.tsx", "audit-contents-search-preconfigured.tsx", "audit-tabs", "audit-invitation.tsx", "audit-customization-editor.tsx", "voice.tsx", "voice-plan-location.tsx", "audit-native-search-placement.tsx", "add.tsx", "settings.tsx"},
                         {"audit-menu-ownership.tsx", "audit-customization.tsx", "_layout.tsx", "index.tsx", "audit-add.tsx", "audit-add-push.tsx", "audit-add-header.tsx", "audit-inventory-query.tsx", "audit-inventory-switcher.tsx", "audit-home-return.tsx", "audit-home-header.tsx", "home-return-details.tsx", "asset-tag-selection.tsx", "add-destination.tsx", "audit-add-destination.tsx", "audit-checkout-history.tsx", "audit-edit-recovery.tsx", "audit-edit-tags.tsx", "audit-move-here-recovery.tsx", "audit-move-destination.tsx", "audit-command-height.tsx", "audit-footer-appearance.tsx", "audit-sharing.tsx", "audit-account.tsx", "audit-connection.tsx", "audit-provider-editor.tsx", "audit-notice.tsx", "audit-notice-sheet.tsx", "audit-region-recovery.tsx", "audit-contents-search.tsx", "audit-detail-commands.tsx", "audit-sheet-diagnostic.tsx", "audit-browse.tsx", "audit-expiration.tsx", "audit-expiration-medium.tsx"})
        self.assertIn("AssetEditJourneyEditorFixture as default", (self.routes / "assets/[assetId]/edit.tsx").read_text())
        self.assertIn("../../../../native-audit/FixtureApplication", (self.routes / "assets/[assetId]/edit.tsx").read_text())
        self.assertIn("AssetEditJourneyMoveFixture as default", (self.routes / "assets/[assetId]/move.tsx").read_text())
        self.assertIn("../../../../native-audit/FixtureApplication", (self.routes / "assets/[assetId]/move.tsx").read_text())
        self.assertIn("AssetEditJourneyDetailFixture as default", (self.routes / "audit-edit-journey.tsx").read_text())
        self.assertIn("AndroidHeaderCompositionFixture as default", (self.routes / "audit-android-header-composition.tsx").read_text())
        self.assertIn("NotificationInboxFixture as default", (self.routes / "audit-notifications.tsx").read_text())
        self.assertIn("AssetContentsSearchFixture as default", (self.routes / "audit-contents-search-preconfigured.tsx").read_text())
        self.assertIn("FixtureMenu as default", (self.routes / "index.tsx").read_text())
        self.assertIn("FixtureLayout as default", (self.routes / "_layout.tsx").read_text())
        self.assertIn("CustomizationEditorFixture as default", (self.routes / "audit-customization-editor.tsx").read_text())
        self.assertIn("VoiceProposalFixture as default", (self.routes / "voice.tsx").read_text())
        self.assertIn("VoicePlanLocationFixture as default", (self.routes / "voice-plan-location.tsx").read_text())
        self.assertIn("ManagedSearchPlacementFixture as default", (self.routes / "audit-managed-search.tsx").read_text())
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

    def test_filter_diagnostic_installs_geometry_probes_only_for_focused_selection(self):
        result = self.run_script(AUDIT_TEST_CASE="filters")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("BrowseFilterGeometryFixture as default", (self.routes / "audit-browse.tsx").read_text())
        self.assertIn("ExpirationFilterGeometryFixture as default", (self.routes / "audit-expiration.tsx").read_text())
        self.assertIn("ExpirationFilterFixture as default", (self.routes / "audit-expiration-medium.tsx").read_text())

    def test_provider_free_diagnostic_installs_distinct_root_and_keeps_backup(self):
        result = self.run_script(AUDIT_TEST_CASE="text-entry-no-provider")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("FixtureLayoutWithoutKeyboardProvider as default", (self.routes / "_layout.tsx").read_text())
        self.assertEqual((self.runner / "production-mobile-routes/index.tsx").read_text(), "production route\n")

    def test_explicit_disposable_archive_keeps_external_production_backup(self):
        (self.root / ".mobile-audit-archive").write_text("disposable-mobile-audit\n")
        with tempfile.TemporaryDirectory() as backup:
            result = self.run_script(GITHUB_ACTIONS="false", MOBILE_AUDIT_ARCHIVE_ROOT=str(self.root), RUNNER_TEMP=backup)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual((Path(backup) / "production-mobile-routes/index.tsx").read_text(), "production route\n")
            self.assertIn("FixtureMenu", (self.routes / "index.tsx").read_text())

    def test_archive_authorization_rejects_missing_marker_git_and_internal_backup(self):
        with tempfile.TemporaryDirectory() as backup:
            settings = dict(GITHUB_ACTIONS="false", MOBILE_AUDIT_ARCHIVE_ROOT=str(self.root), RUNNER_TEMP=backup)
            self.assertNotEqual(self.run_script(**settings).returncode, 0)
            (self.root / ".mobile-audit-archive").write_text("disposable-mobile-audit\n")
            git_entry = self.root / ".git"
            git_entry.write_text("gitdir: elsewhere\n")
            self.assertNotEqual(self.run_script(**settings).returncode, 0)
            git_entry.unlink()
            self.assertNotEqual(self.run_script(**{**settings, "RUNNER_TEMP": str(self.runner)}).returncode, 0)
            self.assertEqual((self.routes / "index.tsx").read_text(), "production route\n")

    def test_archive_rejects_external_route_ancestor_without_touching_it(self):
        (self.root / ".mobile-audit-archive").write_text("disposable-mobile-audit\n")
        with tempfile.TemporaryDirectory() as external, tempfile.TemporaryDirectory() as backup:
            original = self.root / "apps/mobile"
            redirected = Path(external) / "mobile"
            shutil.move(original, redirected)
            original.symlink_to(redirected, target_is_directory=True)
            result = self.run_script(GITHUB_ACTIONS="false", MOBILE_AUDIT_ARCHIVE_ROOT=str(self.root), RUNNER_TEMP=backup)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual((redirected / "src/app/index.tsx").read_text(), "production route\n")
            self.assertFalse((Path(backup) / "production-mobile-routes").exists())

    def test_archive_rejects_git_ancestor_and_untrusted_marker(self):
        marker = self.root / ".mobile-audit-archive"
        with tempfile.TemporaryDirectory() as backup:
            settings = dict(GITHUB_ACTIONS="false", MOBILE_AUDIT_ARCHIVE_ROOT=str(self.root), RUNNER_TEMP=backup)
            marker.write_text("wrong marker\n")
            self.assertNotEqual(self.run_script(**settings).returncode, 0)
            marker.write_text("disposable-mobile-audit\n")
            self.assertNotEqual(self.run_script(**{**settings, "MOBILE_AUDIT_ARCHIVE_ROOT": backup}).returncode, 0)
            (self.root.parent / ".git").mkdir()
            self.assertNotEqual(self.run_script(**settings).returncode, 0)
            (self.root.parent / ".git").rmdir()
            target = self.root / "marker-source"
            marker.rename(target)
            marker.symlink_to(target)
            self.assertNotEqual(self.run_script(**settings).returncode, 0)
            self.assertEqual((self.routes / "index.tsx").read_text(), "production route\n")


class NativeAuditSelectionTests(unittest.TestCase):
    def test_native_command_enables_bounded_case_execution(self):
        root = Path(__file__).resolve().parents[1]
        workflow = (root / '.github/workflows/mobile-native-audit.yml').read_text()
        command = 'xcodebuild test' + workflow.split('          xcodebuild test', 1)[1].split('      - name: Export screenshots', 1)[0]
        with tempfile.TemporaryDirectory() as runner:
            (Path(runner) / 'native-audit').mkdir()
            # A command-boundary fake records the actual workflow arguments.
            result = subprocess.run(['bash', '-c',
                'set -o pipefail\nxcodebuild() { printf "%s\\n" "$@"; }\naudit_test_args=()\n' + command],
                env={**os.environ, 'RUNNER_TEMP': runner, 'AUDIT_DEVICE': 'iPhone 17'},
                capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            arguments = (Path(runner) / 'native-audit/xcodebuild.log').read_text().splitlines()
        for option, expected in [('-test-timeouts-enabled', 'YES'), ('-maximum-test-execution-time-allowance', '600')]:
            self.assertEqual(arguments.count(option), 1, option)
            self.assertEqual(arguments[arguments.index(option) + 1], expected)

    def test_release_corrections_selects_named_existing_workflows(self):
        root = Path(__file__).resolve().parents[1]
        workflow = (root / '.github/workflows/mobile-native-audit.yml').read_text()
        selection = workflow.split('          audit_test_args=()', 1)[1].split('          printf', 1)[0]
        result = subprocess.run(['bash', '-c', 'audit_test_args=()\n' + selection + '\nprintf "%s\\n" "${audit_test_args[@]}"'],
                                env={**os.environ, 'AUDIT_TEST_CASE': 'release-corrections', 'AUDIT_SUITE': 'fixtures'},
                                capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        expected = {
            'testBrowseTagSearchKeepsActionsAboveKeyboardAccessory',
            'testBrowseLastTagClearsActionFooterAndApplies',
            'testBrowseUsesInPlaceAvailabilityMenuAndReachableActions',
            'testExpirationDatePageKeepsBottomActionsReachable',
            'testExpirationSearchKeepsActionsReachableWithKeyboard',
            'testSharingRecoveryKeepsHeaderAndCommandsReachable',
            'testSettingsCollectionUsesNativeSearchAndAdd',
            'testAddDraftRetainsTextAndRecoversAfterRejectedSave',
            'testColorPickerOpensDirectlyAndClearPreservesParentDraft',
        }
        names = [line.removeprefix('-only-testing:StuffStashAuditTests/FixtureAuditTests/') for line in result.stdout.splitlines()]
        self.assertEqual(set(names), expected)
        self.assertEqual(len(names), len(expected))
        swift = (root / 'apps/mobile/native-audit/FixtureAuditTests.swift').read_text()
        for name in names:
            self.assertIn('func ' + name + '()', swift)


if __name__ == "__main__":
    unittest.main()
