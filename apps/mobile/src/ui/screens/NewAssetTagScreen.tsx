import { t } from '../../presentation/localization';
import { useState } from 'react';
import { Stack } from 'expo-router';
import { Text, View } from 'react-native';
import { assetTagKeyFromDisplayName } from '../../domain/assets/AssetSummary';
import { applyInlineAssetTagResolution, canApplyInlineAssetTagResolution, resolveInlineAssetTag, type CreateAssetTagDraft } from '../../application/assets/AssetTagDraftResolution';
import { DraftTextField } from '../components/DraftTextField';
import { FullSpectrumTagColorPicker } from '../components/FullSpectrumTagColorPicker';
import { NativeFilterSheet } from '../components/NativeFilterSheet';
import { SettingsSection, useSettingsListStyles } from './SettingsList';
import type { AssetTagSelectionOption } from './AssetTagSelectionScreen';

/** A creation form within the selection visit; no server mutation until asset Save. */
export function NewAssetTagScreen({ tags, selectedIds, newTags, onDone, onCancel, available }: {
  readonly tags: readonly AssetTagSelectionOption[];
  readonly selectedIds: readonly string[];
  readonly newTags: readonly CreateAssetTagDraft[];
  readonly onDone: (ids: readonly string[], tags: readonly CreateAssetTagDraft[]) => void;
  readonly onCancel: () => void;
  readonly available: boolean;
}) {
  const [name, setName] = useState('');
  const [color, setColor] = useState('');
  const { palette, styles } = useSettingsListStyles();
  const resolution = resolveInlineAssetTag({ displayName: name, color,
    activeTags: tags.map(tag => ({ id: tag.id, key: tag.key ?? assetTagKeyFromDisplayName(tag.label) })), pendingTags: newTags });
  return <>
    <Stack.Screen options={{ title: t('mobile.NewAssetTagScreen.newTag') }} />
    <NativeFilterSheet title={t('mobile.NewAssetTagScreen.newTag')} footerTestID="new-asset-tag-actions" actions={{ primaryLabel: t('mobile.NewAssetTagScreen.addTag'), secondaryLabel: t('mobile.NewAssetTagScreen.cancel'),
      secondaryAccessibilityLabel: 'Cancel new tag', disabled: !available || !canApplyInlineAssetTagResolution(resolution),
      onApply: () => {
        if (!available) return;
        const next = applyInlineAssetTagResolution({ resolution, selectedTagIds: selectedIds, pendingTags: newTags });
        if (next.shouldClearInputs) onDone(next.selectedTagIds, next.pendingTags);
      }, onBack: onCancel }}>
      <SettingsSection title={t('mobile.NewAssetTagScreen.name')} footer="This tag will be saved with the asset.">
        <View style={styles.navigationRow}><DraftTextField style={[styles.rowLabel, { flex: 1, minHeight: 48 }]} placeholderTextColor={palette.textMuted} accessibilityLabel={t('mobile.NewAssetTagScreen.newTagName')} placeholder={t('mobile.NewAssetTagScreen.tagName')}
          value={name} onChangeText={setName} editable={available} /></View>
        {resolution.status === 'display_name_too_long' ? <Text accessibilityRole="alert" style={styles.rowContext}>{t('mobile.NewAssetTagScreen.useAShorterTagName')}</Text> : null}
      </SettingsSection>
      <SettingsSection title={t('mobile.NewAssetTagScreen.colorOptional')}><View style={styles.navigationRow}>
        <FullSpectrumTagColorPicker disabled={!available} value={color} onChange={setColor} />
      </View></SettingsSection>
    </NativeFilterSheet>
  </>;
}
