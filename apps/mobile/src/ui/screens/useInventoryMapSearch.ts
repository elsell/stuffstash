import { useEffect, useRef } from 'react';
/** Search yields to deliberate map navigation and never reruns for data refresh. */
export function useInventoryMapSearch(query:string,ready:boolean,onSearch:(text:string)=>void){
 const timer=useRef<ReturnType<typeof setTimeout>|undefined>(undefined);
 const completed=useRef<string|undefined>(undefined);
 const callback=useRef(onSearch);callback.current=onSearch;
 function cancel(){clearTimeout(timer.current);timer.current=undefined;completed.current=query;}
 function submit(text:string){clearTimeout(timer.current);timer.current=undefined;if(!ready){completed.current=undefined;return;}completed.current=text;callback.current(text);}
 useEffect(()=>{
  if(completed.current===query)return;
  completed.current=undefined;
  if(!ready)return;
  if(!query.trim()){submit('');return;}
  timer.current=setTimeout(()=>submit(query),300);
  return ()=>clearTimeout(timer.current);
 },[query,ready]);
 useEffect(()=>()=>clearTimeout(timer.current),[]);
 return {cancel,submit};
}
