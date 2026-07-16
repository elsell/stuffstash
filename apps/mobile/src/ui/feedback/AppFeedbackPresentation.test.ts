import { describe, expect, it } from 'vitest';
import { buildAppNoticePresentation } from './AppFeedbackPresentation';
import { darkPalette, lightPalette } from '../theme/tokens';

describe('buildAppNoticePresentation', () => {
  it('uses concise accessible text and a short duration by default', () => {
    expect(buildAppNoticePresentation({
      tone: 'info',
      title: 'Inventory refreshed',
      message: 'Showing the latest items.'
    }, lightPalette)).toMatchObject({
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
    }, lightPalette).durationMs).toBe(6500);
  });

  it('uses semantic dark notice surfaces instead of light-only literals', () => {
    expect(buildAppNoticePresentation({
      tone: 'error',
      title: 'Move failed'
    }, darkPalette)).toMatchObject({
      backgroundColor: darkPalette.dangerSurface,
      borderColor: darkPalette.dangerBorder,
      textColor: darkPalette.text
    });
  });
});
