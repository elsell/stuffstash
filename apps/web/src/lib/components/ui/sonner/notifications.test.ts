import { beforeEach, describe, expect, it, vi } from 'vitest';

type Options = { action?: { onClick?: (event: MouseEvent) => void | Promise<void> } };
const { goto, success } = vi.hoisted(() => ({
  goto: vi.fn(),
  success: vi.fn((_title: string, _options?: Options) => 1)
}));

vi.mock('$app/navigation', () => ({ goto }));
vi.mock('svelte-sonner', () => ({
  toast: { success, warning: vi.fn(), error: vi.fn(), info: vi.fn() }
}));

import { notify } from './notifications';

describe('workspace notifications', () => {
  beforeEach(() => {
    goto.mockClear();
    success.mockClear();
  });

  it('invokes an operation action without navigating', async () => {
    const onClick = vi.fn();
    notify({ kind: 'success', title: 'Saved drill.', action: { label: 'Undo', onClick } });

    await success.mock.calls[0]?.[1]?.action?.onClick?.({} as MouseEvent);

    expect(onClick).toHaveBeenCalledOnce();
    expect(goto).not.toHaveBeenCalled();
  });

  it('keeps route actions as navigation links', async () => {
    notify({ kind: 'success', title: 'Saved drill.', action: { label: 'View', href: '/assets/drill' } });

    await success.mock.calls[0]?.[1]?.action?.onClick?.({} as MouseEvent);

    expect(goto).toHaveBeenCalledWith('/assets/drill');
  });
});
