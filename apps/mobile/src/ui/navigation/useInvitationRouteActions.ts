import { useCallback, useRef } from 'react';
import { useFocusEffect } from 'expo-router';
import type { InventoryInvitationReference } from '../../application/invitations/InventoryInvitationRepository';

export function useInvitationRouteActions({ reference, clear, goHome, selectInventory, startOver }: {
  readonly reference?: InventoryInvitationReference;
  readonly clear: () => void;
  readonly goHome: () => void;
  readonly selectInventory: (id: string) => Promise<unknown>;
  readonly startOver: () => Promise<void>;
}) {
  const session = useRef<object | undefined>(undefined);
  useFocusEffect(useCallback(() => {
    const current = {}; session.current = current;
    return () => { if (session.current === current) session.current = undefined; };
  }, [reference]));
  const complete = async (operation: () => Promise<unknown>) => {
    const owner = session.current;
    await operation();
    if (owner !== undefined && session.current === owner) dismiss();
  };
  const dismiss = () => { clear(); goHome(); };
  return {
    dismiss,
    openInventory: (id: string) => complete(() => selectInventory(id)),
    startOver: () => complete(startOver)
  };
}
