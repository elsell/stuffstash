import { t } from '../../presentation/localization';
import { formatHistoryTimestamp } from './AssetHistoryTimestamp';
import { assertReadActive } from '../shared/ReadRequest';
export type AssetActivityView = 'changes' | 'all';
export type AssetActivityCategory = 'change' | 'read';
export type AssetActivityField = 'title' | 'description' | 'tags' | 'parent' | 'lifecycle_state' | 'checkout_state';
export type AssetActivityEntry = {
  readonly id: string;
  readonly principalId: string;
  readonly principal?: { readonly id: string; readonly email?: string };
  readonly action: string;
  readonly category: AssetActivityCategory;
  readonly source: string;
  readonly occurredAt: string;
  readonly requestId?: string;
  readonly changes: readonly { readonly field: AssetActivityField; readonly previousValue?: string; readonly currentValue?: string }[];
  readonly undo?: { readonly operationId: string; readonly status: 'available' | 'undone' | 'redone' };
  readonly technical: Readonly<Record<string, string>>;
};

export type AssetActivityPage = {
  readonly entries: readonly AssetActivityEntry[];
  readonly nextCursor?: string;
  readonly hasMore: boolean;
};

export interface AssetActivityRepository {
  listAssetActivity(input: {
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId: string;
    readonly view: AssetActivityView;
    readonly limit: number;
    readonly cursor?: string;
    readonly signal?: AbortSignal;
  }): Promise<AssetActivityPage>;
}

export type AssetActivityRecordViewModel = {
  readonly id: string;
  readonly title: string;
  readonly summary: string;
  readonly occurredAtLabel: string;
  readonly occurredAt: string;
  readonly actorLabel: string;
  readonly sourceLabel: string;
};

export type AssetActivityViewModel = {
  readonly entries: readonly AssetActivityEntry[];
  readonly records: readonly AssetActivityRecordViewModel[];
  readonly nextCursor?: string;
  readonly hasMore: boolean;
  readonly emptyTitle: string;
  readonly emptyMessage: string;
};

export class AssetActivityQuery {
  constructor(private readonly repository: AssetActivityRepository) {}

  async execute(input: {
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId: string;
    readonly view?: AssetActivityView;
    readonly limit?: number;
    readonly cursor?: string;
    readonly signal?: AbortSignal;
  }): Promise<AssetActivityViewModel> {
    const tenantId = input.tenantId.trim();
    const inventoryId = input.inventoryId.trim();
    const assetId = input.assetId.trim();
    if (!tenantId || !inventoryId || !assetId) {
      throw new Error('History scope is required.');
    }
    const view = input.view ?? 'changes';
    const page = await this.repository.listAssetActivity({
      tenantId,
      inventoryId,
      assetId,
      view,
      limit: input.limit ?? 20,
      cursor: input.cursor,
      ...(input.signal ? { signal: input.signal } : {})
    });
    const entries = page.entries.map((entry) => ({ ...entry, technical: safeTechnicalMetadata(entry.technical) }));
    return {
      entries,
      records: entries.map(toActivityRecordViewModel),
      nextCursor: page.nextCursor,
      hasMore: page.hasMore,
      emptyTitle: view === 'changes' ? t('history.activity.noChangesYet') : t('history.activity.noActivityYet'),
      emptyMessage: view === 'changes'
        ? t('history.activity.editsToThisItemWillAppearHere')
        : t('history.activity.technicalReadsAndChangesWillAppearHere')
    };
  }

  async loadEntry(input: { readonly tenantId: string; readonly inventoryId: string; readonly assetId: string; readonly activityId: string; readonly signal?: AbortSignal }): Promise<AssetActivityEntry | undefined> {
    let cursor: string | undefined;
    const visited = new Set<string>();
    do {
      assertReadActive(input.signal);
      const page = await this.execute({ ...input, view: 'all', limit: 100, cursor });
      const entry = page.entries.find((entry) => entry.id === input.activityId);
      if (entry) return entry;
      cursor = page.hasMore ? page.nextCursor : undefined;
      if (cursor && visited.has(cursor)) break;
      if (cursor) visited.add(cursor);
    } while (cursor);
    return undefined;
  }
}

const safeTechnicalKeys = new Set([
  'previous_parent', 'new_parent', 'previous_title', 'updated_title',
  'previous_lifecycle_state', 'new_lifecycle_state', 'kind', 'type',
  'attachment_file_name', 'content_type', 'file_size', 'count'
]);

function safeTechnicalMetadata(metadata: Readonly<Record<string, string>>): Readonly<Record<string, string>> {
  return Object.fromEntries(Object.entries(metadata).filter(([key]) => safeTechnicalKeys.has(key)));
}

function toActivityRecordViewModel(entry: AssetActivityEntry): AssetActivityRecordViewModel {
  return {
    id: entry.id,
    title: activityTitle(entry),
    summary: activitySummary(entry),
    occurredAtLabel: formatHistoryTimestamp(entry.occurredAt, 'activity'),
    occurredAt: entry.occurredAt,
    actorLabel: entry.principal?.email?.trim() || t('history.activity.someoneWithAccess'),
    sourceLabel: sourceLabel(entry.source)
  };
}

function activityTitle(entry: AssetActivityEntry): string {
  const fields = new Set(entry.changes.map((change) => change.field));
  if (fields.size === 1 && fields.has('title')) return t('history.activity.changedName');
  if (fields.size === 1 && fields.has('description')) return t('history.activity.updatedDescription');
  if (fields.size === 1 && fields.has('tags')) return t('history.activity.changedTags');
  if (fields.size === 1 && fields.has('parent')) return t('history.activity.movedItem');
  switch (entry.action) {
    case 'label.provisioned': return t('audit.action.label.provisioned');
    case 'label.viewed': return t('audit.action.label.viewed');
    case 'label.resolved': return t('audit.action.label.resolved');
    case 'label.rendered': return t('audit.action.label.rendered');
    case 'label.content_downloaded': return t('audit.action.label.content_downloaded');
    case 'label.templates_listed': return t('audit.action.label.templates_listed');

    case 'asset.created': return t('history.activity.addedItem');
    case 'asset.archived': return t('history.activity.archivedItem');
    case 'asset.restored': return t('history.activity.restoredItem');
    case 'asset.checked_out': return t('history.activity.checkedOutItem');
    case 'asset.returned': return t('history.activity.returnedItem');
    case 'asset.viewed': return t('history.activity.viewedItem');
    case 'asset.listed': return t('history.activity.includedInAList');
    case 'asset.searched': return t('history.activity.includedInSearch');
    default: return entry.category === 'change' ? t('history.activity.updatedItem') : t('history.activity.accessedItem');
  }
}

function activitySummary(entry: AssetActivityEntry): string {
  if (entry.changes.length === 0) {
    return `${entry.principal?.email?.trim() || t('history.activity.someoneWithAccess')} · ${sourceLabel(entry.source)}`;
  }
  const fields = [...new Set(entry.changes.map(change => change.field))];
  if (fields.length > 1) return fields.map(activityFieldLabel).join(' · ');
  return entry.changes.map((change) => {
    if (change.previousValue !== undefined || change.currentValue !== undefined) {
      return `${displayValue(change.previousValue)} → ${displayValue(change.currentValue)}`;
    }
    return change.field === 'description' ? t('history.activity.descriptionChanged') : labelField(change.field);
  }).join(' · ');
}

function displayValue(value: string | undefined): string {
  return value?.trim() || t('history.activity.none');
}

function labelField(field: AssetActivityEntry['changes'][number]['field']): string {
  switch (field) {
    case 'lifecycle_state': return t('history.activity.statusChanged');
    case 'checkout_state': return t('history.activity.checkoutChanged');
    case 'parent': return t('history.activity.locationChanged');
    case 'tags': return t('history.activity.tagsChanged');
    case 'title': return t('history.activity.nameChanged');
    case 'description': return t('history.activity.descriptionChanged');
  }
}

function sourceLabel(source: string): string {
  switch (source) {
    case 'api': return t('history.activity.app');
    case 'conversation':
    case 'voice': return t('history.activity.voice');
    case 'import': return t('history.activity.import');
    default: return t('history.activity.stuffStash');
  }
}

export function activityFieldLabel(field: AssetActivityField): string {
  switch (field) {
    case 'title': return t('history.activity.name');
    case 'description': return t('history.activity.description');
    case 'tags': return t('history.activity.tags');
    case 'parent': return t('history.activity.location');
    case 'lifecycle_state': return t('history.activity.status');
    case 'checkout_state': return t('history.activity.checkout');
  }
}
