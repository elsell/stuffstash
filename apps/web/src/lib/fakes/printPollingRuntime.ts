import type {PrintPollingRuntime} from '$lib/ports/printPolling';
export class FakePrintPollingRuntime implements PrintPollingRuntime {
    private now=0;
    private isVisible=true;
    private readonly listeners=new Set<()=>void>();
    private readonly timers=new Set<{at:number;callback:()=>void}>();
    visible(){return this.isVisible;}
    get listenerCount(){return this.listeners.size;}
    subscribe(listener:()=>void){this.listeners.add(listener);return()=>{this.listeners.delete(listener);};}
    schedule(delay:number,callback:()=>void){const timer={at:this.now+delay,callback};this.timers.add(timer);return()=>{this.timers.delete(timer);};}
    setVisible(value:boolean){this.isVisible=value;for(const listener of this.listeners)listener();}
    advance(duration:number){
        const until=this.now+duration;
        while(true){const next=[...this.timers].filter(t=>t.at<=until).sort((a,b)=>a.at-b.at)[0];if(!next)break;this.now=next.at;this.timers.delete(next);next.callback();}
        this.now=until;
    }
    async settle(){for(let i=0;i<8;i++)await Promise.resolve();}
}
