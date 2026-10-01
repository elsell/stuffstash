import { t } from '$lib/presentation/localization';
import type { WorkspaceNotification, WorkspaceNotificationAction } from '$lib/components/ui/sonner';

export function operationRefreshWarning(
  operationId: string,
  appliedTitle: string,
  inverseAction: WorkspaceNotificationAction
): WorkspaceNotification {
  return {
    id: `asset-operation-refresh:${operationId}`,
    kind: 'warning',
    title: t('web.workspaceOperationNotifications.changeAppliedButThisViewCouldNotBeRefreshed'),
    description: `${appliedTitle} Reload to see the latest inventory.`,
    important: true,
    duration: Infinity,
    action: inverseAction
  };
}

export function safeOperationFailureDescription(caught: unknown): string {
  const safeForUser = typeof caught === 'object' && caught !== null &&
    (caught as { safeForUser?: unknown }).safeForUser === true;
  if (safeForUser && caught instanceof Error && caught.message.trim()) return caught.message.trim();
  return t('web.workspaceOperationNotifications.theSavedOperationIsNoLongerAvailableOrCan');
}
