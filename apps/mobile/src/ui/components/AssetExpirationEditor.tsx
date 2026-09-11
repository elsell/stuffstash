import { useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import type { EditDraft } from '../screens/AssetDetailEditPresentation';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';
import { ExpirationField } from './ExpirationField';

export function AssetExpirationEditor({ asset, draft, types, disabled, onChange }: {
  readonly asset: Pick<AssetDetailViewModel, 'id' | 'title' | 'description' | 'customAssetTypeId' | 'expiration'>;
  readonly draft?: EditDraft;
  readonly types?: readonly CustomAssetTypeDefinition[];
  readonly disabled: boolean;
  readonly onChange: (draft: EditDraft) => void;
}) {
  const colors = useAppearancePalette();
  const [initialPickerDate] = useState(() => new Date());
  const typeId = asset.customAssetTypeId ?? draft?.customAssetTypeId;
  const selectedType = types?.find((type) => type.id === typeId);
  const expiration = draft?.expiration === undefined ? asset.expiration : draft.expiration;
  const base = { title: asset.title, description: asset.description, ...draft };
  function selectType(id?: string) {
    if (disabled || asset.customAssetTypeId || id === typeId) return;
    onChange({ ...base, customAssetTypeId: id, expiration: null, expirationValid: true });
  }
  if (!types) return <Text style={{ color: colors.textMuted }}>Loading expiration settings…</Text>;
  return <View style={{ gap: spacing.sm }}>
    {!asset.customAssetTypeId && types.length ? <View>
      <Text style={{ color: colors.text }}>Custom type</Text>
      {[{ id: undefined, displayName: 'Base asset' }, ...types].map((type) => <Pressable key={type.id ?? 'base'} accessibilityRole="radio" accessibilityState={{ checked: typeId === type.id, disabled }} accessibilityLabel={type.displayName} disabled={disabled} onPress={() => selectType(type.id)} style={{ minHeight: 44, justifyContent: 'center' }}>
        <Text style={{ color: typeId === type.id ? colors.action : colors.text }}>{type.displayName}</Text>
      </Pressable>)}
    </View> : null}
    {selectedType?.expirationEnabled ? <ExpirationField key={`${asset.id}:${typeId}`} initialValue={expiration ?? undefined} initialPickerDate={initialPickerDate} disabled={disabled}
      onChange={(value, valid) => onChange({ ...base, expiration: value ?? null, expirationValid: valid })} />
      : asset.expiration || draft?.expiration !== undefined || draft?.expirationValid === false ? <View>
        <Text style={{ color: colors.textMuted }}>Expiration: {expiration?.date ?? 'Cleared'}. Tracking is disabled for this type.</Text>
        {expiration || draft?.expirationValid === false ? <Pressable accessibilityRole="button" accessibilityLabel="Clear expiration" disabled={disabled} onPress={() => onChange({ ...base, expiration: null, expirationValid: true })} style={{ minHeight: 44, justifyContent: 'center' }}><Text style={{ color: colors.action }}>Clear expiration</Text></Pressable> : null}
      </View> : null}
  </View>;
}
