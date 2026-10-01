import type {PerformanceContext, PerformanceObserver, PerformanceOutcome} from '$lib/ports/performanceObserver';

export interface VisibleImageOptions {
 source: string;
 observer: PerformanceObserver;
 surface: PerformanceContext['surface'];
 variant: PerformanceContext['variant'];
}

// Source is used only for lifecycle identity; never send it to the observer.
export function visibleImage(node: HTMLImageElement, initial: VisibleImageOptions) {
 let current = initial;
 let cleanup = observe(initial);
 function observe(options: VisibleImageOptions): () => void {
  if (!options.source) return () => {};
  let complete = false;
  let finish: (outcome: PerformanceOutcome) => void = () => {};
  try { finish = options.observer.start({operation: 'image', surface: options.surface, variant: options.variant}); } catch { /* Best effort only. */ }
  function settle(outcome: PerformanceOutcome) {
   if (complete) return;
   complete = true;
   try { finish(outcome); } catch { /* Never interfere with image display. */ }
  }
  const loaded = () => settle('success');
  const failed = () => settle('failure');
  node.addEventListener('load', loaded);
  node.addEventListener('error', failed);
  if (node.complete) settle(node.naturalWidth > 0 ? 'success' : 'failure');
  return () => {
   node.removeEventListener('load', loaded);
   node.removeEventListener('error', failed);
   settle('cancelled');
  };
 }
 return {
  update(next: VisibleImageOptions) {
   if (next.source === current.source && next.observer === current.observer && next.surface === current.surface && next.variant === current.variant) return;
   cleanup(); current = next; cleanup = observe(next);
  },
  destroy() { cleanup(); }
 };
}
