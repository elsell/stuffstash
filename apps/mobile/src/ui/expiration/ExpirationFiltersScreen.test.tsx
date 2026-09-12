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
