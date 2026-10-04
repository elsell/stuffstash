import { expect, it } from 'vitest';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { PrinterDetailScreen } from './PrinterDetailScreen';
import { PrinterDefaultsScreen } from './PrinterDefaultsScreen';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };

it('saves only changed media, retains its selection on conflict, and uses committed defaults for test labels', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  fake.printer = { ...fake.printer, media: { ...fake.printer.media, presetId: 'previous-media' } };
  const detail = () => <PrinterDetailScreen workspace={workspace} scope={scope} printerId="printer" canConfigure canPrint onJob={() => {}} />;
  try {
    await h.render(detail());
    expect(h.byLabel('Save label size')).toBeUndefined();
    await h.press(h.byLabel('Label size')); await h.press(h.byLabel('29 × 90 mm'));
    fake.printer = { ...fake.printer, revision: 2, name: 'Changed elsewhere' };
    await h.press(h.byLabel('Save label size')); await h.settle();
    expect(fake.printer.revision).toBe(2);
    expect(h.byLabel('Reload printer settings')).toBeDefined();
    expect(h.byLabel('Save label size')).toBeDefined();
    await h.press(h.byLabel('Reload printer settings')); await h.settle();
    await h.press(h.byLabel('Label size')); await h.press(h.byLabel('29 × 90 mm'));
    await h.press(h.byLabel('Save label size')); await h.settle();
    expect(fake.printer.revision).toBe(3);
    expect(fake.printer.name).toBe('Changed elsewhere');
    expect(h.byLabel('Save label size')).toBeUndefined();

    await h.render(<PrinterDefaultsScreen workspace={workspace} scope={scope} canConfigure />);
    await h.run(() => h.byLabel('Show reference')?.props.onValueChange(false));
    await h.press(h.byLabel('Save')); await h.settle();
    await h.render(detail());
    await h.press(h.byLabel('Print test label')); await h.settle();
    expect(fake.queuedSelections.get('request')?.template.showReference).toBe(false);

    await h.render(<></>); fake.printer = { ...fake.printer, mediaName: '' };
    await h.render(<PrinterDetailScreen workspace={workspace} scope={scope} printerId="printer" canConfigure={false} canPrint={false} onJob={() => {}} />);
    expect(h.byLabel('Save label size')).toBeUndefined();
    expect(h.byLabel('Print test label')).toBeUndefined();
    expect(h.allText()).toContain('29 × 89.8 mm');
  } finally { await h.unmount(); }
});
