import type { EvaluationRun, RunResult } from '$lib/domain/conversationRun';

export interface ComparedCaseResult { passed: boolean; modelCalls: number; durationMilliseconds: number }
export interface ComparedRunTotals { passedCases: number; modelCalls: number; durationMilliseconds: number }
export type ConversationRunComparison = { compatible: false; reason: 'same' | 'incomplete' | 'cases' | 'providers' } | {
  compatible: true; baseline: ComparedRunTotals; candidate: ComparedRunTotals;
  cases: { title: string; baseline: ComparedCaseResult; candidate: ComparedCaseResult }[];
};
export const conversationRunHasCompleteResults = (run: EvaluationRun) => (run.state === 'succeeded' || run.state === 'failed') && !run.failureCode && run.totalCases > 0 && run.completedCases === run.totalCases && run.cases.length === run.totalCases && run.results.length === run.totalCases;
const pins = (run: EvaluationRun) => run.cases.map(pin => JSON.stringify([pin.caseId, pin.revisionId])).sort();
const providers = (run: EvaluationRun) => run.providers.map(pin => JSON.stringify([pin.step, pin.profileId, pin.configurationId])).sort();
const same = (left: string[], right: string[]) => left.length === right.length && left.every((value, index) => value === right[index]);
const result = (value: RunResult): ComparedCaseResult => ({ passed: value.verdict.passed, modelCalls: value.modelCalls, durationMilliseconds: value.durationMilliseconds });
function totals(values: ComparedCaseResult[]): ComparedRunTotals {
  return values.reduce((sum, value) => ({ passedCases: sum.passedCases + Number(value.passed), modelCalls: sum.modelCalls + value.modelCalls, durationMilliseconds: sum.durationMilliseconds + value.durationMilliseconds }), { passedCases: 0, modelCalls: 0, durationMilliseconds: 0 });
}
export function compareConversationRuns(baseline: EvaluationRun, candidate: EvaluationRun): ConversationRunComparison {
  if (baseline.id === candidate.id) return { compatible: false, reason: 'same' };
  if (!conversationRunHasCompleteResults(baseline) || !conversationRunHasCompleteResults(candidate)) return { compatible: false, reason: 'incomplete' };
  if (!same(pins(baseline), pins(candidate))) return { compatible: false, reason: 'cases' };
  if (!same(providers(baseline), providers(candidate))) return { compatible: false, reason: 'providers' };
  const left = new Map(baseline.results.map(value => [value.caseRevisionId, value]));
  const right = new Map(candidate.results.map(value => [value.caseRevisionId, value]));
  if (left.size !== baseline.totalCases || right.size !== candidate.totalCases || candidate.cases.some(pin => !left.has(pin.revisionId) || !right.has(pin.revisionId))) return { compatible: false, reason: 'incomplete' };
  const cases = candidate.cases.map(pin => ({ title: pin.title, baseline: result(left.get(pin.revisionId)!), candidate: result(right.get(pin.revisionId)!) }));
  return { compatible: true, cases, baseline: totals(cases.map(value => value.baseline)), candidate: totals(cases.map(value => value.candidate)) };
}
