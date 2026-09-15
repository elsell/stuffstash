import { useCallback, useEffect, useRef } from 'react';
import { Stack, useFocusEffect } from 'expo-router';
import type { SearchBarCommands } from 'react-native-screens';

/** Platform search field; callers own search scheduling and result semantics. */
export function NativeNavigationSearch({query,placeholder,onChange,onSubmit,onClear,enabled=true}:{
 readonly enabled?:boolean; readonly query:string; readonly placeholder:string;
 readonly onChange:(text:string)=>void; readonly onSubmit:(text:string)=>void; readonly onClear:()=>void;
}) {
 const ref=useRef<SearchBarCommands|null>(null);
 const nativeText=useRef(query);
 const active=useRef(false);
 useFocusEffect(useCallback(()=>{active.current=enabled;return()=>{active.current=false;};},[enabled]));
 useEffect(()=>{ref.current?.setText(query);},[]);
 useEffect(()=>{if(query!==nativeText.current){nativeText.current=query;ref.current?.setText(query);}},[query]);
 function clear(){if(!active.current)return;nativeText.current='';ref.current?.clearText();onClear();}
 return <Stack.Screen options={{headerSearchBarOptions:enabled?{
  ref,placeholder,placement:'integratedButton',allowToolbarIntegration:false,hideWhenScrolling:false,hideNavigationBar:false,obscureBackground:false,autoCapitalize:'none',
  onChangeText:event=>{if(!active.current)return;nativeText.current=event.nativeEvent.text;if(!event.nativeEvent.text.trim())onClear();else onChange(event.nativeEvent.text);},
  onSearchButtonPress:event=>{if(!active.current)return;nativeText.current=event.nativeEvent.text;onSubmit(event.nativeEvent.text);ref.current?.blur();},
  onCancelButtonPress:clear,onClose:clear,
 }:undefined}} />;
}
