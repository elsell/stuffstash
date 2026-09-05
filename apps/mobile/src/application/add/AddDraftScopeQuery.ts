import type { ReadRequest } from '../shared/ReadRequest';
import type { CurrentPrincipalRepository } from '../settings/SettingsQuery';

export type AddDraftScopeViewModel = {
  readonly principalId: string;
};

export class AddDraftScopeQuery {
  constructor(private readonly principals: CurrentPrincipalRepository) {}

  getPrincipal(request: ReadRequest = {}) { return this.principals.getCurrentPrincipal(request); }

  async execute(): Promise<AddDraftScopeViewModel> {
    const principal = await this.getPrincipal();

    return {
      principalId: principal.id
    };
  }
}
