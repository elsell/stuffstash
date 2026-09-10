import { assertReadActive } from '../../application/shared/ReadRequest';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { StuffStashAPIError, type StuffStashClient, type ExpirationNotification as WireNotification, type NotificationPreferences as WirePreferences } from '@stuff-stash/api-client';
import type { ExpirationNotification, ExpirationReminderPolicy, NotificationPreferences, NotificationPreferencesUpdate } from '../../domain/notifications/Notification';
import type { InboxOptions, NotificationRepository } from '../../application/notifications/NotificationRepository';

export class ApiNotificationRepository implements NotificationRepository {
  constructor(private readonly client: StuffStashClient) {}
  async getPreferences(tenantId: string, inventoryId: string, signal?: AbortSignal) {
    return mapPreferences(await this.safe(() => this.client.notifications.getPreferences(tenantId, inventoryId, signal), signal));
  }
  async initializePreferences(tenantId: string, inventoryId: string, timezone: string, signal?: AbortSignal) {
    return mapPreferences(await this.safe(() => this.client.notifications.initializePreferences(tenantId, inventoryId, timezone, signal), signal));
  }
  async updatePreferences(tenantId: string, inventoryId: string, input: NotificationPreferencesUpdate, signal?: AbortSignal) {
    return mapPreferences(await this.safe(() => this.client.notifications.updatePreferences(tenantId, inventoryId, input, signal), signal));
  }
  async setTypeOverride(tenantId: string, inventoryId: string, typeId: string, revision: number, settings: ExpirationReminderPolicy, signal?: AbortSignal) {
    return mapPreferences(await this.safe(() => this.client.notifications.setTypeOverride(tenantId, inventoryId, typeId, revision, settings, signal), signal));
  }
  async removeTypeOverride(tenantId: string, inventoryId: string, typeId: string, revision: number, signal?: AbortSignal) {
    return mapPreferences(await this.safe(() => this.client.notifications.removeTypeOverride(tenantId, inventoryId, typeId, revision, signal), signal));
  }
  async listInbox(tenantId: string, inventoryId: string, options?: InboxOptions) {
    const page = await this.safe(() => this.client.notifications.listInbox(tenantId, inventoryId, options), options?.signal);
    return { items: page.items.map(mapNotification), pagination: page.pagination };
  }
  async getNotification(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal) {
    return mapNotification(await this.safe(() => this.client.notifications.getNotification(tenantId, inventoryId, notificationId, signal), signal));
  }
  countUnreadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal) {
    return this.safe(() => this.client.notifications.countUnreadPage(tenantId, inventoryId, cursor, signal), signal);
  }
  markAllReadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal) {
    return this.safe(() => this.client.notifications.markAllReadPage(tenantId, inventoryId, cursor, signal), signal);
  }
  markRead(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal) {
    return this.safe(() => this.client.notifications.markRead(tenantId, inventoryId, notificationId, signal), signal);
  }
  private async safe<T>(operation: () => Promise<T>, signal?: AbortSignal): Promise<T> {
    assertReadActive(signal);
    try { const result = await operation(); assertReadActive(signal); return result; }
    catch (error) {
      assertReadActive(signal);
      if (error instanceof Error && error.name === 'AbortError') throw error;
      throw new NotificationFailure(error instanceof StuffStashAPIError
        ? error.status === 401 ? 'authentication-required' : error.status === 403 ? 'permission-denied'
        : error.status === 404 ? 'not-found' : error.status === 409 ? 'conflict'
        : error.status === 400 || error.status === 422 ? 'invalid' : 'unavailable'
        : 'unavailable');
    }
  }

}
function mapPreferences(value: WirePreferences): NotificationPreferences {
  return { revision: value.revision, defaults: { ...value.defaults }, timezone: value.timezone, pushEnabled: value.pushEnabled,
    overrides: (value.overrides ?? []).map((override) => ({ customAssetTypeId: override.customAssetTypeId, settings: { ...override.settings } })) };
}
function mapNotification(value: WireNotification): ExpirationNotification {
  return { id: value.id, assetId: value.assetId, title: value.title, parentAssetId: value.parentAssetId, customAssetTypeId: value.customAssetTypeId,
    expiration: { date: value.expirationDate, precision: value.expirationPrecision }, milestone: value.milestone, createdAt: value.createdAt, readAt: value.readAt };
}
