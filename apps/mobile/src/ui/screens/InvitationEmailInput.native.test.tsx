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
    h.byType('SwiftUITextField')!.props.ref.current = { blur: async () => {} };
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

it('blurs the native field when submission disables editing without discarding the draft', async () => {
  const h = new MobileRenderHarness(); let blurs = 0;
  const view = (editable: boolean) => <InvitationEmailInput email="audit@example.invalid" editable={editable} onChangeText={() => {}} />;
  try {
    await h.render(view(true));
    const nativeField = h.byType('SwiftUITextField');
    expect(nativeField?.props.ref).toBeDefined();
    nativeField!.props.ref.current = { blur: async () => { blurs += 1; } };
    await h.render(view(false));
    expect(blurs).toBe(1);
    await h.render(view(true));
    expect(blurs).toBe(1);
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('audit@example.invalid');
  } finally { await h.unmount(); }
});
