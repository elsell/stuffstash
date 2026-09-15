export type PreventedAction = { readonly type?: string; readonly [key: string]: unknown };
type PreventCallback = (event: { data: { action: PreventedAction } }) => void;

let active = false;
let preventCallback: PreventCallback | undefined;
let dispatching = false;
const dispatched: PreventedAction[] = [];
const options: unknown[] = [];
const optionListeners = new Set<() => void>();
export function subscribeNavigationOptions(listener: () => void) { optionListeners.add(listener); return () => { optionListeners.delete(listener); }; }

export const fakeNavigation = {
  dispatch(action: PreventedAction) {
    if (active && preventCallback && !dispatching) {
      dispatching = true;
      try { preventCallback({ data: { action } }); } finally { dispatching = false; }
      return;
    }
    dispatched.push(action);
  },
  setOptions(value: unknown) {
    const previous = options.at(-1) as Record<string, unknown> | undefined;
    options.push(value);
    const next = value as Record<string, unknown> | undefined;
    const keys = new Set([...Object.keys(previous ?? {}), ...Object.keys(next ?? {})]);
    if ([...keys].some(key => previous?.[key] !== next?.[key])) optionListeners.forEach(listener => listener());
  }
};

export function installPreventRemove(enabled: boolean, callback: PreventCallback) {
  active = enabled;
  preventCallback = callback;
  return () => { if (preventCallback === callback) { active = false; preventCallback = undefined; } };
}
export function attemptNavigation(action: PreventedAction) { fakeNavigation.dispatch(action); }
export function dispatchedActions() { return [...dispatched]; }
export function navigationOptions() { return [...options]; }
export function resetNavigation() { active = false; preventCallback = undefined; dispatching = false; dispatched.length = 0; options.length = 0; }

let canGoBack = true;
export function setCanGoBack(value: boolean) { canGoBack = value; }
export function getCanGoBack() { return canGoBack; }

let screenFocused = true;
const focusListeners = new Set<() => void>();
export const getScreenFocused = () => screenFocused;
export function subscribeScreenFocus(listener: () => void) { focusListeners.add(listener); return () => { focusListeners.delete(listener); }; }
export function setScreenFocused(value: boolean) { screenFocused = value; focusListeners.forEach(listener => listener()); }
