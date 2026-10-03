import { createContext, useContext, type ReactNode } from 'react';
import * as Linking from 'expo-linking';
import { parseMobileLabelLink } from '../../adapters/labels/LabelLinkParser';
import { useLabelLinkSnapshot, type LabelLinkSource } from './useLabelLinkSnapshot';
const source: LabelLinkSource = {
  getInitialURL: () => Linking.getInitialURL(),
  subscribe: callback => { const subscription = Linking.addEventListener('url', ({ url }) => callback(url)); return () => subscription.remove(); }
};
const Context = createContext<ReturnType<typeof useLabelLinkSnapshot> | undefined>(undefined);
/** Deliberately outside the authenticated composition: no URL-driven server changes. */
export function LabelLinkProvider({ children }: { readonly children: ReactNode }) {
  const pending = useLabelLinkSnapshot(source, parseMobileLabelLink);
  return <Context.Provider value={pending}>{children}</Context.Provider>;
}
export function usePendingLabel() { const value = useContext(Context); if (!value) throw new Error('Label link provider unavailable.'); return value; }
