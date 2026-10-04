import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { PrintingFake } from '../../test-support/PrintingFake';
import type { LabelWorkspace } from '../../application/labels/LabelWorkspace';
import { AssetLabelTask } from './AssetLabelTask';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
function fixture() {
  const fake = new PrintingFake(); const printing = fake.workspace();
  const labels: LabelWorkspace = {
    parse: () => { throw new Error('Not a scan'); },
    repository: {
      instance: async () => 'instance', resolve: async () => ({ ...scope, assetId: 'asset', archived: false }),
      catalog: async () => ({ profiles: [{ id: 'media', name: '29 × 90 mm', media: fake.printer.media }], templates: (await fake.catalog()).templates }),
      render: async () => ({ bytes: new Uint8Array([1]), format: 'png', width: 306, height: 991, rotation: 270 })
    }, files: printing.files
  };
  return { fake, printing, labels };
}
it('offers registered printing directly without submitting on label-sheet entry', async () => {
  const h = new MobileRenderHarness(); const f = fixture();
  try {
    await h.render(<AssetLabelTask {...f} scope={scope} assetId="asset" onQueued={() => {}} />);
    expect(h.byLabel('Printer')).toBeDefined();
    expect(h.byLabel('Print label')).toBeDefined();
    expect(h.byLabel('Print options')).toBeUndefined();
    expect(f.fake.submitted.size).toBe(0);
  } finally { await h.unmount(); }
});
it('keeps the export-only reader independent of printer access', async () => {
  const h = new MobileRenderHarness(); const f = fixture();
  try {
    await h.render(<AssetLabelTask labels={f.labels} scope={scope} assetId="asset" onQueued={() => {}} />);
    expect(h.byLabel('Save or share PNG')).toBeDefined();
    expect(h.byLabel('Print label')).toBeUndefined();
  } finally { await h.unmount(); }
});
it('exposes catalog failure recovery rather than claiming no printers exist', async () => {
  const h = new MobileRenderHarness(); const f = fixture(); f.fake.deny = true;
  try {
    await h.render(<AssetLabelTask {...f} scope={scope} assetId="asset" onQueued={() => {}} />);
    expect(h.byLabel('Try again')).toBeDefined();
    expect(h.byLabel('Save or share PNG')).toBeUndefined();
    f.fake.deny = false; await h.press(h.byLabel('Try again'));
    expect(h.byLabel('Print label')).toBeDefined();
  } finally { await h.unmount(); }
});
it('keeps sharing available when every registered printer is retired', async () => {
  const h = new MobileRenderHarness(); const f = fixture(); f.fake.printer = { ...f.fake.printer, retired: true };
  try {
    await h.render(<AssetLabelTask {...f} scope={scope} assetId="asset" onQueued={() => {}} />);
    expect(h.byLabel('Save or share PNG')).toBeDefined();
    expect(h.byLabel('Print label')).toBeUndefined();
    expect(f.fake.submitted.size).toBe(0);
  } finally { await h.unmount(); }
});
