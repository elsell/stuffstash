import { NativeTabs } from 'expo-router/unstable-native-tabs';
import { VoiceBottomAccessory } from '../../ui/navigation/VoiceBottomAccessory';

export default function TabLayout() {
  return (
    <NativeTabs>
      <NativeTabs.BottomAccessory>
        <VoiceBottomAccessory />
      </NativeTabs.BottomAccessory>
      <NativeTabs.Trigger name="(home)">
        <NativeTabs.Trigger.Label>Home</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon md="home" sf={{ default: 'house', selected: 'house.fill' }} />
      </NativeTabs.Trigger>
      <NativeTabs.Trigger name="(search)">
        <NativeTabs.Trigger.Label>Browse</NativeTabs.Trigger.Label>
        <NativeTabs.Trigger.Icon md="grid_view" sf={{ default: 'square.grid.2x2', selected: 'square.grid.2x2.fill' }} />
      </NativeTabs.Trigger>
    </NativeTabs>
  );
}
