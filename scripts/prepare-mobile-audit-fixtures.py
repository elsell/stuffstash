#!/usr/bin/env python3
"""Install an isolated fixture route root in an ephemeral macOS audit checkout."""
import os
from pathlib import Path
import shutil

if os.environ.get("GITHUB_ACTIONS") != "true" or not os.environ.get("RUNNER_TEMP"):
    raise SystemExit("Fixture routes may only be installed in a GitHub Actions runner checkout")
if os.environ.get("AUDIT_SUITE") != "fixtures":
    raise SystemExit("Explicit AUDIT_SUITE=fixtures is required")

root = Path(__file__).resolve().parents[1]
routes = root / "apps/mobile/src/app"
backup = Path(os.environ["RUNNER_TEMP"]) / "production-mobile-routes"
if backup.exists():
    raise SystemExit("Production route backup already exists; refusing to replace it")
shutil.copytree(routes, backup)
shutil.rmtree(routes)
routes.mkdir()
exports = {
    "_layout": "FixtureLayout",
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
    "audit-detail-commands": "AssetDetailCommandsFixture",
    "audit-contents-search": "AssetContentsSearchFixture",
    "audit-contents-search-preconfigured": "AssetContentsSearchFixture",
    "audit-region-recovery": "AssetRegionRecoveryFixture",
    "audit-command-height": "CommandHeightFixture",
    "audit-notice": "NoticePlacementFixture",
    "audit-notice-sheet": "NoticePlacementFixture",
    "audit-provider-editor": "ProviderEditorFixture",
    "audit-account": "AccountConnectionFixture",
    "audit-invitation": "InvitationAcceptanceFixture",
    "audit-connection": "AccountConnectionFixture",
    "audit-sharing": "InventorySharingFixture",
    "audit-customization": "CustomizationCollectionFixture",
    "audit-native-search-placement": "NativeSearchPlacementFixture",
    "audit-customization-editor": "CustomizationEditorFixture",
    "voice": "VoiceProposalFixture",
    "voice-plan-location": "VoicePlanLocationFixture",
    "audit-footer-appearance": "FooterAppearanceFixture",
    "audit-move-here-recovery": "MoveHereRecoveryFixture",
    "audit-edit-tags": "AssetEditTagsFixture",
    "audit-edit-recovery": "AssetEditRecoveryFixture",
    "audit-checkout-history": "CheckoutHistoryFixture",
    "audit-browse": "BrowseFilterFixture",
    "audit-expiration": "ExpirationFilterFixture",
    "audit-expiration-medium": "ExpirationFilterFixture",
}
for route, component in exports.items():
    (routes / f"{route}.tsx").write_text(
        f"export {{ {component} as default }} from '../../native-audit/FixtureApplication';\n"
    )

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
