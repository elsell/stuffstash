import { useCallback, useEffect, useLayoutEffect, useMemo, useRef } from 'react';
import { Stack, useFocusEffect } from 'expo-router';
import type { SearchBarCommands } from 'react-native-screens';
import type { HeaderOptions } from './NativeHeaderActions.types';

/** Platform search field; callers own search scheduling and result semantics. */
export function NativeNavigationSearch({query,placeholder,onChange,onSubmit,onClear,enabled=true}:{
 readonly enabled?:boolean; readonly query:string; readonly placeholder:string;
 readonly onChange:(text:string)=>void; readonly onSubmit:(text:string)=>void; readonly onClear:()=>void;
}) {
 const ref=useRef<SearchBarCommands|null>(null);
 const nativeText=useRef(query);
 const active=useRef(false);
 const current=useRef<{onChange:typeof onChange;onSubmit:typeof onSubmit;onClear:typeof onClear;enabled:boolean}|undefined>(undefined);
 useLayoutEffect(()=>{
  current.current={onChange,onSubmit,onClear,enabled};
  return()=>{current.current=undefined;};
 },[onChange,onSubmit,onClear,enabled]);
 useFocusEffect(useCallback(()=>{active.current=enabled;return()=>{active.current=false;};},[enabled]));
 useEffect(()=>{ref.current?.setText(query);},[]);
 useEffect(()=>{if(query!==nativeText.current){nativeText.current=query;ref.current?.setText(query);}},[query]);
 // Keep presentation stable while reading the current committed caller for events.
 const options=useMemo<HeaderOptions>(()=>{
 const owner=()=>active.current&&current.current?.enabled?current.current:undefined;
 function clear(){const handlers=owner();if(!handlers)return;nativeText.current='';ref.current?.clearText();handlers.onClear();}
 return {headerSearchBarOptions:enabled?{
  ref,placeholder,placement:'integratedButton',allowToolbarIntegration:false,hideWhenScrolling:false,hideNavigationBar:false,obscureBackground:false,autoCapitalize:'none',
  onChangeText:event=>{const handlers=owner();if(!handlers)return;nativeText.current=event.nativeEvent.text;if(!event.nativeEvent.text.trim())handlers.onClear();else handlers.onChange(event.nativeEvent.text);},
  onSearchButtonPress:event=>{const handlers=owner();if(!handlers)return;nativeText.current=event.nativeEvent.text;handlers.onSubmit(event.nativeEvent.text);ref.current?.blur();},
  onCancelButtonPress:clear,onClose:clear,
 }:undefined};
 },[enabled,placeholder]);
 return <Stack.Screen options={options} />;
}
