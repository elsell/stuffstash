import { t } from '../../presentation/localization';
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
  if (kind === 'permission-denied') return t('mobile.CustomizationErrors.yourAccessChangedThisChangeWasNotSaved');
  if (kind === 'not-found') return t('mobile.CustomizationErrors.thisSettingIsNoLongerAvailable');
  if (kind === 'conflict') return t('mobile.CustomizationErrors.thisSettingConflictsWithAnotherActiveSetting');
  if (kind === 'invalid') return t('mobile.CustomizationErrors.someInformationIsNoLongerValidReviewTheForm');
  return t('mobile.CustomizationErrors.stuffStashCouldNotCompleteThisRequestTryAgain');
}

export function safeCustomizationMessage(error: unknown, fallback: string): string {
  return error instanceof CustomizationFailure || error instanceof CustomizationValidationError ? error.message : fallback;
}
