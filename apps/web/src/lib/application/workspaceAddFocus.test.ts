import { afterEach, describe, expect, it } from 'vitest';
import { addReturnFocusTarget } from './workspaceAddFocus';

describe('Add return focus', () => {
  afterEach(() => { document.body.innerHTML = ''; });

  it('keeps a connected opener and falls back to the durable responsive Add trigger', () => {
    const desktop = document.createElement('button');
    desktop.dataset.workspaceAddTrigger = 'desktop';
    const mobile = document.createElement('button');
    mobile.dataset.workspaceAddTrigger = 'mobile';
    document.body.append(desktop, mobile);
    const transient = document.createElement('a');
    const replacedLocalOpener = document.createElement('a');
    replacedLocalOpener.dataset.workspaceAddReturnFocus = 'location-item';
    const connectedLocalOpener = document.createElement('a');
    connectedLocalOpener.dataset.workspaceAddReturnFocus = 'location-item';
    document.body.append(connectedLocalOpener);

    expect(addReturnFocusTarget(desktop, document, false)).toBe(desktop);
    expect(addReturnFocusTarget(desktop, document, true)).toBe(mobile);
    expect(addReturnFocusTarget(transient, document, false)).toBe(desktop);
    expect(addReturnFocusTarget(transient, document, true)).toBe(mobile);
    expect(addReturnFocusTarget(document.body, document, false)).toBe(desktop);
    expect(addReturnFocusTarget(replacedLocalOpener, document, false)).toBe(connectedLocalOpener);
  });

  it('uses the mobile trigger through the shell breakpoint', () => {
    const originalMatchMedia = window.matchMedia;
    Object.defineProperty(window, 'matchMedia', { configurable: true, value: () => ({ matches: true }) });
    const mobile = document.createElement('button');
    mobile.dataset.workspaceAddTrigger = 'mobile';
    document.body.append(mobile);
    try {
      expect(addReturnFocusTarget(document.createElement('a'))).toBe(mobile);
    } finally {
      Object.defineProperty(window, 'matchMedia', { configurable: true, value: originalMatchMedia });
    }
  });

  it('prefers the focused result after a successful save', () => {
    const opener = document.createElement('button');
    opener.dataset.workspaceAddTrigger = 'mobile';
    const result = document.createElement('h1');
    result.dataset.workspaceAddResultFocus = '';
    document.body.append(opener, result);

    expect(addReturnFocusTarget(opener, document, true, true)).toBe(result);
  });
});
