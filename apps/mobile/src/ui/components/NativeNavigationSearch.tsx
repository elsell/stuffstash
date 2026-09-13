import { useEffect, useRef } from 'react';
import { Stack } from 'expo-router';
import type { SearchBarCommands } from 'react-native-screens';

/** Platform search field; callers own search scheduling and result semantics. */
export function NativeNavigationSearch({query,placeholder,onChange,onSubmit,onClear}:{
 readonly query:string; readonly placeholder:string;
 readonly onChange:(text:string)=>void; readonly onSubmit:(text:string)=>void; readonly onClear:()=>void;
}) {
 const ref=useRef<SearchBarCommands|null>(null);
 const nativeText=useRef(query);
 useEffect(()=>{ref.current?.setText(query);},[]);
 useEffect(()=>{if(query!==nativeText.current){nativeText.current=query;ref.current?.setText(query);}},[query]);
 function clear(){nativeText.current='';ref.current?.clearText();onClear();}
 return <Stack.Screen options={{headerSearchBarOptions:{
  ref,placeholder,placement:'stacked',hideWhenScrolling:false,hideNavigationBar:false,obscureBackground:false,autoCapitalize:'none',
  onChangeText:event=>{nativeText.current=event.nativeEvent.text;if(!event.nativeEvent.text.trim())onClear();else onChange(event.nativeEvent.text);},
  onSearchButtonPress:event=>{nativeText.current=event.nativeEvent.text;onSubmit(event.nativeEvent.text);ref.current?.blur();},
  onCancelButtonPress:clear,onClose:clear,
 }}} />;
}
