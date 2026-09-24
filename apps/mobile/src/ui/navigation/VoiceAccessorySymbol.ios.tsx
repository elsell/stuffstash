import { Host, Image } from '@expo/ui/swift-ui';
import { foregroundStyle } from '@expo/ui/swift-ui/modifiers';
import { View } from 'react-native';
import type { VoiceAccessorySymbolProps } from './VoiceAccessorySymbol.types';

export function VoiceAccessorySymbol({ name, color, size }: VoiceAccessorySymbolProps) {
  return <View pointerEvents="none" accessibilityElementsHidden style={{ width: size, height: size }}>
    <Host style={{ width: size, height: size }}>
      <Image systemName={name === 'microphone' ? 'mic' : 'paperplane.fill'} size={size}
        modifiers={[foregroundStyle(color)]} />
    </Host>
  </View>;
}
