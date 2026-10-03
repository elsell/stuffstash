import type { CaseOperation, CaseOutcome } from '$lib/domain/conversationCase';
import { t } from './localization';

export function caseOperationLabel(operation: CaseOperation): string {
  return t(`case.operation.${operation}`);
}
export function caseOutcomeLabel(outcome: CaseOutcome): string {
  return t(`case.outcome.${outcome}`);
}
