import { useCommittedCommand } from './useCommittedCommand';
import { Button, Host, HStack, Spacer, Text } from '@expo/ui/swift-ui';
import { accessibilityLabel as nativeLabel, buttonStyle, contentShape, disabled as nativeDisabled, foregroundStyle, frame, padding, shapes } from '@expo/ui/swift-ui/modifiers';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';
import { useAppearanceAwarePalette } from '../theme/appearance';

export function NativeActionRow({ label, accessibilityLabel = label, disabled, role, onPress }: NativeCommandButtonProps) {
  const press = useCommittedCommand(onPress, disabled);
  const palette = useAppearanceAwarePalette();
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 48 }}>
    <Button onPress={press} role={role === 'destructive' ? 'destructive' : undefined}
      modifiers={[buttonStyle('plain'), nativeDisabled(!!disabled), nativeLabel(accessibilityLabel)]}>
      <HStack modifiers={[padding({ horizontal: 16, vertical: 12 }), frame({ minHeight: 48, maxWidth: Infinity }), contentShape(shapes.rectangle())]}>
        <Text modifiers={[foregroundStyle(role === 'destructive' ? palette.danger : palette.action)]}>{label}</Text><Spacer />
      </HStack>
    </Button>
  </Host>;
}
