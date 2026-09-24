import { useState } from 'react';
import { Text } from 'react-native';
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
    const modifierTypes = native?.props.modifiers.map((modifier: { type: string }) => modifier.type);
    expect(native?.props.modifiers).toContainEqual({ type: 'accessibilityHint', value: '' });
    await h.render(field('This option already exists.'));
    expect(native?.props.modifiers.map((modifier: { type: string }) => modifier.type)).toEqual(modifierTypes);
    expect(h.byType('SwiftUITextField')).toBe(native);
    expect(native?.props.modifiers).toContainEqual({ type: 'accessibilityHint', value: 'This option already exists.' });
    await h.render(field());
    expect(h.byType('SwiftUITextField')).toBe(native);
    expect(native?.props.modifiers).toContainEqual({ type: 'accessibilityHint', value: '' });
    expect(native?.props.modifiers.map((modifier: { type: string }) => modifier.type)).toEqual(modifierTypes);
    expect(native?.props.defaultValue).toBe('Saved');
  } finally { await h.unmount(); }
});

// Expo UI 55.0.17 TextFieldView.onAppear reapplies defaultValue and emits the change.
it('retains the committed draft when a native selection visit hides and restores the field', async () => {
  const h = new MobileRenderHarness();
  function Editor() {
    const [value, setValue] = useState('');
    return <><DraftTextField accessibilityLabel="Asset name" value={value} onChangeText={setValue} />
      <Text>{`Draft: ${value}`}</Text></>;
  }
  try {
    await h.render(<Editor />);
    const field = h.byType('SwiftUITextField');
    await h.change(field, 'Tent');
    expect(h.byText('Draft: Tent')).toBeDefined();
    expect(h.byType('SwiftUITextField')).toBe(field);
    await h.change(field, field!.props.defaultValue);
    expect(h.byText('Draft: Tent')).toBeDefined();
  } finally { await h.unmount(); }
});
