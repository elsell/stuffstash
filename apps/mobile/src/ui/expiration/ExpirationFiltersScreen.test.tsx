import { navigationOptions, resetNavigation } from '../../test-support/navigation';
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

it('chooses availability in place and only commits with Apply', async () => {
 const h = new MobileRenderHarness(); const applied: unknown[] = [];
 try {
  await h.render(<ExpirationFiltersScreen initial={{mode:'all'}} choices={{types:[],tags:[],locations:[]}} onApply={value=>applied.push(value)} onCancel={()=>{}} />);
  await h.press(h.byLabel('Choose availability'));
  expect(h.byLabel('Choose date range')).toBeDefined();
  await h.press(h.byLabel('Checked out'));
  expect(h.byLabel('Choose date range')).toBeDefined();
  expect(applied).toEqual([]);
  await h.press(h.byLabel('Apply expiration filters'));
  expect(applied).toEqual([{mode:'all',checkoutState:'checked_out'}]);
 } finally { await h.unmount(); }
});

it('exposes tags as independent checkbox selections and applies the remaining draft', async () => {
 const h = new MobileRenderHarness(); const applied: unknown[] = [];
 try {
  await h.render(<ExpirationFiltersScreen initial={{mode:'all'}} choices={{types:[],tags:[{id:'one',label:'Medicine'},{id:'two',label:'Travel'}],locations:[]}} onApply={value=>applied.push(value)} onCancel={()=>{}} />);
  await h.press(h.byLabel('Choose tags'));
  expect(h.byLabel('Medicine')?.props.accessibilityRole).toBe('checkbox');
  await h.press(h.byLabel('Medicine')); await h.press(h.byLabel('Travel'));
  expect(h.byLabel('Medicine')?.props.accessibilityState.checked).toBe(true);
  expect(h.byLabel('Travel')?.props.accessibilityState.checked).toBe(true);
  await h.press(h.byLabel('Medicine'));
  expect(h.byLabel('Travel')?.props.accessibilityState.checked).toBe(true);
  expect(applied).toEqual([]);
  await h.press(h.byLabel('Apply expiration filters'));
  expect(applied).toEqual([{mode:'all',tagIds:['two']}]);
 } finally { await h.unmount(); }
});

it('clears compact selection search without losing staged tags and resets it on Back', async () => {
 resetNavigation(); const h = new MobileRenderHarness(); const applied: unknown[] = [];
 type Search = { placement: string; onChangeText: (event: { nativeEvent: { text: string } }) => void; onClose: () => void };
 const search = () => (Object.assign({}, ...navigationOptions()) as { headerSearchBarOptions?: Search }).headerSearchBarOptions!;
 try {
  await h.render(<ExpirationFiltersScreen initial={{mode:'all'}} choices={{types:[],locations:[],tags:[{id:'one',label:'Medicine'},{id:'two',label:'Travel'}]}} onApply={value=>applied.push(value)} onCancel={()=>{}} />);
  await h.press(h.byLabel('Choose tags'));
  await h.press(h.byLabel('Medicine'));
  expect(search().placement).toBe('integratedButton');
  await h.run(()=>search().onChangeText({nativeEvent:{text:'Travel'}}));
  expect(h.byLabel('Medicine')).toBeUndefined();
  await h.run(()=>search().onClose());
  expect(h.byLabel('Medicine')?.props.accessibilityState.checked).toBe(true);
  await h.run(()=>search().onChangeText({nativeEvent:{text:'missing'}}));
  await h.press(h.byLabel('Cancel or return to filters'));
  expect(search()).toBeUndefined();
  await h.press(h.byLabel('Choose tags'));
  expect(search().placement).toBe('integratedButton');
  expect(h.byLabel('Medicine')?.props.accessibilityState.checked).toBe(true);
  expect(h.byLabel('Travel')).toBeDefined();
  expect(applied).toEqual([]);
  await h.press(h.byLabel('Apply expiration filters'));
  expect(applied).toEqual([{mode:'all',tagIds:['one']}]);
 } finally { await h.unmount(); resetNavigation(); }
});
