import { useSyncExternalStore } from 'react';

let headerHeight = 144;
const listeners = new Set<() => void>();
export function setNativeHeaderHeight(height: number) {
  headerHeight = height;
  listeners.forEach(listener => listener());
}
export function useHeaderHeight() {
  return useSyncExternalStore(listener => {
    listeners.add(listener);
    return () => { listeners.delete(listener); };
  }, () => headerHeight);
}
