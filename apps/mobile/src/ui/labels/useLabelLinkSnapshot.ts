import { useCallback, useEffect, useRef, useState } from 'react';
import type { LabelReference } from '../../application/labels/LabelWorkspace';
export type LabelLinkSource = { getInitialURL(): Promise<string | null>; subscribe(listener: (url: string) => void): () => void };
export function isLabelLink(source: string) {
  try { const url = new URL(source); return url.protocol === 'stuffstash:' && url.hostname === 'labels' || /\/l\/v[^/]+\//.test(url.pathname); }
  catch { return false; }
}
export function useLabelLinkSnapshot(source: LabelLinkSource, parse: (source: string) => LabelReference) {
  const generation = useRef(0);
  const [pending, setPending] = useState<{ reference?: LabelReference; invalid: boolean; revision: number }>({ invalid: false, revision: 0 });
  const capture = useCallback((value: string) => {
    const revision = ++generation.current;
    try { setPending({ reference: parse(value), invalid: false, revision }); }
    catch { setPending({ invalid: true, revision }); }
  }, [parse]);
  const clear = useCallback(() => { const revision = ++generation.current; setPending({ invalid: false, revision }); }, []);
  useEffect(() => {
    let active = true; const initialGeneration = generation.current;
    void source.getInitialURL().then(value => {
      if (active && generation.current === initialGeneration && value && isLabelLink(value)) capture(value);
    }).catch(() => undefined);
    const remove = source.subscribe(value => { if (active && isLabelLink(value)) capture(value); });
    return () => { active = false; remove(); };
  }, [source, capture]);
  return { ...pending, capture, clear };
}
