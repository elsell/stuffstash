import type {PrintPollingRuntime} from '$lib/ports/printPolling';
import type {PrintJob} from '$lib/domain/printing';
const INITIAL_POLL_MS=2000;
const MAX_POLL_MS=30000;
export function printJobNeedsPolling(job:PrintJob|null){return Boolean(job&&!['completed','failed','canceled'].includes(job.status));}
export class PrintStatusPoller<T> {
    private timer?:()=>void;
    private unsubscribe?:()=>void;
    private controller?:AbortController;
    private generation=0;
    private failures=0;
    private pending=false;
    private stopped=false;
    private refreshRequested=false;
    constructor(private readonly runtime:PrintPollingRuntime,private readonly read:(signal:AbortSignal)=>Promise<T>,private readonly publish:(value:T)=>void,private readonly failed:(error:unknown)=>void){}
    start(){if(this.unsubscribe||this.stopped)return;this.unsubscribe=this.runtime.subscribe(()=>this.visibilityChanged());this.schedule();}
    stop(){this.stopped=true;this.invalidate();this.unsubscribe?.();this.unsubscribe=undefined;}
    private invalidate(){this.generation++;this.timer?.();this.timer=undefined;this.controller?.abort();this.refreshRequested=false;}
    private visibilityChanged(){
        this.invalidate();
        if(!this.stopped&&this.runtime.visible()){
            this.refreshRequested=true;
            void this.poll();
        }
    }
    private schedule(){
        if(this.stopped||!this.runtime.visible())return;
        this.timer?.();
        this.timer=this.runtime.schedule(Math.min(MAX_POLL_MS,INITIAL_POLL_MS*2**Math.min(this.failures,4)),()=>{this.timer=undefined;void this.poll();});
    }
    private async poll(){
        if(this.pending||this.stopped||!this.runtime.visible())return;
        this.refreshRequested=false;this.pending=true;
        const generation=this.generation;const controller=new AbortController();this.controller=controller;
        const current=()=>!this.stopped&&!controller.signal.aborted&&generation===this.generation&&this.runtime.visible();
        try{const value=await this.read(controller.signal);if(current()){this.failures=0;this.publish(value);}}
        catch(error){if(current()){this.failures++;this.failed(error);}}
        finally{
            this.pending=false;
            if(this.controller===controller)this.controller=undefined;
            if(this.refreshRequested&&!this.stopped&&this.runtime.visible())void this.poll();
            else if(current())this.schedule();
        }
    }
}
