import React, { useLayoutEffect, useRef } from 'react';
import { Button, Host } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, controlSize, disabled as nativeDisabled, labelStyle, tint } from '@expo/ui/swift-ui/modifiers';
import { photoViewerActions, type PhotoViewerActionButtonProps } from './PhotoViewerActionButton.types';

export function PhotoViewerActionButton({ action, disabled = false, onPress }: PhotoViewerActionButtonProps) {
  const current = useRef<(() => void) | null>(null);
  useLayoutEffect(() => {
    current.current = disabled ? null : onPress;
    return () => { current.current = null; };
  }, [disabled, onPress]);
  const { label, symbol } = photoViewerActions[action];
  return <Host style={{ width: 54, height: 48 }}>
    <Button label={label} systemImage={symbol} role={action === 'remove' ? 'destructive' : undefined}
      modifiers={[accessibilityLabel(label), buttonStyle('bordered'), controlSize('large'), labelStyle('iconOnly'), tint('#FFFFFF'), nativeDisabled(disabled)]}
      onPress={() => current.current?.()} />
  </Host>;
}
