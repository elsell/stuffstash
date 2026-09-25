import type { ComponentProps } from 'react';
import { Image } from '@expo/ui/swift-ui';
import { buttonStyle, controlSize, frame, tint } from '@expo/ui/swift-ui/modifiers';

// The same native label and control layout keeps Menu and Button equal despite
// different SF Symbol glyph bounds (for example ellipsis versus xmark).
export const nativePhotoControlHost = { width: 54, height: 48 };
export const nativePhotoControlModifiers = [buttonStyle('bordered'), controlSize('large'), tint('#FFFFFF')];
export function NativePhotoSymbol({ symbol }: { readonly symbol: ComponentProps<typeof Image>['systemName'] }) {
  return <Image color="#FFFFFF" size={20} systemName={symbol}
    modifiers={[frame({ width: 20, height: 20 })]} />;
}
