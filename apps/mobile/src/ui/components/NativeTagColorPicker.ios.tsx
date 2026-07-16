import { ColorPicker, Host } from '@expo/ui/swift-ui';
import { StyleSheet, View } from 'react-native';
import { nativeTagColorInteraction, nativeTagColorSelection } from './NativeTagColorPickerPresentation';

export function NativeTagColorPicker({ disabled, onChange, value }: { readonly disabled: boolean; readonly onChange: (value: string) => void; readonly value: string }) {
  const interaction = nativeTagColorInteraction(disabled, onChange);
  return <View accessibilityLabel="Full color picker" accessibilityState={{ disabled }} pointerEvents={interaction.pointerEvents} style={[styles.host, disabled && styles.disabled]}>
    <Host matchContents style={styles.host}>
      <ColorPicker label="Choose any color" selection={nativeTagColorSelection(value)} supportsOpacity={false} onSelectionChange={interaction.onSelectionChange} />
    </Host>
  </View>;
}

const styles = StyleSheet.create({ disabled: { opacity: 0.55 }, host: { minHeight: 44, width: '100%' } });
