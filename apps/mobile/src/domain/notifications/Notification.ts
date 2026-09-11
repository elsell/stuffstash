import type { AssetExpiration } from '../assets/AssetSummary';

export interface ExpirationReminderPolicy {
  enabled: boolean;
  upcoming: boolean;
  expired: boolean;
  advanceDays: number;
}
export interface NotificationPreferences {
  revision: number;
  defaults: ExpirationReminderPolicy;
  timezone: string;
  pushEnabled: boolean;
  overrides: { customAssetTypeId: string; settings: ExpirationReminderPolicy }[];
}
export type NotificationPreferencesUpdate = Omit<NotificationPreferences, 'overrides'>;
export interface ExpirationNotification {
  parentTrail?: readonly {assetId:string;title:string;kind:'item'|'container'|'location'}[];
  parentTrailIncomplete?: boolean;
  id: string;
  assetId: string;
  title: string;
  parentAssetId: string;
  customAssetTypeId: string;
  expiration: AssetExpiration;
  milestone: 'upcoming' | 'expired';
  createdAt: string;
  readAt?: string;
}
