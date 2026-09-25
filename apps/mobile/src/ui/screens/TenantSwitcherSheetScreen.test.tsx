import { CreateWorkspace } from '../../application/inventories/CreateWorkspace';
import { expect, it } from 'vitest';
import { HomeDashboardQuery, type HomeDashboardViewModel } from '../../application/home/HomeDashboardQuery';
import { SelectInventoryCommand } from '../../application/home/SelectInventoryCommand';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { dispatchedActions, resetNavigation, setCanGoBack, setScreenFocused } from '../../test-support/navigation';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { TenantSwitcherSheetScreen } from './TenantSwitcherSheetScreen';

const dashboard: HomeDashboardViewModel = {
  tenantId: 'second', tenantName: 'Home', inventoryId: 'selected', inventoryName: 'Main', canAdd: true, canReturn: true, recentAssets: [], checkedOutAssets: [],
  tenants: [{ id: 'first', name: 'Home' }, { id: 'second', name: 'Home' }],
  inventories: [{ id: 'wrong', tenantId: 'first', tenantName: 'Home', name: 'Other household inventory', roleLabel: 'Owner', updatedAtLabel: 'Today' },
    { id: 'selected', tenantId: 'second', tenantName: 'Home', name: 'Main', roleLabel: 'Owner', updatedAtLabel: 'Today' }]
};

it('opens the current household by identity when names collide, with scroll and explicit close', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(fixture(new SelectInventoryCommand({ async selectInventory() {} })));
    expect(h.allText()).toContain('Main');
    expect(h.allText()).not.toContain('Other household inventory');
    expect(h.byType('ScrollView')).toBeDefined();
    expect(h.byLabel('Close inventory switcher')).toBeDefined();
  } finally { await h.unmount(); }
});

it('retires retained selection callbacks across departure and return', async () => {
  resetNavigation(); setScreenFocused(true);
  const h = new MobileRenderHarness(); let calls = 0;
  try {
    await h.render(fixture(new SelectInventoryCommand({ async selectInventory() { calls++; } })));
    const retained = h.byLabel('Switch to inventory Main')!.props.onPress as () => Promise<void>;
    await h.run(() => setScreenFocused(false));
    await h.run(retained);
    expect(calls).toBe(0);
    await h.run(() => setScreenFocused(true));
    await h.run(retained);
    expect(calls).toBe(0);
    expect(dispatchedActions()).toEqual([]);
    await h.press(h.byLabel('Switch to inventory Main'));
    expect(calls).toBe(1);
    expect(dispatchedActions()).toEqual([{ type: 'back' }]);
  } finally { await h.unmount(); setScreenFocused(true); resetNavigation(); }
});

it('retires selection immediately when Close is pressed', async () => {
  resetNavigation(); setScreenFocused(true);
  const h = new MobileRenderHarness(); let calls = 0;
  try {
    await h.render(fixture(new SelectInventoryCommand({ async selectInventory() { calls++; } })));
    const retained = h.byLabel('Switch to inventory Main')!.props.onPress as () => Promise<void>;
    await h.press(h.byLabel('Close inventory switcher'));
    await h.run(retained);
    await h.press(h.byLabel('Close inventory switcher'));
    expect(calls).toBe(0);
    expect(dispatchedActions()).toEqual([{ type: 'back' }]);
  } finally { await h.unmount(); setScreenFocused(true); resetNavigation(); }
});

it('retains the switcher with safe retry feedback when selecting an inventory fails', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(fixture(new SelectInventoryCommand({ async selectInventory() { throw new Error('private transport detail'); } })));
    await h.press(h.byLabel('Switch to inventory Main'));
    expect(h.allText()).toContain('Could not switch inventories. Try again.');
    expect(h.allText()).not.toContain('private transport detail');
    expect(h.byLabel('Switch to inventory Main')?.props.disabled).toBe(false);
  } finally { await h.unmount(); }
});

function fixture(command: SelectInventoryCommand, snapshot = dashboard, creation?: CreateWorkspace, scopeId = 'session') {
  const client = createMobileQueryClient();
  const scope = { tenantId: 'second', inventoryId: 'selected' };
  client.setQueryData(mobileQueryKeys.inventoryScope(scopeId), scope);
  client.setQueryData(mobileQueryKeys.home(scopeId, 'second', 'selected'), snapshot);
  return <MobileServerStateProvider client={client} scopeId={scopeId} loadInventoryScope={async () => scope}>
    <TenantSwitcherSheetScreen createWorkspace={creation} dashboardQuery={new HomeDashboardQuery({ async getHomeDashboardSnapshot() { throw new Error('Fresh cache must be used'); } })} selectInventoryCommand={command} />
  </MobileServerStateProvider>;
}

it('ignores duplicate switches and never navigates back after the sheet was dismissed', async () => {
  const h = new MobileRenderHarness(); let calls = 0; let finish!: () => void;
  resetNavigation();
  const command = new SelectInventoryCommand({ async selectInventory() { calls++; await new Promise<void>(resolve => { finish = resolve; }); } });
  try {
    await h.render(fixture(command));
    await h.run(() => {
      const row = h.byLabel('Switch to inventory Main');
      void row?.props.onPress(); void row?.props.onPress();
    });
    expect(calls).toBe(1);
    expect(h.byLabel('Switch to inventory Main')?.props.disabled).toBe(true);
    await h.unmount();
    await h.run(() => finish()); await h.settle();
    expect(dispatchedActions()).toEqual([]);
  } finally { await h.unmount(); resetNavigation(); }
});

it.each([true, false])('recovers after a departed selection with refocus before completion: %s', async returnBeforeCompletion => {
  resetNavigation(); const h = new MobileRenderHarness(); let finish!: () => void; let calls = 0;
  const command = new SelectInventoryCommand({ async selectInventory() {
    calls++; if (calls === 1) await new Promise<void>(resolve => { finish = resolve; });
  } });
  try {
    await h.render(fixture(command));
    await h.run(() => { void h.byLabel('Switch to inventory Main')?.props.onPress(); });
    await h.run(() => setScreenFocused(false));
    if (returnBeforeCompletion) await h.run(() => setScreenFocused(true));
    await h.run(() => finish()); await h.settle();
    if (!returnBeforeCompletion) await h.run(() => setScreenFocused(true));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byLabel('Switch to inventory Main')?.props.disabled).toBe(false);
    await h.press(h.byLabel('Switch to inventory Main'));
    expect(calls).toBe(2);
    expect(dispatchedActions()).toHaveLength(1);
  } finally { await h.unmount(); resetNavigation(); }
});


it.each(['close', 'select'])('returns Home after %s from a root inventory switcher', async action => {
  resetNavigation(); setCanGoBack(false); setScreenFocused(true);
  const h = new MobileRenderHarness();
  try {
    await h.render(fixture(new SelectInventoryCommand({ async selectInventory() {} })));
    await h.press(h.byLabel(action === 'close' ? 'Close inventory switcher' : 'Switch to inventory Main'));
    expect(dispatchedActions()).toEqual([{ type: 'replace', href: '/' }]);
  } finally { await h.unmount(); setCanGoBack(true); resetNavigation(); }
});


it.each([0, 1, 2])('shows the household inventory count with correct wording: %s', async count => {
  const h = new MobileRenderHarness();
  const snapshot = { ...dashboard, inventories: [
    dashboard.inventories[1]!,
    ...Array.from({ length: count }, (_, index) => ({ ...dashboard.inventories[0]!, id: `other-${index}` }))
  ] };
  try {
    await h.render(fixture(new SelectInventoryCommand({ async selectInventory() {} }), snapshot));
    await h.press(h.byLabel('Switch household'));
    expect(h.allText().join('')).toContain(`${count} ${count === 1 ? 'inventory' : 'inventories'}`);
    expect(h.allText().join('')).not.toContain('1 inventories');
  } finally { await h.unmount(); }
});


it('creates a household and inventory without changing the current inventory', async () => {
  const h = new MobileRenderHarness(); const created: string[] = []; let switched = 0;
  const creation = new CreateWorkspace({
    async canCreateInventory(id) { return id === 'new-household'; },
    async createHousehold(name) { created.push(name); return { id: 'new-household', name, canCreateInventory: true }; },
    async createInventory(tenantId, name) { created.push(name); return { id: 'new-inventory', tenantId, name }; }
  }, { created() {} });
  try {
    await h.render(fixture(new SelectInventoryCommand({ async selectInventory() { switched++; } }), dashboard, creation));
    expect(h.byLabel('New inventory')).toBeUndefined();
    await h.press(h.byLabel('Switch household'));
    await h.press(h.byLabel('New household'));
    await h.changeText(h.byLabel('Household name'), 'Lake house');
    await h.press(h.byLabel('Create household'));
    expect(h.byLabel('New inventory')).toBeDefined();
    await h.press(h.byLabel('New inventory'));
    await h.changeText(h.byLabel('Inventory name'), 'Garage');
    await h.press(h.byLabel('Create inventory'));
    expect(created).toEqual(['Lake house', 'Garage']);
    expect(h.byLabel('Switch to inventory Garage')).toBeDefined();
    expect(switched).toBe(0);
    await h.press(h.byLabel('Switch to inventory Garage'));
    expect(switched).toBe(1);
  } finally { await h.unmount(); }
});


it.each(['blur', 'scope'] as const)('retires creation callbacks and deferred results after %s', async departure => {
  const h = new MobileRenderHarness(); resetNavigation(); setScreenFocused(true);
  let resolve!: (value: { id: string; name: string; canCreateInventory: boolean }) => void;
  const creation = new CreateWorkspace({
    async canCreateInventory() { return true; },
    createHousehold: () => new Promise(done => { resolve = done; }),
    async createInventory() { throw new Error('Not requested'); }
  }, { created() {} });
  const selection = new SelectInventoryCommand({ async selectInventory() {} });
  try {
    await h.render(fixture(selection, dashboard, creation));
    await h.press(h.byLabel('Switch household'));
    const open = h.byLabel('New household')!.props.onPress;
    await h.press(h.byLabel('New household'));
    await h.changeText(h.byLabel('Household name'), 'Retired');
    await h.press(h.byLabel('Create household'));
    if (departure === 'blur') await h.run(() => setScreenFocused(false));
    else await h.render(fixture(selection, dashboard, creation, 'new-session'));
    await h.run(() => resolve({ id: 'retired', name: 'Retired', canCreateInventory: true }));
    await h.run(open);
    expect(h.byLabel('Household name')).toBeUndefined();
    expect(h.allText()).not.toContain('Retired');
    expect(dispatchedActions()).toEqual([]);
  } finally { await h.unmount(); resetNavigation(); setScreenFocused(true); }
});
