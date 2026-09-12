import React from 'react';
import { expect, it } from 'vitest';
import { Platform } from 'react-native';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationDateRange } from './ExpirationDateRange';
it('opens only the chosen Android date dialog and allows reopening after dismissal', async()=>{
 const harness=new MobileRenderHarness();const original=Platform.OS;Object.assign(Platform,{OS:'android'});
 try{
  await harness.render(<ExpirationDateRange fromDate="2026-09-12" throughDate="2026-10-15" onChange={()=>{}} />);
  expect(harness.byLabel('First expiration date')).toBeUndefined();expect(harness.byLabel('Last expiration date')).toBeUndefined();
  await harness.press(harness.byLabel('Change last expiration date'));
  const picker=harness.byLabel('Last expiration date'); expect(picker).toBeDefined();expect(harness.byLabel('First expiration date')).toBeUndefined();
  await harness.run(()=>picker?.props.onChange({type:'dismissed'}));expect(harness.byLabel('Last expiration date')).toBeUndefined();
  await harness.press(harness.byLabel('Change last expiration date'));expect(harness.byLabel('Last expiration date')).toBeDefined();
 }finally{await harness.unmount();Object.assign(Platform,{OS:original});}
});
