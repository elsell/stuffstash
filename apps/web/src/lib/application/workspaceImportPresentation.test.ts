import { describe, expect, it } from 'vitest';
import type { ImportMessage, ImportPreview } from '$lib/domain/inventory';
import {
  buildLegacyHomeboxImportRequest,
  importAppliedDescription,
  importApplyMessagesPresentation,
  importApplyStatus,
  importDeniedPresentation,
  importEmptyPreviewPresentation,
  importMessageDetail,
  importMessageTone,
  importMissingInventoryPresentation,
  importPreviewSourceSummary,
  importPlannedCountLabel,
  importSourceOptions,
  importSourceSummary,
  isImportPreviewReady
} from './workspaceImportPresentation';

describe('workspace import presentation helpers', () => {
  it('builds route-backed import source options', () => {
    expect(importSourceOptions('tenant-one', 'inventory-one')).toEqual([
      {
        value: 'legacy_homebox',
        label: 'Connect',
        href: '/tenants/tenant-one/inventories/inventory-one/import/legacy-homebox'
      },
      {
        value: 'legacy_homebox_csv',
        label: 'CSV',
        href: '/tenants/tenant-one/inventories/inventory-one/import/legacy-homebox-csv'
      }
    ]);
  });

  it('summarizes the selected import source', () => {
    expect(importSourceSummary('legacy_homebox', '')).toBe('Live Homebox API');
    expect(importSourceSummary('legacy_homebox_csv', '')).toBe('Homebox CSV export');
    expect(importSourceSummary('legacy_homebox_csv', 'homebox.csv')).toBe('homebox.csv');
  });

  it('checks whether each import source is ready for preview', () => {
    expect(
      isImportPreviewReady({
        hasInventory: false,
        sourceType: 'legacy_homebox',
        baseUrl: 'https://homebox.local',
        username: 'owner',
        password: 'secret',
        contentBase64: ''
      })
    ).toBe(false);
    expect(
      isImportPreviewReady({
        hasInventory: true,
        sourceType: 'legacy_homebox',
        baseUrl: ' https://homebox.local ',
        username: ' owner ',
        password: 'secret',
        contentBase64: ''
      })
    ).toBe(true);
    expect(
      isImportPreviewReady({
        hasInventory: true,
        sourceType: 'legacy_homebox_csv',
        baseUrl: '',
        username: '',
        password: '',
        contentBase64: ''
      })
    ).toBe(false);
    expect(
      isImportPreviewReady({
        hasInventory: true,
        sourceType: 'legacy_homebox_csv',
        baseUrl: '',
        username: '',
        password: '',
        contentBase64: 'YWJj'
      })
    ).toBe(true);
  });

  it('builds clear apply status copy for each disabled or ready state', () => {
    expect(importApplyStatus({ busy: true, hasPreview: false, blockingErrorCount: 0, canImport: true })).toBe('Import action is running.');
    expect(importApplyStatus({ busy: false, hasPreview: false, blockingErrorCount: 0, canImport: true })).toBe(
      'Preview the import before applying changes.'
    );
    expect(importApplyStatus({ busy: false, hasPreview: true, blockingErrorCount: 1, canImport: true })).toBe(
      'Resolve preview errors before applying changes.'
    );
    expect(importApplyStatus({ busy: false, hasPreview: true, blockingErrorCount: 0, canImport: false })).toBe(
      'Inventory configuration access is required.'
    );
    expect(importApplyStatus({ busy: false, hasPreview: true, blockingErrorCount: 0, canImport: true })).toBe('Preview is ready to apply.');
  });

  it('derives import panel fallback, count, and applied-result presentation', () => {
    expect(importMissingInventoryPresentation()).toEqual({ title: 'Select an inventory' });
    expect(importDeniedPresentation()).toEqual({
      title: 'Import unavailable',
      description: 'Inventory configuration access is required.'
    });
    expect(importEmptyPreviewPresentation()).toEqual({
      title: 'Preview an import',
      description: 'Review planned records before anything is saved.'
    });
    expect(importPlannedCountLabel({ counts: { fields: 1, locations: 2, assets: 3, attachments: 4, warnings: 0, errors: 0 } })).toBe(
      '10 planned records'
    );
    expect(importPlannedCountLabel({ counts: { fields: 0, locations: 1, assets: 0, attachments: 0, warnings: 0, errors: 0 } })).toBe(
      '1 planned record'
    );
    expect(
      importAppliedDescription({
        counts: {
          locationsCreated: 1,
          assetsCreated: 0,
          attachmentsCreated: 3,
          fieldsCreated: 2,
          fieldsExisting: 1,
          assetsSkipped: 1,
          attachmentsSkipped: 2
        }
      })
    ).toBe(
      'Created 2 field definitions, 1 location, and 3 attachments. Reused 1 field definition. Skipped 1 item and 2 attachments.'
    );
    expect(
      importAppliedDescription({
        counts: {
          fieldsCreated: 0,
          fieldsExisting: 0,
          locationsCreated: 0,
          assetsCreated: 0,
          assetsSkipped: 0,
          attachmentsCreated: 0,
          attachmentsSkipped: 0
        }
      })
    ).toBe('Import finished without creating records.');
    expect(importApplyMessagesPresentation()).toEqual({ title: 'Apply messages' });
  });

  it('summarizes preview source identity and maps message tone', () => {
    expect(importPreviewSourceSummary(importSource({ version: '0.10.3', imageImport: 'enabled' }))).toBe('Homebox 0.10.3');
    expect(importPreviewSourceSummary(importSource({ imageImport: 'disabled' }))).toBe('disabled');
    expect(importMessageDetail({ code: 'missing_parent', severity: 'warning', summary: 'Missing parent' })).toBe('missing_parent');
    expect(
      importMessageDetail({
        code: 'attachment_skipped',
        severity: 'warning',
        summary: 'Attachment skipped',
        detail: 'One attachment could not be downloaded.',
        sourceName: 'Camera'
      })
    ).toBe('Camera: One attachment could not be downloaded.');
    expect(importMessageTone(importMessage('error'))).toBe('destructive');
    expect(importMessageTone(importMessage('warning'))).toBe('secondary');
    expect(importMessageTone(importMessage('info'))).toBe('outline');
  });

  it('builds trimmed live Homebox requests and CSV upload requests', () => {
    expect(
      buildLegacyHomeboxImportRequest({
        sourceType: 'legacy_homebox',
        baseUrl: ' https://homebox.local ',
        username: ' owner ',
        password: 'secret',
        includeImages: false,
        allowInsecureTLS: true,
        allowPrivateNetwork: true,
        fileName: '',
        contentBase64: ''
      })
    ).toEqual({
      sourceType: 'legacy_homebox',
      baseUrl: 'https://homebox.local',
      username: 'owner',
      password: 'secret',
      includeImages: false,
      allowInsecureTLS: true,
      allowPrivateNetwork: true
    });

    expect(
      buildLegacyHomeboxImportRequest({
        sourceType: 'legacy_homebox_csv',
        baseUrl: '',
        username: '',
        password: '',
        includeImages: true,
        allowInsecureTLS: false,
        allowPrivateNetwork: false,
        fileName: 'homebox.csv',
        contentBase64: 'YWJj'
      })
    ).toEqual({
      sourceType: 'legacy_homebox_csv',
      fileName: 'homebox.csv',
      contentBase64: 'YWJj'
    });
  });
});

function importSource(source: Partial<ImportPreview['source']>): ImportPreview['source'] {
  return {
    type: 'legacy_homebox',
    name: 'Homebox',
    imageImport: 'disabled',
    ...source
  };
}

function importMessage(severity: string): ImportMessage {
  return {
    code: 'message_code',
    severity,
    summary: 'Message'
  };
}
