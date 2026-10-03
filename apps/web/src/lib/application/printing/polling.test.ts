import {expect,it} from 'vitest';
import {PrintStatusPoller} from './polling';
import {FakePrintPollingRuntime} from '$lib/fakes/printPollingRuntime';
it('pauses hidden views, ignores late reads, refreshes on return and cancels on teardown',async()=>{
 const runtime=new FakePrintPollingRuntime();let finish:((value:number)=>void)|undefined;let reads=0;const published:number[]=[];
 const poller=new PrintStatusPoller(runtime,()=>{reads++;return new Promise<number>(resolve=>{finish=resolve;});},value=>published.push(value),()=>{});
 poller.start();runtime.advance(2000);expect(reads).toBe(1);runtime.setVisible(false);finish!(1);await runtime.settle();expect(published).toEqual([]);runtime.advance(60000);expect(reads).toBe(1);
 runtime.setVisible(true);expect(reads).toBe(2);finish!(2);await runtime.settle();expect(published).toEqual([2]);poller.stop();runtime.advance(60000);expect(reads).toBe(2);expect(runtime.listenerCount).toBe(0);
});
it('backs off failures and never overlaps an outstanding read',async()=>{
 const runtime=new FakePrintPollingRuntime();let attempts=0;const errors:unknown[]=[];const poller=new PrintStatusPoller(runtime,async()=>{attempts++;throw new Error('Offline');},()=>{},e=>errors.push(e));
 poller.start();runtime.advance(2000);runtime.advance(60000);expect(attempts).toBe(1);await runtime.settle();expect(errors).toHaveLength(1);runtime.advance(3999);expect(attempts).toBe(1);runtime.advance(1);await runtime.settle();expect(attempts).toBe(2);runtime.advance(7999);expect(attempts).toBe(2);runtime.advance(1);await runtime.settle();expect(attempts).toBe(3);poller.stop();
});
it('fences an old scope response after a view is replaced',async()=>{
 const runtime=new FakePrintPollingRuntime();let finish:((value:string)=>void)|undefined;let oldSignal:AbortSignal|undefined;const published:string[]=[];
 const old=new PrintStatusPoller(runtime,signal=>{oldSignal=signal;return new Promise<string>(resolve=>{finish=resolve;});},value=>published.push(value),()=>{});
 old.start();runtime.advance(2000);old.stop();expect(oldSignal?.aborted).toBe(true);
 const current=new PrintStatusPoller(runtime,async()=> 'new scope',value=>published.push(value),()=>{});current.start();runtime.advance(2000);await runtime.settle();finish!('old scope');await runtime.settle();expect(published).toEqual(['new scope']);current.stop();
});
