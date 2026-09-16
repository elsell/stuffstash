import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { VoicePlanLocationScreen } from './VoicePlanLocationScreen';
import type { VoicePlanParentDraft } from './VoicePlanEdits';

it('distinguishes loading, retry and empty lookup while preserving the selected destination', async () => {
  const h = new MobileRenderHarness(); let retries = 0; const selected: VoicePlanParentDraft[] = [];
  const common = { current: { kind: 'root', label: 'Inventory root' } as const, proposed: [], matches: [], query: '', onQuery: () => {}, onRetry: () => { retries++; }, onSelect: (parent: VoicePlanParentDraft) => selected.push(parent) };
  try {
    await h.render(<VoicePlanLocationScreen {...common} loading error={false} />);
    expect(h.allText()).toContain('Loading locations');
    expect(h.allText()).not.toContain('No matching locations');
    await h.render(<VoicePlanLocationScreen {...common} loading={false} error />);
    expect(h.allText()).toContain('Could not load locations.');
    await h.press(h.byLabel('Retry locations')); expect(retries).toBe(1);
    expect(h.allText()).not.toContain('No matching locations');
    await h.render(<VoicePlanLocationScreen {...common} loading={false} error={false} />);
    expect(h.allText()).toContain('No matching locations');
    expect(h.byLabel('Select Inventory root')?.props.accessibilityState.checked).toBe(true);
    await h.press(h.byLabel('Select Inventory root'));
    expect(selected).toEqual([common.current]);
  } finally { await h.unmount(); }
});

it('marks the existing destination and keeps disabled reasons without allowing selection', async () => {
  const h = new MobileRenderHarness(); const selected: VoicePlanParentDraft[] = [];
  const matches = ['bin', 'blocked'].map(id => ({ id, title: id, kind: 'container' as const, subtitle: 'Container', pathLabel: `Garage / ${id}`, selectionHint: '', willPromoteToContainer: false,
    canSelectAsParent: id !== 'blocked', disabledReason: id === 'blocked' ? 'This location is archived' : undefined }));
  try {
    await h.render(<VoicePlanLocationScreen current={{ kind: 'asset', id: 'bin', label: 'Garage / bin' }} proposed={[]} matches={matches}
      query="bin" loading={false} error={false} onQuery={() => {}} onRetry={() => {}} onSelect={parent => selected.push(parent)} />);
    expect(h.byLabel('Select bin, Garage / bin')?.props.accessibilityState.checked).toBe(true);
    expect(h.allText()).toContain('This location is archived');
    await h.press(h.byLabel('Select blocked, This location is archived'));
    expect(selected).toEqual([]);
    await h.press(h.byLabel('Select bin, Garage / bin'));
    expect(selected).toEqual([{ kind: 'asset', id: 'bin', label: 'Garage / bin' }]);
  } finally { await h.unmount(); }
});
