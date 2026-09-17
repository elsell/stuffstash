import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AddDraftNameField } from './AddDraftNameField.ios';

it('retains the native seed during editing and replaces it only for a new draft revision', async () => {
  const h = new MobileRenderHarness(); const changes: string[] = [];
  const field = (revision: number, value: string, editable = true) => <AddDraftNameField key={revision}
    accessibilityLabel="Asset name" editable={editable} value={value} onChangeText={text => changes.push(text)} />;
  try {
    await h.render(field(0, 'Restored tent'));
    await h.change(h.byType('SwiftUITextField'), 'Native draft name');
    expect(changes).toEqual(['Native draft name']);
    await h.render(field(0, 'Native draft name'));
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('Restored tent');
    await h.render(field(0, 'Native draft name', false));
    await h.change(h.byType('SwiftUITextField'), 'Busy callback');
    expect(changes).toEqual(['Native draft name']);
    await h.render(field(0, 'Native draft name'));
    await h.change(h.byType('SwiftUITextField'), 'Retry draft');
    expect(changes).toEqual(['Native draft name', 'Retry draft']);
    await h.render(field(1, ''));
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('');
    await h.render(field(2, 'Other inventory draft'));
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('Other inventory draft');
  } finally { await h.unmount(); }
});
