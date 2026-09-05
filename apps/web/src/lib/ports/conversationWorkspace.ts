import type { ConversationWorkflowRepository } from './conversationWorkflowRepository';
import type { ConversationCaseRepository } from './conversationCaseRepository';
import type { ConversationRunRepository } from './conversationRunRepository';
import type { ConversationProviderRepository } from './conversationProviderRepository';
export const conversationWorkspaceContext = Symbol('conversationWorkspace');
export interface ConversationWorkspaceRepositories {
  apiIdentity: string;
  workflows: ConversationWorkflowRepository;
  cases: ConversationCaseRepository;
  runs: ConversationRunRepository;
  providers: ConversationProviderRepository;
}
