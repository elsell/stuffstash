import type { PushTransport } from '../../domain/notifications/NotificationDevice';
export interface PushDevicePort {
 permissionGranted(): Promise<boolean>;
 requestPermission(): Promise<boolean>;
 nativeToken(): Promise<{transport: PushTransport; token: string}>;
}
export interface PushRegistrationScope {
 readonly serverId: string;
 readonly principalId: string;
 readonly tenantId: string;
 readonly inventoryId: string;
}
export interface PushRegistrationJournal {
 installationId(): Promise<string>;
 remember(scope: PushRegistrationScope): Promise<void>;
 list(): Promise<readonly PushRegistrationScope[]>;
 forget(scope: PushRegistrationScope): Promise<void>;
}
