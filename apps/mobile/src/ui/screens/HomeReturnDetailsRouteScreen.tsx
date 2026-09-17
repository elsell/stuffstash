import { useCallback, useRef } from 'react';
import { router, Stack, useFocusEffect, useNavigation } from 'expo-router';
import { usePreventRemove } from '@react-navigation/native';
import { useHomeReturnTask } from '../navigation/HomeReturnTaskPresentation';

export default function HomeReturnDetailsRoute() {
  const task = useHomeReturnTask();
  const navigation = useNavigation();
  const current = useRef(task); current.current = task;
  const hasTask = Boolean(task);
  useFocusEffect(useCallback(() => {
    const active = current.current;
    if (!active) { if (router.canGoBack()) router.back(); else router.replace('/'); return; }
    active.focusChanged(true);
    return () => active.focusChanged(false);
  }, [hasTask]));
  // Keep native removal configuration stable while the completed task exits.
  usePreventRemove(true, ({ data }) => {
    if (current.current) current.current.requestClose();
    else navigation.dispatch(data.action);
  });
  return <><Stack.Screen options={returnDetailsOptions} />{task?.content}</>;
}

const returnDetailsOptions = { title: 'Return details', headerShown: true, gestureEnabled: false, headerBackVisible: false } as const;
