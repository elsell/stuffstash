import { useLayoutEffect, useMemo, useRef } from 'react';
import { requireNativeView } from 'expo';
import { StyleSheet, Text, View, type ViewProps } from 'react-native';
import { nativeTagColorSelection } from './NativeTagColorPickerPresentation';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { minimumTouchTargetSize } from '../theme/tokens';

type ColorWellProps = ViewProps & {
  readonly selection: string | null;
  readonly enabled: boolean;
  readonly onSelectionChange: (event: { readonly nativeEvent: { readonly value: string } }) => void;
};

export function NativeTagColorPicker({ disabled, onChange, value }: { readonly disabled: boolean; readonly onChange: (value: string) => void; readonly value: string }) {
  // Resolve only when the availability-gated adapter mounts, preserving stale-binary fallback.
  const ColorWell = useMemo(() => requireNativeView<ColorWellProps>('StuffStashColorWell'), []);
  const palette = useAppearancePalette();
  const committed = useRef<{ disabled: boolean; onChange: (value: string) => void } | null>(null);
  useLayoutEffect(() => {
    committed.current = { disabled, onChange };
    return () => { committed.current = null; };
  }, [disabled, onChange]);
  return <View style={styles.row}>
    <Text accessible={false} style={[styles.label, { color: palette.text }, disabled && styles.disabled]}>Choose any color</Text>
    <ColorWell style={styles.well} selection={nativeTagColorSelection(value)} enabled={!disabled}
      onSelectionChange={event => {
        const selected = nativeTagColorSelection(event.nativeEvent.value);
        const owner = committed.current;
        if (owner && !owner.disabled && selected) owner.onChange(selected);
      }} />
  </View>;
}

const styles = StyleSheet.create({
  row: { minHeight: minimumTouchTargetSize, width: '100%', flexDirection: 'row', alignItems: 'center', gap: 12 },
  label: { flex: 1, fontSize: 17 },
  well: { width: minimumTouchTargetSize, height: minimumTouchTargetSize },
  disabled: { opacity: 0.55 }
});
