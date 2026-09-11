export type NotificationEvent = {
  readonly operation: 'read-state' | 'list' | 'count' | 'open' | 'mark-all' | 'preferences' | 'push-setup' | 'push-cleanup' | 'push-reconcile' | 'push-open';
  readonly outcome: 'succeeded' | 'failed';
};
export interface NotificationObservability { record(event: NotificationEvent): void }
