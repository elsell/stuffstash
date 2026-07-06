#!/usr/bin/env python3
import datetime as dt
import importlib.util
import json
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MODULE_PATH = ROOT / "scripts" / "check-dependency-age.py"
UTC = dt.timezone.utc


def load_module():
    spec = importlib.util.spec_from_file_location("check_dependency_age", MODULE_PATH)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def main():
    module = load_module()
    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        (root / "dependency-age-allowlist.json").write_text(
            json.dumps(
                [
                    {
                        "kind": "npm",
                        "name": "expo-auth-session",
                        "version": "55.0.17",
                        "reason": "Required for mobile OIDC.",
                        "compensatingVerification": "Mobile tests and typecheck pass.",
                    }
                ]
            ),
            encoding="utf-8",
        )

        allowlist, failures = module.load_allowlist(root)
        assert not failures, failures
        assert ("npm", "expo-auth-session", "55.0.17") in allowlist

    cutoff = dt.datetime(2026, 6, 20, tzinfo=UTC)
    published_at = dt.datetime(2026, 6, 25, tzinfo=UTC)
    allowed = {("npm", "expo-auth-session", "55.0.17")}
    assert module.check_age(
        "npm",
        "expo-auth-session",
        "55.0.17",
        published_at,
        "pnpm-lock.yaml",
        cutoff,
        allowed,
    ) is None
    assert module.check_age(
        "npm",
        "other-package",
        "1.0.0",
        published_at,
        "pnpm-lock.yaml",
        cutoff,
        allowed,
    )

    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        (root / "dependency-age-allowlist.json").write_text(
            json.dumps([{"kind": "npm", "name": "expo-auth-session"}]),
            encoding="utf-8",
        )

        _, failures = module.load_allowlist(root)
        assert failures, "expected malformed allowlist entry to fail closed"

    print("dependency age allowlist tests passed")


if __name__ == "__main__":
    main()
