<script lang="ts" generics="T">
import {untrack} from 'svelte';
import {PrintStatusPoller} from '$lib/application/printing/polling';
import {browserPrintPolling} from '$lib/adapters/printing/browserPrintPolling';
import type {PrintPollingRuntime} from '$lib/ports/printPolling';
let {identity,enabled,read,publish,failed,runtime=browserPrintPolling}:{identity:string;enabled:boolean;read:(signal:AbortSignal)=>Promise<T>;publish:(value:T)=>void;failed:(error:unknown)=>void;runtime?:PrintPollingRuntime}=$props();
$effect(()=>{
 identity;
 if(!enabled)return;
 const scheduler=runtime;
 const poller=untrack(()=>new PrintStatusPoller(scheduler,signal=>read(signal),value=>publish(value),error=>failed(error)));
 poller.start();return()=>poller.stop();
});
</script>
