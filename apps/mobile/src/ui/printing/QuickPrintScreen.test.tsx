import { expect, it } from 'vitest';
import { QuickPrintScreen } from './QuickPrintScreen';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
it('queues one default label without preview confirmation and retains a lost response through reopening', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace(); let queued = '';
  fake.drop = true;
  try {
    await h.render(<QuickPrintScreen workspace={workspace} scope={scope} assetId="asset" onQueued={id => { queued = id; }} />);
    expect(fake.submitted.size).toBe(1); expect(queued).toBe(''); expect(fake.previews).toBe(1);
    await h.render(<></>); fake.settings = { ...fake.settings, defaultPrinterId: null };
    await h.render(<QuickPrintScreen workspace={workspace} scope={scope} assetId="asset" onQueued={id => { queued = id; }} />);
    expect(queued).toBe(''); expect(fake.submitted.size).toBe(1);
    await h.press(h.byLabel('Try again')); await h.settle();
    expect(queued).toBe('request'); expect(fake.submitted.size).toBe(1); expect(fake.previews).toBe(1);
  } finally { await h.unmount(); }
});
it('opens custom options without printing when no default is configured', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); fake.settings = { ...fake.settings, defaultPrinterId: null };
  try {
    await h.render(<QuickPrintScreen workspace={fake.workspace()} scope={scope} assetId="asset" onQueued={() => {}} />);
    expect(fake.submitted.size).toBe(0); expect(h.byLabel('Preview label')).toBeDefined();
  } finally { await h.unmount(); }
});
it('does not enqueue when the scope task leaves while its render is outstanding', async () => {
  let release!: () => void;
  class DelayedPrinting extends PrintingFake { override async preview() { await new Promise<void>(resolve => { release = resolve; }); return super.preview(); } }
  const h = new MobileRenderHarness(); const fake = new DelayedPrinting();
  await h.render(<QuickPrintScreen workspace={fake.workspace()} scope={scope} assetId="asset" onQueued={() => {}} />);
  await h.unmount(); release(); await h.settle(); expect(fake.submitted.size).toBe(0);
});
it('restores edited fallback options after route reauthorization without automatically printing defaults', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace(); const draftState = {};
  fake.settings = { ...fake.settings, defaultPrinterId: null };
  const screen = () => <QuickPrintScreen workspace={workspace} scope={scope} assetId="asset" draftState={draftState} onQueued={() => {}} />;
  try {
    await h.render(screen());
    await h.run(() => h.byLabel('Copies')!.props.onChangeText('3'));
    await h.render(<></>);
    fake.settings = { ...fake.settings, defaultPrinterId: fake.printer.id };
    await h.render(screen());
    expect(h.byLabel('Copies')!.props.value).toBe('3');
    expect(h.byLabel('Preview label')).toBeDefined();
    expect(fake.previews).toBe(0); expect(fake.submitted.size).toBe(0);
  } finally { await h.unmount(); }
});
