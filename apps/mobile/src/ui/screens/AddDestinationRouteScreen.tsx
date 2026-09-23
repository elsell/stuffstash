import { useCallback, useEffect, useLayoutEffect, useRef } from 'react';
import { router, useFocusEffect, useNavigation } from 'expo-router';
import { usePreventRemove } from '@react-navigation/native';
import { useAddDestinationTask } from '../navigation/AddDestinationTask';
export default function AddDestinationRouteScreen() {
  const task = useAddDestinationTask();
  const navigation = useNavigation();
  const current = useRef(task);
  useLayoutEffect(() => { current.current = task; }, [task]);
  useEffect(() => () => current.current?.cancel(), []);
  // Keep native removal configuration stable while a completed task exits.
  usePreventRemove(true, ({ data }) => {
    if (current.current) current.current.cancel();
    else navigation.dispatch(data.action);
  });
  const hasTask = Boolean(task);
  useFocusEffect(useCallback(() => {
    if (!current.current) { if (router.canGoBack()) router.back(); else router.replace('/'); }
  }, [hasTask]));
  return task?.content ?? null;
}
