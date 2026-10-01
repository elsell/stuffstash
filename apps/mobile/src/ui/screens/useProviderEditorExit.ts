import { t } from '../../presentation/localization';
import { useEffect, useRef, useState } from 'react';
import { Alert } from 'react-native';
import { useNavigation } from 'expo-router';
import { usePreventRemove } from '@react-navigation/native';

type Exit = { readonly canPresent: () => boolean; readonly leave: () => void };

/** Disarm the native removal guard before a confirmed or successfully saved exit. */
export function useProviderEditorExit({ dirty, isSaving, capturePresentation }: {
  readonly dirty: boolean;
  readonly isSaving: () => boolean;
  readonly capturePresentation: () => () => boolean;
}) {
  const navigation = useNavigation();
  const [exit, setExit] = useState<Exit>();
  const consumed = useRef<Exit | undefined>(undefined);
  usePreventRemove(!exit, ({ data }) => {
    if (isSaving()) return;
    const canPresent = capturePresentation();
    if (!canPresent()) return;
    if (!dirty) {
      setExit({ canPresent, leave: () => navigation.dispatch(data.action) });
      return;
    }
    let confirmed = false;
    Alert.alert(t('mobile.useProviderEditorExit.discardChanges'), t('mobile.useProviderEditorExit.yourUnsavedReplacementWillBeLost'), [
      { text: t('mobile.useProviderEditorExit.keepEditing'), style: 'cancel' },
      { text: t('mobile.useProviderEditorExit.discard'), style: 'destructive', onPress: () => {
        if (confirmed || isSaving() || !canPresent()) return;
        confirmed = true;
        setExit({ canPresent, leave: () => navigation.dispatch(data.action) });
      } }
    ]);
  });
  useEffect(() => {
    if (!exit || consumed.current === exit) return;
    consumed.current = exit;
    if (!exit.canPresent()) { setExit(undefined); return; }
    try { exit.leave(); }
    catch (error) { setExit(undefined); throw error; }
  }, [exit]);
  return (canPresent: () => boolean, leave: () => void) => {
    if (canPresent()) setExit({ canPresent, leave });
  };
}
