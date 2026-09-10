import type { Client } from 'openapi-fetch';
import type { components, paths } from './generated/schema';
import type { Page } from './stuffStashClient';

export type ExpirationReminderPolicy = components['schemas']['ExpirationPolicy'];
export type NotificationPreferences = components['schemas']['PreferencesResponse'];
export type ExpirationNotification = components['schemas']['NotificationResponse'];
export type UpdateNotificationPreferences = components['schemas']['UpdateBody'];
type Result<T> = { data?: T; error?: components['schemas']['ErrorEnvelope']; response: Response };
export interface NotificationClientTransport {
  headers(): Promise<Record<string, string>>;
  unwrap<T>(request: Promise<Result<T>>): Promise<T>;
}
const preferencesPath = '/tenants/{tenantId}/inventories/{inventoryId}/notification-preferences';
const inboxPath = '/tenants/{tenantId}/inventories/{inventoryId}/notifications';

export class NotificationsClient {
  constructor(private readonly client: Client<paths>, private readonly transport: NotificationClientTransport) {}
  async getPreferences(tenantId: string, inventoryId: string, signal?: AbortSignal): Promise<NotificationPreferences> {
    const result = await this.transport.unwrap(this.client.GET(preferencesPath, { params: { path: { tenantId, inventoryId } }, headers: await this.transport.headers(), signal }));
    return result.data;
  }
  async initializePreferences(tenantId: string, inventoryId: string, timezone: string, signal?: AbortSignal): Promise<NotificationPreferences> {
    const result = await this.transport.unwrap(this.client.POST(`${preferencesPath}/initialize`, { params: { path: { tenantId, inventoryId } }, body: { timezone }, headers: await this.transport.headers(), signal }));
    return result.data;
  }
  async updatePreferences(tenantId: string, inventoryId: string, input: UpdateNotificationPreferences, signal?: AbortSignal): Promise<NotificationPreferences> {
    const { revision, defaults, timezone, pushEnabled } = input;
    const result = await this.transport.unwrap(this.client.PUT(preferencesPath, { params: { path: { tenantId, inventoryId } }, body: { revision, defaults, timezone, pushEnabled }, headers: await this.transport.headers(), signal }));
    return result.data;
  }
  async setTypeOverride(tenantId: string, inventoryId: string, customAssetTypeId: string, revision: number, settings: ExpirationReminderPolicy, signal?: AbortSignal): Promise<NotificationPreferences> {
    const result = await this.transport.unwrap(this.client.PUT(`${preferencesPath}/types/{customAssetTypeId}`, { params: { path: { tenantId, inventoryId, customAssetTypeId } }, body: { revision, settings }, headers: await this.transport.headers(), signal }));
    return result.data;
  }
  async removeTypeOverride(tenantId: string, inventoryId: string, customAssetTypeId: string, revision: number, signal?: AbortSignal): Promise<NotificationPreferences> {
    const result = await this.transport.unwrap(this.client.DELETE(`${preferencesPath}/types/{customAssetTypeId}`, { params: { path: { tenantId, inventoryId, customAssetTypeId }, query: { revision } }, headers: await this.transport.headers(), signal }));
    return result.data;
  }
  async listInbox(tenantId: string, inventoryId: string, options: { cursor?: string; limit?: number; unreadOnly?: boolean; signal?: AbortSignal } = {}): Promise<Page<ExpirationNotification>> {
    const { signal, ...query } = options;
    const result = await this.transport.unwrap(this.client.GET(inboxPath, { params: { path: { tenantId, inventoryId }, query }, headers: await this.transport.headers(), signal }));
    return { items: result.data ?? [], pagination: { limit: result.meta.pagination?.limit ?? 0, nextCursor: result.meta.pagination?.nextCursor ?? null, hasMore: result.meta.pagination?.hasMore ?? false } };
  }
  async getNotification(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal): Promise<ExpirationNotification> {
    const result = await this.transport.unwrap(this.client.GET(`${inboxPath}/{notificationId}`, { params: { path: { tenantId, inventoryId, notificationId } }, headers: await this.transport.headers(), signal }));
    return result.data;
  }
  async countUnreadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal): Promise<{ count: number; nextCursor: string | null }> {
    const result = await this.transport.unwrap(this.client.GET(`${inboxPath}/unread-count`, { params: { path: { tenantId, inventoryId }, query: { cursor } }, headers: await this.transport.headers(), signal }));
    return { count: result.data.count, nextCursor: result.meta.pagination?.nextCursor ?? null };
  }
  async markAllReadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal): Promise<{ complete: boolean; nextCursor: string | null }> {
    const result = await this.transport.unwrap(this.client.PUT(`${inboxPath}/read-all`, { params: { path: { tenantId, inventoryId }, query: { cursor } }, headers: await this.transport.headers(), signal }));
    return { complete: result.data.complete, nextCursor: result.meta.pagination?.nextCursor ?? null };
  }
  async markRead(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal): Promise<void> {
    await this.transport.unwrap(this.client.PUT(`${inboxPath}/{notificationId}/read`, { params: { path: { tenantId, inventoryId, notificationId } }, headers: await this.transport.headers(), signal }));
  }
}
