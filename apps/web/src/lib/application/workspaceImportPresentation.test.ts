import { describe, expect, it } from 'vitest';
import type { ImportMessage, ImportPreview } from '$lib/domain/inventory';
import {
  buildLegacyHomeboxImportRequest,
  importApplyStatus,
  importMessageTone,
  importPreviewSourceSummary,
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

  it('summarizes preview source identity and maps message tone', () => {
    expect(importPreviewSourceSummary(importSource({ version: '0.10.3', imageImport: 'enabled' }))).toBe('Homebox 0.10.3');
    expect(importPreviewSourceSummary(importSource({ imageImport: 'disabled' }))).toBe('disabled');
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
