import type { NotificationRepository } from './notificationRepository';
export const notificationWorkspaceContext = Symbol('notificationWorkspace');
export interface NotificationWorkspace {
  apiIdentity: string;
  repository: NotificationRepository;
}
