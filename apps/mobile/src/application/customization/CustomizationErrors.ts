export type CustomizationFailureKind = 'permission-denied' | 'not-found' | 'conflict' | 'invalid' | 'unavailable';

export class CustomizationFailure extends Error {
  constructor(readonly kind: CustomizationFailureKind) {
    super(customizationFailureMessage(kind));
    this.name = 'CustomizationFailure';
  }
}

export class CustomizationValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'CustomizationValidationError';
  }
}

export function customizationFailureMessage(kind: CustomizationFailureKind): string {
  if (kind === 'permission-denied') return 'Your access changed. This change was not saved.';
  if (kind === 'not-found') return 'This setting is no longer available.';
  if (kind === 'conflict') return 'This setting conflicts with another active setting.';
  if (kind === 'invalid') return 'Some information is no longer valid. Review the form and try again.';
  return 'Stuff Stash could not complete this request. Try again.';
}

export function safeCustomizationMessage(error: unknown, fallback: string): string {
  return error instanceof CustomizationFailure || error instanceof CustomizationValidationError ? error.message : fallback;
}
