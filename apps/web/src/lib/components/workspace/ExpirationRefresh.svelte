<script lang="ts">
 import { onMount } from 'svelte';
 import type { Asset } from '$lib/domain/inventory';
 let {assets, timezone, scope, onRefresh}: {assets: Asset[]; timezone?: string; scope: string; onRefresh: () => Promise<boolean>} = $props();
 onMount(() => {
  let previous = '', pending = false, retry = false;
  async function check(force = false) {
   if (document.hidden || pending) return;
   const zones = [...new Set([...(timezone ? [timezone] : []), ...assets.flatMap(asset => asset.expirationContext ? [asset.expirationContext.timezone] : [])])].sort();
   const key = JSON.stringify([scope, ...zones.map(timeZone => new Intl.DateTimeFormat('en-CA', {timeZone,year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date()))]);
   const changed = !!previous && previous !== key;
   if (!previous) previous = key;
   if (!zones.length || (!changed && !force && !retry)) return;
   pending = true;
   try { retry = !(await onRefresh()); if (!retry) previous = key; } catch { retry = true; } finally { pending = false; }
  }
  const focus = () => { void check(true); };
  void check();
  const timer = setInterval(() => { void check(); }, 60_000);
  window.addEventListener('focus', focus); document.addEventListener('visibilitychange', focus);
  return () => { clearInterval(timer); window.removeEventListener('focus', focus); document.removeEventListener('visibilitychange', focus); };
 });
</script>
