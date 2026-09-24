import { typeInNativeSearch } from '../../test-support/NativeSearchDriver';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { navigationOptions, resetNavigation } from '../../test-support/navigation';
import { AddDestinationSelectionScreen, type AddDestinationSelectionProps } from './AddDestinationSelectionScreen';
import type { ParentSelection } from './AddAssetResolution';
const garage: ParentSelection = { id: 'garage', title: 'Garage', kind: 'location', subtitle: '', pathLabel: 'House / Garage', selectionHint: 'Place', willPromoteToContainer: false };
const base: AddDestinationSelectionProps = { query: '', selected: garage, matches: [garage], disabled: false, loading: false, failed: false, creating: false, canCreate: false, onQuery: () => {}, onRetry: () => {}, onSelect: () => {}, onCreate: () => {}, onClose: () => {} };
it('keeps the current destination through search and applies one eligible choice directly', async () => {
  const h = new MobileRenderHarness(); const selected: unknown[] = []; const queries: string[] = []; resetNavigation();
  try {
    await h.render(<AddDestinationSelectionScreen {...base} onQuery={q => queries.push(q)} onSelect={p => selected.push(p?.id)} />);
    await h.run(() => typeInNativeSearch('another'));
    expect(queries).toEqual(['another']); expect(selected).toEqual([]);
    await h.render(<AddDestinationSelectionScreen {...base} query="another" matches={[]} onSelect={p => selected.push(p?.id)} />);
    expect(h.byText('Current: House / Garage')).toBeDefined();
    await h.press(h.byLabel('Choose inventory top level'));
    expect(selected).toEqual([undefined]);
  } finally { await h.unmount(); resetNavigation(); }
});
it('rejects retained destination callbacks after the option is removed or the task locks', async () => {
  const h = new MobileRenderHarness(); const selected: unknown[] = [];
  const render = (props: Partial<AddDestinationSelectionProps>) => h.render(<AddDestinationSelectionScreen {...base} onSelect={p => selected.push(p)} {...props} />);
  try {
    await render({}); const choose = h.byLabel('Choose destination Garage')!.props.onPress;
    await render({ matches: [] }); await h.run(choose);
    await render({ disabled: true }); await h.run(choose);
    expect(selected).toEqual([]);
    await render({}); await h.run(choose); expect(selected).toEqual([garage]);
  } finally { await h.unmount(); }
});
it('does not offer a new place while lookup is unresolved or failed', async () => {
  const h = new MobileRenderHarness();
  try {
    for (const state of [{ loading: true }, { failed: true }]) {
      await h.render(<AddDestinationSelectionScreen {...base} {...state} />);
      expect(h.byLabel('New place')?.props.disabled).toBe(true);
    }
  } finally { await h.unmount(); }
});
it('keeps failed creation editable and rejects late name edits while creation is pending', async () => {
  const h = new MobileRenderHarness(); const names: string[] = []; let created = 0;
  const render = (props: Partial<AddDestinationSelectionProps>) => h.render(<AddDestinationSelectionScreen {...base} query="Shed" canCreate onQuery={name => names.push(name)} onCreate={() => created++} {...props} />);
  try {
    await render({}); await h.press(h.byLabel('New place'));
    const change = h.byLabel('New place name')!.props.onChangeText;
    await h.press(h.byLabel('Create place')); expect(created).toBe(1);
    await render({ disabled: true, creating: true }); await h.run(() => change('Changed while saving'));
    expect(names).toEqual([]);
    await render({ error: 'Could not create place' });
    expect(h.byText('Could not create place')).toBeDefined();
    expect(h.byLabel('New place name')!.props.value).toBe('Shed');
    await h.run(() => change('Garden shed')); expect(names).toEqual(['Garden shed']);
    await render({ loading: true });
    await h.run(h.byLabel('Create place')!.props.onPress); expect(created).toBe(1);
    await render({});
    await h.press(h.byLabel('Create place')); expect(created).toBe(2);
  } finally { await h.unmount(); }
});

it('separates destination selection from creation and returns without closing the picker', async () => {
 resetNavigation(); const h = new MobileRenderHarness(); let closed = 0; let created = 0;
 const title = () => (Object.assign({}, ...navigationOptions()) as {title?:string}).title;
 try {
  await h.render(<AddDestinationSelectionScreen {...base} query="Shed" canCreate onClose={()=>closed++} onCreate={()=>created++} />);
  await h.press(h.byLabel('New place'));
  expect(title()).toBe('New place');
  expect(h.byLabel('Choose inventory top level')).toBeUndefined();
  expect(h.byLabel('Choose destination Garage')).toBeUndefined();
  expect(h.byLabel('New place name')?.props.value).toBe('Shed');
  const create = h.byLabel('Create place')!.props.onPress;
  await h.press(h.byLabel('Cancel new place'));
  expect(title()).toBe('Put in');
  expect(h.byLabel('Choose destination Garage')).toBeDefined();
  expect(closed).toBe(0);
  await h.run(create); expect(created).toBe(0);
 } finally {await h.unmount();resetNavigation();}
});

it('rejects retained destination rows while creating a place', async () => {
  const h = new MobileRenderHarness(); const selected: unknown[] = [];
  try {
    await h.render(<AddDestinationSelectionScreen {...base} onSelect={parent => selected.push(parent)} />);
    const choose = h.byLabel('Choose destination Garage')!.props.onPress;
    const topLevel = h.byLabel('Choose inventory top level')!.props.onPress;
    await h.press(h.byLabel('New place'));
    await h.run(choose); await h.run(topLevel);
    expect(selected).toEqual([]);
    await h.press(h.byLabel('Cancel new place'));
    await h.press(h.byLabel('Choose destination Garage'));
    expect(selected).toEqual([garage]);
  } finally { await h.unmount(); }
});
