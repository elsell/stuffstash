import { describe, expect, it } from 'vitest';
import { navigateAfterTransientDismissal } from './TransientNavigation';

describe('transient navigation', () => {
  it('dismisses the transient surface before opening the destination', () => {
    const events: string[] = [];

    navigateAfterTransientDismissal(
      () => events.push('dismiss'),
      () => events.push('navigate')
    );

    expect(events).toEqual(['dismiss', 'navigate']);
  });
});
