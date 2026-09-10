export type NotificationEvent = {
  readonly operation: 'list' | 'count' | 'open' | 'mark-all' | 'preferences';
  readonly outcome: 'succeeded' | 'failed';
};
export interface NotificationObservability { record(event: NotificationEvent): void }
