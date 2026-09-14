import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationReminderEditor, reminderSummary } from './ExpirationReminderEditor';
const policy = { enabled: true, upcoming: true, expired: true, advanceDays: 30 };
it('uses a focused timing destination instead of expanding an inline form', async () => {
 const harness = new MobileRenderHarness(); let opened = false;
 try {
  await harness.render(<ExpirationReminderEditor initialPolicy={{...policy,advanceDays:1}} onEditDays={()=>{opened=true;}} onSave={async()=>{}} />);
  expect(harness.byLabel('Before expiration')?.props.accessibilityValue?.text).toBe('1 day');
  await harness.press(harness.byLabel('Before expiration'));
  expect(opened).toBe(true);
  expect(harness.byLabel('Days before expiration')).toBeUndefined();
  expect(reminderSummary({...policy,advanceDays:1})).toBe('1 day before and when expired');
 }finally{await harness.unmount();}
});
it('lets a type override disabled defaults and restore inheritance',async()=>{
 const harness=new MobileRenderHarness(); const saved:unknown[]=[];
 try{
  await harness.render(<ExpirationReminderEditor initialPolicy={null} inheritedPolicy={{...policy,enabled:false}} onEditDays={()=>{}} onSave={async value=>{saved.push(value);}}/>);
  await harness.press(harness.byLabel('Choose reminder mode'));
  await harness.press(harness.byLabel('Custom'));
  expect(saved).toEqual([policy]);
  await harness.press(harness.byLabel('Choose reminder mode'));
  await harness.press(harness.byLabel('Use defaults'));
  expect(saved).toEqual([policy,null]);
 }finally{await harness.unmount();}
});
it('retains and retries failed changes, and can discard them',async()=>{
 const harness=new MobileRenderHarness(); const saved:unknown[]=[];let fails=true;
 try{
  await harness.render(<ExpirationReminderEditor initialPolicy={policy} onEditDays={()=>{}} onSave={async value=>{saved.push(value);if(fails)throw new Error('private');}}/>);
  await harness.run(()=>harness.byLabel('When expired')!.props.onValueChange(false));
  expect(harness.byLabel('When expired')!.props.value).toBe(false);
  expect(harness.allText().join(' ')).not.toContain('private');
  fails=false; await harness.press(harness.byLabel('Retry saving reminders'));
  expect(saved).toEqual([{...policy,expired:false},{...policy,expired:false}]);
 }finally{await harness.unmount();}
});

it('retains a failed native mode change and discards back to the saved inheritance',async()=>{
 const h=new MobileRenderHarness();const saved:unknown[]=[];
 try {
  await h.render(<ExpirationReminderEditor initialPolicy={null} inheritedPolicy={policy} onEditDays={()=>{}} onSave={async value=>{saved.push(value);throw new Error('offline');}}/>);
  await h.press(h.byLabel('Choose reminder mode'));await h.press(h.byLabel('Use defaults'));
  expect(saved).toEqual([]);
  await h.press(h.byLabel('Choose reminder mode'));await h.press(h.byLabel('Off'));
  expect(saved).toEqual([{...policy,enabled:false}]);
  expect(h.byLabel('Choose reminder mode')?.props.disabled).toBe(true);
  expect(h.allText()).toContain('Off');
  await h.press(h.byLabel('Discard reminder changes'));
  expect(h.byLabel('Choose reminder mode')?.props.disabled).toBe(false);
  expect(h.allText()).toContain('Use defaults');
  expect(saved).toHaveLength(1);
 } finally {await h.unmount();}
});
