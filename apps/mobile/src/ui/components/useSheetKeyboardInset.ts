import { useCallback, useEffect, useRef, useState, type RefObject } from 'react';
import { Keyboard, Platform, type KeyboardEvent, type KeyboardMetrics, type View } from 'react-native';
import { keyboardBoundaryInset } from './keyboardBoundaryInset';

/** Measure the unmoved sheet edge, never the footer whose position this controls. */
export function useSheetKeyboardInset(boundaryRef: RefObject<Pick<View, 'measureInWindow'> | null>) {
  const [bottomInset, setBottomInset] = useState(0);
  const keyboardFrame = useRef<KeyboardMetrics | undefined>(undefined);
  const generation = useRef(0);
  const mounted = useRef(false);
  const measure = useCallback(() => {
    const request = ++generation.current;
    const frame = keyboardFrame.current;
    if (!mounted.current) return;
    if (!frame) { setBottomInset(0); return; }
    boundaryRef.current?.measureInWindow((x, y, width) => {
      if (mounted.current && request === generation.current) {
        setBottomInset(keyboardBoundaryInset({ x, y, width }, frame));
      }
    });
  }, [boundaryRef]);

  useEffect(() => {
    mounted.current = true;
    const changed = (event: KeyboardEvent) => {
      keyboardFrame.current = event.endCoordinates;
      Keyboard.scheduleLayoutAnimation(event);
      measure();
    };
    const hidden = () => { keyboardFrame.current = undefined; measure(); };
    const subscriptions = [
      Keyboard.addListener(Platform.OS === 'ios' ? 'keyboardWillChangeFrame' : 'keyboardDidShow', changed),
      Keyboard.addListener('keyboardDidHide', hidden)
    ];
    keyboardFrame.current = Keyboard.metrics();
    measure();
    return () => {
      mounted.current = false;
      generation.current++;
      subscriptions.forEach(subscription => subscription.remove());
    };
  }, [measure]);

  return { bottomInset, measure };
}
