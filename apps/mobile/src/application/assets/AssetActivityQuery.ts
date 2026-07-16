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
  readonly records: readonly AssetActivityRecordViewModel[];
  readonly nextCursor?: string;
  readonly hasMore: boolean;
  readonly emptyTitle: string;
  readonly emptyMessage: string;
};

export class AssetActivityQuery {
  private readonly entries = new Map<string, AssetActivityEntry>();

  constructor(private readonly repository: AssetActivityRepository) {}

  async execute(input: {
    readonly tenantId: string;
    readonly inventoryId: string;
    readonly assetId: string;
    readonly view?: AssetActivityView;
    readonly limit?: number;
    readonly cursor?: string;
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
      cursor: input.cursor
    });
    for (const entry of page.entries) {
      this.entries.set(activityCacheKey({ tenantId, inventoryId, assetId, activityId: entry.id }), { ...entry, technical: safeTechnicalMetadata(entry.technical) });
    }
    return {
      records: page.entries.map((entry) => toActivityRecordViewModel(this.cachedEntry({ tenantId, inventoryId, assetId, activityId: entry.id }) ?? entry)),
      nextCursor: page.nextCursor,
      hasMore: page.hasMore,
      emptyTitle: view === 'changes' ? 'No changes yet' : 'No activity yet',
      emptyMessage: view === 'changes'
        ? 'Edits to this item will appear here.'
        : 'Technical reads and changes will appear here.'
    };
  }

  cachedEntry(scope: { readonly tenantId: string; readonly inventoryId: string; readonly assetId: string; readonly activityId: string }): AssetActivityEntry | undefined {
    return this.entries.get(activityCacheKey(scope));
  }

  invalidateEntry(scope: { readonly tenantId: string; readonly inventoryId: string; readonly assetId: string; readonly activityId: string }): void {
    this.entries.delete(activityCacheKey(scope));
  }

  async loadEntry(input: { readonly tenantId: string; readonly inventoryId: string; readonly assetId: string; readonly activityId: string }): Promise<AssetActivityEntry | undefined> {
    const cached = this.cachedEntry(input);
    if (cached) return cached;
    let cursor: string | undefined;
    const visited = new Set<string>();
    do {
      const page = await this.execute({ ...input, view: 'all', limit: 100, cursor });
      const entry = this.cachedEntry(input);
      if (entry) return entry;
      cursor = page.hasMore ? page.nextCursor : undefined;
      if (cursor && visited.has(cursor)) break;
      if (cursor) visited.add(cursor);
    } while (cursor);
    return undefined;
  }
}

function activityCacheKey(scope: { readonly tenantId: string; readonly inventoryId: string; readonly assetId: string; readonly activityId: string }): string {
  return `${scope.tenantId}\u0000${scope.inventoryId}\u0000${scope.assetId}\u0000${scope.activityId}`;
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
    occurredAtLabel: formatActivityTime(entry.occurredAt),
    occurredAt: entry.occurredAt,
    actorLabel: entry.principal?.email?.trim() || 'Someone with access',
    sourceLabel: sourceLabel(entry.source)
  };
}

function activityTitle(entry: AssetActivityEntry): string {
  const fields = new Set(entry.changes.map((change) => change.field));
  if (fields.size === 1 && fields.has('title')) return 'Changed name';
  if (fields.size === 1 && fields.has('description')) return 'Updated description';
  if (fields.size === 1 && fields.has('tags')) return 'Changed tags';
  if (fields.size === 1 && fields.has('parent')) return 'Moved item';
  switch (entry.action) {
    case 'asset.created': return 'Added item';
    case 'asset.archived': return 'Archived item';
    case 'asset.restored': return 'Restored item';
    case 'asset.checked_out': return 'Checked out item';
    case 'asset.returned': return 'Returned item';
    case 'asset.viewed': return 'Viewed item';
    case 'asset.listed': return 'Included in a list';
    case 'asset.searched': return 'Included in search';
    default: return entry.category === 'change' ? 'Updated item' : 'Accessed item';
  }
}

function activitySummary(entry: AssetActivityEntry): string {
  if (entry.changes.length === 0) {
    return `${entry.principal?.email?.trim() || 'Someone with access'} · ${sourceLabel(entry.source)}`;
  }
  return entry.changes.map((change) => {
    if (change.previousValue !== undefined || change.currentValue !== undefined) {
      return `${displayValue(change.previousValue)} → ${displayValue(change.currentValue)}`;
    }
    return change.field === 'description' ? 'Description changed' : labelField(change.field);
  }).join(' · ');
}

function displayValue(value: string | undefined): string {
  return value?.trim() || 'None';
}

function labelField(field: AssetActivityEntry['changes'][number]['field']): string {
  switch (field) {
    case 'lifecycle_state': return 'Status changed';
    case 'checkout_state': return 'Checkout changed';
    case 'parent': return 'Location changed';
    case 'tags': return 'Tags changed';
    case 'title': return 'Name changed';
    case 'description': return 'Description changed';
  }
}

function sourceLabel(source: string): string {
  switch (source) {
    case 'api': return 'App';
    case 'conversation':
    case 'voice': return 'Voice';
    case 'import': return 'Import';
    default: return 'Stuff Stash';
  }
}

function formatActivityTime(value: string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return value;
  return new Intl.DateTimeFormat('en-US', {
    month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit'
  }).format(new Date(timestamp));
}
