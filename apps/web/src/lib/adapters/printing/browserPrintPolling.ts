import type {PrintPollingRuntime} from '$lib/ports/printPolling';
export const browserPrintPolling:PrintPollingRuntime={
 visible:()=>!document.hidden,
 subscribe(listener){document.addEventListener('visibilitychange',listener);window.addEventListener('focus',listener);return()=>{document.removeEventListener('visibilitychange',listener);window.removeEventListener('focus',listener);};},
 schedule(delay,callback){const timer=window.setTimeout(callback,delay);return()=>window.clearTimeout(timer);}
};
