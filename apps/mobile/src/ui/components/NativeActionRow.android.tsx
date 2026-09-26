import { useCommittedCommand } from './useCommittedCommand';
import { NativeComposeHost as Host } from './NativeComposeHost.android';
import { Row, Text, TextButton } from '@expo/ui/jetpack-compose';
import { fillMaxWidth } from '@expo/ui/jetpack-compose/modifiers';
import { nativeContentDescription } from './NativeComposeAccessibility.android';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';
import { useAppearanceAwarePalette } from '../theme/appearance';

export function NativeActionRow({ label, accessibilityLabel = label, disabled, role, onPress }: NativeCommandButtonProps) {
  const press = useCommittedCommand(onPress, disabled);
  const palette = useAppearanceAwarePalette();
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 48 }}>
    <TextButton enabled={!disabled} modifiers={[fillMaxWidth(), nativeContentDescription(accessibilityLabel)]}
      contentPadding={{ start: 16, top: 12, end: 16, bottom: 12 }}
      colors={{ contentColor: role === 'destructive' ? palette.danger : palette.action }} onClick={press}>
      <Row modifiers={[fillMaxWidth()]}><Text style={{ fontSize: 17 }}>{label}</Text></Row>
    </TextButton>
  </Host>;
}
