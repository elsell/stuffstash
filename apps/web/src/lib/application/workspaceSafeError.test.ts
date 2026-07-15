import { describe, expect, it } from 'vitest';
import { safeWorkspaceErrorMessage } from './workspaceSafeError';

describe('safeWorkspaceErrorMessage', () => {
  it('uses fixed recovery copy for untyped failures', () => {
    expect(safeWorkspaceErrorMessage(new Error('RAW_SENTINEL provider stack'), 'Sharing could not be loaded. Try again.')).toBe(
      'Sharing could not be loaded. Try again.'
    );
  });

  it('preserves adapter errors explicitly marked safe for users', () => {
    const error = Object.assign(new Error('That invitation has already expired.'), {
      safeForUser: true,
      status: 400,
      code: 'invalid_request'
    });
    expect(safeWorkspaceErrorMessage(error, 'Invitation could not be updated. Try again.')).toBe(
      'That invitation has already expired.'
    );
  });

  it.each(['Invalid request.', 'validation failed'])('suppresses generic safe validation text: %s', (message) => {
    const error = Object.assign(new Error(message), {
      safeForUser: true,
      status: message === 'Invalid request.' ? 400 : 422,
      code: 'invalid_request'
    });
    expect(safeWorkspaceErrorMessage(error, 'Sharing change could not be saved. Try again.')).toBe(
      'Sharing change could not be saved. Try again.'
    );
  });
});
