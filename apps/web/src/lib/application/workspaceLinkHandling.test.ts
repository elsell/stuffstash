import { describe, expect, it } from 'vitest';
import { shouldHandleWorkspaceLinkClick } from './workspaceLinkHandling';

describe('workspace link handling', () => {
  it('handles ordinary primary clicks in app', () => {
    expect(shouldHandleWorkspaceLinkClick(click())).toBe(true);
  });

  it('preserves modified and non-primary browser navigation', () => {
    expect(shouldHandleWorkspaceLinkClick(click({ metaKey: true }))).toBe(false);
    expect(shouldHandleWorkspaceLinkClick(click({ ctrlKey: true }))).toBe(false);
    expect(shouldHandleWorkspaceLinkClick(click({ shiftKey: true }))).toBe(false);
    expect(shouldHandleWorkspaceLinkClick(click({ altKey: true }))).toBe(false);
    expect(shouldHandleWorkspaceLinkClick(click({ button: 1 }))).toBe(false);
  });

  it('does not intercept an event another listener already handled', () => {
    const event = click();
    event.preventDefault();

    expect(shouldHandleWorkspaceLinkClick(event)).toBe(false);
  });
});

function click(init: MouseEventInit = {}): MouseEvent {
  return new MouseEvent('click', { bubbles: true, cancelable: true, ...init });
}
