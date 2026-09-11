import { expect,it } from 'vitest';
import { PushReconciliationController } from './PushReconciliationController';
it('coalesces events, reruns after changes during reconciliation, and cancels on stop',async()=>{
 let trigger=()=>{};let removed=false;const pending: (()=>void)[]=[];const signals:AbortSignal[]=[];
 const controller=new PushReconciliationController({subscribe(listener){trigger=listener;return ()=>{removed=true;};}}, {async reconcile(request){signals.push(request.signal!);await new Promise<void>(resolve=>pending.push(resolve));}},{record(){}});
 const stop=controller.start();expect(signals).toHaveLength(1);
 trigger();trigger();expect(signals).toHaveLength(1);
 pending.shift()!();await new Promise(resolve=>setTimeout(resolve,0));expect(signals).toHaveLength(2);
 stop();expect(removed).toBe(true);expect(signals[1].aborted).toBe(true);
 pending.shift()!();await new Promise(resolve=>setTimeout(resolve,0));trigger();expect(signals).toHaveLength(2);
});
