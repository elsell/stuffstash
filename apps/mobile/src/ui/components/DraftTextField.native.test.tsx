import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { DraftTextField } from './DraftTextField.ios';

it('updates the native validation hint without replacing the editing seed', async () => {
  const h = new MobileRenderHarness();
  const field = (hint?: string) => <DraftTextField accessibilityLabel="New enum option" accessibilityHint={hint}
    value="Saved" onChangeText={() => {}} />;
  try {
    await h.render(field());
    const native = h.byType('SwiftUITextField');
    await h.render(field('This option already exists.'));
    expect(h.byType('SwiftUITextField')).toBe(native);
    expect(native?.props.modifiers).toContainEqual({ type: 'accessibilityHint', value: 'This option already exists.' });
    await h.render(field());
    expect(h.byType('SwiftUITextField')).toBe(native);
    expect(native?.props.modifiers.some((modifier: { type: string }) => modifier.type === 'accessibilityHint')).toBe(false);
    expect(native?.props.defaultValue).toBe('Saved');
  } finally { await h.unmount(); }
});
