import { useCallback, useEffect, useRef, useState, type RefObject } from 'react';
import { Keyboard, Platform, type KeyboardEvent, type KeyboardMetrics } from 'react-native';
import { keyboardBoundaryInset } from './keyboardBoundaryInset';
import type { SheetBoundaryPort } from './SheetBoundaryPort';

/** Measure the unmoved sheet edge, never the footer whose position this controls. */
export function useSheetKeyboardInset(boundaryRef: RefObject<SheetBoundaryPort | null>) {
  const [bottomInset, setBottomInset] = useState(0);
  const keyboardFrame = useRef<KeyboardMetrics | undefined>(undefined);
  const generation = useRef(0);
  const mounted = useRef(false);
  const measure = useCallback(() => {
    const request = ++generation.current;
    const frame = keyboardFrame.current;
    if (!mounted.current) return;
    if (!frame) { setBottomInset(0); return; }
    const boundary = boundaryRef.current;
    if (!boundary) { setBottomInset(0); return; }
    void boundary.measureInKeyboardWindow().then(measured => {
      if (mounted.current && request === generation.current) {
        setBottomInset(measured ? keyboardBoundaryInset(measured, frame) : 0);
      }
    }, () => {
      if (mounted.current && request === generation.current) setBottomInset(0);
    });
  }, [boundaryRef]);

  useEffect(() => {
    mounted.current = true;
    const changed = (event: KeyboardEvent) => {
      keyboardFrame.current = event.endCoordinates;
      Keyboard.scheduleLayoutAnimation(event);
      measure();
    };
    const settled = (event: KeyboardEvent) => {
      keyboardFrame.current = event.endCoordinates;
      measure();
    };
    const hidden = () => { keyboardFrame.current = undefined; measure(); };
    const subscriptions = [
      Keyboard.addListener(Platform.OS === 'ios' ? 'keyboardWillChangeFrame' : 'keyboardDidShow', changed),
      Keyboard.addListener('keyboardDidHide', hidden)
    ];
    if (Platform.OS === 'ios') {
      subscriptions.push(
        Keyboard.addListener('keyboardDidChangeFrame', settled),
        Keyboard.addListener('keyboardDidShow', settled)
      );
    }
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
