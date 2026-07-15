import { describe, expect, it } from 'vitest';
import { BrowserAuthObserver, authFailureAttributes } from './authObserver';

describe('auth observer', () => {
  it('emits a structured operator event without exposing the caught message', () => {
    const events: Event[] = [];
    const observer = new BrowserAuthObserver({
      dispatchEvent(event) {
        events.push(event);
        return true;
      }
    });

    observer.record(
      'auth.sign_in_start_failed',
      authFailureAttributes(new Error('Dex client secret URL'), 'sign_in_navigation')
    );

    const detail = (events[0] as CustomEvent).detail;
    expect(detail).toEqual({
      eventName: 'auth.sign_in_start_failed',
      attributes: { failureType: 'Error', reason: 'sign_in_navigation' }
    });
    expect(JSON.stringify(detail)).not.toContain('Dex client secret URL');
  });

  it('classifies non-Error failures without stringifying their value', () => {
    expect(authFailureAttributes('OIDC raw failure', 'callback_completion')).toEqual({
      failureType: 'string',
      reason: 'callback_completion'
    });
  });
});
