import type { TokenProvider } from '@stuff-stash/api-client';
import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
import { WorkflowAPIRepository } from './workflowRepository';
import { CaseAPIRepository } from './caseRepository';
import { RunAPIRepository } from './runRepository';
import { ConversationProviderAPIRepository } from './providerRepository';
export function conversationWorkspaceRepositories(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch): ConversationWorkspaceRepositories {
  return { apiIdentity: baseUrl.replace(/\/+$/, ''),
    workflows: new WorkflowAPIRepository(baseUrl, tokenProvider, fetchImpl),
    cases: new CaseAPIRepository(baseUrl, tokenProvider, fetchImpl),
    runs: new RunAPIRepository(baseUrl, tokenProvider, fetchImpl),
    providers: new ConversationProviderAPIRepository(baseUrl, tokenProvider, fetchImpl) };
}
