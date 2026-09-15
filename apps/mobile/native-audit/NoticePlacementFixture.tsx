import { useState } from 'react';
import { Stack, usePathname, useRouter } from 'expo-router';
import { useNativeHeaderActionOptions } from '../src/ui/components/useNativeHeaderActionOptions';
import { Button, ScrollView, Text } from 'react-native';
import { useAppFeedback } from '../src/ui/feedback/AppFeedback';

/** Exercises the shared production presenter; no network, clipboard or external action. */
export function NoticePlacementFixture() {
  const router = useRouter();
  const pathname = usePathname();
  const closeOptions = useNativeHeaderActionOptions([{ kind: 'close', label: 'Close notice fixture', onPress: () => router.back() }]);
  const feedback = useAppFeedback();
  const [actions, setActions] = useState(0);
  return <>
    {pathname.endsWith('audit-notice-sheet') ? <Stack.Screen options={closeOptions} /> : null}
    <ScrollView testID="notice-placement-content" contentInsetAdjustmentBehavior="automatic">
    <Text>Notice actions completed: {actions}</Text>
    <Button title="Show placement notice" onPress={() => feedback.showNotice({
      tone: 'info', title: 'Audit notice', message: 'A retained action must leave navigation reachable',
      action: { label: 'Complete audit action', onPress: () => setActions(current => current + 1) }
    })} />
  </ScrollView></>;
}
