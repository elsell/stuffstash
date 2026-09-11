import type { PushDevicePort } from '../../application/notifications/PushDevicePort';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';

type NativeNotifications = {
  setNotificationChannelAsync(id: string, options: { name: string; importance: number }): Promise<unknown>;
  getPermissionsAsync(): Promise<{ granted: boolean }>;
  requestPermissionsAsync(): Promise<{ granted: boolean }>;
  getDevicePushTokenAsync(): Promise<{ type: string; data: unknown }>;
};

export class ExpoPushDevice implements PushDevicePort {
  private tokenGeneration = 0;
  private cachedToken: Awaited<ReturnType<PushDevicePort['nativeToken']>> | undefined;
  constructor(
    private readonly notifications: NativeNotifications,
    private readonly platform: string,
    private readonly channelImportance: number
  ) {}

  async permissionGranted(): Promise<boolean> {
    try { this.assertSupported(); return (await this.notifications.getPermissionsAsync()).granted; }
    catch { throw new NotificationFailure('unavailable'); }
  }

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
      if(this.cachedToken) return this.cachedToken;
      const generation=this.tokenGeneration;
      const result = await this.notifications.getDevicePushTokenAsync();
      if(generation!==this.tokenGeneration) {
        if(this.cachedToken) return this.cachedToken;
        throw new NotificationFailure('unavailable');
      }
      if(!this.acceptNativeToken(result)) throw new NotificationFailure('unavailable');
      return this.cachedToken!;
    } catch {
      throw new NotificationFailure('unavailable');
    }
  }

  invalidateNativeToken(): void { this.tokenGeneration++; this.cachedToken=undefined; }

  acceptNativeToken(result: {type:string;data:unknown}): boolean {
    this.tokenGeneration++;
    if((this.platform!=='ios' && this.platform!=='android') || result.type!==this.platform || typeof result.data!=='string' || !/^[\x21-\x7e]{1,4096}$/.test(result.data)) {
      this.cachedToken=undefined; return false;
    }
    this.cachedToken={transport:this.platform==='ios'?'apns':'fcm',token:result.data};
    return true;
  }

  private assertSupported(): void {
    if (this.platform !== 'ios' && this.platform !== 'android') {
      throw new NotificationFailure('unavailable');
    }
  }
}
