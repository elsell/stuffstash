#!/usr/bin/env python3
"""Reject domain audit actions absent from the migrated PostgreSQL constraint."""
from pathlib import Path
import re


def missing_actions(root: Path) -> set[str]:
    domain = (root / 'apps/api/internal/domain/audit/audit.go').read_text()
    actions = set(re.findall(r'\bAction\w+\s+Action\s*=\s*"([^"]+)"', domain))
    if not actions:
        raise ValueError('No domain audit actions found')
    allowed = None
    for migration in sorted((root / 'apps/api/migrations').glob('*.up.sql')):
        for match in re.finditer(
            r'ADD\s+CONSTRAINT\s+chk_audit_records_action\s+CHECK\s*\(\s*action\s+IN\s*\((.*?)\)\s*\)',
            migration.read_text(), re.IGNORECASE | re.DOTALL,
        ):
            allowed = set(re.findall(r"'([^']+)'", match[1]))
    if not allowed:
        raise ValueError('No migrated audit action constraint found')
    return actions - allowed


if __name__ == '__main__':
    missing = missing_actions(Path(__file__).resolve().parent.parent)
    if missing:
        raise SystemExit('Domain audit actions need a forward constraint migration: ' + ', '.join(sorted(missing)))
