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
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, value]) => ({
        label: labelMetadataKey(key),
        value
      }))
  };
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
