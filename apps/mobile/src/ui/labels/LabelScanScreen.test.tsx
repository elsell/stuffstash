import { afterEach, expect, it } from 'vitest';
import { LabelScanScreen } from './LabelScanScreen';
import { MobileRenderHarness } from '../../test-support/render';
import { setCameraAllowed } from '../../test-support/expo-camera';
import { setScreenFocused } from '../../test-support/navigation';
import { LabelFailure } from '../../application/labels/LabelWorkspace';
const reference = { instanceId: 'instance', labelId: 'label' };
const target = { tenantId: 'tenant', inventoryId: 'inventory', assetId: 'asset', archived: true };
afterEach(() => { setCameraAllowed(true); setScreenFocused(true); });
it('resolves a repeated QR once and never navigates after leaving the scanner', async () => {
  const h = new MobileRenderHarness(); let reads = 0; let release!: (value: typeof target) => void; const navigated: string[] = [];
  try {
    await h.render(<LabelScanScreen open={{ execute: async () => { reads++; return new Promise(resolve => { release = resolve; }); } }} parse={() => reference}
      invalid={false} onResolved={id => navigated.push(id)} onAccount={() => {}} onServer={() => {}} />);
    const frame = h.byType('CameraView')?.props.onBarcodeScanned;
    expect(frame).toBeTypeOf('function');
    await h.run(() => { frame({ data: 'code' }); frame({ data: 'code' }); });
    expect(reads).toBe(1);
    await h.run(() => setScreenFocused(false));
    await h.run(() => release(target));
    expect(navigated).toEqual([]);
    await h.run(() => setScreenFocused(true));
    await h.press(h.byLabel('Use camera'));
    await h.run(() => h.byType('CameraView')?.props.onBarcodeScanned({ data: 'another-code' }));
    expect(reads).toBe(2);
    await h.run(() => release(target));
    expect(navigated).toEqual(['asset']);
  } finally { await h.unmount(); }
});
it('allows paste when camera access is denied and keeps a foreign label recoverable', async () => {
  setCameraAllowed(false); const h = new MobileRenderHarness(); let changes = 0;
  try {
    await h.render(<LabelScanScreen open={{ execute: async () => { throw new LabelFailure('wrong_instance'); } }} parse={() => reference}
      invalid={false} onResolved={() => { throw new Error('Must not navigate'); }} onAccount={() => {}} onServer={() => { changes++; }} />);
    expect(h.byType('CameraView')).toBeUndefined();
    await h.changeText(h.byLabel('Paste label link'), 'https://old.example/l/v1/instance/label');
    await h.press(h.byLabel('Open label')); await h.settle();
    expect(h.byText('This label belongs to another Stuff Stash instance. Connect to that server and try again.')).toBeDefined();
    await h.press(h.byLabel('Change server')); expect(changes).toBe(1);
  } finally { await h.unmount(); }
});

it('cancels an active valid resolution when a malformed incoming label supersedes it', async () => {
  const h = new MobileRenderHarness(); let release!: (value: typeof target) => void;
  const navigated: string[] = []; let signal: AbortSignal | undefined;
  const open = { execute: async (_reference: typeof reference, next: AbortSignal) => { signal = next; return new Promise<typeof target>(resolve => { release = resolve; }); } };
  const props = { open, parse: () => reference, onResolved: (id: string) => navigated.push(id), onAccount: () => {}, onServer: () => {} };
  try {
    await h.render(<LabelScanScreen {...props} pending={reference} invalid={false} />);
    await h.press(h.byLabel('Open label'));
    await h.render(<LabelScanScreen {...props} invalid={true} />);
    expect(signal?.aborted).toBe(true);
    await h.run(() => release(target));
    expect(navigated).toEqual([]);
  } finally { await h.unmount(); }
});
