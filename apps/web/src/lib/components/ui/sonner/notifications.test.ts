import { beforeEach, describe, expect, it, vi } from 'vitest';

type NotificationOptions = {
  action?: { onClick?: (event: MouseEvent) => void | Promise<void> };
};

type SuccessNotification = (title: string, options?: NotificationOptions) => number;

const { goto, success } = vi.hoisted(() => ({
  goto: vi.fn(),
  success: vi.fn<SuccessNotification>(() => 1)
}));

vi.mock('$app/navigation', () => ({ goto }));
vi.mock('svelte-sonner', () => ({
  toast: {
    success,
    warning: vi.fn(() => 2),
    error: vi.fn(() => 3),
    info: vi.fn(() => 4)
  }
}));

import { notify } from './notifications';

describe('workspace notifications', () => {
  beforeEach(() => {
    goto.mockClear();
    success.mockClear();
  });

  it('invokes an operation action without navigating away from the current workspace', async () => {
    const onClick = vi.fn();
    notify({ kind: 'success', title: 'Saved drill.', action: { label: 'Undo', onClick } });

    const options = success.mock.calls[0]?.[1];
    await options?.action?.onClick?.({} as MouseEvent);

    expect(onClick).toHaveBeenCalledOnce();
    expect(goto).not.toHaveBeenCalled();
  });

  it('keeps route actions as navigation links', async () => {
    notify({ kind: 'success', title: 'Saved drill.', action: { label: 'View', href: '/assets/drill' } });

    const options = success.mock.calls[0]?.[1];
    await options?.action?.onClick?.({} as MouseEvent);

    expect(goto).toHaveBeenCalledWith('/assets/drill');
  });
});
