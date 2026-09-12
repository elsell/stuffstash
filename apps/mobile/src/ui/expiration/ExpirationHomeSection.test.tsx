import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationHomeSection } from './ExpirationHomeSection';
it('opens each count in its own mode and See all in all dates',async()=>{
 const harness=new MobileRenderHarness();const modes:string[]=[];
 try {
  await harness.render(<ExpirationHomeSection data={{items:[],counts:{soon:3,expired:2,all:50},timezone:'UTC'}} onOpen={mode=>modes.push(mode)} onOpenAsset={()=>{}} onRetry={()=>{}} />);
  await harness.press(harness.byLabel('View 2 expired items'));await harness.press(harness.byLabel('View 3 items expiring soon'));await harness.press(harness.byLabel('View all expiration dates'));
  expect(modes).toEqual(['expired','soon','all']);
 }finally{await harness.unmount()}
});
it('shows query failure instead of claiming that no items expire',async()=>{
 const harness=new MobileRenderHarness();let retries=0;
 try{
  await harness.render(<ExpirationHomeSection error="Could not load expiration" onOpen={()=>{}} onOpenAsset={()=>{}} onRetry={()=>{retries++}} />);
  expect(harness.allText().join(' ')).toContain('Could not load expiration');
  expect(harness.allText().join(' ')).not.toContain('None expiring soon');
  await harness.press(harness.byLabel('Retry expiration'));expect(retries).toBe(1);
 }finally{await harness.unmount()}
});
