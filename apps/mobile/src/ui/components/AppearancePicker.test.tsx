import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { AppearancePreferenceController, type AppearancePreference } from '../../application/settings/AppearancePreference';
import { AppearanceProvider } from '../theme/AppearanceContext';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { AppearancePicker } from './AppearancePicker';

it.each(['departed', 'returned', 'superseded', 'current'] as const)('owns appearance failure feedback for the %s selection', async mode => {
  setScreenFocused(true);
  const h = new MobileRenderHarness();
  let reject!: (error: Error) => void;
  const saved: AppearancePreference[] = [];
  const controller = new AppearancePreferenceController({ load: async () => 'system', save: async value => {
    saved.push(value);
    if (saved.length === 1) await new Promise<void>((_resolve, fail) => { reject = fail; });
  } });
  try {
    await h.render(<AppearanceProvider controller={controller}><AppFeedbackProvider><AppearancePicker /></AppFeedbackProvider></AppearanceProvider>);
    await h.settle();
    await h.press(h.byLabel('Choose appearance')); await h.press(h.byLabel('Dark'));
    await h.settle();
    if (mode === 'departed' || mode === 'returned') await h.run(() => setScreenFocused(false));
    if (mode === 'returned') await h.run(() => setScreenFocused(true));
    if (mode === 'superseded') {
      await h.press(h.byLabel('Choose appearance')); await h.press(h.byLabel('Light'));
    }
    await h.run(() => reject(new Error('Save unavailable'))); await h.settle();
    expect(h.allText().includes('Appearance not saved')).toBe(mode === 'current');
    expect(saved).toEqual(mode === 'superseded' ? ['dark', 'light'] : ['dark']);
  } finally { await h.unmount(); setScreenFocused(true); }
});
