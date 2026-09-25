import React, { useLayoutEffect, useRef } from 'react';
import { Button, Host } from '@expo/ui/swift-ui';
import { accessibilityLabel, disabled as nativeDisabled } from '@expo/ui/swift-ui/modifiers';
import { NativePhotoSymbol, nativePhotoControlModifiers } from './NativePhotoSymbol.ios';
import { photoViewerActions, type PhotoViewerActionButtonProps } from './PhotoViewerActionButton.types';

export function PhotoViewerActionButton({ action, disabled = false, onPress }: PhotoViewerActionButtonProps) {
  const current = useRef<(() => void) | null>(null);
  useLayoutEffect(() => {
    current.current = disabled ? null : onPress;
    return () => { current.current = null; };
  }, [disabled, onPress]);
  const { label, symbol } = photoViewerActions[action];
  return <Host matchContents>
    <Button role={action === 'remove' ? 'destructive' : undefined}
      modifiers={[accessibilityLabel(label), ...nativePhotoControlModifiers, nativeDisabled(disabled)]}
      onPress={() => current.current?.()}><NativePhotoSymbol symbol={symbol} /></Button>
  </Host>;
}
