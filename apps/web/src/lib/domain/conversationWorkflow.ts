export type WorkflowStepKind = 'interpret' | 'assess' | 'respond';
export interface WorkflowStep {
  kind: WorkflowStepKind;
  providerProfileId: string | null;
  instructions: string;
  attempts: number;
}
export interface WorkflowDefinition {
  name: string;
  retrieval: 'precise_first' | 'expanded';
  response: 'generated_with_grounded_fallback' | 'grounded';
  budget: { evidenceRounds: number; modelCalls: number; elapsedSeconds: number; followUpTurns: number };
  steps: WorkflowStep[];
}
export interface WorkflowRevision {
  id: string;
  workflowId: string;
  number: number;
  authorId: string;
  createdAt: string;
  definition: WorkflowDefinition;
}
export interface WorkflowHead {
  id: string;
  name: string;
  latestRevisionId: string;
  latestRevision: number;
  activeRevisionId: string | null;
  createdAt: string;
  updatedAt: string;
}
export interface WorkflowSelection { workflowId: string; revisionId: string }
export interface EvaluationCasePin { caseId: string; revisionId: string }
export interface WorkflowActivation {
  revisionId: string;
  runId: string;
  cases: EvaluationCasePin[];
  expected: WorkflowSelection | null;
}
