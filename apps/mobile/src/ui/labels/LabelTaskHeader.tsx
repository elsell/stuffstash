import { useMemo } from 'react';
import { Stack, useRouter } from 'expo-router';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { t } from '../../presentation/localization';
export function LabelTaskHeader({ onCancel }: { readonly onCancel?: () => void }) {
  const router = useRouter();
  const actions = useNativeHeaderActionOptions([{ kind: 'close', label: t('labels.mobile.cancel'), onPress: () => {
    onCancel?.(); if (router.canGoBack()) router.back(); else router.replace('/');
  } }], 'left');
  const options = useMemo(() => ({ ...actions }), [actions]);
  return <Stack.Screen options={options} />;
}
