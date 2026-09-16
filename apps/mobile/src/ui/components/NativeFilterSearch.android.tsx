import { useCallback, useLayoutEffect, useRef } from 'react';
import { useFocusEffect } from 'expo-router';
import { AppTextInput } from './AppTextInput';
import type { NativeFilterSheetProps } from './NativeFilterSheet.types';
import { useAppearancePalette } from '../theme/AppearanceContext';

/** A keyed page owns its input events until blur or removal. */
export function NativeFilterSearch(search: NonNullable<NativeFilterSheetProps['search']>) {
  const palette = useAppearancePalette();
  const current = useRef<typeof search | undefined>(undefined);
  const active = useRef(false);
  useLayoutEffect(() => {
    current.current = search;
    return () => { current.current = undefined; };
  }, [search]);
  useFocusEffect(useCallback(() => {
    active.current = true;
    return () => { active.current = false; };
  }, []));
  const owner = () => active.current ? current.current : undefined;
  return <AppTextInput accessibilityLabel={search.placeholder} placeholder={search.placeholder}
    value={search.query} onChangeText={text => {
      const latest = owner(); if (!latest) return;
      if (text) latest.onChange(text); else latest.onClear();
    }} onSubmitEditing={event => owner()?.onSubmit(event.nativeEvent.text)}
    returnKeyType="search" autoCapitalize="none" autoCorrect={false} multiline={false}
    style={{ marginHorizontal: 20, marginTop: 16, paddingHorizontal: 16, paddingVertical: 12,
      minHeight: 48, fontSize: 16, borderWidth: 1, borderRadius: 8,
      color: palette.text, borderColor: palette.textMuted }}
    placeholderTextColor={palette.textMuted} />;
}
