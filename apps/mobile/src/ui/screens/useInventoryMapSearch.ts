import { useCallback, useRef } from 'react';
import { useFocusEffect } from 'expo-router';
/** Search yields to deliberate map navigation and never reruns for data refresh. */
export function useInventoryMapSearch(query:string,ready:boolean,onSearch:(text:string)=>void){
 const timer=useRef<ReturnType<typeof setTimeout>|undefined>(undefined);
 const focused=useRef(false);
 const completed=useRef<string|undefined>(undefined);
 const callback=useRef(onSearch);callback.current=onSearch;
 function cancel(){clearTimeout(timer.current);timer.current=undefined;completed.current=query;}
 function submit(text:string){if(!focused.current)return;clearTimeout(timer.current);timer.current=undefined;if(!ready){completed.current=undefined;return;}completed.current=text;callback.current(text);}
 useFocusEffect(useCallback(()=>{
  focused.current=true;
  if(completed.current!==query){
   completed.current=undefined;
   if(ready){
    if(!query.trim())submit('');
    else timer.current=setTimeout(()=>submit(query),300);
   }
  }
  return ()=>{focused.current=false;clearTimeout(timer.current);timer.current=undefined;};
 },[query,ready]));
 return {cancel,submit};
}
