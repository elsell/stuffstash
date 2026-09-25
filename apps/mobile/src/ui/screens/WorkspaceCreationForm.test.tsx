import { expect, it } from 'vitest';
import { CreateWorkspace } from '../../application/inventories/CreateWorkspace';
import { MobileRenderHarness } from '../../test-support/render';
import { WorkspaceCreationForm } from './WorkspaceCreationForm';

it('locks duplicate creation, preserves failed input, and allows a deliberate retry', async () => {
  const h = new MobileRenderHarness(); let reject!: (error: Error) => void; let calls = 0; const results: unknown[] = [];
  const command = new CreateWorkspace({
    async canCreateInventory() { return true; },
    async createHousehold(name) { calls++; if (calls === 1) await new Promise((_, fail) => { reject = fail; }); return { id: 'home', name, canCreateInventory: true }; },
    async createInventory() { throw new Error('Not requested'); }
  }, { created() {} });
  try {
    await h.render(<WorkspaceCreationForm task={{ kind: 'household' }} command={command}
      onBusy={() => {}} onCancel={() => { throw new Error('Must remain while submitting'); }} onCreated={result => results.push(result)} />);
    await h.changeText(h.byLabel('Household name'), 'Family');
    const submit = h.byLabel('Create household')!.props.onPress;
    await h.run(() => { submit(); submit(); });
    expect(calls).toBe(1);
    await h.press(h.byLabel('Cancel creation'));
    await h.run(() => reject(new Error('Private transport detail')));
    expect(h.byLabel('Household name')?.props.value).toBe('Family');
    expect(h.allText().join(' ')).not.toContain('Private transport detail');
    await h.press(h.byLabel('Create household'));
    expect(calls).toBe(2); expect(results).toHaveLength(1);
  } finally { await h.unmount(); }
});
