import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { ExpirationReminderEditor } from './ExpirationReminderEditor';
import { ReminderTimingEditor } from './ReminderTimingEditor';
import { TimeZonePicker } from './TimeZonePicker';

const policy = { enabled: true, upcoming: true, expired: true, advanceDays: 30 };
it.each((['mode', 'timing', 'timezone'] as const).flatMap(kind => (['success', 'failure'] as const).map(outcome => [kind, outcome] as const)))('keeps departed %s completion out of the returning visit: %s', async (kind, outcome) => {
  const h = new MobileRenderHarness(); let done = 0; let attempts = 0;
  let resolve!: () => void; let reject!: (error: Error) => void;
  const pending = new Promise<void>((yes, no) => { resolve = yes; reject = no; });
  const save = async () => { if (++attempts === 1) await pending; };
  async function activate() {
    if (kind === 'mode') await h.run(() => h.byLabel('When expired')?.props.onValueChange(false));
    else await h.press(h.byLabel(kind === 'timing' ? '7 days before' : 'UTC'));
  }
  try {
    await h.render(kind === 'mode' ? <ExpirationReminderEditor initialPolicy={policy} onSave={save} onEditDays={() => {}} />
      : kind === 'timing' ? <ReminderTimingEditor policy={policy} onSave={save} onDone={() => { done++; }} />
      : <TimeZonePicker value="America/New_York" onChange={save} />);
    await activate(); expect(attempts).toBe(1);
    await h.run(() => setScreenFocused(false));
    await h.run(() => setScreenFocused(true));
    await activate(); expect(attempts).toBe(1);
    await h.run(() => outcome === 'success' ? resolve() : reject(new Error('Late failure')));
    expect(h.allText().join(' ')).not.toContain('Could not save');
    expect(done).toBe(0);
    if (kind === 'mode') expect(h.byLabel('When expired')?.props.value).toBe(true);
    if (kind === 'timing') expect(h.byLabel('30 days before')?.props.accessibilityState.checked).toBe(true);
    await activate(); expect(attempts).toBe(2);
    expect(done).toBe(kind === 'timing' ? 1 : 0);
  } finally { await h.unmount(); setScreenFocused(true); }
});
