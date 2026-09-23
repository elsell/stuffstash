#!/usr/bin/env python3
"""Install fixture routes in a runner checkout or an explicit disposable archive."""
import os
from pathlib import Path
import shutil

root = Path(__file__).resolve().parents[1]
runner_temp = os.environ.get("RUNNER_TEMP")
if not runner_temp:
    raise SystemExit("A production-route backup directory is required")
if os.environ.get("GITHUB_ACTIONS") != "true":
    archive_root = os.environ.get("MOBILE_AUDIT_ARCHIVE_ROOT")
    marker = root / ".mobile-audit-archive"
    if (not archive_root or Path(archive_root).resolve() != root
            or marker.is_symlink() or not marker.is_file()
            or marker.read_text() != "disposable-mobile-audit\n"
            or any((parent / ".git").exists() for parent in (root, *root.parents))
            or Path(runner_temp).resolve().is_relative_to(root)):
        raise SystemExit("Fixture routes require a runner checkout or an explicit disposable source archive")
if os.environ.get("AUDIT_SUITE") != "fixtures":
    raise SystemExit("Explicit AUDIT_SUITE=fixtures is required")

routes = root / "apps/mobile/src/app"
for component in (routes, *routes.parents):
    if component == root:
        break
    if component.is_symlink():
        raise SystemExit("Fixture route paths must not contain symlinks")
backup = Path(runner_temp) / "production-mobile-routes"
if backup.exists():
    raise SystemExit("Production route backup already exists; refusing to replace it")
shutil.copytree(routes, backup)
shutil.rmtree(routes)
routes.mkdir()
exports = {
    "_layout": "FixtureLayoutWithoutKeyboardProvider" if os.environ.get("AUDIT_TEST_CASE") == "text-entry-no-provider" else "FixtureLayout",
    "index": "FixtureMenu",
    "audit-sheet-diagnostic": "SheetLayoutFixture",
    "audit-add": "AddAssetFixture",
    "audit-add-push": "AddAssetFixture",
    "audit-add-header": "AddAssetFixture",
    "audit-inventory-query": "InventoryQueryFixture",
    "audit-inventory-switcher": "InventorySwitcherFixture",
    "audit-home-return": "HomeReturnFixture",
    "audit-home-header": "HomeHeaderFixture",
    "add": "HomeAddProbeDestination",
    "settings": "HomeProfileProbeDestination",
    "home-return-details": "HomeReturnDetailsRoute",
    "audit-add-destination": "AddDestinationFixture",
    "add-destination": "AddDestinationRoute",
    "asset-tag-selection": "AssetTagSelectionRoute",
    "audit-detail-commands": "AssetDetailCommandsFixture",
    "audit-contents-search": "AssetContentsSearchFixture",
    "audit-contents-search-preconfigured": "AssetContentsSearchFixture",
    "audit-region-recovery": "AssetRegionRecoveryFixture",
    "audit-menu-ownership": "NativeMenuOwnershipFixture",
    "audit-command-height": "CommandHeightFixture",
    "audit-notice": "NoticePlacementFixture",
    "audit-notice-sheet": "NoticePlacementFixture",
    "audit-provider-editor": "ProviderEditorFixture",
    "audit-account": "AccountConnectionFixture",
    "audit-invitation": "InvitationAcceptanceFixture",
    "audit-connection": "AccountConnectionFixture",
    "audit-sharing": "InventorySharingFixture",
    "audit-notifications": "NotificationInboxFixture",
    "audit-notification-target": "NotificationTargetFixture",
    "audit-customization": "CustomizationCollectionFixture",
    "audit-settings-readback": "SettingsReadbackFixture",
    "audit-managed-search": "ManagedSearchPlacementFixture",
    "audit-browse-journey": "BrowseJourneyFixture",
    "search": "BrowseFilterJourneySearch",
    "browse-filters": "BrowseFilterJourneyFilters",
    "expiration": "BrowseFilterJourneyExpiration",
    "assets/[assetId]/index": "BrowseFilterJourneyDetail",
    "audit-android-header-composition": "AndroidHeaderCompositionFixture",
    "audit-native-search-placement": "NativeSearchPlacementFixture",
    "audit-customization-editor": "CustomizationEditorFixture",
    "voice": "VoiceProposalFixture",
    "voice-plan-location": "VoicePlanLocationFixture",
    "audit-footer-appearance": "FooterAppearanceFixture",
    "audit-move-here-recovery": "MoveHereRecoveryFixture",
    "audit-move-destination": "MoveDestinationFixture",
    "audit-edit-journey": "AssetEditJourneyDetailFixture",
    "assets/[assetId]/edit": "AssetEditJourneyEditorFixture",
    "assets/[assetId]/move": "AssetEditJourneyMoveFixture",
    "audit-edit-tags": "AssetEditTagsFixture",
    "audit-edit-recovery": "AssetEditRecoveryFixture",
    "audit-checkout-history": "CheckoutHistoryFixture",
    "audit-browse": "BrowseFilterFixture",
    "audit-expiration": "ExpirationFilterFixture",
    "audit-expiration-medium": "ExpirationFilterFixture",
}
if os.environ.get("AUDIT_TEST_CASE") == "filters":
    exports["audit-browse"] = "BrowseFilterGeometryFixture"
    exports["audit-expiration"] = "ExpirationFilterGeometryFixture"
for route, component in exports.items():
    target = routes / f"{route}.tsx"
    target.parent.mkdir(parents=True, exist_ok=True)
    source = os.path.relpath(root / "apps/mobile/native-audit/FixtureApplication", target.parent)
    target.write_text(f"export {{ {component} as default }} from '{source}';\n")

# Preserve the exact production shell and nested stack layouts. Only their data
# screens are replaced; no production services, session, or route root is mounted.
for layout in ("(tabs)/_layout.tsx", "(tabs)/(home)/_layout.tsx", "(tabs)/search/_layout.tsx"):
    target = routes / layout.replace("(tabs)/", "audit-tabs/", 1)
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(backup / layout, target)
for route, component in {
    "audit-tabs/(home)/index.tsx": "HomeTabShellFixture",
    "audit-tabs/search/index.tsx": "TabShellBrowsePlaceholder",
}.items():
    (routes / route).write_text(
        f"export {{ {component} as default }} from '../../../../native-audit/FixtureApplication';\n"
    )
