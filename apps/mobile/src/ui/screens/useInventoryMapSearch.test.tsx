import { setScreenFocused } from '../../test-support/navigation';
import React from 'react';
import { expect,it,vi } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { useInventoryMapSearch } from './useInventoryMapSearch';
it('waits for data, yields to manual navigation, and ignores refreshed map callbacks',async()=>{
 vi.useFakeTimers();const h=new MobileRenderHarness();const found:string[]=[];let controls:ReturnType<typeof useInventoryMapSearch>;
 function Surface({query,ready}:{query:string;ready:boolean}){controls=useInventoryMapSearch(query,ready,text=>found.push(text));return null;}
 try{
  await h.render(<Surface query="tent" ready={false}/>);await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual([]);
  await h.render(<Surface query="tent" ready/>);await h.run(()=>controls.cancel());await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual([]);
  await h.render(<Surface query="tent" ready/>);await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual([]);
  await h.render(<Surface query="drill" ready/>);await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual(['drill']);
  await h.run(()=>controls.submit('box'));await h.render(<Surface query="box" ready/>);await h.run(()=>controls.cancel());await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual(['drill','box']);
  await h.render(<Surface query="" ready/>);expect(found.at(-1)).toBe('');
 }finally{await h.unmount();vi.useRealTimers();}
});
it('retains keyboard submission until the map is ready',async()=>{
 vi.useFakeTimers();const h=new MobileRenderHarness();const found:string[]=[];let controls:ReturnType<typeof useInventoryMapSearch>;
 function Surface({ready}:{ready:boolean}){controls=useInventoryMapSearch('tent',ready,text=>found.push(text));return null;}
 try{
  await h.render(<Surface ready={false}/>);await h.run(()=>controls.submit('tent'));expect(found).toEqual([]);
  await h.render(<Surface ready/>);await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual(['tent']);
 }finally{await h.unmount();vi.useRealTimers();}
});

it('pauses map path search while hidden and preserves deliberate cancellation on return',async()=>{
 vi.useFakeTimers();setScreenFocused(true);const h=new MobileRenderHarness();const found:string[]=[];let controls!:ReturnType<typeof useInventoryMapSearch>;
 function Surface({query}:{query:string}){controls=useInventoryMapSearch(query,true,text=>found.push(text));return null;}
 try{
  await h.render(<Surface query="tent"/>);await h.run(()=>setScreenFocused(false));
  await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual([]);
  await h.run(()=>controls.submit('hidden'));expect(found).toEqual([]);
  await h.run(()=>setScreenFocused(true));await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual(['tent']);
  await h.render(<Surface query="drill"/>);await h.run(()=>controls.cancel());
  await h.run(()=>setScreenFocused(false));await h.run(()=>setScreenFocused(true));
  await h.run(()=>vi.advanceTimersByTime(300));expect(found).toEqual(['tent']);
 }finally{await h.unmount();setScreenFocused(true);vi.useRealTimers();}
});
