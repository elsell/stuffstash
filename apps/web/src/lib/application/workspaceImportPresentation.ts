import type { ImportApplyResult, ImportMessage, ImportPreview, ImportSourceType, LegacyHomeboxImportRequest } from '$lib/domain/inventory';
import { importSourceHref } from './workspaceShellNavigation';

export interface ImportSourceOption {
  value: ImportSourceType;
  label: string;
  href: string;
}

export interface ImportRequestInput {
  sourceType: ImportSourceType;
  baseUrl: string;
  username: string;
  password: string;
  includeImages: boolean;
  allowInsecureTLS: boolean;
  allowPrivateNetwork: boolean;
  fileName: string;
  contentBase64: string;
}

export interface ImportApplyStatusInput {
  busy: boolean;
  hasPreview: boolean;
  blockingErrorCount: number;
  canImport: boolean;
}

export interface ImportPanelMessage {
  title: string;
  description?: string;
}

export const importSourceChoices: Array<Omit<ImportSourceOption, 'href'>> = [
  { value: 'legacy_homebox', label: 'Connect' },
  { value: 'legacy_homebox_csv', label: 'CSV' }
];

export function importSourceOptions(tenantId: string, inventoryId: string | null): ImportSourceOption[] {
  return importSourceChoices.map((option) => ({
    ...option,
    href: importSourceHref(tenantId, inventoryId, option.value)
  }));
}

export function importSourceSummary(sourceType: ImportSourceType, fileName: string): string {
  if (sourceType === 'legacy_homebox') {
    return 'Live Homebox API';
  }
  return fileName || 'Homebox CSV export';
}

export function isImportPreviewReady(input: Pick<ImportRequestInput, 'sourceType' | 'baseUrl' | 'username' | 'password' | 'contentBase64'> & {
  hasInventory: boolean;
}): boolean {
  if (!input.hasInventory) {
    return false;
  }
  if (input.sourceType === 'legacy_homebox_csv') {
    return input.contentBase64.length > 0;
  }
  return input.baseUrl.trim().length > 0 && input.username.trim().length > 0 && input.password.length > 0;
}

export function importApplyStatus(input: ImportApplyStatusInput): string {
  if (input.busy) {
    return 'Import action is running.';
  }
  if (!input.hasPreview) {
    return 'Preview the import before applying changes.';
  }
  if (input.blockingErrorCount > 0) {
    return 'Resolve preview errors before applying changes.';
  }
  if (!input.canImport) {
    return 'Inventory configuration access is required.';
  }
  return 'Preview is ready to apply.';
}

export function importMissingInventoryPresentation(): ImportPanelMessage {
  return { title: 'Select an inventory' };
}

export function importDeniedPresentation(): Required<ImportPanelMessage> {
  return {
    title: 'Import unavailable',
    description: 'Inventory configuration access is required.'
  };
}

export function importEmptyPreviewPresentation(): Required<ImportPanelMessage> {
  return {
    title: 'Preview an import',
    description: 'Review planned records before anything is saved.'
  };
}

export function importPlannedCountLabel(preview: Pick<ImportPreview, 'counts'>): string {
  const plannedCount = preview.counts.fields + preview.counts.locations + preview.counts.assets + preview.counts.attachments;
  return `${plannedCount} planned ${pluralize(plannedCount, 'record')}`;
}

export function importAppliedDescription(result: Pick<ImportApplyResult, 'counts'>): string {
  const created = compactList([
    countPhrase(result.counts.fieldsCreated, 'field definition'),
    countPhrase(result.counts.locationsCreated, 'location'),
    countPhrase(result.counts.assetsCreated, 'item'),
    countPhrase(result.counts.attachmentsCreated, 'attachment')
  ]);
  const skipped = compactList([
    countPhrase(result.counts.assetsSkipped, 'item'),
    countPhrase(result.counts.attachmentsSkipped, 'attachment')
  ]);
  const reused = countPhrase(result.counts.fieldsExisting, 'field definition');
  const sentences: string[] = [];

  if (created) {
    sentences.push(`Created ${created}.`);
  }
  if (reused) {
    sentences.push(`Reused ${reused}.`);
  }
  if (skipped) {
    sentences.push(`Skipped ${skipped}.`);
  }
  return sentences.join(' ') || 'Import finished without creating records.';
}

export function importApplyMessagesPresentation(): ImportPanelMessage {
  return { title: 'Apply messages' };
}

export function importPreviewSourceSummary(source: ImportPreview['source']): string {
  return source.version ? `Homebox ${source.version}` : source.imageImport;
}

export function importMessageDetail(message: ImportMessage): string {
  const detail = message.detail ?? message.code;
  return message.sourceName ? `${message.sourceName}: ${detail}` : detail;
}

export function importMessageTone(message: ImportMessage): 'destructive' | 'secondary' | 'outline' {
  return message.severity === 'error' ? 'destructive' : message.severity === 'warning' ? 'secondary' : 'outline';
}

export function buildLegacyHomeboxImportRequest(input: ImportRequestInput): LegacyHomeboxImportRequest {
  if (input.sourceType === 'legacy_homebox_csv') {
    return {
      sourceType: input.sourceType,
      fileName: input.fileName,
      contentBase64: input.contentBase64
    };
  }

  return {
    sourceType: input.sourceType,
    baseUrl: input.baseUrl.trim(),
    username: input.username.trim(),
    password: input.password,
    includeImages: input.includeImages,
    allowInsecureTLS: input.allowInsecureTLS,
    allowPrivateNetwork: input.allowPrivateNetwork
  };
}

function countPhrase(count: number, noun: string): string {
  return count > 0 ? `${count} ${pluralize(count, noun)}` : '';
}

function pluralize(count: number, noun: string): string {
  return count === 1 ? noun : `${noun}s`;
}

function compactList(parts: string[]): string {
  const values = parts.filter(Boolean);
  if (values.length === 0) {
    return '';
  }
  if (values.length === 1) {
    return values[0];
  }
  if (values.length === 2) {
    return `${values[0]} and ${values[1]}`;
  }
  return `${values.slice(0, -1).join(', ')}, and ${values[values.length - 1]}`;
}
