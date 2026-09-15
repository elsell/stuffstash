import { useCallback, useMemo } from 'react';
import { useFocusEffect } from 'expo-router';

/** Mutation completion may outlive navigation; presentation belongs to one visit. */
export function useProviderTaskPresentation(command: object, identity = '') {
  const owner = useMemo<{ session?: object }>(() => ({}), [command, identity]);
  useFocusEffect(useCallback(() => {
    const session = {};
    owner.session = session;
    return () => { if (owner.session === session) owner.session = undefined; };
  }, [owner]));
  return () => {
    const session = owner.session;
    return () => session !== undefined && owner.session === session;
  };
}
