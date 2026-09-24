import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { navigationOptions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { NativeNavigationSearch } from './NativeNavigationSearch';
import { NavigationOptionFeedback } from '../../test-support/NavigationOptionFeedback';

it('settles navigation updates while keeping current handlers and presentation changes', async () => {
 resetNavigation(); const h=new MobileRenderHarness(); const events:string[]=[];
 const render=(version:string,enabled=true,placeholder='Search')=><NavigationOptionFeedback render={()=><NativeNavigationSearch
   query={version} enabled={enabled} placeholder={placeholder} onChange={text=>events.push(`${version}:${text}`)}
   onSubmit={text=>events.push(`${version}:submit:${text}`)} onClear={()=>events.push(`${version}:clear`)} />} />;
 type Search={onFocus:()=>void;onChangeText:(event:{nativeEvent:{text:string}})=>void;onSearchButtonPress:(event:{nativeEvent:{text:string}})=>void;onClose:()=>void;placeholder:string};
 const options=()=>navigationOptions().at(-1) as {headerSearchBarOptions:Search|undefined};
 try {
  await h.render(render('first'));
  const original=options().headerSearchBarOptions!;
  await h.render(render('latest'));
  expect(options().headerSearchBarOptions).toBe(original);
  await h.run(()=>{original.onFocus();original.onChangeText({nativeEvent:{text:'draft'}});original.onSearchButtonPress({nativeEvent:{text:'draft'}});original.onClose();});
  expect(events).toEqual(['latest:draft','latest:submit:draft','latest:clear']);
  await h.render(render('disabled',false));
  expect(options().headerSearchBarOptions).toBeUndefined();
  await h.run(()=>{original.onChangeText({nativeEvent:{text:'late'}});original.onClose();});
  expect(events).toHaveLength(3);
  await h.render(render('returned',true,'Find a place'));
  expect(options().headerSearchBarOptions?.placeholder).toBe('Find a place');
  await h.run(()=>{options().headerSearchBarOptions?.onFocus();options().headerSearchBarOptions?.onChangeText({nativeEvent:{text:'box'}});});
  expect(events.at(-1)).toBe('returned:box');
 } finally {await h.unmount();resetNavigation();}
});
it('uses native search callbacks, submits current event text and clears immediately',async()=>{
 resetNavigation();const h=new MobileRenderHarness();const events:string[]=[];
 try{
  await h.render(<NativeNavigationSearch query="saved" placeholder="Search inventory" onChange={q=>events.push(`change:${q}`)} onSubmit={q=>events.push(`submit:${q}`)} onClear={()=>events.push('clear')} />);
  const options=navigationOptions().at(-1) as {headerSearchBarOptions:{placement:string;hideWhenScrolling:boolean;allowToolbarIntegration:boolean;onFocus:()=>void;onChangeText:(event:{nativeEvent:{text:string}})=>void;onSearchButtonPress:(event:{nativeEvent:{text:string}})=>void;onCancelButtonPress:()=>void}};const search=options.headerSearchBarOptions;
  expect(search.placement).toBe('integratedButton');expect(search.allowToolbarIntegration).toBe(false);expect(search.hideWhenScrolling).toBe(false);
  await h.run(()=>{search.onFocus();search.onChangeText({nativeEvent:{text:'new'}});});
  await h.run(()=>search.onSearchButtonPress({nativeEvent:{text:'newest'}}));
  await h.run(()=>search.onChangeText({nativeEvent:{text:''}}));
  await h.run(()=>search.onCancelButtonPress());
  expect(events).toEqual(['change:new','submit:newest','clear','clear']);
  expect(h.allByType('TextInput')).toHaveLength(0);
  await h.render(<></>);
  await h.run(()=>{search.onChangeText({nativeEvent:{text:'late'}});search.onSearchButtonPress({nativeEvent:{text:'late'}});search.onCancelButtonPress();});
  expect(events).toEqual(['change:new','submit:newest','clear','clear']);
 }finally{await h.unmount()}
});

it('ignores hidden native callbacks and accepts input again after return', async () => {
 resetNavigation();setScreenFocused(true);const h=new MobileRenderHarness();const values:string[]=[];
 try {
  await h.render(<NativeNavigationSearch query="saved" placeholder="Search" onChange={value=>values.push(value)} onSubmit={value=>values.push(value)} onClear={()=>values.push('clear')} />);
  const search=(navigationOptions().at(-1) as {headerSearchBarOptions:{onFocus:()=>void;onChangeText:(event:{nativeEvent:{text:string}})=>void;onSearchButtonPress:(event:{nativeEvent:{text:string}})=>void;onClose:()=>void}}).headerSearchBarOptions;
  await h.run(()=>setScreenFocused(false));
  await h.run(()=>{search.onChangeText({nativeEvent:{text:'hidden'}});search.onSearchButtonPress({nativeEvent:{text:'hidden'}});search.onClose();});
  expect(values).toEqual([]);
  await h.run(()=>setScreenFocused(true));
  await h.run(()=>{search.onFocus();search.onChangeText({nativeEvent:{text:'returned'}});});
  expect(values).toEqual(['returned']);
 }finally{await h.unmount();resetNavigation();setScreenFocused(true);}
});

it('seeds a newly enabled native field without rewriting native edit echoes', async () => {
 resetNavigation(); const h = new MobileRenderHarness();
 const { subscribeNavigationOptions } = await import('../../test-support/navigation');
 const writes:string[]=[];
 let ref: { current: { setText: (text:string)=>void } | null } | undefined;
 const unsubscribe = subscribeNavigationOptions(() => {
  const options = navigationOptions().at(-1) as {headerSearchBarOptions?:{ref:NonNullable<typeof ref>}};
  if (ref) ref.current = null;
  ref = options.headerSearchBarOptions?.ref;
  if (ref) ref.current = {setText:text=>writes.push(text)};
 });
 const render=(enabled:boolean,query:string)=><NativeNavigationSearch enabled={enabled} query={query}
  placeholder="Search" onChange={()=>{}} onSubmit={()=>{}} onClear={()=>{}} />;
 try {
  await h.render(render(false,'retained')); expect(writes).toEqual([]);
  await h.render(render(true,'retained')); expect(writes).toEqual(['retained']);
  const search=(navigationOptions().at(-1) as {headerSearchBarOptions:{onFocus:()=>void;onChangeText:(event:{nativeEvent:{text:string}})=>void}}).headerSearchBarOptions;
  await h.run(()=>{search.onFocus();search.onChangeText({nativeEvent:{text:'typing'}});});
  await h.render(render(true,'typing')); expect(writes).toEqual(['retained','retained']);
  await h.render(render(false,'replacement')); expect(writes).toEqual(['retained','retained']);
  await h.render(render(true,'replacement')); expect(writes).toEqual(['retained','retained','replacement']);
 } finally {await h.unmount();unsubscribe();resetNavigation();}
});

it('preserves retained search through inactive header events before and after navigation', async () => {
 resetNavigation(); const h = new MobileRenderHarness(); const events: string[] = [];
 try {
  await h.render(<NativeNavigationSearch query="Camping" placeholder="Search" onChange={text=>events.push(text)} onSubmit={text=>events.push(text)} onClear={()=>events.push('clear')} />);
  const search = (navigationOptions().at(-1) as {headerSearchBarOptions:{onChangeText:(event:{nativeEvent:{text:string}})=>void;onClose:()=>void}}).headerSearchBarOptions;
  await h.run(()=>{search.onChangeText({nativeEvent:{text:''}});search.onClose();});
  expect(events).toEqual([]);
  const focusedSearch = search as typeof search & {onFocus:()=>void};
  await h.run(()=>{focusedSearch.onFocus();search.onChangeText({nativeEvent:{text:'draft'}});});
  expect(events).toEqual(['draft']);
  await h.run(()=>setScreenFocused(false));
  await h.run(()=>setScreenFocused(true));
  await h.run(()=>{search.onChangeText({nativeEvent:{text:''}});search.onClose();});
  expect(events).toEqual(['draft']);
 } finally {await h.unmount();resetNavigation();}
});

it('allows selection tasks to show stacked native search without changing the default', async () => {
 resetNavigation(); const h = new MobileRenderHarness();
 try {
  await h.render(<NativeNavigationSearch placement="stacked" query="" placeholder="Find a destination" onChange={()=>{}} onSubmit={()=>{}} onClear={()=>{}} />);
  expect((navigationOptions().at(-1) as {headerSearchBarOptions:{placement:string}}).headerSearchBarOptions.placement).toBe('stacked');
 } finally { await h.unmount(); resetNavigation(); }

});


it('updates Android search colors without replacing the retained query', async () => {
 const { AppearanceProvider } = await import('../theme/AppearanceContext');
 const { AppearancePreferenceController } = await import('../../application/settings/AppearancePreference');
 const { setSystemColorSchemeForTest, resetNativeTestState } = await import('../../test-support/react-native');
 const { darkPalette, lightPalette } = await import('../theme/tokens');
 const { Platform } = await import('react-native');
 const original = Platform.OS; (Platform as { OS: string }).OS = 'android';
 const controller = new AppearancePreferenceController({ load: async () => 'system', save: async () => {} });
 const h = new MobileRenderHarness(); const events: string[] = [];
 const render = () => <AppearanceProvider controller={controller}><NativeNavigationSearch query="Camping" placeholder="Search"
   onChange={value=>events.push(value)} onSubmit={()=>{}} onClear={()=>events.push('clear')} /></AppearanceProvider>;
 const options = () => (navigationOptions().at(-1) as {headerSearchBarOptions:{textColor:string;hintTextColor:string;headerIconColor:string;onFocus:()=>void;onChangeText:(event:{nativeEvent:{text:string}})=>void}}).headerSearchBarOptions;
 try {
  resetNavigation(); setSystemColorSchemeForTest('light'); await h.render(render());
  expect(options().textColor).toBe(lightPalette.text);
  await h.run(()=>options().onFocus());
  setSystemColorSchemeForTest('dark'); await h.render(render());
  expect(options()).toMatchObject({textColor:darkPalette.text,headerIconColor:darkPalette.text,hintTextColor:darkPalette.textMuted});
  expect(events).toEqual([]);
  await h.run(()=>options().onChangeText({nativeEvent:{text:'Camping gear'}}));
  expect(events).toEqual(['Camping gear']);
 } finally { await h.unmount(); (Platform as { OS: string }).OS = original; resetNativeTestState(); resetNavigation(); }
});
