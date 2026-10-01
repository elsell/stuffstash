import {getContext, hasContext} from 'svelte';
import {noopPerformanceObserver, performanceObserverContext, type PerformanceObserver} from '$lib/ports/performanceObserver';

export function imageObserverContext(): PerformanceObserver {
 return hasContext(performanceObserverContext) ? getContext<PerformanceObserver>(performanceObserverContext) : noopPerformanceObserver;
}
