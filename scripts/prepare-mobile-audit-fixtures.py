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
    "audit-browse": "BrowseFilterFixture",
    "audit-expiration": "ExpirationFilterFixture",
    "audit-expiration-medium": "ExpirationFilterFixture",
}
for route, component in exports.items():
    (routes / f"{route}.tsx").write_text(
        f"export {{ {component} as default }} from '../../native-audit/FixtureApplication';\n"
    )
