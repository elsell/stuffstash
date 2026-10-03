import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { LabelOptionsScreen } from './LabelOptionsScreen';
import type { LabelFile, LabelWorkspace } from '../../application/labels/LabelWorkspace';
import { setScreenFocused } from '../../test-support/navigation';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
const file: LabelFile = { bytes: new Uint8Array([1]), format: 'png', width: 10, height: 20, rotation: 90 };
const media = { widthMicrometers: 29000, heightMicrometers: 90000, margins: { left: 1, right: 1, top: 1, bottom: 1 }, resolutionDPI: 300, rasterWidth: 306, rasterHeight: 991, orientation: 'portrait', colorMode: 'monochrome', cutPolicy: 'cut', displayRotation: 90 };
function fixture() {
  let renders = 0; let shares = 0; let releases = 0;
  const workspace: LabelWorkspace = {
    parse: () => ({ instanceId: 'instance', labelId: 'label' }),
    repository: { instance: async () => 'instance', resolve: async () => ({ ...scope, assetId: 'asset', archived: false }),
      catalog: async () => ({ profiles: [{ id: 'profile', name: '29 × 90 mm', media }], templates: [{ id: 'qr-title', version: 1, name: 'QR and title', supportsReference: true, showReference: false }] }),
      render: async () => { renders++; return file; } },
    files: { preview: async () => ({ uri: 'file:///private/label.png', release: () => { releases++; } }), deliver: async () => { shares++; } }
  };
  return { workspace, counts: () => ({ renders, shares, releases }) };
}
it('renders an authenticated preview, invalidates it when options change and reauthorizes each delivery', async () => {
  const h = new MobileRenderHarness(); const f = fixture();
  try {
    await h.render(<LabelOptionsScreen workspace={f.workspace} scope={scope} assetId="asset" />);
    await h.press(h.byLabel('Preview label')); await h.settle();
    expect(h.byLabel('Label preview')).toBeDefined();
    await h.run(() => h.byLabel('Show reference')?.props.onValueChange(true));
    expect(h.byLabel('Label preview')).toBeUndefined();
    expect(f.counts().releases).toBe(1);
    await h.press(h.byLabel('Save or share PNG')); await h.settle();
    expect(f.counts()).toEqual({ renders: 2, shares: 1, releases: 1 });
  } finally { await h.unmount(); }
});
it('does not hand off a download completed after leaving the task', async () => {
  const h = new MobileRenderHarness(); const f = fixture(); let complete!: (file: LabelFile) => void;
  f.workspace.repository.render = async () => new Promise(resolve => { complete = resolve; });
  try {
    await h.render(<LabelOptionsScreen workspace={f.workspace} scope={scope} assetId="asset" />);
    await h.press(h.byLabel('Save or share PNG'));
    await h.run(() => setScreenFocused(false)); await h.run(() => complete(file));
    expect(f.counts().shares).toBe(0);
  } finally { await h.unmount(); setScreenFocused(true); }
});
