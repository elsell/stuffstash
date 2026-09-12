import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationFiltersScreen } from './ExpirationFiltersScreen';
it('stages type selection until Apply and clears existing filters explicitly',async()=>{
 const harness=new MobileRenderHarness();const applied:unknown[]=[];
 try{
  await harness.render(<ExpirationFiltersScreen initial={{mode:'all',typeId:'type'}} choices={{types:[{id:'type',label:'Medicine'}],tags:[],locations:[]}} onApply={filter=>applied.push(filter)} onCancel={()=>{}} />);
  await harness.press(harness.byLabel('Clear expiration filters'));expect(applied).toHaveLength(0);
  await harness.press(harness.byLabel('Apply expiration filters'));expect(applied).toEqual([{mode:'all'}]);
 }finally{await harness.unmount()}
});
it('keeps sheet actions outside the scroll area and preserves staged choices on Back',async()=>{
 const harness=new MobileRenderHarness();let cancelled=0;const applied:unknown[]=[];
 try{
  await harness.render(<ExpirationFiltersScreen initial={{mode:'all'}} choices={{types:[],tags:[],locations:[]}} onApply={filter=>applied.push(filter)} onCancel={()=>{cancelled++}} />);
  expect(harness.byTestId('expiration-filter-footer')).toBeDefined();
  await harness.press(harness.byLabel('Choose date range'));
  expect(harness.allByType('DateTimePicker')).toHaveLength(0);
  const enable=harness.byLabel('From date');
  await harness.run(()=>enable?.props.onValueChange(true));
  expect(harness.byLabel('First expiration date')?.props.display).not.toBe('inline');
  await harness.press(harness.byLabel('Cancel or return to filters'));
  expect(cancelled).toBe(0);expect(applied).toEqual([]);
  await harness.press(harness.byLabel('Apply expiration filters'));
  expect(applied).toHaveLength(1);
 }finally{await harness.unmount()}
});
