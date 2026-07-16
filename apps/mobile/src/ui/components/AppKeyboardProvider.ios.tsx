import type { PropsWithChildren } from 'react';
import { KeyboardProvider } from 'react-native-keyboard-controller';

export function AppKeyboardProvider({ children }: PropsWithChildren) {
  return <KeyboardProvider preload={false}>{children}</KeyboardProvider>;
}
