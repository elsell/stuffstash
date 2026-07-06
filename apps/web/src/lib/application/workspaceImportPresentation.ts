import type { ImportApplyResult, ImportMessage, ImportPreview, ImportSourceType, LegacyHomeboxImportRequest } from '$lib/domain/inventory';
import { importSourceHref } from './workspaceShellNavigation';

export interface ImportSourceOption {
  value: ImportSourceType;
  label: string;
  href: string;
  disabled?: boolean;
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
  activeOperation: ImportOperation | null;
  hasPreview: boolean;
  blockingErrorCount: number;
  canImport: boolean;
}

export interface ImportPanelMessage {
  title: string;
  description?: string;
}

export type ImportFailureOperation = 'preview' | 'apply';
export type ImportOperation = 'preview' | 'apply';

export type ImportWorkflowStepState = 'complete' | 'current' | 'blocked' | 'pending';

export interface ImportWorkflowStep {
  label: string;
  description: string;
  state: ImportWorkflowStepState;
}

export interface ImportWorkflowInput {
  ready: boolean;
  hasPreview: boolean;
  hasBlockingErrors: boolean;
  hasResult: boolean;
  activeOperation: ImportOperation | null;
}

export const importPreviewDisplayLimits = {
  messages: 12,
  fields: 10,
  assets: 8,
  images: 6
} as const;

export const maxHomeboxCSVBytes = 10 * 1024 * 1024;

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
  if (input.activeOperation === 'preview') {
    return 'Preview is reading the source.';
  }
  if (input.activeOperation === 'apply') {
    return 'Import is applying. Keep this tab open.';
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

export function importOperationPresentation(operation: ImportOperation): Required<ImportPanelMessage> {
  if (operation === 'preview') {
    return {
      title: 'Previewing source',
      description: 'Stuff Stash is reading Homebox and building a plan. Nothing has been saved.'
    };
  }
  return {
    title: 'Applying import',
    description: 'Stuff Stash is creating records from the current preview. Keep this tab open.'
  };
}

export function importWorkflowSteps(input: ImportWorkflowInput): ImportWorkflowStep[] {
  const sourceComplete = input.ready || input.hasPreview || input.hasResult;
  const previewComplete = input.hasPreview && !input.hasBlockingErrors;
  return [
    {
      label: 'Source',
      description: sourceComplete ? 'Ready' : 'Needs connection details',
      state: sourceComplete ? 'complete' : 'current'
    },
    {
      label: 'Preview',
      description:
        input.activeOperation === 'preview'
          ? 'Reading source'
          : input.hasBlockingErrors
            ? 'Has errors'
            : input.hasPreview
              ? 'Reviewed'
              : 'Not run',
      state:
        input.activeOperation === 'preview'
          ? 'current'
          : input.hasBlockingErrors
            ? 'blocked'
            : input.hasPreview
              ? 'complete'
              : sourceComplete
                ? 'current'
                : 'pending'
    },
    {
      label: 'Apply',
      description:
        input.activeOperation === 'apply'
          ? 'Saving records'
          : input.hasResult
            ? 'Applied'
            : previewComplete
              ? 'Ready'
              : 'Locked',
      state:
        input.activeOperation === 'apply'
          ? 'current'
          : input.hasResult
            ? 'complete'
            : input.hasBlockingErrors
              ? 'blocked'
              : previewComplete
                ? 'current'
                : 'pending'
    }
  ];
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

export function importPreviewStatus(preview: ImportPreview, blockingErrorCount: number): Required<ImportPanelMessage> {
  if (blockingErrorCount > 0) {
    return {
      title: 'Preview needs attention',
      description: `${blockingErrorCount} blocking ${pluralize(blockingErrorCount, 'error')} must be resolved before apply.`
    };
  }
  const planned = importPlannedCountLabel(preview);
  return {
    title: 'Preview ready',
    description: `${planned} from ${importPreviewSourceSummary(preview.source)}. Nothing has been saved yet.`
  };
}

export function importSampleLimitLabel(total: number, shown: number, noun: string): string {
  if (total <= shown) {
    return `${total} ${pluralize(total, noun)}`;
  }
  return `Showing ${shown} of ${total} ${pluralize(total, noun)}`;
}

export function importFailurePresentation(
  operation: ImportFailureOperation,
  sourceType: ImportSourceType,
  caught: unknown
): Required<ImportPanelMessage> {
  const title = operation === 'preview' ? 'Preview failed' : 'Import failed';
  const message = caught instanceof Error ? caught.message.trim() : '';
  if (isFetchFailureMessage(message)) {
    return {
      title,
      description:
        operation === 'preview'
          ? sourceType === 'legacy_homebox'
            ? 'The preview request could not complete. Check that Stuff Stash is reachable, then verify the Homebox URL. For a local Homebox server, enable Private network address and try again.'
            : 'The preview request could not complete. Check that Stuff Stash is reachable and try again.'
          : 'The apply request could not complete. Check that Stuff Stash is reachable and try again.'
    };
  }
  if (!isUserSafeError(caught)) {
    return {
      title,
      description:
        operation === 'preview'
          ? 'The preview could not be completed. Check the source details and try again.'
          : 'The import could not be completed. Review the preview and try again.'
    };
  }
  return {
    title,
    description: message || (operation === 'preview' ? 'Import preview failed.' : 'Import failed.')
  };
}

export function legacyHomeboxImportRequestKey(request: LegacyHomeboxImportRequest, csvVersion = ''): string {
  if (request.sourceType === 'legacy_homebox_csv') {
    return JSON.stringify({
      sourceType: request.sourceType,
      fileName: request.fileName ?? '',
      csvVersion
    });
  }
  return JSON.stringify({
    sourceType: request.sourceType,
    baseUrl: request.baseUrl ?? '',
    username: request.username ?? '',
    password: request.password ?? '',
    includeImages: request.includeImages ?? false,
    allowInsecureTLS: request.allowInsecureTLS ?? false,
    allowPrivateNetwork: request.allowPrivateNetwork ?? false
  });
}

export function importMessagesForDisplay(messages: ImportMessage[], limit = importPreviewDisplayLimits.messages): ImportMessage[] {
  return [...messages].sort(compareImportMessages).slice(0, limit);
}

export function importHiddenCount(total: number, shown: number): number {
  return Math.max(0, total - shown);
}

export function importHiddenLabel(hidden: number, noun: string): string {
  return hidden > 0 ? `${hidden} more ${pluralize(hidden, noun)} not shown.` : '';
}

export function csvFileTooLargePresentation(fileSize: number): Required<ImportPanelMessage> {
  return {
    title: 'CSV is too large',
    description: `Choose a Homebox CSV under ${formatBytes(maxHomeboxCSVBytes)}. This file is ${formatBytes(fileSize)}.`
  };
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

function isFetchFailureMessage(message: string): boolean {
  const normalized = message.toLowerCase();
  return (
    normalized === 'load failed' ||
    normalized === 'failed to fetch' ||
    normalized === 'networkerror when attempting to fetch resource.'
  );
}

function isUserSafeError(caught: unknown): boolean {
  return typeof caught === 'object' && caught !== null && (caught as { safeForUser?: unknown }).safeForUser === true;
}

function compareImportMessages(left: ImportMessage, right: ImportMessage): number {
  return severityRank(left.severity) - severityRank(right.severity);
}

function severityRank(severity: string): number {
  if (severity === 'error') {
    return 0;
  }
  if (severity === 'warning') {
    return 1;
  }
  return 2;
}

function formatBytes(bytes: number): string {
  const mib = bytes / (1024 * 1024);
  if (mib >= 1) {
    return `${mib.toFixed(mib >= 10 ? 0 : 1)} MB`;
  }
  const kib = bytes / 1024;
  return `${Math.max(1, Math.round(kib))} KB`;
}
