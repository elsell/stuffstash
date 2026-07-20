import { createElement, type ReactNode } from 'react';

let dismissals = 0;

export function KeyboardProvider({ children, ...props }: { readonly children?: ReactNode; readonly preload?: boolean }) {
  return createElement('KeyboardProvider', props, children);
}

export function KeyboardExtender({ children, ...props }: { readonly children?: ReactNode; readonly enabled?: boolean }) {
  return createElement('KeyboardExtender', props, children);
}

export const KeyboardController = {
  dismiss: async () => { dismissals += 1; }
};

export function keyboardControllerDismissCount() { return dismissals; }
export function resetKeyboardControllerTestState() { dismissals = 0; }
