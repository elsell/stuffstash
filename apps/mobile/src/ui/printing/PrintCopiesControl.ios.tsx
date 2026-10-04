import { useCallback, useLayoutEffect, useRef } from 'react';
import { Host, Stepper } from '@expo/ui/swift-ui';
import { accessibilityLabel, accessibilityValue, disabled as nativeDisabled, fixedSize, padding } from '@expo/ui/swift-ui/modifiers';
import type { PrintCopiesControlProps } from './PrintCopiesControl';
export function PrintCopiesControl({ label, value, disabled, onChange }: PrintCopiesControlProps) {
  const current = useRef<PrintCopiesControlProps | undefined>(undefined);
  useLayoutEffect(() => { current.current = { label, value, disabled, onChange }; return () => { current.current = undefined; }; }, [label, value, disabled, onChange]);
  const change = useCallback((next: number) => {
    const owner = current.current;
    if (owner && !owner.disabled && Number.isSafeInteger(next) && next > 0) owner.onChange(String(next));
  }, []);
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 52 }}>
    <Stepper label={`${label}: ${value}`} value={Number(value)} min={1} max={Number.MAX_SAFE_INTEGER} step={1} onValueChange={change}
      modifiers={[fixedSize({ horizontal: false, vertical: true }), padding({ horizontal: 16, vertical: 12 }), nativeDisabled(disabled), accessibilityLabel(label), accessibilityValue(value)]} />
  </Host>;
}
