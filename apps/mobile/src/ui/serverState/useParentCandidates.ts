import { useEffect, useState } from 'react';
import type { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from './useMobileInventoryServerQuery';

const parentSearchDebounceMs = 250;

export function useParentCandidates(query: string, lookup: Pick<ParentLookupQuery, 'execute'>, enabled = true) {
  const normalized = query.trim();
  const [settled, setSettled] = useState(normalized);
  useEffect(() => {
    const timer = setTimeout(() => setSettled(normalized), parentSearchDebounceMs);
    return () => clearTimeout(timer);
  }, [normalized]);
  const result = useMobileInventoryServerQuery({
    key: (scope, tenant, inventory) => mobileQueryKeys.parentCandidates(scope, tenant, inventory, normalized),
    query: (signal) => lookup.execute(normalized, { signal }),
    enabled: enabled && normalized === settled
  });
  return { ...result, data: enabled && normalized === settled ? result.data : undefined };
}
