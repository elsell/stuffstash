import { describe, expect, it } from 'vitest';
import { CustomizationFailure, safeCustomizationMessage } from './CustomizationErrors';

describe('safeCustomizationMessage', () => {
  it('never renders raw unknown or transport errors', () => {
    expect(safeCustomizationMessage(new Error('token=secret request-id=raw'), 'Could not save.')).toBe('Could not save.');
    expect(safeCustomizationMessage(new CustomizationFailure('permission-denied'), 'Could not save.')).toBe('Your access changed. This change was not saved.');
  });
});
