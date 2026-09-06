export interface WorkflowDefinition {
  name: string;
  providerProfileId: string | null;
  instructions: string;
  budget: { toolCalls: number; modelCalls: number; elapsedSeconds: number; followUpTurns: number };
}
export interface WorkflowRevision {
  settingsMigration?: 'legacy-investigation-v1';
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
