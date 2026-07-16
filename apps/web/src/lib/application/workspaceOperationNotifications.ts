import type { WorkspaceNotification, WorkspaceNotificationAction } from '$lib/components/ui/sonner';

export function operationRefreshWarning(
  operationId: string,
  appliedTitle: string,
  inverseAction: WorkspaceNotificationAction
): WorkspaceNotification {
  return {
    id: `asset-operation-refresh:${operationId}`,
    kind: 'warning',
    title: 'Change applied, but this view could not be refreshed.',
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
  return 'The saved operation is no longer available or can’t be applied safely.';
}
