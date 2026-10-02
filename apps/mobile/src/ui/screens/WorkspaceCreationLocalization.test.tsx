import { expect, it } from 'vitest';

it('localizes creation failure as a complete sentence while preserving the draft', async () => {
  const previous = process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
  process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = 'en-XA';
  try {
    const { MobileRenderHarness } = await import('../../test-support/render');
    const { CreateWorkspace } = await import('../../application/inventories/CreateWorkspace');
    const { WorkspaceCreationForm } = await import('./WorkspaceCreationForm');
    const { t } = await import('../../presentation/localization');
    const command = new CreateWorkspace({
      async canCreateInventory() { return true; },
      async createHousehold() { throw new Error('Private transport diagnostic'); },
      async createInventory() { throw new Error('Private transport diagnostic'); }
    }, { created() {} });
    for (const task of [{ kind: 'household' } as const, { kind: 'inventory', household: { id: 'home', name: 'My household' } } as const]) {
      const h = new MobileRenderHarness();
      try {
        await h.render(<WorkspaceCreationForm task={task} command={command} onBusy={() => {}} onCancel={() => {}} onCreated={() => { throw new Error('Must not succeed'); }} />);
        const nameLabel = t(task.kind === 'household' ? 'mobile.WorkspaceCreationForm.householdName' : 'mobile.WorkspaceCreationForm.inventoryName');
        await h.changeText(h.byLabel(nameLabel), 'My untouched name');
        await h.press(h.byLabel(t(task.kind === 'household' ? 'mobile.WorkspaceCreationForm.createHousehold' : 'mobile.WorkspaceCreationForm.createInventory')));
        const text = h.allText().join(' ');
        expect(text).toContain(t(`workspace.creationFailed.${task.kind}`));
        expect(text).not.toContain('Private transport diagnostic');
        expect(text).not.toContain(task.kind);
        expect(h.byLabel(nameLabel)?.props.value).toBe('My untouched name');
      } finally { await h.unmount(); }
    }
  } finally {
    if (previous === undefined) delete process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
    else process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = previous;
  }
});
