import { t } from './localization';

const failureMessages = {
  invalid_observation: 'evaluation.failure.invalid_observation',
  unexpected_outcome: 'evaluation.failure.unexpected_outcome',
  missing_reference: 'evaluation.failure.missing_reference',
  missing_location: 'evaluation.failure.missing_location',
  missing_proposal: 'evaluation.failure.missing_proposal',
  forbidden_operation: 'evaluation.failure.forbidden_operation',
  unexpected_mutation: 'evaluation.failure.unexpected_mutation',
  unexpected_proposal: 'evaluation.failure.unexpected_proposal'
} as const;

export function evaluationFailureLabel(code: string): string {
  return t(Object.hasOwn(failureMessages, code)
    ? failureMessages[code as keyof typeof failureMessages]
    : 'evaluation.failure.unknown');
}
