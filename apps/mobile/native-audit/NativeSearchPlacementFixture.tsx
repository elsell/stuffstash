import { ScrollView, Text } from 'react-native';

/** Search is configured only at route registration, before this view mounts. */
export function NativeSearchPlacementFixture() {
  return <ScrollView contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled">
    <Text>Static native search placement comparison</Text>
  </ScrollView>;
}
