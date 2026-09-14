import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { CustomizationFieldControls } from './CustomizationEditorFields';

it('does not change a field type from an option opened before the form became read-only', async () => {
  const h = new MobileRenderHarness(); const changes: string[] = [];
  const form = (canMutate: boolean) => <CustomizationFieldControls applicability="all_assets" canMutate={canMutate} eligibleTypes={[]} enumOptions={[]} fieldType="text" mode="create" newOption=""
    onApplicability={() => {}} onEnumOptions={() => {}} onFieldType={value => changes.push(value)} onNewOption={() => {}} onTargets={() => {}} targetIds={[]} />;
  try {
    await h.render(form(true));
    await h.press(h.byLabel('Choose Type. Current value Text'));
    await h.render(form(false));
    await h.press(h.byText('Number')?.parent ?? undefined);
    expect(changes).toEqual([]);
  } finally { await h.unmount(); }
});
