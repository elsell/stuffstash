import { NativeTabs } from 'expo-router/unstable-native-tabs';
import { VoiceAccessoryContent } from './VoiceAccessoryContent';

export function VoiceBottomAccessory() {
  const placement = NativeTabs.BottomAccessory.usePlacement();
  return <VoiceAccessoryContent placement={placement} />;
}
