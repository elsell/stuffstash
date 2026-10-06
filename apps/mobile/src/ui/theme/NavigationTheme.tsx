import { useMemo, type ReactNode } from 'react';
import { DarkTheme, DefaultTheme, ThemeProvider, type Theme } from '@react-navigation/native';
import { useAppearance } from './AppearanceContext';

/** Native-stack derives UIKit appearance from this navigation theme. */
export function NavigationTheme({ children }: { readonly children: ReactNode }) {
  const { palette, resolvedColorScheme } = useAppearance();
  const theme = useMemo<Theme>(() => ({
    ...(resolvedColorScheme === 'dark' ? DarkTheme : DefaultTheme),
    colors: {
      primary: palette.action,
      background: palette.background,
      card: palette.surface,
      text: palette.text,
      border: palette.border,
      notification: palette.danger
    }
  }), [palette, resolvedColorScheme]);
  return <ThemeProvider value={theme}>{children}</ThemeProvider>;
}
