import { useRef } from 'react';
import { ScrollView, StyleSheet, Text, View, type TextInputProps } from 'react-native';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { PendingHomeReturn } from './useHomeReturnActions';

export function HomeReturnDetailsSheet({ pendingReturn, canReturn, onCancel, onClose, onChangeDetails, onSave }: {
  readonly pendingReturn: PendingHomeReturn | undefined;
  readonly canReturn: boolean;
  readonly onCancel: () => void;
  readonly onClose: () => void;
  readonly onChangeDetails: (details: string) => void;
  readonly onSave: () => void;
}) {
  const colors = useAppearanceAwarePalette();
  if (!pendingReturn) return null;
  const busy = pendingReturn.isSaving;
  const close = () => { if (!busy) { if (canReturn) onCancel(); else onClose(); } };
  return <ScrollView style={{ backgroundColor: colors.background }} contentContainerStyle={styles.content} contentInsetAdjustmentBehavior="automatic" automaticallyAdjustKeyboardInsets
        keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
        <Text style={[styles.asset, { color: colors.textMuted }]}>{pendingReturn.asset.title}</Text>
        {!canReturn ? <Text accessibilityRole="alert" style={{ color: colors.textMuted }}>
          Your access changed. The item is already returned. You can close this sheet, but cannot save details or cancel the return.
        </Text> : null}
        {pendingReturn.error ? <Text accessibilityLabel="Return details error" accessibilityRole="alert" style={{ color: colors.danger }}>
          <Text>{pendingReturn.error.title}</Text>{'\n'}{pendingReturn.error.message}
        </Text> : null}
        {!pendingReturn.undoableOperationId ? <Text style={{ color: colors.textMuted }}>This return cannot be canceled.</Text> : null}
        <Text style={{ color: colors.text }}>Optional return details</Text>
        <ReturnNoteInput key={pendingReturn.sessionId} accessibilityLabel="Optional return details" multiline editable={canReturn && !busy}
          initialValue={pendingReturn.details} onChangeText={onChangeDetails} textAlignVertical="top"
          style={[styles.input, { color: colors.text, backgroundColor: colors.surface, borderColor: colors.controlBorder }]} />
        <View style={styles.actions}>
          <View style={styles.action}><NativeCommandButton label={canReturn && pendingReturn.undoableOperationId ? 'Cancel return' : 'Close'}
            disabled={busy} onPress={close} /></View>
          {canReturn ? <View style={styles.action}><NativeCommandButton label={busy ? (pendingReturn.operation === 'undo' ? 'Canceling return...' : 'Saving...') : 'Save'} disabled={busy} onPress={onSave} /></View> : null}
        </View>
      </ScrollView>;
}

const styles = StyleSheet.create({
  content: { padding: 20, gap: 16, flexGrow: 1 },
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: 16 },
  action: { flexGrow: 1, flexBasis: 120 },
  asset: { fontSize: 17 },
  input: { minHeight: 160, borderWidth: 1, borderRadius: 10, padding: 12, fontSize: 17 }
});

function ReturnNoteInput({ initialValue, ...props }: TextInputProps & { readonly initialValue: string }) {
  const seed = useRef(initialValue);
  return <AppTextInput {...props} defaultValue={seed.current} />;
}
