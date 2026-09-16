import { navigationOptions, resetNavigation } from '../../test-support/navigation';
import React from 'react';
import { expect, it, vi } from 'vitest';
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

it('uses compact native search and carries pending text into filters before debounce', async () => {
 resetNavigation(); vi.useFakeTimers(); const h=new MobileRenderHarness(); const queries:string[]=[]; const filters:string[]=[];
 try {
  await h.render(<ExpirationWorkspaceScreen mode="all" items={[]} loading={false} refreshing={false} hasMore={false} onMode={()=>{}} onSearch={value=>queries.push(value)} onFilters={value=>filters.push(value)} onRefresh={()=>{}} onMore={()=>{}} onOpenAsset={()=>{}} />);
  const options=()=> (Object.assign({}, ...navigationOptions()) as {headerSearchBarOptions?:{placement:string;allowToolbarIntegration:boolean;onChangeText:(event:{nativeEvent:{text:string}})=>void;onClose:()=>void}}).headerSearchBarOptions!;
  expect(options().placement).toBe('integratedButton');
  expect(options().allowToolbarIntegration).toBe(false);
  await h.run(()=>options().onChangeText({nativeEvent:{text:' medicine '}}));
  expect(queries).toEqual([]);
  await h.press(h.byLabel('Filter expiration items'));
  expect(filters).toEqual(['medicine']); expect(queries).toEqual(['medicine']);
  await h.run(()=>options().onClose());
  await h.run(()=>vi.advanceTimersByTime(300));
  expect(queries).toEqual(['medicine','']);
 } finally { await h.unmount(); resetNavigation(); vi.useRealTimers(); }
});

it('routes error recovery separately from pull refresh and supports returning Home', async () => {
 const h=new MobileRenderHarness(); let retry=0; let pulls=0; let home=0;
 const props={mode:'all' as const,items:[],loading:false,refreshing:false,hasMore:false,error:'Unavailable',onMode:()=>{},onSearch:()=>{},onFilters:()=>{},onRefresh:()=>{pulls++},onMore:()=>{},onOpenAsset:()=>{}};
 try {
  await h.render(<ExpirationWorkspaceScreen {...props} recovery={{label:'Retry expiration',onPress:()=>{retry++}}} />);
  await h.press(h.byLabel('Retry expiration'));
  expect(retry).toBe(1); expect(pulls).toBe(0);
  await h.render(<ExpirationWorkspaceScreen {...props} recovery={{label:'Retrying expiration',disabled:true,onPress:()=>{retry++}}} />);
  expect(h.byLabel('Retrying expiration')?.props.accessibilityState.disabled).toBe(true);
  await h.render(<ExpirationWorkspaceScreen {...props} recovery={{label:'Return to Home',onPress:()=>{home++}}} />);
  expect(h.byLabel('Retry expiration')).toBeUndefined();
  await h.press(h.byLabel('Return to Home'));
  expect(home).toBe(1); expect(pulls).toBe(0);
 } finally { await h.unmount(); }
});
