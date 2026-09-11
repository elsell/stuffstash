export type NotificationEvent = {
  readonly operation: 'list' | 'count' | 'open' | 'mark-all' | 'preferences' | 'push-setup' | 'push-cleanup' | 'push-reconcile';
  readonly outcome: 'succeeded' | 'failed';
};
export interface NotificationObservability { record(event: NotificationEvent): void }
