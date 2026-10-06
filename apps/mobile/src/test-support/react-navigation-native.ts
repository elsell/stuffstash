import { useEffect, useSyncExternalStore } from 'react';
import { getScreenFocused, subscribeScreenFocus } from './navigation';
export function useIsFocused() { return useSyncExternalStore(subscribeScreenFocus, getScreenFocused); }
import { installPreventRemove } from './navigation';

export function usePreventRemove(active: boolean, callback: (event: { data: { action: Record<string, unknown> } }) => void) {
  useEffect(() => installPreventRemove(active, callback), [active, callback]);
}

// Exercise real navigation theme context without loading unrelated native views.
const { createRequire } = await import('node:module');
const { pathToFileURL } = await import('node:url');
const themeRequire = createRequire(import.meta.url);
const nativeRoot = themeRequire.resolve('@react-navigation/native/package.json');
const coreRoot = createRequire(nativeRoot).resolve('@react-navigation/core/package.json');
const themeModule = (root: string, name: string) => pathToFileURL(
  new URL(`./src/theming/${name}.tsx`, pathToFileURL(root)).pathname
).href;
export const { DarkTheme } = await import(themeModule(nativeRoot, 'DarkTheme'));
export const { DefaultTheme } = await import(themeModule(nativeRoot, 'DefaultTheme'));
export const { ThemeProvider } = await import(themeModule(coreRoot, 'ThemeProvider'));
export const { useTheme } = await import(themeModule(coreRoot, 'useTheme'));
