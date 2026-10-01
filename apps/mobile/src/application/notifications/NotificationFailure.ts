import { t } from '../../presentation/localization';
export type NotificationFailureKind = 'authentication-required' | 'permission-denied' | 'not-found' | 'conflict' | 'invalid' | 'unavailable' | 'wrong-recipient' | 'invalid-payload';
const messages: Record<NotificationFailureKind, string> = {
  'authentication-required': t('mobile.NotificationFailure.signInAgainToAccessYourNotifications'),
  'permission-denied': t('mobile.NotificationFailure.youNoLongerHaveAccessToNotificationsForThis'),
  'not-found': t('mobile.NotificationFailure.thisNotificationOrSettingIsNoLongerAvailable'),
  'conflict': t('mobile.NotificationFailure.yourSettingsChangedOnAnotherDeviceRefreshBeforeSaving'),
  'invalid-payload': t('mobile.NotificationFailure.thisNotificationCouldNotBeOpenedOpenYourNotifications'),
  'invalid': t('mobile.NotificationFailure.reviewYourNotificationSettingsAndTryAgain'),
  'wrong-recipient': t('mobile.NotificationFailure.thisNotificationBelongsToADifferentAccountOrServer'),
  'unavailable': t('mobile.NotificationFailure.notificationsCouldNotBeUpdatedTryAgain')
};
export class NotificationFailure extends Error {
  constructor(readonly kind: NotificationFailureKind) {
    super(messages[kind]); this.name = 'NotificationFailure';
  }
}
