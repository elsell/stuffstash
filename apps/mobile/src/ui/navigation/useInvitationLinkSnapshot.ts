import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { PendingInventoryInvitation } from '../../application/invitations/PendingInventoryInvitation';

export interface InvitationLinkSource {
  getInitialURL(): Promise<string | null>;
  subscribe(listener: (url: string) => void): () => void;
}

export function useInvitationLinkSnapshot(source: InvitationLinkSource, origin?: string, allowInsecureLocalHTTP = false) {
  const pending = useMemo(() => new PendingInventoryInvitation(), []);
  const lookupGeneration = useRef(0);
  const [snapshot, setSnapshot] = useState(pending.current());
  const capture = useCallback((url: string | null) => {
    if (!url || !isInvitationRoute(url)) return;
    setSnapshot(pending.capture(url, origin, allowInsecureLocalHTTP));
  }, [pending, origin, allowInsecureLocalHTTP]);
  useEffect(() => {
    const generation = ++lookupGeneration.current;
    let active = true;
    let foregroundCaptured = false;
    const finishWithoutLink = () => {
      if (active && generation === lookupGeneration.current && !foregroundCaptured) setSnapshot(current => current.initialized ? current : { invalid: false, initialized: true });
    };
    void source.getInitialURL().then(url => {
      if (!active || generation !== lookupGeneration.current || foregroundCaptured) return;
      if (url && isInvitationRoute(url)) { capture(url); return; }
      finishWithoutLink();
    }).catch(finishWithoutLink);
    const remove = source.subscribe(url => {
      if (!active || !isInvitationRoute(url)) return;
      foregroundCaptured = true;
      capture(url);
    });
    return () => { active = false; remove(); };
  }, [source, capture]);
  const clear = useCallback(() => { lookupGeneration.current++; setSnapshot(pending.clear()); }, [pending]);
  return { snapshot, clear };
}

function isInvitationRoute(source: string): boolean {
  try {
    const url = new URL(source);
    return (url.protocol === 'stuffstash:' && url.host === 'invitations') || url.pathname.replace(/\/$/, '') === '/invitations/accept';
  } catch { return false; }
}
