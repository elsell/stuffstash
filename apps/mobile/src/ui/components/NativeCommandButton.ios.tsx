import { useLayoutEffect, useRef, useState } from 'react';
import { requireNativeView } from 'expo';
import type { ViewStyle } from 'react-native';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

type NativeProps = Omit<NativeCommandButtonProps, 'onPress'> & {
  onPress: () => void;
  onSizeChange: (event: { nativeEvent: { height: number } }) => void;
  style: ViewStyle;
};
const Command = requireNativeView<NativeProps>('StuffStashCommandButton');
const minimumTarget = 48;

export function NativeCommandButton({ label, accessibilityLabel = label, disabled = false, onPress, prominence = 'secondary', role = 'default' }: NativeCommandButtonProps) {
  const [height, setHeight] = useState(minimumTarget);
  const current = useRef<(() => void) | null>(null);
  useLayoutEffect(() => {
    current.current = disabled ? null : onPress;
    return () => { current.current = null; };
  }, [disabled, onPress]);
  return <Command label={label} accessibilityLabel={accessibilityLabel} disabled={disabled}
    prominence={prominence} role={role} onPress={() => current.current?.()}
    onSizeChange={({ nativeEvent }) => {
      if (Number.isFinite(nativeEvent.height) && nativeEvent.height >= minimumTarget) setHeight(nativeEvent.height);
    }} style={{ width: '100%', height }} />;
}
