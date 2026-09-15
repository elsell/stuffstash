import { useEffect, useState } from 'react';
import { AccessibilityInfo } from 'react-native';

export function useNoticeAccessibility() {
  // Keep feedback still and available until the platform preferences arrive.
  const [reduceMotion, setReduceMotion] = useState(true);
  const [screenReader, setScreenReader] = useState(true);
  useEffect(() => {
    let active = true;
    let motionChanged = false;
    let readerChanged = false;
    const motion = AccessibilityInfo.addEventListener('reduceMotionChanged', value => {
      motionChanged = true;
      setReduceMotion(value);
    });
    const reader = AccessibilityInfo.addEventListener('screenReaderChanged', value => {
      readerChanged = true;
      setScreenReader(value);
    });
    void AccessibilityInfo.isReduceMotionEnabled().then(value => {
      if (active && !motionChanged) setReduceMotion(value);
    }).catch(() => {});
    void AccessibilityInfo.isScreenReaderEnabled().then(value => {
      if (active && !readerChanged) setScreenReader(value);
    }).catch(() => {});
    return () => { active = false; motion.remove(); reader.remove(); };
  }, []);
  return { reduceMotion, screenReader };
}
