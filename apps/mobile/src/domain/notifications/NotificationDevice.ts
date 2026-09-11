export type PushTransport = 'apns' | 'fcm';
export interface NotificationDevice {
 readonly id: string;
 readonly installationId: string;
 readonly transport: PushTransport;
 readonly revision: number;
 readonly active: boolean;
}
export interface RegisterNotificationDevice {
 readonly installationId: string;
 readonly transport: PushTransport;
 readonly token: string;
 readonly revision: number;
}
