export type NotificationFailureKind = 'authentication-required' | 'permission-denied' | 'not-found' | 'conflict' | 'invalid' | 'unavailable';
const messages: Record<NotificationFailureKind, string> = {
  'authentication-required': 'Sign in again to access your notifications.',
  'permission-denied': 'You no longer have access to notifications for this inventory.',
  'not-found': 'This notification or setting is no longer available.',
  'conflict': 'Your settings changed on another device. Refresh before saving again.',
  'invalid': 'Review your notification settings and try again.',
  'unavailable': 'Notifications could not be updated. Try again.'
};
export class NotificationFailure extends Error {
  constructor(readonly kind: NotificationFailureKind) {
    super(messages[kind]); this.name = 'NotificationFailure';
  }
}
