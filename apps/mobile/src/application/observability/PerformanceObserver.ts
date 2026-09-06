export interface PerformanceContext {
  operation: 'request' | 'image';
  surface: 'application' | 'home' | 'list' | 'detail' | 'gallery' | 'fullscreen' | 'upload';
  variant: 'none' | 'small' | 'medium' | 'large' | 'original';
}
export type PerformanceOutcome = 'success' | 'failure' | 'cancelled';
export interface PerformanceObserver {
  start(context: PerformanceContext): (outcome: PerformanceOutcome) => void;
}
export const noopPerformanceObserver: PerformanceObserver = { start: () => () => {} };
