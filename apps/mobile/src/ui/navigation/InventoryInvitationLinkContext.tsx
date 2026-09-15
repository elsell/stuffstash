import * as Linking from 'expo-linking';
import { createContext, type ReactNode, useContext, useMemo } from 'react';
import type { InventoryInvitationReference } from '../../application/invitations/InventoryInvitationRepository';
import type { PendingInventoryInvitationSnapshot } from '../../application/invitations/PendingInventoryInvitation';
import { loadMobileRuntimeConfigSeed } from '../../config/mobileRuntimeConfig';
import { useInvitationLinkSnapshot, type InvitationLinkSource } from './useInvitationLinkSnapshot';

type InventoryInvitationLinkState = PendingInventoryInvitationSnapshot & { readonly clear: () => void };
const InventoryInvitationLinkContext = createContext<InventoryInvitationLinkState | null>(null);
const systemLinks: InvitationLinkSource = {
  getInitialURL: () => Linking.getInitialURL(),
  subscribe: listener => {
    const subscription = Linking.addEventListener('url', ({ url }) => listener(url));
    return () => subscription.remove();
  }
};

export function InventoryInvitationLinkProvider({ children }: { readonly children: ReactNode }) {
  const config = useMemo(() => loadMobileRuntimeConfigSeed(), []);
  const { snapshot, clear } = useInvitationLinkSnapshot(systemLinks, config.invitationOrigin, config.invitationAllowInsecureLocalHTTP);
  return <InventoryInvitationLinkContext.Provider value={{ ...snapshot, clear }}>{children}</InventoryInvitationLinkContext.Provider>;
}

export function useInventoryInvitationLink(): InventoryInvitationLinkState {
  const value = useContext(InventoryInvitationLinkContext);
  if (!value) throw new Error('Inventory invitation link state is unavailable.');
  return value;
}

export function invitationReferenceFromState(
  state: InventoryInvitationLinkState
): InventoryInvitationReference | undefined {
  return state.reference;
}

