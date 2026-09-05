import type { CaseOutcome, CaseOperation, CaseProposal } from './conversationCase';
import type { ConversationRunState } from './conversation';
import type { EvaluationCasePin } from './conversationWorkflow';
export interface RunQueue { workflowId: string; revisionId: string; cases: EvaluationCasePin[] }
export interface RunHead {
  id: string; state: ConversationRunState; version: number; workflowId: string; revisionId: string;
  totalCases: number; completedCases: number; passedCases: number; createdAt: string; updatedAt: string;
}
export interface RunObservation {
  kind: CaseOutcome; referencedAssets: string[]; locations: { assetId: string; ancestorId: string }[];
  proposals: CaseProposal[];
  executedOperations: CaseOperation[];
}
export interface RunResult {
  caseRevisionId: string; observation: RunObservation;
  verdict: { passed: boolean; failures: { code: string; fixtureId: string; operation: CaseOperation | '' }[] };
  modelCalls: number; durationMilliseconds: number; completedAt: string;
}
export interface EvaluationRun extends RunHead {
  authorId: string; coverage: 'text_only'; cases: (EvaluationCasePin & { title: string })[];
  providers: { step: string; profileId: string; configurationId: string }[];
  results: RunResult[]; startedAt: string | null; finishedAt: string | null; failureCode: string;
}
