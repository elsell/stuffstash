import { expect, it } from 'vitest';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { PrinterSettingsScreen } from './PrinterSettingsScreen';
import { PrinterDefaultsScreen } from './PrinterDefaultsScreen';
import { attemptNavigation, dispatchedActions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { latestAlert, pressAlertButton } from '../../test-support/react-native';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };

it('leads with printer navigation and keeps defaults/history out of the landing form', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const opened: string[] = [];
  try {
    await h.render(<PrinterSettingsScreen workspace={fake.workspace()} scope={scope} onPrinter={id => opened.push(id)} onDefaults={() => opened.push('defaults')} onHistory={() => opened.push('history')} />);
    expect(h.byLabel('Show reference')).toBeUndefined();
    expect(h.byLabel('Save')).toBeUndefined();
    await h.press(h.byLabel('Garage'));
    await h.press(h.byLabel('Print defaults'));
    await h.press(h.byLabel('Print history'));
    expect(opened).toEqual(['printer', 'defaults', 'history']);
  } finally { await h.unmount(); }
});

it('saves the latest draft through the header and preserves unsaved defaults across focus refresh', async () => {
  const h = new MobileRenderHarness(); const fake = new PrintingFake(); const workspace = fake.workspace();
  try {
    await h.render(<PrinterDefaultsScreen workspace={workspace} scope={scope} canConfigure />);
    expect(h.byLabel('Save')?.props.disabled).toBe(true);
    await h.run(() => h.byLabel('Print label when adding an item')?.props.onValueChange(true));
    const retainedSave = h.byLabel('Save')!.props.onPress;
    await h.run(() => h.byLabel('Show reference')?.props.onValueChange(false));
    await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); await h.settle();
    expect(h.byLabel('Show reference')?.props.value).toBe(false);
    await h.run(() => { retainedSave(); retainedSave(); }); await h.settle();
    expect(fake.settings.revision).toBe(2);
    expect(fake.settings.printOnCreateDefault).toBe(true);
    expect(fake.settings.template.showReference).toBe(false);
    await h.render(<PrinterDefaultsScreen workspace={workspace} scope={scope} canConfigure={false} />);
    await h.run(() => retainedSave());
    expect(fake.settings.revision).toBe(2);
    expect(h.byLabel('Save')).toBeUndefined();
  } finally { await h.unmount(); setScreenFocused(true); }
});


it('keeps dirty defaults on canceled navigation and expires discarded-focus confirmations', async () => {
  resetNavigation();
  const h = new MobileRenderHarness(); const fake = new PrintingFake();
  try {
    await h.render(<PrinterDefaultsScreen workspace={fake.workspace()} scope={scope} canConfigure />);
    await h.run(() => h.byLabel('Show reference')?.props.onValueChange(false));
    await h.run(() => attemptNavigation({ type: 'GO_BACK' }));
    const staleDiscard = latestAlert()!.buttons.find(button => button.text === 'Discard')!.onPress!;
    await h.run(() => pressAlertButton('Keep editing'));
    expect(dispatchedActions()).toHaveLength(0);
    expect(h.byLabel('Show reference')?.props.value).toBe(false);
    await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); await h.settle();
    await h.run(staleDiscard);
    expect(dispatchedActions()).toHaveLength(0);
    await h.run(() => attemptNavigation({ type: 'GO_BACK' }));
    await h.run(() => pressAlertButton('Discard'));
    expect(dispatchedActions()).toEqual([{ type: 'GO_BACK' }]);
    expect(fake.settings.template.showReference).toBe(true);
  } finally { await h.unmount(); setScreenFocused(true); resetNavigation(); }
});
