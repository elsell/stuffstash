import { expect, it } from 'vitest';
import { AssetPrintScreen } from './AssetPrintScreen';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import type { LabelFile, LabelFiles } from '../../application/labels/LabelWorkspace';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
it('automatically refreshes rendering choices, keeps copies independent, and releases obsolete native files', async () => {
  class DeferredFiles implements LabelFiles {
    readonly pending: Array<(local: { uri: string; release(): void }) => void> = [];
    readonly released: string[] = [];
    preview(_file: LabelFile, _signal: AbortSignal) { return new Promise<{ uri: string; release(): void }>(resolve => this.pending.push(resolve)); }
    async deliver() {}
    finish(index: number) { this.pending[index]({ uri: `file://${index}`, release: () => this.released.push(String(index)) }); }
  }
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const files = new DeferredFiles(); const workspace = { ...fake.workspace(), files };
  try {
    await h.render(<AssetPrintScreen workspace={workspace} scope={scope} assetId="asset" onQueued={() => {}} />);
    expect(files.pending).toHaveLength(1); expect(h.byLabel('Preview label')).toBeUndefined();
    await h.run(() => h.byLabel('Show reference')!.props.onValueChange(false));
    expect(files.pending).toHaveLength(2);
    await h.run(() => files.finish(0));
    expect(files.released).toEqual(['0']); expect(h.byLabel('Print label')!.props.disabled).toBe(true);
    await h.run(() => files.finish(1));
    expect(h.byLabel('Print label')!.props.disabled).toBe(false);
    await h.run(() => h.byLabel('Copies')!.props.onChangeText('3'));
    expect(files.pending).toHaveLength(2); expect(h.byLabel('Print label')!.props.disabled).toBe(false);
    await h.run(() => setScreenFocused(false));
    expect(files.released).toEqual(['0', '1']); expect(fake.submitted.size).toBe(0);
  } finally { await h.unmount(); setScreenFocused(true); }
});

it('exports the selected printer media without output and drops a share after blur', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  const exported: Array<{ format: string; selection: import('../../application/labels/LabelWorkspace').LabelSelection }> = [];
  const delivered: string[] = []; let finish: ((file: LabelFile) => void) | undefined;
  const labelWorkspace: import('../../application/labels/LabelWorkspace').LabelWorkspace = {
    parse: () => ({ instanceId: 'instance', labelId: 'label' }),
    repository: {
      instance: async () => 'instance', resolve: async () => ({ ...scope, assetId: 'asset', archived: false }),
      catalog: async () => ({ profiles: [], templates: [] }),
      render: async (_scope, _asset, selection, format) => { exported.push({ selection, format }); return new Promise(resolve => { finish = resolve; }); }
    },
    files: { preview: workspace.files.preview, deliver: async (file, action) => { delivered.push(`${file.format}:${action}`); } }
  };
  try {
    await h.render(<AssetPrintScreen workspace={workspace} labelWorkspace={labelWorkspace} scope={scope} assetId="asset" onQueued={() => {}} />);
    await h.run(() => h.byLabel('Show reference')!.props.onValueChange(false));
    await h.run(() => { void h.byLabel('Save or share PDF')!.props.onPress(); });
    expect(exported[0].selection.media).toEqual(fake.printer.media);
    expect(exported[0].selection.showReference).toBe(false); expect(exported[0].format).toBe('pdf');
    await h.run(() => finish!({ bytes: new Uint8Array([1]), format: 'pdf', width: 10, height: 20, rotation: 0 }));
    expect(delivered).toEqual(['pdf:share']);
    await h.run(() => { void h.byLabel('Save or share PNG')!.props.onPress(); });
    await h.run(() => setScreenFocused(false));
    await h.run(() => finish!({ bytes: new Uint8Array([1]), format: 'png', width: 10, height: 20, rotation: 0 }));
    expect(delivered).toEqual(['pdf:share']); expect(fake.submitted.size).toBe(0);
  } finally { await h.unmount(); setScreenFocused(true); }
});
