import { useReducedMotionPreference } from '../accessibility/useReducedMotionPreference';
import { useEffect, useState } from 'react';
import { AccessibilityInfo } from 'react-native';

export function useNoticeAccessibility() {
  // Keep feedback still and available until the platform preferences arrive.
  const reduceMotion = useReducedMotionPreference();
  const [screenReader, setScreenReader] = useState(true);
  useEffect(() => {
    let active = true;
    let readerChanged = false;
    const reader = AccessibilityInfo.addEventListener('screenReaderChanged', value => {
      readerChanged = true;
      setScreenReader(value);
    });
    void AccessibilityInfo.isScreenReaderEnabled().then(value => {
      if (active && !readerChanged) setScreenReader(value);
    }).catch(() => {});
    return () => { active = false; reader.remove(); };
  }, []);
  return { reduceMotion, screenReader };
}
