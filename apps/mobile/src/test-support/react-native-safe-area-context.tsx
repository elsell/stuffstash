import React, { type ReactNode } from 'react';

export function SafeAreaView(props: { readonly children?: ReactNode; readonly [key: string]: unknown }) {
  return React.createElement('SafeAreaView', props, props.children);
}
export const useSafeAreaInsets = () => ({ top: 0, right: 0, bottom: 0, left: 0 });
