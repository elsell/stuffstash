import { expect, it } from 'vitest';
import { importCountAction, importCountIcon } from './importCountMetrics';
import { visibleCountCells } from './importWorkspacePresentation';

it('uses semantic metrics for import navigation, icons and zero-error visibility regardless of label language', () => {
  const cells = [
    { metric: 'blockingIssue' as const, value: 0, label: '問題' },
    { metric: 'warning' as const, value: 2, label: 'Avertissements' },
    { metric: 'assetCreated' as const, value: 3, label: 'Créés' },
  ];
  expect(visibleCountCells(cells)).toEqual(cells);
  expect(importCountAction('warning')).toBe('issues');
  expect(importCountAction('assetCreated')).toBe('records');
  expect(importCountAction('fieldCreated')).toBeUndefined();
  expect(importCountIcon('locationCreated')).toBe('location');
  expect(importCountIcon('photoFileImported')).toBe('attachment');
});
