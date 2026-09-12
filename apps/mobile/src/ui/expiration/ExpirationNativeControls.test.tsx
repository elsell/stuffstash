import React from 'react';
import { navigationOptions } from '../../test-support/navigation';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationWorkspaceScreen } from './ExpirationWorkspaceScreen';
import { ExpirationHomeSection } from './ExpirationHomeSection';

it('uses native search options and removes loose content search/filter actions', async () => {
 const harness = new MobileRenderHarness();
 try {
  await harness.render(<ExpirationWorkspaceScreen mode="all" items={[]} loading={false} refreshing={false} hasMore={false} onMode={()=>{}} onSearch={()=>{}} onFilters={()=>{}} onRefresh={()=>{}} onMore={()=>{}} onOpenAsset={()=>{}} />);
  expect(harness.allByType('TextInput')).toHaveLength(0);
  expect(harness.allText()).not.toContain('Search');
  const screen = navigationOptions().at(-1) as {headerSearchBarOptions?: {placeholder?:string}};
  expect(screen?.headerSearchBarOptions?.placeholder).toBe('Search items');
  expect(harness.byType('FlatList')?.props.contentInsetAdjustmentBehavior).toBe('automatic');
 } finally { await harness.unmount(); }
});
it('gives Home status disclosures a separately announced count and whole-row action', async () => {
 const harness = new MobileRenderHarness(); const modes: string[] = [];
 try {
  await harness.render(<ExpirationHomeSection data={{items:[],counts:{soon:3,expired:2,all:50},timezone:'UTC'}} onOpen={mode=>modes.push(mode)} onOpenAsset={()=>{}} onRetry={()=>{}} />);
  const expired = harness.byLabel('View 2 expired items');
  expect(expired?.props.accessibilityValue).toEqual({text:'2'});
  expect(harness.allText()).not.toContain('Expired · 2');
  await harness.press(expired); expect(modes).toEqual(['expired']);
 } finally { await harness.unmount(); }
});
it('settles typed search before opening filters and does not write after navigation', async () => {
 const harness=new MobileRenderHarness(); const events:string[]=[];
 try {
  await harness.render(<ExpirationWorkspaceScreen mode="all" items={[]} loading={false} refreshing={false} hasMore={false} onMode={()=>{}} onSearch={value=>events.push(`search:${value}`)} onFilters={value=>events.push(`filters:${value}`)} onRefresh={()=>{}} onMore={()=>{}} onOpenAsset={()=>{}} />);
  const options=navigationOptions().at(-1) as {headerSearchBarOptions:{onChangeText:(event:{nativeEvent:{text:string}})=>void}};
  await harness.run(()=>options.headerSearchBarOptions.onChangeText({nativeEvent:{text:'Medicine'}}));
  await harness.press(harness.byLabel('Filter expiration items'));
  expect(events).toEqual(['search:Medicine','filters:Medicine']);
 } finally { await harness.unmount(); }
});
