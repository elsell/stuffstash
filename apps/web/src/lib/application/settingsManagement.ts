import type { CustomizationPage } from '$lib/ports/inventoryCustomizationRepository';

type SharedSettingsLoad = { promise: Promise<unknown>; resolvedAt: number | null };
const sharedPendingSettingsLoads = new WeakMap<object, Map<string, SharedSettingsLoad>>();

export function sharePendingSettingsLoad<T>(owner: object, key: string, load: () => Promise<T>, cacheResolved = false): Promise<T> {
  let pending = sharedPendingSettingsLoads.get(owner);
  if (!pending) {
    pending = new Map();
    sharedPendingSettingsLoads.set(owner, pending);
  }
  const existing = pending.get(key);
  if (existing) return existing.promise as Promise<T>;
  const entry: SharedSettingsLoad = { promise: Promise.resolve(undefined), resolvedAt: null };
  const promise = load().then((value) => {
    if (pending?.get(key) === entry) {
      if (cacheResolved) entry.resolvedAt = Date.now();
      else pending.delete(key);
    }
    return value;
  }, (caught) => {
    if (pending?.get(key) === entry) pending.delete(key);
    throw caught;
  });
  entry.promise = promise;
  pending.set(key, entry);
  return promise;
}

export function invalidateSharedSettingsLoads(owner: object, keyPrefix: string): void {
  const pending = sharedPendingSettingsLoads.get(owner);
  if (!pending) return;
  for (const key of pending.keys()) {
    if (key.startsWith(keyPrefix)) pending.delete(key);
  }
}

export async function collectSettingsPages<T>(load: (cursor?: string) => Promise<CustomizationPage<T>>, maxPages = 100): Promise<T[]> {
  const items: T[] = [];
  let cursor: string | undefined;
  const seen = new Set<string>();
  for (let pageIndex = 0; pageIndex < maxPages; pageIndex += 1) {
    const page = await load(cursor);
    items.push(...page.items);
    if (!page.pagination.hasMore) return items;
    const next = page.pagination.nextCursor;
    if (!next || seen.has(next)) throw new Error('Settings collection could not be fully loaded.');
    seen.add(next);
    cursor = next;
  }
  throw new Error('Settings collection could not be fully loaded.');
}

const collator = new Intl.Collator(undefined, { sensitivity: 'base', numeric: true });

export function sortSettingsRecords<T extends { id: string; displayName: string }>(items: T[]): T[] {
  return [...items].sort((left, right) => collator.compare(left.displayName, right.displayName) || left.id.localeCompare(right.id));
}

export function mergeCanonicalSettingsRecord<T extends { id: string; displayName: string }>(items: T[], record: T): T[] {
  return sortSettingsRecords([...items.filter((item) => item.id !== record.id), record]);
}

export function removeCanonicalSettingsRecord<T extends { id: string }>(items: T[], recordId: string): T[] {
  return items.filter((item) => item.id !== recordId);
}

export function filterSettingsRecords<T extends { displayName: string; key?: string }>(items: T[], query: string): T[] {
  const normalized = query.trim().toLocaleLowerCase();
  if (!normalized) return items;
  return items.filter((item) => item.displayName.toLocaleLowerCase().includes(normalized) || item.key?.toLocaleLowerCase().includes(normalized));
}

export function settingsKeyFromName(value: string): string {
  return value.trim().toLocaleLowerCase().normalize('NFKD').replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 80).replace(/-+$/g, '');
}

export function normalizeTagColor(value: string): string | undefined | null {
  const raw = value.trim();
  if (!raw) return undefined;
  const color = raw.startsWith('#') ? raw : `#${raw}`;
  return /^#[0-9a-fA-F]{6}$/.test(color) ? color.toUpperCase() : null;
}

export function utf8ByteLength(value: string): number {
  return new TextEncoder().encode(value).length;
}

export function isSettingsPermissionDenied(caught: unknown): boolean {
  return typeof caught === 'object' && caught !== null && (caught as { status?: unknown }).status === 403;
}

const namedColors = [
  { name: 'Red', rgb: [220, 38, 38] }, { name: 'Orange', rgb: [234, 88, 12] },
  { name: 'Yellow', rgb: [202, 138, 4] }, { name: 'Green', rgb: [22, 163, 74] },
  { name: 'Blue', rgb: [37, 99, 235] }, { name: 'Purple', rgb: [124, 58, 237] },
  { name: 'Pink', rgb: [219, 39, 119] }, { name: 'Gray', rgb: [107, 114, 128] },
  { name: 'Black', rgb: [0, 0, 0] }, { name: 'White', rgb: [255, 255, 255] }
] as const;

export function tagColorAccessibleLabel(color?: string): string {
  const normalized = color ? normalizeTagColor(color) : undefined;
  if (!normalized) return 'No color';
  const rgb = [1, 3, 5].map((index) => Number.parseInt(normalized.slice(index, index + 2), 16));
  const closest = namedColors.reduce((best, candidate) => {
    const distance = candidate.rgb.reduce<number>((total, channel, index) => total + ((channel - rgb[index]) ** 2), 0);
    return distance < best.distance ? { name: candidate.name, distance } : best;
  }, { name: 'Color', distance: Number.POSITIVE_INFINITY });
  return `${closest.name} color (${normalized})`;
}
