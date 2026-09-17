import { useState } from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { CustomizationFieldControls } from './CustomizationEditorFields';

it.each([
  ['!!!', 'Use letters to start the option, then letters, numbers, or hyphens.'],
  ['Saved', 'This option already exists.']
])('preserves rejected enum draft %s and allows correction', async (draft, message) => {
  const h = new MobileRenderHarness();
  function Form() {
    const [options, setOptions] = useState<readonly string[]>(['saved']);
    const [value, setValue] = useState(draft);
    return <CustomizationFieldControls persistedEnumOptions={['saved']} persistedTargetIds={[]} applicability="all_assets" canMutate eligibleTypes={[]} enumOptions={options} fieldType="enum" mode="edit" newOption={value}
      onApplicability={() => {}} onEnumOptions={setOptions} onFieldType={() => {}} onNewOption={setValue} onTargets={() => {}} targetIds={[]} />;
  }
  try {
    await h.render(<Form />);
    await h.press(h.byLabel('Add option'));
    expect(h.byLabel('New enum option')?.props.value).toBe(draft);
    expect(h.byText(message)).toBeDefined();
    await h.changeText(h.byLabel('New enum option'), 'Corrected');
    expect(h.byText(message)).toBeUndefined();
    await h.press(h.byLabel('Add option'));
    expect(h.byLabel('Remove corrected')).toBeDefined();
    expect(h.byLabel('New enum option')?.props.value).toBe('');
    expect(h.byText('saved · Existing')).toBeDefined();
  } finally { await h.unmount(); }
});

it('does not change a field type from an option opened before the form became read-only', async () => {
  const h = new MobileRenderHarness(); const changes: string[] = [];
  const form = (canMutate: boolean) => <CustomizationFieldControls persistedEnumOptions={[]} persistedTargetIds={[]} applicability="all_assets" canMutate={canMutate} eligibleTypes={[]} enumOptions={[]} fieldType="text" mode="create" newOption=""
    onApplicability={() => {}} onEnumOptions={() => {}} onFieldType={value => changes.push(value)} onNewOption={() => {}} onTargets={() => {}} targetIds={[]} />;
  try {
    await h.render(form(true));
    await h.press(h.byLabel('Choose Type. Current value Text'));
    await h.render(form(false));
    await h.press(h.byText('Number')?.parent ?? undefined);
    expect(changes).toEqual([]);
  } finally { await h.unmount(); }
});


it('allows a newly selected asset type to be deselected before creating a field', async () => {
  const h = new MobileRenderHarness();
  function Form() {
    const [targets, setTargets] = useState<readonly string[]>([]);
    return <CustomizationFieldControls persistedEnumOptions={[]} persistedTargetIds={[]} applicability="custom_asset_types" canMutate eligibleTypes={[{ kind: 'asset-type', id: 'type-1', tenantId: 'tenant', scope: 'inventory', key: 'tools', displayName: 'Tools', description: '', lifecycle: 'active' }]} enumOptions={[]} fieldType="text" mode="create" newOption=""
      onApplicability={() => {}} onEnumOptions={() => {}} onFieldType={() => {}} onNewOption={() => {}} onTargets={setTargets} targetIds={targets} />;
  }
  try {
    await h.render(<Form />);
    await h.press(h.byLabel('Tools'));
    expect(h.byText('Choose at least one asset type.')).toBeUndefined();
    await h.press(h.byLabel('Tools'));
    expect(h.byText('Choose at least one asset type.')).toBeDefined();
  } finally { await h.unmount(); }
});


it('preserves saved targets while allowing draft additions to be removed during editing', async () => {
  const h = new MobileRenderHarness();
  function Form() {
    const [targets, setTargets] = useState<readonly string[]>(['saved']);
    return <CustomizationFieldControls persistedEnumOptions={[]} applicability="custom_asset_types" canMutate
      eligibleTypes={['saved', 'draft'].map(id => ({ kind: 'asset-type', id, tenantId: 'tenant', scope: 'inventory', key: id, displayName: id, description: '', lifecycle: 'active' }))}
      enumOptions={[]} fieldType="text" mode="edit" newOption="" persistedTargetIds={['saved']}
      onApplicability={() => {}} onEnumOptions={() => {}} onFieldType={() => {}} onNewOption={() => {}} onTargets={setTargets} targetIds={targets} />;
  }
  try {
    await h.render(<Form />);
    expect(h.byText('saved · Existing')).toBeDefined();
    expect(h.byLabel('saved')).toBeUndefined();
    await h.press(h.byLabel('draft'));
    expect(h.byLabel('draft')?.props.accessibilityState.checked).toBe(true);
    await h.press(h.byLabel('draft'));
    expect(h.byLabel('draft')?.props.accessibilityState.checked).toBe(false);
    expect(h.byText('saved · Existing')).toBeDefined();
    expect(h.byText('Choose at least one asset type.')).toBeUndefined();
  } finally { await h.unmount(); }
});


it('can remove unavailable draft targets without removing saved targets or revealing IDs', async () => {
  const h = new MobileRenderHarness();
  function Form() {
    const [targets, setTargets] = useState<readonly string[]>(['private-saved', 'private-draft']);
    return <CustomizationFieldControls persistedEnumOptions={[]} applicability="custom_asset_types" canMutate eligibleTypes={[]}
      enumOptions={[]} fieldType="text" mode="edit" newOption="" persistedTargetIds={['private-saved']}
      onApplicability={() => {}} onEnumOptions={() => {}} onFieldType={() => {}} onNewOption={() => {}} onTargets={setTargets} targetIds={targets} />;
  }
  try {
    await h.render(<Form />);
    await h.press(h.byLabel('Include unavailable draft selections'));
    expect(h.byLabel('Include unavailable draft selections')).toBeUndefined();
    expect(h.byText('1 existing asset type is unavailable')).toBeDefined();
    expect(h.byText('Choose at least one asset type.')).toBeUndefined();
    expect(h.allText()).not.toContain('private-');
  } finally { await h.unmount(); }
});


it('removes unsaved enum additions while preserving saved options', async () => {
  const h = new MobileRenderHarness();
  function Form() {
    const [options, setOptions] = useState<readonly string[]>(['saved', 'draft']);
    return <CustomizationFieldControls persistedEnumOptions={['saved']} persistedTargetIds={[]} applicability="all_assets" canMutate
      eligibleTypes={[]} enumOptions={options} fieldType="enum" mode="edit" newOption=""
      onApplicability={() => {}} onEnumOptions={setOptions} onFieldType={() => {}} onNewOption={() => {}} onTargets={() => {}} targetIds={[]} />;
  }
  try {
    await h.render(<Form />);
    expect(h.byText('saved · Existing')).toBeDefined();
    expect(h.byLabel('Remove saved')).toBeUndefined();
    await h.press(h.byLabel('Remove draft'));
    expect(h.byText('draft')).toBeUndefined();
    expect(h.byLabel('Remove draft')).toBeUndefined();
    expect(h.byText('saved · Existing')).toBeDefined();
  } finally { await h.unmount(); }
});

it('can remove the last option from a create draft and shows validation', async () => {
  const h = new MobileRenderHarness();
  function Form() {
    const [options, setOptions] = useState<readonly string[]>(['draft']);
    return <CustomizationFieldControls persistedEnumOptions={[]} persistedTargetIds={[]} applicability="all_assets" canMutate
      eligibleTypes={[]} enumOptions={options} fieldType="enum" mode="create" newOption=""
      onApplicability={() => {}} onEnumOptions={setOptions} onFieldType={() => {}} onNewOption={() => {}} onTargets={() => {}} targetIds={[]} />;
  }
  try {
    await h.render(<Form />);
    await h.press(h.byLabel('Remove draft'));
    expect(h.byText('Add at least one option.')).toBeDefined();
  } finally { await h.unmount(); }
});

it('changes draft applicability in place while preserving the saved constraint', async () => {
  const h = new MobileRenderHarness();
  function Form() {
    const [applicability, setApplicability] = useState<'all_assets' | 'custom_asset_types'>('custom_asset_types');
    return <CustomizationFieldControls persistedApplicability="custom_asset_types" persistedEnumOptions={[]} persistedTargetIds={[]} applicability={applicability} canMutate eligibleTypes={[]} enumOptions={[]} fieldType="text" mode="edit" newOption=""
      onApplicability={setApplicability} onEnumOptions={() => {}} onFieldType={() => {}} onNewOption={() => {}} onTargets={() => {}} targetIds={[]} />;
  }
  try {
    await h.render(<Form />);
    await h.press(h.byLabel('Choose Applies to. Current value Selected asset types'));
    await h.press(h.byLabel('All assets'));
    await h.press(h.byLabel('Choose Applies to. Current value All assets'));
    await h.press(h.byLabel('Selected asset types'));
    expect(h.byLabel('Choose Applies to. Current value Selected asset types')).toBeDefined();
  } finally { await h.unmount(); }
});

it('adds a normalized option without changing persisted options and clears its input', async () => {
  const h = new MobileRenderHarness();
  function Form() {
    const [options, setOptions] = useState<readonly string[]>(['saved']);
    const [draft, setDraft] = useState('');
    return <CustomizationFieldControls persistedEnumOptions={['saved']} persistedTargetIds={[]} applicability="all_assets" canMutate eligibleTypes={[]} enumOptions={options} fieldType="enum" mode="edit" newOption={draft}
      onApplicability={() => {}} onEnumOptions={setOptions} onFieldType={() => {}} onNewOption={setDraft} onTargets={() => {}} targetIds={[]} />;
  }
  try {
    await h.render(<Form />);
    await h.changeText(h.byLabel('New enum option'), 'New Option');
    await h.press(h.byLabel('Add option'));
    expect(h.byLabel('Remove new-option')).toBeDefined();
    expect(h.byText('saved · Existing')).toBeDefined();
    expect(h.byLabel('New enum option')?.props.value).toBe('');
    await h.changeText(h.byLabel('New enum option'), 'New Option');
    await h.press(h.byLabel('Add option'));
    expect(h.all().filter(node => node.props.accessibilityLabel === 'Remove new-option')).toHaveLength(1);
  } finally { await h.unmount(); }
});
