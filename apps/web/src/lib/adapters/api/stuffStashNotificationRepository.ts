import { StuffStashClient, type TokenProvider, type ExpirationNotification as WireNotification, type NotificationPreferences as WirePreferences } from '@stuff-stash/api-client';
import type { ExpirationNotification, ExpirationReminderPolicy, NotificationPreferences, NotificationPreferencesUpdate } from '$lib/domain/notification';
import type { InboxOptions, NotificationRepository } from '$lib/ports/notificationRepository';

export class StuffStashNotificationRepository implements NotificationRepository {
  private readonly client: StuffStashClient;
  constructor(apiBaseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) {
    this.client = new StuffStashClient({ baseUrl: apiBaseUrl, tokenProvider, fetch: fetchImpl });
  }
  async getPreferences(tenantId: string, inventoryId: string, signal?: AbortSignal) {
    return mapPreferences(await this.client.notifications.getPreferences(tenantId, inventoryId, signal));
  }
  async initializePreferences(tenantId: string, inventoryId: string, timezone: string, signal?: AbortSignal) {
    return mapPreferences(await this.client.notifications.initializePreferences(tenantId, inventoryId, timezone, signal));
  }
  async updatePreferences(tenantId: string, inventoryId: string, input: NotificationPreferencesUpdate, signal?: AbortSignal) {
    return mapPreferences(await this.client.notifications.updatePreferences(tenantId, inventoryId, input, signal));
  }
  async setTypeOverride(tenantId: string, inventoryId: string, typeId: string, revision: number, settings: ExpirationReminderPolicy, signal?: AbortSignal) {
    return mapPreferences(await this.client.notifications.setTypeOverride(tenantId, inventoryId, typeId, revision, settings, signal));
  }
  async removeTypeOverride(tenantId: string, inventoryId: string, typeId: string, revision: number, signal?: AbortSignal) {
    return mapPreferences(await this.client.notifications.removeTypeOverride(tenantId, inventoryId, typeId, revision, signal));
  }
  async listInbox(tenantId: string, inventoryId: string, options?: InboxOptions) {
    const page = await this.client.notifications.listInbox(tenantId, inventoryId, options);
    return { items: page.items.map(mapNotification), pagination: page.pagination };
  }
  async getNotification(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal) {
    return mapNotification(await this.client.notifications.getNotification(tenantId, inventoryId, notificationId, signal));
  }
  countUnreadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal) {
    return this.client.notifications.countUnreadPage(tenantId, inventoryId, cursor, signal);
  }
  markAllReadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal) {
    return this.client.notifications.markAllReadPage(tenantId, inventoryId, cursor, signal);
  }
  markRead(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal) {
    return this.client.notifications.markRead(tenantId, inventoryId, notificationId, signal);
  }
  markUnread(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal) {
    return this.client.notifications.markUnread(tenantId, inventoryId, notificationId, signal);
  }
}
function mapPreferences(value: WirePreferences): NotificationPreferences {
  return { revision: value.revision, defaults: { ...value.defaults }, timezone: value.timezone, pushEnabled: value.pushEnabled,
    overrides: (value.overrides ?? []).map((override) => ({ customAssetTypeId: override.customAssetTypeId, settings: { ...override.settings } })) };
}
function mapNotification(value: WireNotification): ExpirationNotification {
  return { parentTrail: (value.parentTrail ?? []).map(entry => ({assetId:entry.assetId,title:entry.title,kind:entry.kind})), parentTrailIncomplete: value.parentTrailIncomplete ?? false, id: value.id, assetId: value.assetId, title: value.title, parentAssetId: value.parentAssetId, customAssetTypeId: value.customAssetTypeId,
    expiration: { date: value.expirationDate, precision: value.expirationPrecision }, milestone: value.milestone, createdAt: value.createdAt, readAt: value.readAt };
}
