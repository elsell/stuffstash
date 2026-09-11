import type { NotificationDevice, RegisterNotificationDevice } from '../../domain/notifications/NotificationDevice';
export interface NotificationDeviceRepository {
 register(tenantId: string, inventoryId: string, input: RegisterNotificationDevice, signal?: AbortSignal): Promise<NotificationDevice>;
 getByInstallation(tenantId: string, inventoryId: string, installationId: string, signal?: AbortSignal): Promise<NotificationDevice>;
 revoke(tenantId: string, inventoryId: string, deviceId: string, revision: number, signal?: AbortSignal): Promise<NotificationDevice>;
}
