import type { PushDevicePort } from '../../application/notifications/PushDevicePort';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';

type NativeNotifications = {
  setNotificationChannelAsync(id: string, options: { name: string; importance: number }): Promise<unknown>;
  getPermissionsAsync(): Promise<{ granted: boolean }>;
  requestPermissionsAsync(): Promise<{ granted: boolean }>;
  getDevicePushTokenAsync(): Promise<{ type: string; data: unknown }>;
};

export class ExpoPushDevice implements PushDevicePort {
  constructor(
    private readonly notifications: NativeNotifications,
    private readonly platform: string,
    private readonly channelImportance: number
  ) {}

  async requestPermission(): Promise<boolean> {
    try {
      this.assertSupported();
      if (this.platform === 'android') {
        await this.notifications.setNotificationChannelAsync('expiration', {
          name: 'Expiration reminders', importance: this.channelImportance
        });
      }
      const existing = await this.notifications.getPermissionsAsync();
      if (existing.granted) return true;
      return (await this.notifications.requestPermissionsAsync()).granted;
    } catch {
      throw new NotificationFailure('unavailable');
    }
  }

  async nativeToken(): ReturnType<PushDevicePort['nativeToken']> {
    try {
      this.assertSupported();
      const result = await this.notifications.getDevicePushTokenAsync();
      if (result.type !== this.platform || typeof result.data !== 'string' ||
          !/^[\x21-\x7e]{1,4096}$/.test(result.data)) {
        throw new NotificationFailure('unavailable');
      }
      return { transport: this.platform === 'ios' ? 'apns' : 'fcm', token: result.data };
    } catch {
      throw new NotificationFailure('unavailable');
    }
  }

  private assertSupported(): void {
    if (this.platform !== 'ios' && this.platform !== 'android') {
      throw new NotificationFailure('unavailable');
    }
  }
}
