import { useCallback, useEffect, useState } from 'react';
import type { SettingsQuery, SettingsViewModel } from '../../application/settings/SettingsQuery';

export type SettingsLoadState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly settings: SettingsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function useSettingsModel(query: SettingsQuery) {
  const [state, setState] = useState<SettingsLoadState>({ status: 'loading' });
  const load = useCallback(async () => {
    setState({ status: 'loading' });
    try {
      setState({ status: 'ready', settings: await query.execute() });
    } catch (error) {
      setState({
        status: 'error',
        message: error instanceof Error ? error.message : 'Stuff Stash could not load settings.'
      });
    }
  }, [query]);

  useEffect(() => {
    void load();
  }, [load]);

  return { load, state };
}
