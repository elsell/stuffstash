type AlertButton = { readonly text: string; readonly style?: string; readonly onPress?: () => unknown };
type AlertRecord = { readonly title: string; readonly message?: string; readonly buttons: readonly AlertButton[]; readonly options?: { readonly onDismiss?: () => void } };
const alerts: AlertRecord[] = [];
const focusHandles: unknown[] = [];
const focusedInputs: string[] = [];
let animationStarts = 0;
let animationStops = 0;
let deferAnimations = false;
const runningAnimations = new Set<object>();
let keyboardDismissals = 0;
let keyboardVisible = false;
const keyboardListeners = new Map<string, Set<(event?: unknown) => void>>();
const accessibilityListeners = new Map<string, Set<(enabled: boolean) => void>>();
let reduceMotionEnabled = false;
let reducedMotionSnapshot: Promise<boolean> | undefined;
let screenReaderEnabled = false;
const announcements: string[] = [];
let darkerSystemColorsEnabled = false;
let highTextContrastEnabled = false;
let systemColorScheme: 'light' | 'dark' = 'light';

export const View = 'View';
export const Switch = 'Switch';
type ImageSizeRequest = {
  readonly uri: string;
  readonly headers: unknown;
  readonly succeed: (width: number, height: number) => void;
  readonly fail: () => void;
};
const imageSizeRequests: ImageSizeRequest[] = [];
export const Image = Object.assign(
  (props: Record<string, unknown>) => createElement('Image', props),
  { getSizeWithHeaders(uri: string, headers: unknown, succeed: ImageSizeRequest['succeed'], fail: () => void) {
      imageSizeRequests.push({ uri, headers, succeed, fail });
    },
    resolveAssetSource: (asset: number) => ({ uri: `asset:${asset}`, width: 100, height: 100 }) }
);
export function imageSizeRequestsForTest() { return [...imageSizeRequests]; }
export function FlatList(props: Record<string, unknown>) {
  const rows = (props.data ?? []) as readonly unknown[];
  const render = props.renderItem as ((input: { item: unknown; index: number }) => ReactNode) | undefined;
  return createElement('FlatList', props,
    props.ListHeaderComponent as ReactNode,
    props.refreshControl as ReactNode,
    rows.length ? rows.map((item, index) => createElement('Row', { key: index }, render?.({ item, index }))) : props.ListEmptyComponent as ReactNode,
    props.ListFooterComponent as ReactNode);
}
export function SectionList(props: Record<string, unknown>) {
  const sections = props.sections as readonly { data: readonly unknown[] }[];
  const render = props.renderItem as (input: { item: unknown }) => ReactNode;
  return createElement('SectionList', props,
    props.refreshControl as ReactNode,
    sections.flatMap((section) => section.data).map((item, index) => createElement('Row', { key: index }, render({ item }))),
    props.ListFooterComponent as ReactNode);
}
const appStateListeners = new Set<(state: string) => void>();
export const deviceSettingsFake = {
  attempts: 0,
  completion: undefined as Promise<void> | undefined,
  reset() { this.attempts = 0; this.completion = undefined; }
};
export const Linking = {
  async openSettings() { deviceSettingsFake.attempts++; await deviceSettingsFake.completion; }
};
export const AppState = {
  currentState: 'active',
  addEventListener(_event: string, callback: (state: string) => void) {
    appStateListeners.add(callback);
    return { remove: () => appStateListeners.delete(callback) };
  }
};
export function setAppStateForTest(state: string) {
  AppState.currentState = state;
  for (const listener of appStateListeners) listener(state);
}
let actionSheetCallback: ((index: number) => void) | undefined;
export function latestActionSheetCallback() { return actionSheetCallback; }
export const ActionSheetIOS = {
  showActionSheetWithOptions(_options: unknown, callback: (index: number) => void) {
    actionSheetCallback = callback;
  }
};
export const Text = 'Text';
export const Pressable = 'Pressable';
const scrollCommands: { y?: number; x?: number; animated?: boolean }[] = [];
export const ScrollView = forwardRef((props: Record<string, unknown>, ref) => {
  useImperativeHandle(ref, () => ({
    scrollTo: (options: { y?: number; x?: number; animated?: boolean }) => { scrollCommands.push(options); },
    scrollToEnd: (options: { animated?: boolean }) => { scrollCommands.push(options); }
  }), []);
  return createElement('ScrollView', props, props.children as ReactNode);
});
export function scrollCommandsForTest() { return [...scrollCommands]; }
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
  isReduceMotionEnabled: async () => reducedMotionSnapshot ?? reduceMotionEnabled,
  isScreenReaderEnabled: async () => screenReaderEnabled,
  announceForAccessibility(message: string) { announcements.push(message); },
  addEventListener(event: string, listener: (enabled: boolean) => void) {
    const listeners = accessibilityListeners.get(event) ?? new Set<(enabled: boolean) => void>();
    listeners.add(listener);
    accessibilityListeners.set(event, listeners);
    return { remove() { listeners.delete(listener); } };
  },
  isDarkerSystemColorsEnabled: async () => darkerSystemColorsEnabled,
  isHighTextContrastEnabled: async () => highTextContrastEnabled,
  setAccessibilityFocus(handle: unknown) { focusHandles.push(handle); }
};
export const Appearance = { setColorScheme() {} };
export const Platform = { OS: 'ios', select: <T>(values: { ios?: T; default?: T }) => values.ios ?? values.default };
export const PlatformColor = (name: string) => `platform:${name}`;
export const Keyboard = {
  metrics() { return undefined; },
  scheduleLayoutAnimation() {},
  addListener(event: string, listener: (event?: unknown) => void) {
    const listeners = keyboardListeners.get(event) ?? new Set<(event?: unknown) => void>();
    listeners.add(listener);
    keyboardListeners.set(event, listeners);
    return { remove() { listeners.delete(listener); } };
  },
  dismiss() { keyboardDismissals += 1; },
  isVisible() { return keyboardVisible; }
};
export const StyleSheet = { create: <T>(styles: T) => styles, hairlineWidth: 1 };
export const findNodeHandle = () => 1;
let windowFontScale = 1;
export function setWindowFontScaleForTest(value: number) { windowFontScale = value; }
export const Dimensions = { get: () => ({ fontScale: windowFontScale, height: 844, width: 390, scale: 1 }) };
export const useWindowDimensions = () => ({ fontScale: windowFontScale, height: 844, width: 390 });
export const useColorScheme = () => systemColorScheme;
class AnimatedValue {
  private value: number;
  private listeners = new Map<string, (event: { value: number }) => void>();
  constructor(readonly initial: number) { this.value = initial; }
  setValue(value: number) { this.value = value; for (const listener of this.listeners.values()) listener({ value }); }
  addListener(listener: (event: { value: number }) => void) { const id = String(this.listeners.size); this.listeners.set(id, listener); return id; }
  removeListener(id: string) { this.listeners.delete(id); }
  removeAllListeners() { this.listeners.clear(); }
  __getValue() { return this.value; }
  stopAnimation() {}
  interpolate() { return this.value; }
}
const animation = (value?: AnimatedValue, options?: { toValue: number }) => {
  const handle = {
    stop() { animationStops++; runningAnimations.delete(handle); },
    start(callback?: (result: { finished: boolean }) => void) {
      animationStarts++;
      if (deferAnimations) { runningAnimations.add(handle); return; }
      if (value && options) value.setValue(options.toValue);
      callback?.({ finished: true });
    }
  };
  return handle;
};
const parallelAnimation = (children: ReturnType<typeof animation>[]) => ({
  stop() { children.forEach(child => child.stop()); },
  start(callback?: (result: { finished: boolean }) => void) { children.forEach(child => child.start()); if (!deferAnimations) callback?.({ finished: true }); }
});
class AnimatedValueXY {
  readonly x: AnimatedValue;
  readonly y: AnimatedValue;
  constructor(initial: { x: number; y: number }) { this.x = new AnimatedValue(initial.x); this.y = new AnimatedValue(initial.y); }
  getTranslateTransform() { return [{ translateX: this.x }, { translateY: this.y }]; }
}
export const Animated = { ValueXY: AnimatedValueXY, Value: AnimatedValue, View: 'AnimatedView', multiply: (value: AnimatedValue, factor: number) => ({ value, factor }), parallel: parallelAnimation, spring: animation, timing: animation };
export const PanResponder = { create: (handlers: Record<string, unknown>) => ({ panHandlers: handlers }) };

export function resetNativeTestState() {
  deviceSettingsFake.reset();
  scrollCommands.length = 0;
  imageSizeRequests.length = 0;
  reducedMotionSnapshot = undefined;
  reduceMotionEnabled = false;
  screenReaderEnabled = false;
  announcements.length = 0;
  animationStarts = 0; animationStops = 0; deferAnimations = false; runningAnimations.clear();
  alerts.length = 0;
  focusHandles.length = 0;
  focusedInputs.length = 0;
  keyboardDismissals = 0;
  keyboardVisible = false;
  darkerSystemColorsEnabled = false;
  highTextContrastEnabled = false;
  systemColorScheme = 'light';
  keyboardListeners.clear();
  accessibilityListeners.clear();
}
export function latestAlert() { return alerts.at(-1); }
export function alertCount() { return alerts.length; }
export async function pressAlertButton(label: string) { return latestAlert()?.buttons.find((button) => button.text === label)?.onPress?.(); }
export function focusedAccessibilityHandles() { return [...focusHandles]; }
export function focusedInputLabels() { return [...focusedInputs]; }
export function keyboardDismissCount() { return keyboardDismissals; }
export function setKeyboardVisibleForTest(visible: boolean) {
  keyboardVisible = visible;
  const event = visible ? 'keyboardWillShow' : 'keyboardDidHide';
  keyboardListeners.get(event)?.forEach((listener) => listener());
}
export function setDarkerSystemColorsEnabledForTest(enabled: boolean) {
  darkerSystemColorsEnabled = enabled;
  accessibilityListeners.get('darkerSystemColorsChanged')?.forEach((listener) => listener(enabled));
}
export function setHighTextContrastEnabledForTest(enabled: boolean) {
  highTextContrastEnabled = enabled;
  accessibilityListeners.get('highTextContrastChanged')?.forEach((listener) => listener(enabled));
}
export function setSystemColorSchemeForTest(colorScheme: 'light' | 'dark') {
  systemColorScheme = colorScheme;
}
import { createElement, forwardRef, useImperativeHandle, type ReactNode } from 'react';

export function setScreenReaderEnabledForTest(enabled: boolean) {
  screenReaderEnabled = enabled;
  accessibilityListeners.get('screenReaderChanged')?.forEach(listener => listener(enabled));
}
export function setReduceMotionEnabledForTest(enabled: boolean) {
  reduceMotionEnabled = enabled;
  accessibilityListeners.get('reduceMotionChanged')?.forEach(listener => listener(enabled));
}
export function accessibilityAnnouncements() { return [...announcements]; }

export function animationStartCount() { return animationStarts; }

export function holdReduceMotionSnapshotForTest() {
  let resolve!: (value: boolean) => void;
  let reject!: (cause: Error) => void;
  reducedMotionSnapshot = new Promise<boolean>((accept, fail) => { resolve = accept; reject = fail; });
  return { resolve, reject };
}

export function emitKeyboardEventForTest(name: string, event?: unknown) {
  for (const listener of keyboardListeners.get(name) ?? []) listener(event);
}

export function deferAnimationsForTest(value: boolean) { deferAnimations = value; }
export function animationStopCount() { return animationStops; }
export function pendingAnimationCount() { return runningAnimations.size; }
export function animatedValueForTest(value: unknown) { if (!(value instanceof AnimatedValue)) throw new Error("Not an animated value"); return value.__getValue(); }
