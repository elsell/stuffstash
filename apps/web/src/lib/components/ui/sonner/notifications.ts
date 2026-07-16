import { goto } from '$app/navigation';
import { toast, type ExternalToast } from 'svelte-sonner';

export type WorkspaceNotificationKind = 'success' | 'info' | 'warning' | 'error';

export interface WorkspaceNotificationAction {
  label: string;
  href?: string;
  onClick?: () => void | Promise<void>;
}

export interface WorkspaceNotification {
  kind: WorkspaceNotificationKind;
  title: string;
  description?: string;
  important?: boolean;
  duration?: number;
  id?: string | number;
  action?: WorkspaceNotificationAction;
}

export function notify(notification: WorkspaceNotification): string | number {
  const options: ExternalToast = {
    description: notification.description,
    important: notification.important ?? notification.kind === 'error',
    duration: notification.duration,
    id: notification.id,
    action: notification.action
      ? {
          label: notification.action.label,
          onClick: () => {
            if (notification.action?.onClick) {
              void notification.action.onClick();
              return;
            }
            if (notification.action?.href) {
              void goto(notification.action.href);
            }
          }
        }
      : undefined
  };

  if (notification.kind === 'success') {
    return toast.success(notification.title, options);
  }
  if (notification.kind === 'warning') {
    return toast.warning(notification.title, options);
  }
  if (notification.kind === 'error') {
    return toast.error(notification.title, options);
  }
  return toast.info(notification.title, options);
}

export function notifySuccess(
  title: string,
  options: { description?: string; action?: WorkspaceNotificationAction } = {}
): string | number {
  return notify({ kind: 'success', title, description: options.description, action: options.action });
}

export function notifyError(title: string, description?: string): string | number {
  return notify({ kind: 'error', title, description });
}
