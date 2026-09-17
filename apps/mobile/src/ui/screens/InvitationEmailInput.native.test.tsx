import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { InvitationEmailInput } from './InvitationEmailInput.ios';

it('preserves the email draft across echoes and rejected attempts, rejects busy edits, and resets by revision', async () => {
  const h = new MobileRenderHarness(); const changes: string[] = [];
  const field = (revision: string, email: string, editable = true) => <InvitationEmailInput key={revision}
    email={email} editable={editable} onChangeText={text => changes.push(text)} />;
  try {
    await h.render(field('scope-a:0', 'restored@example.invalid'));
    await h.change(h.byType('SwiftUITextField'), 'audit@example.invalid');
    expect(changes).toEqual(['audit@example.invalid']);
    await h.render(field('scope-a:0', 'audit@example.invalid', false));
    await h.change(h.byType('SwiftUITextField'), 'late@example.invalid');
    expect(changes).toEqual(['audit@example.invalid']);
    await h.render(field('scope-a:0', 'audit@example.invalid'));
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('restored@example.invalid');
    await h.change(h.byType('SwiftUITextField'), 'retry@example.invalid');
    expect(changes).toEqual(['audit@example.invalid', 'retry@example.invalid']);
    await h.render(field('scope-a:1', ''));
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('');
    await h.render(field('scope-b:0', 'other@example.invalid'));
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('other@example.invalid');
  } finally { await h.unmount(); }
});
