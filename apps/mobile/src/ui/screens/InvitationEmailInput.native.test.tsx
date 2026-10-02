import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { InvitationEmailInput } from './InvitationEmailInput.ios';

it('preserves native email editing through rejected submissions and resets only for a new draft', async () => {
  const h = new MobileRenderHarness(); const changes: string[] = [];
  const view = (revision: string, email: string, editable = true) => <InvitationEmailInput key={revision}
    email={email} editable={editable} onChangeText={value => changes.push(value)} />;
  try {
    await h.render(view('scope-a:0', 'restored@example.invalid'));
    const field = h.byType('TextInput');
    expect(field, 'use the UIKit-backed editing session that Keyboard.dismiss ends').toBeDefined();
    await h.run(() => field!.props.onChangeText('audit@example.invalid'));
    expect(changes).toEqual(['audit@example.invalid']);
    await h.render(view('scope-a:0', 'audit@example.invalid', false));
    expect(h.byType('TextInput')?.props.editable).toBe(false);
    await h.run(() => h.byType('TextInput')!.props.onChangeText('late@example.invalid'));
    expect(changes).toEqual(['audit@example.invalid']);
    await h.render(view('scope-a:0', 'audit@example.invalid'));
    expect(h.byType('TextInput')?.props.defaultValue).toBe('restored@example.invalid');
    expect(h.byType('TextInput')?.props.value).toBeUndefined();
    await h.run(() => h.byType('TextInput')!.props.onChangeText('retry@example.invalid'));
    expect(changes).toEqual(['audit@example.invalid', 'retry@example.invalid']);
    await h.render(view('scope-a:1', ''));
    expect(h.byType('TextInput')?.props.defaultValue).toBe('');
    await h.render(view('scope-b:0', 'other@example.invalid'));
    expect(h.byType('TextInput')?.props.defaultValue).toBe('other@example.invalid');
  } finally { await h.unmount(); }
});
