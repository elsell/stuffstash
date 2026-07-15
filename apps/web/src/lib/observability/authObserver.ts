export type AuthEventName =
  | 'auth.runtime_configuration_failed'
  | 'auth.workspace_load_failed'
  | 'auth.sign_in_start_failed'
  | 'auth.session_invalidated'
  | 'auth.callback_failed';

export type AuthFailureReason =
  | 'runtime_configuration'
  | 'workspace_transport'
  | 'sign_in_navigation'
  | 'session_expired'
  | 'post_callback_rejected'
  | 'callback_completion';

export type AuthEventAttributes = Record<string, string | number | boolean>;

export interface AuthObserver {
  record(eventName: AuthEventName, attributes?: AuthEventAttributes): void;
}

interface AuthEventTarget {
  dispatchEvent(event: Event): boolean;
}

export class BrowserAuthObserver implements AuthObserver {
  constructor(private readonly target: AuthEventTarget = window) {}

  record(eventName: AuthEventName, attributes: AuthEventAttributes = {}): void {
    this.target.dispatchEvent(
      new CustomEvent('stuffstash:auth-observability', {
        detail: { eventName, attributes }
      })
    );
  }
}

export function authFailureAttributes(failure: unknown, reason: AuthFailureReason): AuthEventAttributes {
  return {
    failureType: failure instanceof Error ? failure.name : failure === undefined ? 'unavailable' : typeof failure,
    reason
  };
}
