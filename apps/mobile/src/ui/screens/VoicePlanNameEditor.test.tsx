import { useState } from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { VoicePlanNameEditor } from './VoicePlanNameEditor';

it.each(['save', 'done', 'cancel', 'blank'] as const)('edits a proposal name through %s', async action => {
  const h = new MobileRenderHarness(); const saved: string[] = []; let cancelled = 0;
  function Surface() {
    const [value, setValue] = useState('Old name');
    return <VoicePlanNameEditor value={value} onChange={setValue} onSave={name => saved.push(name)} onCancel={() => { cancelled++; }} />;
  }
  try {
    await h.render(<Surface />);
    await h.run(() => h.byLabel('Proposed item name')?.props.onChangeText(action === 'blank' ? '   ' : '  New   name  '));
    if (action === 'done' || action === 'blank') await h.run(() => h.byLabel('Proposed item name')?.props.onSubmitEditing());
    if (action === 'save' || action === 'blank') await h.press(h.byLabel('Save'));
    if (action === 'blank') expect(h.byLabel('Save')?.props.accessibilityState.disabled).toBe(true);
    if (action === 'cancel' || action === 'blank') await h.press(h.byLabel('Cancel'));
    expect(saved).toEqual(action === 'save' || action === 'done' ? ['New name'] : []);
    expect(cancelled).toBe(action === 'cancel' || action === 'blank' ? 1 : 0);
  } finally { await h.unmount(); }
});
