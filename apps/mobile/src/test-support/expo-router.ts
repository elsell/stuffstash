import { fakeNavigation } from './navigation';
import { useEffect } from 'react';
export function useNavigation() { return fakeNavigation; }
export const Stack = { Screen: () => null };
export function useFocusEffect(effect: () => void | (() => void)) { useEffect(effect, [effect]); }
export const router = {
  push: (href: unknown) => fakeNavigation.dispatch({ type: 'push', href }),
  navigate: (href: unknown) => fakeNavigation.dispatch({ type: 'navigate', href }),
  back: () => fakeNavigation.dispatch({ type: 'back' }),
  replace: (href: unknown) => fakeNavigation.dispatch({ type: 'replace', href }),
  setParams: (params: unknown) => fakeNavigation.dispatch({ type: 'setParams', params })
};
