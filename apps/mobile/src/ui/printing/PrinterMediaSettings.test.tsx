import { expect, it } from 'vitest';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { PrinterSettingsScreen } from './PrinterSettingsScreen';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };

it('preserves media selection after a stale revision and does not erase unsaved inventory defaults on successful update', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  try {
    await h.render(<PrinterSettingsScreen workspace={workspace} scope={scope} canConfigure canPrint onJob={() => {}} />);
    await h.settle();
    await h.run(() => h.byLabel('Print label when adding an item')?.props.onValueChange(true));
    fake.printer = { ...fake.printer, revision: 2, name: 'Changed elsewhere' };
    await h.press(h.byLabel('Save label size')); await h.settle();
    expect(fake.printer.revision).toBe(2);
    expect(h.byLabel('Reload printer settings')).toBeDefined();
    await h.press(h.byLabel('Reload printer settings')); await h.settle();
    await h.run(() => h.byLabel('Print label when adding an item')?.props.onValueChange(true));
    await h.press(h.byLabel('Save label size')); await h.settle();
    expect(fake.printer.revision).toBe(3); expect(fake.printer.name).toBe('Changed elsewhere');
    expect(h.byLabel('Print label when adding an item')?.props.value).toBe(true);
    expect(fake.settings.printOnCreateDefault).toBe(false);
    await h.run(() => h.byLabel('Show reference')?.props.onValueChange(false));
    await h.press(h.byLabel('Save defaults')); await h.settle();
    await h.run(() => h.byLabel('Show reference')?.props.onValueChange(true));
    await h.press(h.byLabel('Print test label')); await h.settle();
    expect(fake.queuedSelections.get('request')?.template.showReference).toBe(false);

    await h.render(<></>); fake.printer = { ...fake.printer, mediaName: '' };
    await h.render(<PrinterSettingsScreen workspace={workspace} scope={scope} canConfigure={false} onJob={() => {}} />);
    expect(h.byLabel('Save label size')).toBeUndefined();
    expect(h.allText()).toContain('29 × 89.8 mm');
  } finally { await h.unmount(); }
});
