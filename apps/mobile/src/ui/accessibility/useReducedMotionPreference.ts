import { useEffect, useState } from 'react';
import { AccessibilityInfo } from 'react-native';

/** Keep custom motion still until the native preference is known. */
export function useReducedMotionPreference(): boolean {
  const [reduced, setReduced] = useState(true);
  useEffect(() => {
    let active = true;
    let changed = false;
    const subscription = AccessibilityInfo.addEventListener('reduceMotionChanged', value => {
      changed = true;
      if (active) setReduced(value);
    });
    void AccessibilityInfo.isReduceMotionEnabled().then(value => {
      if (active && !changed) setReduced(value);
    }).catch(() => { /* Keep the conservative default or latest live preference. */ });
    return () => { active = false; subscription.remove(); };
  }, []);
  return reduced;
}
