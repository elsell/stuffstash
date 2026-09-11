export type NotificationEvent = {
  readonly operation: 'list' | 'count' | 'open' | 'mark-all' | 'preferences' | 'push-setup' | 'push-cleanup';
  readonly outcome: 'succeeded' | 'failed';
};
export interface NotificationObservability { record(event: NotificationEvent): void }
