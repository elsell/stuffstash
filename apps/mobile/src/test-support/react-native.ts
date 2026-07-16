type AlertButton = { readonly text: string; readonly style?: string; readonly onPress?: () => unknown };
type AlertRecord = { readonly title: string; readonly message?: string; readonly buttons: readonly AlertButton[]; readonly options?: { readonly onDismiss?: () => void } };
const alerts: AlertRecord[] = [];
const focusHandles: unknown[] = [];
const focusedInputs: string[] = [];

export const View = 'View';
export const Text = 'Text';
export const Pressable = 'Pressable';
export const ScrollView = 'ScrollView';
export const KeyboardAvoidingView = 'KeyboardAvoidingView';
export const ActivityIndicator = 'ActivityIndicator';
export const Modal = 'Modal';
export const RefreshControl = 'RefreshControl';
export const TextInput = forwardRef<{ focus(): void }, Record<string, unknown>>((props, ref) => {
  useImperativeHandle(ref, () => ({ focus() { focusedInputs.push(String(props.accessibilityLabel ?? '')); } }));
  return createElement('TextInput', props);
});
export const Alert = { alert(title: string, message?: string, buttons: readonly AlertButton[] = [], options?: AlertRecord['options']) { alerts.push({ title, message, buttons, options }); } };
export const AccessibilityInfo = {
  addEventListener: () => ({ remove() {} }),
  isHighTextContrastEnabled: async () => false,
  setAccessibilityFocus(handle: unknown) { focusHandles.push(handle); }
};
export const Appearance = { setColorScheme() {} };
export const Platform = { OS: 'ios', select: <T>(values: { ios?: T; default?: T }) => values.ios ?? values.default };
export const StyleSheet = { create: <T>(styles: T) => styles, hairlineWidth: 1 };
export const findNodeHandle = () => 1;
export const useWindowDimensions = () => ({ fontScale: 1, height: 844, width: 390 });
export const useColorScheme = () => 'light' as const;
class AnimatedValue { constructor(readonly initial: number) {} setValue() {} }
const animation = () => ({ start(callback?: () => void) { callback?.(); } });
export const Animated = { Value: AnimatedValue, View: 'AnimatedView', parallel: animation, spring: animation, timing: animation };
export const PanResponder = { create: (handlers: Record<string, unknown>) => ({ panHandlers: handlers }) };

export function resetNativeTestState() { alerts.length = 0; focusHandles.length = 0; focusedInputs.length = 0; }
export function latestAlert() { return alerts.at(-1); }
export function alertCount() { return alerts.length; }
export async function pressAlertButton(label: string) { return latestAlert()?.buttons.find((button) => button.text === label)?.onPress?.(); }
export function focusedAccessibilityHandles() { return [...focusHandles]; }
export function focusedInputLabels() { return [...focusedInputs]; }
import { createElement, forwardRef, useImperativeHandle } from 'react';
