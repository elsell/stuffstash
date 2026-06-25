import type { CurrentPrincipalRepository } from '../settings/SettingsQuery';

export type AddDraftScopeViewModel = {
  readonly principalId: string;
};

export class AddDraftScopeQuery {
  constructor(private readonly principals: CurrentPrincipalRepository) {}

  async execute(): Promise<AddDraftScopeViewModel> {
    const principal = await this.principals.getCurrentPrincipal();

    return {
      principalId: principal.id
    };
  }
}
