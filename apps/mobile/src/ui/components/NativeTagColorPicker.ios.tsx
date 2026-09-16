import { ColorPicker, Host, HStack, Spacer, Text } from '@expo/ui/swift-ui';
import { controlSize, fixedSize, frame, labelsHidden } from '@expo/ui/swift-ui/modifiers';
import { StyleSheet, View } from 'react-native';
import { nativeTagColorInteraction, nativeTagColorSelection } from './NativeTagColorPickerPresentation';

export function NativeTagColorPicker({ disabled, onChange, value }: { readonly disabled: boolean; readonly onChange: (value: string) => void; readonly value: string }) {
  const interaction = nativeTagColorInteraction(disabled, onChange);
  return <View accessibilityLabel="Full color picker" accessibilityState={{ disabled }} pointerEvents={interaction.pointerEvents} style={[styles.host, disabled && styles.disabled]}>
    <Host matchContents style={styles.host}>
      <HStack>
        <Text modifiers={[fixedSize({ horizontal: false, vertical: true })]}>Choose any color</Text>
        <Spacer />
        <ColorPicker label="Choose any color" selection={nativeTagColorSelection(value)} supportsOpacity={false}
          onSelectionChange={interaction.onSelectionChange}
          modifiers={[labelsHidden(), controlSize('large'), frame({ minWidth: 44, minHeight: 44 })]} />
      </HStack>
    </Host>
  </View>;
}

const styles = StyleSheet.create({ disabled: { opacity: 0.55 }, host: { minHeight: 44, width: '100%' } });
