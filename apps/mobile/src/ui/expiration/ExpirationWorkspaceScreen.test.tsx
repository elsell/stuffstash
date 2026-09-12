import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationWorkspaceScreen } from './ExpirationWorkspaceScreen';
it('keeps empty expiration review refreshable and distinguishes filtered emptiness',async()=>{
 const harness=new MobileRenderHarness();let filters=0;
 try{
  await harness.render(<ExpirationWorkspaceScreen mode="all" items={[]} filtered loading={false} refreshing={false} hasMore={false} onMode={()=>{}} onSearch={()=>{}} onFilters={()=>{filters++}} onRefresh={()=>{}} onMore={()=>{}} onOpenAsset={()=>{}} />);
  expect(harness.allText().join(' ')).toContain('No matching expiration dates');
  await harness.press(harness.byLabel('Filter expiration items'));expect(filters).toBe(1);
 }finally{await harness.unmount()}
});
