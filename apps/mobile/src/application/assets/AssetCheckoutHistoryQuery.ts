export type AssetCheckoutRecord = {
  readonly id: string;
  readonly state: string;
  readonly checkedOutAt: string;
  readonly checkedOutByPrincipalId: string;
  readonly checkoutDetails?: string;
  readonly returnedAt?: string;
  readonly returnedByPrincipalId?: string;
  readonly returnDetails?: string;
};

export type AssetCheckoutHistoryPage = {
  readonly records: readonly AssetCheckoutRecord[];
  readonly hasMore: boolean;
};

export interface AssetCheckoutHistoryRepository {
  listAssetCheckoutHistory(input: {
    readonly assetId: string;
    readonly limit: number;
  }): Promise<AssetCheckoutHistoryPage>;
}

export type AssetCheckoutHistoryViewModel = {
  readonly assetId: string;
  readonly records: readonly AssetCheckoutRecordViewModel[];
  readonly hasMore: boolean;
  readonly emptyTitle: string;
  readonly emptyMessage: string;
};

export type AssetCheckoutRecordViewModel = {
  readonly id: string;
  readonly title: string;
  readonly subtitle: string;
  readonly statusLabel: string;
  readonly checkedOutLabel: string;
  readonly returnedLabel?: string;
  readonly checkoutDetails?: string;
  readonly returnDetails?: string;
};

const maxDetailsLength = 180;

export class AssetCheckoutHistoryQuery {
  constructor(private readonly repository: AssetCheckoutHistoryRepository) {}

  async execute(input: {
    readonly assetId: string;
    readonly limit?: number;
  }): Promise<AssetCheckoutHistoryViewModel> {
    const assetId = input.assetId.trim();
    if (assetId.length === 0) {
      throw new Error('Asset ID is required.');
    }

    const page = await this.repository.listAssetCheckoutHistory({
      assetId,
      limit: input.limit ?? 20
    });

    return {
      assetId,
      records: page.records.map(toRecordViewModel),
      hasMore: page.hasMore,
      emptyTitle: 'No checkout history yet',
      emptyMessage: 'Checkouts and returns for this asset will appear here.'
    };
  }
}

function toRecordViewModel(record: AssetCheckoutRecord): AssetCheckoutRecordViewModel {
  const returned = record.returnedAt && record.returnedByPrincipalId
    ? `${labelReturnedAt(record.returnedAt)} by ${labelPrincipal(record.returnedByPrincipalId)}`
    : undefined;

  return {
    id: record.id,
    title: record.state === 'returned' ? 'Returned' : 'Checked out',
    subtitle: `${labelCheckedOutAt(record.checkedOutAt)} by ${labelPrincipal(record.checkedOutByPrincipalId)}`,
    statusLabel: labelState(record.state),
    checkedOutLabel: labelCheckedOutAt(record.checkedOutAt),
    returnedLabel: returned,
    checkoutDetails: safeDetails(record.checkoutDetails),
    returnDetails: safeDetails(record.returnDetails)
  };
}

function labelState(state: string): string {
  switch (state) {
    case 'open':
      return 'Checked out';
    case 'returned':
      return 'Returned';
    case 'undone':
      return 'Undone';
    default:
      return state.charAt(0).toUpperCase() + state.slice(1).replaceAll('_', ' ');
  }
}

function labelPrincipal(principalId: string): string {
  return `Principal ${principalId}`;
}

function labelCheckedOutAt(value: string): string {
  return `Checked out ${labelTimestamp(value)}`;
}

function labelReturnedAt(value: string): string {
  return `Returned ${labelTimestamp(value)}`;
}

function labelTimestamp(value: string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) {
    return value;
  }

  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  }).format(new Date(timestamp));
}

function safeDetails(value: string | undefined): string | undefined {
  const trimmed = value?.trim() ?? '';
  if (trimmed.length === 0) {
    return undefined;
  }
  return trimmed.length > maxDetailsLength
    ? `${trimmed.slice(0, maxDetailsLength - 3)}...`
    : trimmed;
}
