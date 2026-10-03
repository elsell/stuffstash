import { PrintJobScreen } from './PrintJobScreen';
import { PrintingFake } from '../../test-support/PrintingFake';
import { PrintRequests } from '../../application/printing/PrintSubmission';
import { expect, it, vi } from 'vitest';
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

it('requires latest idle evidence and explicit acknowledgement, then recovers a lost resolution without reporting completion', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  fake.submitted.set('uncertain', { id: 'uncertain', printerId: 'printer', status: 'uncertain', revision: 4, copies: 1, completedCopies: 0 });
  try {
    await h.render(<PrintJobScreen workspace={workspace} scope={scope} jobId="uncertain" canPrint />);
    expect(h.byLabel('Resolve job')?.props.disabled).toBe(true);
    fake.submitted.set('uncertain', { ...fake.submitted.get('uncertain')!, idleConfirmed: true });
    await h.press(h.byLabel('Refresh status')); await h.settle();
    await h.press(h.byLabel('What happened at the printer?')); await h.press(h.byLabel('A label printed'));
    expect(h.byLabel('Resolve job')?.props.disabled).toBe(true);
    await h.run(() => h.byLabel('I understand the print outcome remains unconfirmed')?.props.onValueChange(true));
    fake.dropResolution = true;
    await h.press(h.byLabel('Resolve job')); await h.settle();
    expect(fake.submitted.get('uncertain')?.status).toBe('failed');
    expect(h.byLabel('What happened at the printer?')?.props.disabled).toBe(true);
    await h.press(h.byLabel('Retry acknowledgement')); await h.settle();
    expect(fake.submitted.get('uncertain')?.resolution?.reportedOutcome).toBe('printed');
    expect(fake.submitted.size).toBe(1);
    expect(h.allText()).toContain('Uncertainty acknowledged');
  } finally { await h.unmount(); }
});

it('keeps uncertainty recovery unavailable to a viewer', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake();
  fake.submitted.set('uncertain', { id: 'uncertain', printerId: 'printer', status: 'uncertain', revision: 1, copies: 1, completedCopies: 0, idleConfirmed: true });
  try {
    await h.render(<PrintJobScreen workspace={fake.workspace()} scope={scope} jobId="uncertain" canPrint={false} />);
    expect(h.byLabel('What happened at the printer?')).toBeUndefined();
    expect(h.byLabel('Resolve job')).toBeUndefined();
  } finally { await h.unmount(); }
});

it('keeps the exact ambiguous acknowledgement on same-revision refresh and never rebases it during polling', async () => {
  vi.useFakeTimers();
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  const old = { id: 'job', printerId: 'printer', status: 'uncertain', revision: 4, copies: 1, completedCopies: 0, idleConfirmed: true, latestAttemptId: 'old' };
  fake.submitted.set(old.id, old);
  try {
    await h.render(<PrintJobScreen workspace={workspace} scope={scope} jobId="job" canPrint />);
    await h.press(h.byLabel('What happened at the printer?')); await h.press(h.byLabel('A label printed'));
    await h.run(() => h.byLabel('I understand the print outcome remains unconfirmed')?.props.onValueChange(true));
    fake.loseResolutionBeforeCommit = true;
    await h.press(h.byLabel('Resolve job')); await h.settle();
    await h.press(h.byLabel('Refresh status')); await h.settle();
    expect(h.byLabel('What happened at the printer?')?.props.disabled).toBe(true);
    expect(h.byLabel('Retry acknowledgement')).toBeDefined();
    fake.submitted.set(old.id, { ...old, revision: 8, latestAttemptId: 'new' });
    await h.run(() => vi.advanceTimersByTimeAsync(5000)); await h.settle();
    expect(h.byLabel('What happened at the printer?')?.props.disabled).toBe(true);
    await h.press(h.byLabel('Retry acknowledgement')); await h.settle();
    expect(fake.submitted.get(old.id)?.status).toBe('uncertain');
    await h.press(h.byLabel('Refresh status')); await h.settle();
    expect(h.byLabel('What happened at the printer?')?.props.disabled).toBe(false);
    expect(h.byLabel('I understand the print outcome remains unconfirmed')?.props.value).toBe(false);
    expect(h.byLabel('Resolve job')?.props.disabled).toBe(true);
    await h.press(h.byLabel('What happened at the printer?')); await h.press(h.byLabel('No label printed'));
    await h.run(() => h.byLabel('I understand the print outcome remains unconfirmed')?.props.onValueChange(true));
    await h.press(h.byLabel('Resolve job')); await h.settle();
    expect(fake.submitted.get(old.id)?.revision).toBe(9);
    expect(fake.submitted.get(old.id)?.resolution?.reportedOutcome).toBe('not_printed');
  } finally { await h.unmount(); vi.useRealTimers(); }
});
