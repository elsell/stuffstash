import * as Linking from 'expo-linking';
import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState
} from 'react';
import type { InventoryInvitationReference } from '../../application/invitations/InventoryInvitationRepository';
import {
  PendingInventoryInvitation,
  type PendingInventoryInvitationSnapshot
} from '../../application/invitations/PendingInventoryInvitation';
import { loadMobileRuntimeConfigSeed } from '../../config/mobileRuntimeConfig';

type InventoryInvitationLinkState = PendingInventoryInvitationSnapshot & {
  readonly clear: () => void;
};

const InventoryInvitationLinkContext = createContext<InventoryInvitationLinkState | null>(null);

export function InventoryInvitationLinkProvider({ children }: { readonly children: ReactNode }) {
  const pending = useMemo(() => new PendingInventoryInvitation(), []);
  const invitationConfig = useMemo(() => loadMobileRuntimeConfigSeed(), []);
  const [snapshot, setSnapshot] = useState(pending.current());

  const capture = useCallback((url: string | null) => {
    if (!url || !isInvitationRoute(url)) return;
    setSnapshot(pending.capture(
      url,
      invitationConfig.invitationOrigin,
      invitationConfig.invitationAllowInsecureLocalHTTP
    ));
  }, [invitationConfig, pending]);

  useEffect(() => {
    let foregroundInvitationCaptured = false;
    void Linking.getInitialURL().then((url) => {
      if (foregroundInvitationCaptured) return;
      if (url && isInvitationRoute(url)) {
        capture(url);
        return;
      }
      setSnapshot((current) => current.initialized ? current : { invalid: false, initialized: true });
    });
    const subscription = Linking.addEventListener('url', ({ url }) => {
      if (!isInvitationRoute(url)) return;
      foregroundInvitationCaptured = true;
      capture(url);
    });
    return () => subscription.remove();
  }, [capture]);

  const clear = useCallback(() => setSnapshot(pending.clear()), [pending]);
  return (
    <InventoryInvitationLinkContext.Provider value={{ ...snapshot, clear }}>
      {children}
    </InventoryInvitationLinkContext.Provider>
  );
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

function isInvitationRoute(source: string): boolean {
  try {
    const url = new URL(source);
    return (url.protocol === 'stuffstash:' && url.host === 'invitations') ||
      url.pathname.replace(/\/$/, '') === '/invitations/accept';
  } catch {
    return false;
  }
}
