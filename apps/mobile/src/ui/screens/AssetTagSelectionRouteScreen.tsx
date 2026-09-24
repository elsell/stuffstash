import { useCallback, useEffect, useLayoutEffect, useRef } from 'react';
import { router, useFocusEffect } from 'expo-router';
import { useAssetTagSelectionTask } from '../navigation/AssetTagSelectionTask';

export default function AssetTagSelectionRouteScreen() {
  const task = useAssetTagSelectionTask();
  const current = useRef(task);
  useLayoutEffect(() => { current.current = task; }, [task]);
  useEffect(() => () => current.current?.cancel(), []);
  const hasTask = Boolean(task);
  useFocusEffect(useCallback(() => {
    if (!current.current) {
      if (router.canGoBack()) router.back(); else router.replace('/');
    }
  }, [hasTask]));
  return task?.content ?? null;
}
