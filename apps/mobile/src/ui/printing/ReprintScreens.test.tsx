import { expect, it } from 'vitest';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { ReprintLabelScreen } from './ReprintLabelScreen';
import { PrinterDetailScreen } from './PrinterDetailScreen';
import { PrintJobScreen } from './PrintJobScreen';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };

it('links a reprint with fresh preview and retries one retained request after its predecessor expires', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace(); let queued = '';
  fake.submitted.set('original', { id: 'original', assetId: 'asset', printerId: 'printer', status: 'completed', revision: 1, copies: 1, completedCopies: 1 });
  try {
    await h.render(<ReprintLabelScreen workspace={workspace} scope={scope} jobId="original" onQueued={id => { queued = id; }} />);
    expect(h.byLabel('Reprint label')?.props.disabled).toBe(false);
    await h.settle(); await h.settle();
    fake.drop = true; await h.press(h.byLabel('Reprint label')); await h.settle();
    expect(fake.submitted.size).toBe(2); expect(fake.submitted.get('request')?.predecessor).toBe('original');
    await h.render(<></>); fake.submitted.delete('original');
    fake.settings = { ...fake.settings, template: { id: 'qr-only', version: 1, showReference: false } };
    await h.render(<ReprintLabelScreen workspace={workspace} scope={scope} jobId="original" onQueued={id => { queued = id; }} />);
    await h.press(h.byLabel('Try again')); await h.settle();
    expect(queued).toBe('request'); expect(fake.submitted.size).toBe(1); expect(fake.previews).toBe(1);
  } finally { await h.unmount(); }
});

it('does not expose reprint for unsettled jobs and never queues a test label on entry or refresh', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace(); let queued = '';
  fake.submitted.set('uncertain', { id: 'uncertain', printerId: 'printer', status: 'uncertain', revision: 1, copies: 1, completedCopies: 0 });
  try {
    await h.render(<PrintJobScreen workspace={workspace} scope={scope} jobId="uncertain" canPrint onReprint={() => {}} />);
    expect(h.byLabel('Reprint label')).toBeUndefined();
    await h.render(<ReprintLabelScreen workspace={workspace} scope={scope} jobId="uncertain" onQueued={() => {}} />);
    expect(h.byLabel('Reprint label')).toBeUndefined();
    await h.render(<PrinterDetailScreen workspace={workspace} scope={scope} printerId="printer" canConfigure canPrint onJob={id => { queued = id; }} />);
    await h.press(h.byLabel('Refresh status')); await h.settle();
    expect(fake.submitted.size).toBe(1);
    fake.drop = true; await h.press(h.byLabel('Print test label')); await h.settle();
    expect(fake.submitted.get('request')?.kind).toBe('printer_test'); expect(fake.submitted.get('request')?.assetId).toBeUndefined();
    await h.render(<></>);
    fake.settings = { ...fake.settings, template: { id: 'qr-only', version: 1, showReference: false } };
    await h.render(<PrinterDetailScreen workspace={workspace} scope={scope} printerId="printer" canConfigure canPrint onJob={id => { queued = id; }} />);
    await h.press(h.byLabel('Retry test label')); await h.settle();
    expect(queued).toBe('request'); expect(fake.submitted.size).toBe(2);
    await h.render(<PrinterDetailScreen workspace={workspace} scope={scope} printerId="printer" canConfigure={false} canPrint={false} onJob={() => {}} />);
    expect(h.byLabel('Print test label')).toBeUndefined();
  } finally { await h.unmount(); }
});

it('reprints diagnostic content without creating an asset preview', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  fake.submitted.set('diagnostic', { id: 'diagnostic', kind: 'printer_test', printerId: 'printer', status: 'canceled', revision: 2, copies: 1, completedCopies: 0 });
  try {
    await h.render(<ReprintLabelScreen workspace={workspace} scope={scope} jobId="diagnostic" onQueued={() => {}} />);
    expect(h.byLabel('Preview label')).toBeUndefined();
    await h.press(h.byLabel('Reprint label')); await h.settle();
    expect(fake.previews).toBe(0); expect(fake.submitted.get('request')?.predecessor).toBe('diagnostic');
    expect(fake.submitted.get('request')?.assetId).toBeUndefined();
  } finally { await h.unmount(); }
});
