import React, { createRef } from 'react';
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

it('ends native editing explicitly without depending on a disabled render', async () => {
  const h = new MobileRenderHarness(); const handle = createRef<{ blur(): void | Promise<void> }>();
  let resolveBlur: (() => void) | undefined; let blurs = 0; let settled = false;
  try {
    await h.render(<InvitationEmailInput ref={handle} email="audit@example.invalid" editable onChangeText={() => {}} />);
    h.byType('SwiftUITextField')!.props.ref.current = { blur: () => {
      blurs += 1; return new Promise<void>(resolve => { resolveBlur = resolve; });
    } };
    expect(handle.current).not.toBeNull();
    const ending = Promise.resolve(handle.current!.blur()).then(() => { settled = true; });
    expect(blurs).toBe(1); expect(settled).toBe(false);
    resolveBlur!(); await ending;
    expect(settled).toBe(true);
    expect(h.byType('SwiftUITextField')?.props.defaultValue).toBe('audit@example.invalid');
  } finally { await h.unmount(); }
});
