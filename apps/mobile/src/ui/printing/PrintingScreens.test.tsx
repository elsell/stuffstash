import { PrintingFake } from '../../test-support/PrintingFake';
import { PrintRequests } from '../../application/printing/PrintSubmission';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { AssetPrintScreen } from './AssetPrintScreen';
import { PrinterSettingsScreen } from './PrinterSettingsScreen';
import type { PrintCatalog, PrintJob, PrintSettings, PrintingRepository, PrintingWorkspace } from '../../application/printing/PrintingWorkspace';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
it('queues one label to an unavailable printer and retries a lost response without duplicate output', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); fake.drop = true; let queued = '';
  try {
    await h.render(<AssetPrintScreen workspace={fake.workspace()} scope={scope} assetId="asset" onQueued={id => { queued = id; }} />);
    await h.press(h.byLabel('Preview label')); await h.settle();
    await h.press(h.byLabel('Print label')); await h.settle();
    expect(fake.submitted.size).toBe(1); expect(queued).toBe('');
    await h.press(h.byLabel('Try again')); await h.settle();
    expect(queued).toBe('request'); expect(fake.submitted.size).toBe(1);
  } finally { await h.unmount(); }
});
it('retains a settings draft on revision conflict and does not expose cached rows after access loss', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  try {
    await h.render(<PrinterSettingsScreen workspace={workspace} scope={scope} canConfigure onJob={() => {}} />);
    await h.run(() => h.byLabel('Print label when adding an item')?.props.onValueChange(true));
    fake.settings = { ...fake.settings, revision: 2 };
    await h.press(h.byLabel('Save defaults')); await h.settle();
    expect(fake.settings.printOnCreateDefault).toBe(false);
    expect(h.byLabel('Print label when adding an item')?.props.value).toBe(true);
    fake.deny = true;
    await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); await h.settle();
    expect(h.byLabel('Print label when adding an item')).toBeUndefined();
  } finally { await h.unmount(); setScreenFocused(true); }
});

it('recovers a lost submission after leaving the task without generating another request', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); fake.drop = true; const workspace = fake.workspace(); let queued = '';
  try {
    await h.render(<AssetPrintScreen workspace={workspace} scope={scope} assetId="asset" onQueued={id => { queued = id; }} />);
    await h.press(h.byLabel('Preview label')); await h.settle();
    await h.press(h.byLabel('Print label')); await h.settle();
    await h.render(<></>);
    await h.render(<AssetPrintScreen workspace={workspace} scope={scope} assetId="asset" onQueued={id => { queued = id; }} />);
    await h.press(h.byLabel('Try again')); await h.settle();
    expect(queued).toBe('request'); expect(fake.submitted.size).toBe(1); expect(fake.previews).toBe(1);
  } finally { await h.unmount(); }
});
