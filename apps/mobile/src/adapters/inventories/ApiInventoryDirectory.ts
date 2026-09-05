import type { Inventory, Page, StuffStashClient, Tenant } from '@stuff-stash/api-client';
import { SelectedInventoryUnavailableError } from '../../application/shared/SelectedInventoryUnavailableError';

type DirectoryClient = Pick<StuffStashClient, 'listMyTenants' | 'listInventories'>;
export type InventoryDirectory = { readonly tenants: readonly Tenant[]; readonly availableInventories: readonly SelectedInventory[] };
export type SelectedInventory = { readonly tenant: Tenant; readonly inventory: Inventory };
type PendingDirectory = { readonly controller: AbortController; readonly promise: Promise<InventoryDirectory>; consumers: number };
const directoryRetentionMs = 300_000;
const maxDirectoryPages = 1000;

/** Composition-local discovery. Request cancellation belongs to each joining consumer. */
export class ApiInventoryDirectory {
  private selectedId?: string;
  private selectedIdentity?: SelectedInventory;
  private cached?: { readonly value: InventoryDirectory; readonly expiresAt: number };
  private pending?: PendingDirectory;
  constructor(private readonly client: DirectoryClient, private readonly configuredTenantId: string, private readonly now: () => number = Date.now) {}

  async selected(signal?: AbortSignal): Promise<SelectedInventory> {
    const directory = await this.load(signal);
    signal?.throwIfAborted();
    const selected = this.selectedId
      ? directory.availableInventories.find(item => item.inventory.id === this.selectedId)
      : directory.availableInventories.find(item => item.tenant.id === this.configuredTenantId) ?? directory.availableInventories[0];
    if (!selected) { this.selectedIdentity = undefined; throw new SelectedInventoryUnavailableError(); }
    this.selectedId = selected.inventory.id;
    this.selectedIdentity = selected;
    return selected;
  }

  async selectedForCommand(): Promise<SelectedInventory> { return this.selectedIdentity ?? this.selected(); }

  async select(id: string): Promise<void> {
    let directory = await this.load();
    if (!directory.availableInventories.some(item => item.inventory.id === id)) directory = await this.load(undefined, true);
    if (!directory.availableInventories.some(item => item.inventory.id === id)) throw new Error('Selected inventory is not available in the configured tenant.');
    this.selectedId = id;
    this.selectedIdentity = directory.availableInventories.find(item => item.inventory.id === id);
  }

  load(signal?: AbortSignal, refresh = false): Promise<InventoryDirectory> {
    signal?.throwIfAborted();
    if (refresh) this.cached = undefined;
    if (this.cached && this.cached.expiresAt > this.now()) return Promise.resolve(this.cached.value);
    if (!this.pending) {
      const controller = new AbortController();
      const pending: PendingDirectory = { controller, consumers: 0, promise: this.fetch(controller.signal).then(value => {
        controller.signal.throwIfAborted();
        if (this.pending === pending) this.cached = { value, expiresAt: this.now() + directoryRetentionMs };
        return value;
      }).finally(() => { if (this.pending === pending) this.pending = undefined; }) };
      this.pending = pending;
    }
    const pending = this.pending;
    pending.consumers++;
    return new Promise((resolve, reject) => {
      let settled = false;
      const finish = (error?: unknown, value?: InventoryDirectory) => {
        if (settled) return;
        settled = true;
        signal?.removeEventListener('abort', abort);
        pending.consumers--;
        if (pending.consumers === 0 && this.pending === pending) {
          this.pending = undefined;
          pending.controller.abort();
        }
        if (error !== undefined) reject(error); else resolve(value!);
      };
      const abort = () => finish(signal?.reason ?? new Error('Discovery cancelled.'));
      signal?.addEventListener('abort', abort, { once: true });
      pending.promise.then(value => finish(undefined, value), error => finish(error));
    });
  }

  private async fetch(signal: AbortSignal): Promise<InventoryDirectory> {
    const tenants = await collectPages(cursor => this.client.listMyTenants(100, cursor, signal), signal);
    const availableInventories = (await Promise.all(tenants.map(async tenant => {
      const inventories = await collectPages(cursor => this.client.listInventories(tenant.id, 100, cursor, signal), signal);
      return inventories.map(inventory => ({ tenant, inventory }));
    }))).flat();
    return { tenants, availableInventories };
  }
}

async function collectPages<T>(read: (cursor?: string) => Promise<Page<T>>, signal: AbortSignal): Promise<T[]> {
  const rows: T[] = []; const cursors = new Set<string>(); let cursor: string | undefined;
  for (let count = 0; count < maxDirectoryPages; count++) {
    signal.throwIfAborted();
    const page = await read(cursor);
    signal.throwIfAborted();
    rows.push(...page.items);
    cursor = page.pagination.nextCursor ?? undefined;
    if (!cursor) return rows;
    if (cursors.has(cursor)) throw new Error('Invalid inventory discovery cursor.');
    cursors.add(cursor);
  }
  throw new Error('Inventory discovery exceeded the page limit.');
}
