import type { PropsWithChildren } from 'react';
import { KeyboardAvoidingView, Platform, type StyleProp, type ViewStyle } from 'react-native';
import { useHeaderHeight } from '@react-navigation/elements';

/** iOS sheets retain padding; Android stack forms resize below their native header. */
export function AssetActionKeyboardFrame({ children, style }: PropsWithChildren<{ style: StyleProp<ViewStyle> }>) {
  const headerHeight = useHeaderHeight();
  return <KeyboardAvoidingView style={style}
    behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    keyboardVerticalOffset={Platform.OS === 'android' ? headerHeight : 0}>
    {children}
  </KeyboardAvoidingView>;
}
