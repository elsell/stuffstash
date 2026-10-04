import { expect, it } from 'vitest';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { PrinterDefaultsRoute } from './PrinterDefaultsRoute';
import type { SettingsScopeRepository } from '../../application/settings/SettingsQuery';

class SelectedScope implements SettingsScopeRepository {
  denied = false;
  permission = true;
  async getSelectedScope() {
    if (this.denied) throw new Error('Forbidden');
    return { tenant: { id: 'tenant', name: 'Home', permissions: [] }, inventory: { id: 'inventory', name: 'House', permissions: this.permission ? ['configure'] : [] } };
  }
}

it('hides the real route editor during scope reauthorization and restores only its authorized scoped draft', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const query = new SelectedScope(); const workspace = fake.workspace();
  try {
    await h.render(<PrinterDefaultsRoute workspace={workspace} query={query} />);
    await h.run(() => h.byLabel('Show reference')?.props.onValueChange(false));
    await h.run(() => setScreenFocused(false));
    expect(h.byLabel('Show reference')).toBeUndefined();
    fake.settings = { ...fake.settings, revision: 2 };
    await h.run(() => setScreenFocused(true)); await h.settle();
    expect(h.byLabel('Show reference')?.props.value).toBe(false);
    await h.press(h.byLabel('Save')); await h.settle();
    expect(fake.settings.revision).toBe(2);
    expect(fake.settings.template.showReference).toBe(true);
    expect(h.byLabel('Show reference')?.props.value).toBe(false);
    query.denied = true;
    await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); await h.settle();
    expect(h.byLabel('Show reference')).toBeUndefined();
    expect(h.byLabel('Save')).toBeUndefined();
    query.denied = false; query.permission = false;
    await h.press(h.byLabel('Try again')); await h.settle();
    expect(h.byLabel('Show reference')?.props.value).toBe(true);
    expect(h.byLabel('Save')).toBeUndefined();
  } finally { await h.unmount(); setScreenFocused(true); }
});
