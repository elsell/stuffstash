import { describe, expect, it } from 'vitest';
import { buildAppNoticePresentation } from './AppFeedbackPresentation';

describe('buildAppNoticePresentation', () => {
  it('uses concise accessible text and a short duration by default', () => {
    expect(buildAppNoticePresentation({
      tone: 'info',
      title: 'Inventory refreshed',
      message: 'Showing the latest items.'
    })).toMatchObject({
      accessibilityLabel: 'Inventory refreshed. Showing the latest items.',
      durationMs: 4200,
      title: 'Inventory refreshed',
      message: 'Showing the latest items.'
    });
  });

  it('keeps actionable notices visible longer', () => {
    expect(buildAppNoticePresentation({
      tone: 'error',
      title: 'Photo upload failed',
      actionLabel: 'Retry'
    }).durationMs).toBe(6500);
  });
});
