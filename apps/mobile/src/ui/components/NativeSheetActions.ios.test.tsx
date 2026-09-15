import React from 'react';
import { expect,it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeSheetActions } from './NativeSheetActions.ios';
it('gives native sheet actions the full proposed width and content-driven height',async()=>{
 const h=new MobileRenderHarness();const actions:string[]=[];
 try{
  await h.render(<NativeSheetActions primaryLabel="Apply filters" secondaryLabel="Cancel" disabled={false} onApply={()=>actions.push('apply')} onBack={()=>actions.push('cancel')} />);
  expect(h.byType('SwiftUIHost')?.props.ignoreSafeArea).toBeUndefined();
  expect(h.byType('SwiftUIHost')?.props.matchContents).toEqual({vertical:true});
  expect(h.byType('SwiftUIHost')?.props.style).toMatchObject({width:'100%'});
  const buttons=h.allByType('SwiftUIButton');expect(buttons).toHaveLength(2);
  expect(buttons[0].props.modifiers).toContainEqual({type:'buttonStyle',value:'borderedProminent'});
  await h.press(buttons[0]);await h.press(buttons[1]);expect(actions).toEqual(['apply','cancel']);
  expect(h.allText()).toEqual(expect.arrayContaining(['Apply filters','Cancel']));
 }finally{await h.unmount()}
});

it('disables Apply for an invalid range while keeping Back available', async () => {
  const h = new MobileRenderHarness();
  const actions: string[] = [];
  try {
    await h.render(<NativeSheetActions primaryLabel="Apply filters" secondaryLabel="Back" disabled onApply={() => actions.push('apply')} onBack={() => actions.push('back')} />);
    const buttons = h.allByType('SwiftUIButton');
    expect(buttons[0].props.modifiers).toContainEqual({ type: 'disabled', value: true });
    await h.press(buttons[0]);
    await h.press(buttons[1]);
    expect(actions).toEqual(['back']);
    expect(h.allText()).toContain('Back');
  } finally {
    await h.unmount();
  }
});

it('lets a measured container own keyboard avoidance without moving hosted controls again', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<NativeSheetActions primaryLabel="Apply" secondaryLabel="Back" disabled={false}
      keyboardAvoidance="container" onApply={() => {}} onBack={() => {}} />);
    expect(h.byType('SwiftUIHost')?.props.ignoreSafeArea).toBe('keyboard');
  } finally { await h.unmount(); }
});

it('blocks both native actions during a mutation and restores Cancel afterwards', async () => {
  const h = new MobileRenderHarness(); const actions: string[] = [];
  try {
    await h.render(<NativeSheetActions primaryLabel="Saving" secondaryLabel="Cancel" disabled secondaryDisabled onApply={() => actions.push('save')} onBack={() => actions.push('cancel')} />);
    const buttons = h.allByType('SwiftUIButton');
    for (const button of buttons) await h.press(button);
    expect(actions).toEqual([]);
    expect(buttons[1].props.modifiers).toContainEqual({ type: 'disabled', value: true });
    await h.render(<NativeSheetActions primaryLabel="Save" secondaryLabel="Cancel" disabled onApply={() => actions.push('save')} onBack={() => actions.push('cancel')} />);
    await h.press(h.allByType('SwiftUIButton')[1]);
    expect(actions).toEqual(['cancel']);
  } finally { await h.unmount(); }
});
