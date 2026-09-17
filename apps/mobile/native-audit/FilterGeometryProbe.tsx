import { useEffect, useRef, useState, type RefObject } from 'react';
import { Keyboard, Text, View, type KeyboardEvent, type KeyboardMetrics } from 'react-native';
import { keyboardBoundaryInset, type SheetBottomBoundary } from '../src/ui/components/keyboardBoundaryInset';

type Snapshot = { boundary: SheetBottomBoundary; keyboard: KeyboardMetrics; calculatedInset: number };

/** Fixture-only observation; no footer offsets or keyboard behavior are changed. */
export function useFilterGeometryProbe(ref: RefObject<Pick<View, 'measureInWindow'> | null>) {
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  useEffect(() => {
    let generation = 0;
    let active = true;
    const measure = (keyboard: KeyboardMetrics) => {
      const request = ++generation;
      ref.current?.measureInWindow((x, y, width) => {
        if (!active || request !== generation) return;
        const boundary = { x, y, width };
        setSnapshot({ boundary, keyboard, calculatedInset: keyboardBoundaryInset(boundary, keyboard) });
      });
    };
    const changed = (event: KeyboardEvent) => measure(event.endCoordinates);
    const subscriptions = [
      Keyboard.addListener('keyboardDidShow', changed),
      Keyboard.addListener('keyboardDidChangeFrame', changed),
      Keyboard.addListener('keyboardDidHide', () => { generation++; setSnapshot(null); })
    ];
    const current = Keyboard.metrics();
    if (current) measure(current);
    return () => { active = false; generation++; subscriptions.forEach(subscription => subscription.remove()); };
  }, [ref]);
  return snapshot;
}

export function FilterGeometryProbe() {
  const boundary = useRef<View>(null);
  const snapshot = useFilterGeometryProbe(boundary);
  return <>
    <View ref={boundary} collapsable={false} pointerEvents="none" style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 0 }} />
    {snapshot ? <Text testID="audit-filter-geometry" pointerEvents="none"
      style={{ position: 'absolute', top: 140, left: 20, right: 20, fontSize: 8 }}>{JSON.stringify(snapshot)}</Text> : null}
  </>;
}
