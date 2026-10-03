import { expect, it } from 'vitest';
import { AssetPrintScreen } from './AssetPrintScreen';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { PrintRequestRejected } from '../../application/printing/PrintSubmission';
import type { PrintSelection } from '../../application/printing/PrintingWorkspace';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
class CopyLimitedPrinting extends PrintingFake {
  limit = 30; readonly accepted = new Map<string, PrintSelection>();
  override async submit(selected: typeof scope, id: string, selection: PrintSelection, key: string) {
    if (selection.copies > this.limit) throw new PrintRequestRejected();
    this.accepted.set(key, selection); return super.submit(selected, id, selection, key);
  }
}
it('uses the configured server copy limit, preserves rejected drafts and freezes ambiguous retries', async () => {
  const h = new MobileRenderHarness(); const fake = new CopyLimitedPrinting(); const workspace = fake.workspace(); let queued = '';
  try {
    await h.render(<AssetPrintScreen workspace={workspace} scope={scope} assetId="asset" onQueued={id => { queued = id; }} />);
    await h.run(() => h.byLabel('Copies')!.props.onChangeText('31'));
    await h.press(h.byLabel('Preview label')); await h.press(h.byLabel('Print label')); await h.settle();
    expect(fake.submitted.size).toBe(0); expect(h.byLabel('Copies')!.props.value).toBe('31'); expect(h.byLabel('Copies')!.props.editable).toBe(true);
    await h.run(() => h.byLabel('Copies')!.props.onChangeText('25')); expect(h.byLabel('Print label')!.props.disabled).toBe(true);
    await h.press(h.byLabel('Preview label')); fake.drop = true; await h.press(h.byLabel('Print label')); await h.settle();
    expect(fake.accepted.size).toBe(1); expect([...fake.accepted.values()][0].copies).toBe(25); expect(h.byLabel('Copies')!.props.editable).toBe(false);
    await h.render(<></>); await h.render(<AssetPrintScreen workspace={workspace} scope={scope} assetId="asset" onQueued={id => { queued = id; }} />);
    expect(h.byLabel('Copies')!.props.value).toBe('25'); await h.press(h.byLabel('Try again')); await h.settle();
    expect(queued).not.toBe(''); expect(fake.submitted.size).toBe(1); expect(fake.accepted.size).toBe(1);
  } finally { await h.unmount(); }
});
it('does not submit invalid copy counts', async () => {
  const h = new MobileRenderHarness(); const fake = new CopyLimitedPrinting();
  try {
    await h.render(<AssetPrintScreen workspace={fake.workspace()} scope={scope} assetId="asset" onQueued={() => {}} />);
    for (const value of ['', '0', '-1', '1.5', 'NaN', '9007199254740992']) {
      await h.run(() => h.byLabel('Copies')!.props.onChangeText(value));
      expect(h.byLabel('Preview label')!.props.disabled).toBe(true); expect(h.byLabel('Print label')!.props.disabled).toBe(true);
    }
    expect(fake.submitted.size).toBe(0);
  } finally { await h.unmount(); }
});
