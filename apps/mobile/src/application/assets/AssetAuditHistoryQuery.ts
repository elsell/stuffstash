export type AssetAuditRecord = {
  readonly id: string;
  readonly action: string;
  readonly source: string;
  readonly principalId: string;
  readonly targetType: string;
  readonly targetId: string;
  readonly occurredAt: string;
  readonly requestId?: string;
  readonly metadata: Readonly<Record<string, string>>;
};

export type AssetAuditHistoryPage = {
  readonly records: readonly AssetAuditRecord[];
  readonly hasMore: boolean;
};

export interface AssetAuditHistoryRepository {
  listAssetAuditHistory(input: {
    readonly assetId: string;
    readonly limit: number;
  }): Promise<AssetAuditHistoryPage>;
}

export type AssetAuditHistoryViewModel = {
  readonly assetId: string;
  readonly records: readonly AssetAuditRecordViewModel[];
  readonly hasMore: boolean;
  readonly emptyTitle: string;
  readonly emptyMessage: string;
};

export type AssetAuditRecordViewModel = {
  readonly id: string;
  readonly title: string;
  readonly subtitle: string;
  readonly occurredAtLabel: string;
  readonly sourceLabel: string;
  readonly principalLabel: string;
  readonly requestLabel?: string;
  readonly metadataRows: readonly AssetAuditMetadataRowViewModel[];
};

export type AssetAuditMetadataRowViewModel = {
  readonly label: string;
  readonly value: string;
};

const safeMetadataKeys = new Set([
  'attachment_count',
  'content_type',
  'count',
  'file_count',
  'file_name',
  'file_size',
  'from',
  'kind',
  'lifecycle_state',
  'name',
  'photo_count',
  'previous_name',
  'previous_title',
  'size_bytes',
  'status',
  'title',
  'to',
  'type',
  'updated_name',
  'updated_title'
]);

const unsafeMetadataKeyFragments = [
  'authorization',
  'blob',
  'credential',
  'endpoint',
  'internal',
  'key',
  'permission',
  'prompt',
  'provider',
  'relationship',
  'secret',
  'signed',
  'storage',
  'token',
  'transcript',
  'url'
];

const maxMetadataValueLength = 160;

export class AssetAuditHistoryQuery {
  constructor(private readonly repository: AssetAuditHistoryRepository) {}

  async execute(input: {
    readonly assetId: string;
    readonly limit?: number;
  }): Promise<AssetAuditHistoryViewModel> {
    const assetId = input.assetId.trim();
    if (assetId.length === 0) {
      throw new Error('Asset ID is required.');
    }

    const page = await this.repository.listAssetAuditHistory({
      assetId,
      limit: input.limit ?? 20
    });

    return {
      assetId,
      records: page.records.map(toRecordViewModel),
      hasMore: page.hasMore,
      emptyTitle: page.hasMore ? 'No recent history found' : 'No history yet',
      emptyMessage: page.hasMore
        ? 'Older history may be available in the full audit log.'
        : 'Changes to this asset will appear here after they are recorded.'
    };
  }
}

function toRecordViewModel(record: AssetAuditRecord): AssetAuditRecordViewModel {
  return {
    id: record.id,
    title: labelAction(record.action),
    subtitle: `${labelTarget(record.targetType)} ${record.targetId}`,
    occurredAtLabel: labelOccurredAt(record.occurredAt),
    sourceLabel: labelSource(record.source),
    principalLabel: `Principal ${record.principalId}`,
    requestLabel: record.requestId ? `Request ${record.requestId}` : undefined,
    metadataRows: Object.entries(record.metadata)
      .flatMap(toSafeMetadataRow)
      .sort((left, right) => left.label.localeCompare(right.label))
  };
}

function toSafeMetadataRow([key, value]: [string, string]): AssetAuditMetadataRowViewModel[] {
  const normalizedKey = normalizeMetadataKey(key);
  if (!isSafeMetadataKey(normalizedKey)) {
    return [];
  }

  const safeValue = safeMetadataValue(value);
  if (!safeValue) {
    return [];
  }

  return [{
    label: labelMetadataKey(normalizedKey),
    value: safeValue
  }];
}

function isSafeMetadataKey(normalizedKey: string): boolean {
  if (unsafeMetadataKeyFragments.some((fragment) => normalizedKey.includes(fragment))) {
    return false;
  }
  return safeMetadataKeys.has(normalizedKey);
}

function safeMetadataValue(value: string): string | undefined {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    return undefined;
  }
  if (looksUnsafeMetadataValue(trimmed)) {
    return undefined;
  }
  return trimmed.length > maxMetadataValueLength
    ? `${trimmed.slice(0, maxMetadataValueLength - 3)}...`
    : trimmed;
}

function looksUnsafeMetadataValue(value: string): boolean {
  const normalized = value.toLowerCase();
  return normalized.includes('bearer ')
    || normalized.includes('-----begin ')
    || normalized.startsWith('http://')
    || normalized.startsWith('https://')
    || normalized.startsWith('s3://')
    || normalized.startsWith('gs://')
    || normalized.startsWith('garage/');
}

function labelAction(action: string): string {
  return action
    .split('.')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).replaceAll('_', ' '))
    .join(' ');
}

function labelTarget(targetType: string): string {
  return targetType.charAt(0).toUpperCase() + targetType.slice(1).replaceAll('_', ' ');
}

function labelSource(source: string): string {
  switch (source) {
    case 'api':
      return 'API';
    case 'voice':
      return 'Voice';
    default:
      return source.charAt(0).toUpperCase() + source.slice(1).replaceAll('_', ' ');
  }
}

function labelMetadataKey(key: string): string {
  return key
    .replaceAll('_', ' ')
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function normalizeMetadataKey(key: string): string {
  return key.trim().toLowerCase().replaceAll('-', '_');
}

function labelOccurredAt(value: string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) {
    return value;
  }

  return `Recorded ${new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  }).format(new Date(timestamp))}`;
}
