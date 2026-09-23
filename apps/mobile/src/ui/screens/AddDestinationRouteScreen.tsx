import { useCallback, useEffect, useLayoutEffect, useRef } from 'react';
import { router, useFocusEffect } from 'expo-router';
import { usePreventRemove } from '@react-navigation/native';
import { useAddDestinationTask } from '../navigation/AddDestinationTask';
export default function AddDestinationRouteScreen() {
  const task = useAddDestinationTask();
  const current = useRef(task);
  useLayoutEffect(() => { current.current = task; }, [task]);
  useEffect(() => () => current.current?.cancel(), []);
  usePreventRemove(task?.blocked ?? false, () => {});
  const hasTask = Boolean(task);
  useFocusEffect(useCallback(() => {
    if (!current.current) { if (router.canGoBack()) router.back(); else router.replace('/'); }
  }, [hasTask]));
  return task?.content ?? null;
}
