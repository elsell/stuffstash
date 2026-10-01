#!/usr/bin/env python3
import runpy
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest

missing_actions = runpy.run_path(str(Path(__file__).with_name('check-audit-action-migrations.py')))['missing_actions']


class AuditMigrationCheck(unittest.TestCase):
    def test_new_action_requires_forward_migration_without_removing_history(self):
        with TemporaryDirectory() as temporary:
            root = Path(temporary)
            domain = root / 'apps/api/internal/domain/audit/audit.go'
            domain.parent.mkdir(parents=True)
            domain.write_text('const ActionInventoryCreated Action = "inventory.created"\n'
                              'const ActionInventoryExported Action = "inventory.exported"\n')
            migrations = root / 'apps/api/migrations'
            migrations.mkdir(parents=True)
            def constraint(values):
                return 'ALTER TABLE audit_records ADD CONSTRAINT chk_audit_records_action CHECK (action IN (' + values + '));'
            (migrations / '000001_initial.up.sql').write_text(constraint("'inventory.created','historic.action'"))
            # A rollback file must not satisfy a missing forward action.
            (migrations / '000002_export.down.sql').write_text(constraint("'inventory.created','inventory.exported'"))
            self.assertEqual(missing_actions(root), {'inventory.exported'})
            (migrations / '000002_export.up.sql').write_text(constraint("'inventory.created','inventory.exported','historic.action'"))
            self.assertEqual(missing_actions(root), set())
            (migrations / '000003_regression.up.sql').write_text(constraint("'inventory.created'"))
            self.assertEqual(missing_actions(root), {'inventory.exported'})


if __name__ == '__main__':
    unittest.main()
