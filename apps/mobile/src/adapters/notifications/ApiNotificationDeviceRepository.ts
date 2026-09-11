import type { NotificationDevice as WireDevice, StuffStashClient } from '@stuff-stash/api-client';
import type { NotificationDeviceRepository } from '../../application/notifications/NotificationDeviceRepository';
import type { NotificationDevice, RegisterNotificationDevice } from '../../domain/notifications/NotificationDevice';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { safeNotificationRequest } from './safeNotificationRequest';
export class ApiNotificationDeviceRepository implements NotificationDeviceRepository {
 constructor(private readonly client: StuffStashClient) {}
 async register(tenantId: string, inventoryId: string, input: RegisterNotificationDevice, signal?: AbortSignal) {
  return mapDevice(await safeNotificationRequest(() => this.client.notifications.registerDevice(tenantId,inventoryId,input,signal),signal));
 }
 async getByInstallation(tenantId: string, inventoryId: string, installationId: string, signal?: AbortSignal) {
  return mapDevice(await safeNotificationRequest(() => this.client.notifications.getDeviceByInstallation(tenantId,inventoryId,installationId,signal),signal));
 }
 async revoke(tenantId: string, inventoryId: string, deviceId: string, revision: number, signal?: AbortSignal) {
  return mapDevice(await safeNotificationRequest(() => this.client.notifications.revokeDevice(tenantId,inventoryId,deviceId,revision,signal),signal));
 }
}
function mapDevice(value: WireDevice): NotificationDevice {
 if(value.transport !== 'apns' && value.transport !== 'fcm') throw new NotificationFailure('unavailable');
 return { id:value.id, installationId:value.installationId, transport:value.transport, revision:value.revision, active:value.active };
}
