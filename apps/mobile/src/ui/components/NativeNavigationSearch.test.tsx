import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { navigationOptions, resetNavigation } from '../../test-support/navigation';
import { NativeNavigationSearch } from './NativeNavigationSearch';
it('uses native search callbacks, submits current event text and clears immediately',async()=>{
 resetNavigation();const h=new MobileRenderHarness();const events:string[]=[];
 try{
  await h.render(<NativeNavigationSearch query="saved" placeholder="Search inventory" onChange={q=>events.push(`change:${q}`)} onSubmit={q=>events.push(`submit:${q}`)} onClear={()=>events.push('clear')} />);
  const options=navigationOptions().at(-1) as {headerSearchBarOptions:{placement:string;hideWhenScrolling:boolean;onChangeText:(event:{nativeEvent:{text:string}})=>void;onSearchButtonPress:(event:{nativeEvent:{text:string}})=>void;onCancelButtonPress:()=>void}};const search=options.headerSearchBarOptions;
  expect(search.placement).toBe('stacked');expect(search.hideWhenScrolling).toBe(false);
  await h.run(()=>search.onChangeText({nativeEvent:{text:'new'}}));
  await h.run(()=>search.onSearchButtonPress({nativeEvent:{text:'newest'}}));
  await h.run(()=>search.onChangeText({nativeEvent:{text:''}}));
  await h.run(()=>search.onCancelButtonPress());
  expect(events).toEqual(['change:new','submit:newest','clear','clear']);
  expect(h.allByType('TextInput')).toHaveLength(0);
 }finally{await h.unmount()}
});
