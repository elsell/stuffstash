export type PreventedAction = { readonly type?: string; readonly [key: string]: unknown };
type PreventCallback = (event: { data: { action: PreventedAction } }) => void;

let active = false;
let preventCallback: PreventCallback | undefined;
let dispatching = false;
const dispatched: PreventedAction[] = [];
const options: unknown[] = [];

export const fakeNavigation = {
  dispatch(action: PreventedAction) {
    if (active && preventCallback && !dispatching) {
      dispatching = true;
      try { preventCallback({ data: { action } }); } finally { dispatching = false; }
      return;
    }
    dispatched.push(action);
  },
  setOptions(value: unknown) { options.push(value); }
};

export function installPreventRemove(enabled: boolean, callback: PreventCallback) {
  active = enabled;
  preventCallback = callback;
}
export function attemptNavigation(action: PreventedAction) { fakeNavigation.dispatch(action); }
export function dispatchedActions() { return [...dispatched]; }
export function navigationOptions() { return [...options]; }
export function resetNavigation() { active = false; preventCallback = undefined; dispatching = false; dispatched.length = 0; options.length = 0; }
