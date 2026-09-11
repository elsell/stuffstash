import type { ExpirationNotification, ExpirationReminderPolicy, NotificationPreferences, NotificationPreferencesUpdate } from '../../domain/notifications/Notification';
type Pagination = { limit: number; nextCursor: string | null; hasMore: boolean };

export interface NotificationPage { items: ExpirationNotification[]; pagination: Pagination }
export interface InboxOptions { cursor?: string; limit?: number; unreadOnly?: boolean; signal?: AbortSignal }
export interface NotificationRepository {
  countUnreadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal): Promise<{ count: number; nextCursor: string | null }>;
  markAllReadPage(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal): Promise<{ complete: boolean; nextCursor: string | null }>;
  getPreferences(tenantId: string, inventoryId: string, signal?: AbortSignal): Promise<NotificationPreferences>;
  initializePreferences(tenantId: string, inventoryId: string, timezone: string, signal?: AbortSignal): Promise<NotificationPreferences>;
  updatePreferences(tenantId: string, inventoryId: string, input: NotificationPreferencesUpdate, signal?: AbortSignal): Promise<NotificationPreferences>;
  setTypeOverride(tenantId: string, inventoryId: string, typeId: string, revision: number, settings: ExpirationReminderPolicy, signal?: AbortSignal): Promise<NotificationPreferences>;
  removeTypeOverride(tenantId: string, inventoryId: string, typeId: string, revision: number, signal?: AbortSignal): Promise<NotificationPreferences>;
  listInbox(tenantId: string, inventoryId: string, options?: InboxOptions): Promise<NotificationPage>;
  getNotification(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal): Promise<ExpirationNotification>;
  markRead(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal): Promise<void>;
  markUnread(tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal): Promise<void>;
}
