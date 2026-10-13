import { useState } from 'react';
import { afterEach, expect, it } from 'vitest';
import { LabelScanScreen } from './LabelScanScreen';
import { MobileRenderHarness } from '../../test-support/render';
import { cameraPermissionsFake, setCameraAllowed } from '../../test-support/expo-camera';
import { setScreenFocused } from '../../test-support/navigation';
import { LabelFailure } from '../../application/labels/LabelWorkspace';
import { setAppStateForTest } from '../../test-support/react-native';
const reference = { instanceId: 'instance', labelId: 'label' };
const target = { tenantId: 'tenant', inventoryId: 'inventory', assetId: 'asset', archived: true };
afterEach(() => { cameraPermissionsFake.reset(); setAppStateForTest('active'); setScreenFocused(true); });
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
    expect(signal).toBeDefined();
    await h.render(<LabelScanScreen {...props} invalid={true} />);
    expect(signal?.aborted).toBe(true);
    await h.run(() => release(target));
    expect(navigated).toEqual([]);
  } finally { await h.unmount(); }
});

it.each([false, true])('retains the OS permission decision while the app is backgrounded (granted=%s)', async granted => {
  cameraPermissionsFake.reset(); setCameraAllowed(false); cameraPermissionsFake.deferDecision();
  const h = new MobileRenderHarness(); const navigated: string[] = [];
  try {
    await h.render(<LabelScanScreen open={{ execute: async () => target }} parse={() => reference}
      invalid={false} onResolved={id => navigated.push(id)} onAccount={() => {}} onServer={() => {}} />);
    expect(cameraPermissionsFake.requests).toBe(1);
    await h.run(() => setAppStateForTest('background'));
    await h.run(() => cameraPermissionsFake.decide(granted));
    expect(h.byType('CameraView')).toBeUndefined();
    await h.run(() => setAppStateForTest('active'));
    expect(cameraPermissionsFake.requests).toBe(1);
    if (granted) {
      expect(h.byType('CameraView')).toBeDefined();
      await h.run(() => setAppStateForTest('background'));
      expect(h.byType('CameraView')).toBeUndefined();
    } else {
      expect(h.byText('Camera unavailable. Allow camera access in Settings, or paste a label link.')).toBeDefined();
      await h.changeText(h.byLabel('Paste label link'), 'https://old.example/l/v1/instance/label');
      await h.press(h.byLabel('Open label'));
      expect(navigated).toEqual(['asset']);
    }
  } finally { await h.unmount(); }
});

it('opens an external label without camera permission or a second confirmation', async () => {
  const h = new MobileRenderHarness(); const navigated: string[] = []; let reads = 0;
  try {
    await h.render(<LabelScanScreen open={{ execute: async () => { reads++; return target; } }} parse={() => reference}
      pending={reference} invalid={false} onResolved={id => navigated.push(id)} onAccount={() => {}} onServer={() => {}} />);
    expect(reads).toBe(1); expect(navigated).toEqual(['asset']);
    expect(cameraPermissionsFake.requests).toBe(0);
  } finally { await h.unmount(); }
});
it('keeps manual entry secondary and does not offer account changes for an invalid code', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<LabelScanScreen open={{ execute: async () => target }} parse={() => { throw new LabelFailure('invalid_label'); }}
      invalid={false} onResolved={() => {}} onAccount={() => {}} onServer={() => {}} />);
    expect(h.byType('TextInput')).toBeUndefined();
    await h.press(h.byLabel('Enter label link'));
    await h.changeText(h.byLabel('Paste label link'), 'bad-code');
    await h.press(h.byLabel('Open label'));
    expect(h.byLabel('Change server')).toBeUndefined(); expect(h.byLabel('Switch account')).toBeUndefined();
    expect(h.byLabel('Paste label link')).toBeDefined();
  } finally { await h.unmount(); }
});

it('retries a new delivery of the same external link and retains retry after blur', async () => {
  const h = new MobileRenderHarness(); let reads = 0; let signal: AbortSignal | undefined;
  const props = { open: { execute: async (_reference: typeof reference, next: AbortSignal) => { reads++; signal = next; return new Promise<typeof target>(() => {}); } }, parse: () => reference,
    pending: reference, invalid: false, onResolved: () => {}, onAccount: () => {}, onServer: () => {} };
  try {
    await h.render(<LabelScanScreen {...props} revision={1} />);
    const first = signal;
    await h.render(<LabelScanScreen {...props} revision={2} />);
    expect(first?.aborted).toBe(true); expect(reads).toBe(2);
    await h.run(() => setScreenFocused(false));
    expect(signal?.aborted).toBe(true);
    await h.run(() => setScreenFocused(true));
    await h.press(h.byLabel('Try again'));
    expect(reads).toBe(3);
  } finally { await h.unmount(); }
});

it('retains a local scan before lookup without echoing it and resumes after session replacement', async () => {
  const h = new MobileRenderHarness(); let reads = 0; let retained: typeof reference | undefined;
  const open = { execute: async () => { reads++; throw new Error('Connection unavailable'); } };
  function Flow({ session }: { session: string }) {
    const [pending, setPending] = useState<typeof reference>();
    const [revision, setRevision] = useState(0);
    return <LabelScanScreen key={session} open={open} parse={() => reference} pending={pending} revision={revision}
      retain={value => { retained = value; setPending(value); setRevision(1); return 1; }} invalid={false}
      onResolved={() => {}} onAccount={() => {}} onServer={() => {}} />;
  }
  try {
    await h.render(<Flow session="one" />);
    await h.run(() => h.byType('CameraView')?.props.onBarcodeScanned({ data: 'label' }));
    expect(retained).toEqual(reference); expect(reads).toBe(1);
    await h.render(<Flow session="two" />);
    expect(reads).toBe(2);
    expect(h.byLabel('Try again')).toBeDefined();
  } finally { await h.unmount(); }
});
