import React from 'react';
import { Pressable, StyleSheet } from 'react-native';
import { ChevronLeft, ChevronRight, Trash2, X } from 'lucide-react-native';
import { radius } from '../theme/tokens';
import { photoViewerActions, type PhotoViewerActionButtonProps } from './PhotoViewerActionButton.types';

export function PhotoViewerActionButton({ action, disabled = false, onPress }: PhotoViewerActionButtonProps) {
  const Icon = { close: X, previous: ChevronLeft, next: ChevronRight, remove: Trash2 }[action];
  return <Pressable accessibilityLabel={photoViewerActions[action].label} accessibilityRole="button"
    accessibilityState={{ disabled }} disabled={disabled} onPress={disabled ? undefined : onPress}
    hitSlop={10} style={[styles.button, action === 'remove' && styles.destructive, disabled && styles.disabled]}>
    <Icon color={action === 'remove' ? '#F5B95E' : '#FFFFFF'} size={action === 'remove' ? 24 : action === 'close' ? 25 : 27} strokeWidth={action === 'remove' ? 2.4 : 2.5} />
  </Pressable>;
}
const styles = StyleSheet.create({
  button: { alignItems: 'center', justifyContent: 'center', width: 54, height: 46, borderRadius: radius.md },
  destructive: { backgroundColor: 'rgba(249, 189, 73, 0.12)' },
  disabled: { opacity: 0.35 }
});
