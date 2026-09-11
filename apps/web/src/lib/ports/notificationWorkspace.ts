import type { NotificationRepository } from './notificationRepository';
export const notificationWorkspaceContext = Symbol('notificationWorkspace');
export interface NotificationWorkspace {
  apiIdentity: string;
  onPreferencesChanged?: () => Promise<void>;
  repository: NotificationRepository;
}
