import { t } from '../../presentation/localization';
import { formatHistoryTimestamp } from './AssetHistoryTimestamp';
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
  readonly nextCursor?: string;
};

export interface AssetCheckoutHistoryRepository {
  listAssetCheckoutHistory(input: {
    readonly assetId: string;
    readonly cursor?: string;
    readonly signal?: AbortSignal;
    readonly limit: number;
  }): Promise<AssetCheckoutHistoryPage>;
}

export type AssetCheckoutHistoryViewModel = {
  readonly assetId: string;
  readonly records: readonly AssetCheckoutRecordViewModel[];
  readonly hasMore: boolean;
  readonly nextCursor?: string;
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
    readonly cursor?: string;
    readonly signal?: AbortSignal;
    readonly limit?: number;
  }): Promise<AssetCheckoutHistoryViewModel> {
    const assetId = input.assetId.trim();
    if (assetId.length === 0) {
      throw new Error('Asset ID is required.');
    }

    const page = await this.repository.listAssetCheckoutHistory({
      assetId,
      limit: input.limit ?? 20,
      cursor: input.cursor,
      signal: input.signal
    });

    return {
      assetId,
      records: page.records.map(toRecordViewModel),
      hasMore: page.hasMore,
      nextCursor: page.nextCursor,
      emptyTitle: t('mobile.AssetCheckoutHistoryQuery.emptyTitle'),
      emptyMessage: t('mobile.AssetCheckoutHistoryQuery.emptyMessage')
    };
  }
}

function toRecordViewModel(record: AssetCheckoutRecord): AssetCheckoutRecordViewModel {
  const returned = record.returnedAt && record.returnedByPrincipalId
    ? t('mobile.AssetCheckoutHistoryQuery.returnedBy', { time: formatHistoryTimestamp(record.returnedAt, 'checkout'), principal: record.returnedByPrincipalId })
    : undefined;

  return {
    id: record.id,
    title: record.state === 'returned' ? t('mobile.AssetCheckoutHistoryQuery.returned') : t('mobile.AssetCheckoutHistoryQuery.checkedOut'),
    subtitle: t('mobile.AssetCheckoutHistoryQuery.checkedOutBy', { time: formatHistoryTimestamp(record.checkedOutAt, 'checkout'), principal: record.checkedOutByPrincipalId }),
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
      return t('mobile.AssetCheckoutHistoryQuery.checkedOut');
    case 'returned':
      return t('mobile.AssetCheckoutHistoryQuery.returned');
    case 'undone':
      return t('mobile.AssetCheckoutHistoryQuery.undone');
    default:
      return state.charAt(0).toUpperCase() + state.slice(1).replaceAll('_', ' ');
  }
}

function labelCheckedOutAt(value: string): string {
  return t('mobile.AssetCheckoutHistoryQuery.checkedOutAt', { time: formatHistoryTimestamp(value, 'checkout') });
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
