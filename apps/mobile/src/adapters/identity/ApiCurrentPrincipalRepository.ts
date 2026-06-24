import type { StuffStashClient } from '@stuff-stash/api-client';
import type {
  CurrentPrincipalRepository,
  SettingsPrincipal
} from '../../application/settings/SettingsQuery';

type PrincipalApiClient = Pick<StuffStashClient, 'me'>;

export class ApiCurrentPrincipalRepository implements CurrentPrincipalRepository {
  constructor(private readonly client: PrincipalApiClient) {}

  async getCurrentPrincipal(): Promise<SettingsPrincipal> {
    return this.client.me();
  }
}
