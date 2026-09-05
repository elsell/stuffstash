import { isAccessFailure } from '../serverState/isAccessFailure';
import type { SettingsQuery, SettingsViewModel } from '../../application/settings/SettingsQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { useMobileServerQuery } from '../serverState/useMobileServerQuery';

export type SettingsLoadState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly settings: SettingsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function useSettingsModel(query: SettingsQuery) {
  const principal = useMobileServerQuery({ key: mobileQueryKeys.principal, query: (signal) => query.getPrincipal({ signal }) });
  const scope = useMobileInventoryServerQuery({ key: mobileQueryKeys.settingsScope, query: (signal) => query.getSelectedScope({ signal }) });
  const diagnostics = query.getDiagnostics();
  const state: SettingsLoadState = isAccessFailure(scope.error) ? { status: 'error', message: 'These settings are no longer available.' } : scope.data ? { status: 'ready', settings: {
    principal: { id: principal.data?.id ?? '', primaryLabel: principal.data?.email ?? 'Signed in' },
    selectedTenant: scope.data.tenant, selectedInventory: scope.data.inventory,
    serverUrl: diagnostics.apiBaseUrl, appVersion: diagnostics.appVersion, authenticationMode: diagnostics.authenticationMode
  } } : scope.isError ? { status: 'error', message: 'Stuff Stash could not load settings.' } : { status: 'loading' };
  const load = async () => { await Promise.all([scope.refetch(), principal.refetch()]); };
  return { state, load, hasRefreshError: scope.isRefetchError || principal.isError };
}
