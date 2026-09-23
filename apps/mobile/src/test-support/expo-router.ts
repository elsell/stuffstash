import { fakeNavigation, getCanGoBack, getScreenFocused, subscribeScreenFocus } from './navigation';
import { createElement, Fragment, useEffect, useSyncExternalStore } from 'react';
export function useNavigation() { return fakeNavigation; }
export const Stack = { Screen: ({ options }: { options?: { headerTitle?: () => import('react').ReactNode; headerRight?: () => import('react').ReactNode; headerLeft?: () => import('react').ReactNode; unstable_sheetFooter?: () => import('react').ReactNode } }) => { useEffect(() => { fakeNavigation.setOptions(options); }, [options]); return createElement(Fragment, null, options?.headerLeft?.(), options?.headerTitle?.(), options?.headerRight?.(), options?.unstable_sheetFooter?.()); } };
export function useFocusEffect(effect: () => void | (() => void)) {
  const focused = useSyncExternalStore(subscribeScreenFocus, getScreenFocused);
  useEffect(() => focused ? effect() : undefined, [effect, focused]);
}
export const router = {
  canGoBack: getCanGoBack,
  push: (href: unknown) => fakeNavigation.dispatch({ type: 'push', href }),
  navigate: (href: unknown) => fakeNavigation.dispatch({ type: 'navigate', href }),
  back: () => fakeNavigation.dispatch({ type: 'back' }),
  replace: (href: unknown) => fakeNavigation.dispatch({ type: 'replace', href }),
  setParams: (params: unknown) => fakeNavigation.dispatch({ type: 'setParams', params })
};
export function useRouter() { return router; }

let pathname = '/';
const pathnameListeners = new Set<() => void>();
export function setPathname(value: string) { pathname = value; pathnameListeners.forEach(listener => listener()); }
export function usePathname() {
  return useSyncExternalStore(listener => { pathnameListeners.add(listener); return () => { pathnameListeners.delete(listener); }; }, () => pathname);
}
