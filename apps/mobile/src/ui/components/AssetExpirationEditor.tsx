import { Check } from 'lucide-react-native';
import { SelectionRow } from './SelectionRow';
import { AppTextInput } from './AppTextInput';
import { useState } from 'react';
import { Alert, Pressable, Text, View } from 'react-native';
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
  const [choosingType, setChoosingType] = useState(false);
  const [query, setQuery] = useState('');
  const [initialPickerDate] = useState(() => new Date());
  const typeId = asset.customAssetTypeId ?? draft?.customAssetTypeId;
  const selectedType = types?.find((type) => type.id === typeId);
  const expiration = draft?.expiration === undefined ? asset.expiration : draft.expiration;
  const base = { title: asset.title, description: asset.description, ...draft };
  function selectType(id?: string) {
    if (disabled || asset.customAssetTypeId || id === typeId) return;
    const apply = () => { onChange({ ...base, customAssetTypeId: id, expiration: null, expirationValid: true }); setChoosingType(false); };
    if (expiration) Alert.alert('Change item type?', 'Changing type removes the expiration date from this draft.', [{ text: 'Cancel', style: 'cancel' }, { text: 'Change type', onPress: apply }]);
    else apply();
  }
  if (!types) return <Text style={{ color: colors.textMuted }}>Loading expiration settings…</Text>;
  return <View style={{ gap: spacing.sm }}>
    {!asset.customAssetTypeId && types.length ? <SelectionRow label="Item type" value={selectedType?.displayName ?? 'None'} expanded={choosingType} disabled={disabled} onPress={() => setChoosingType(value => !value)}>
      <AppTextInput accessibilityLabel="Search item types" placeholder="Search types" value={query} onChangeText={setQuery} style={{ minHeight: 44, padding: spacing.sm, color: colors.text }} />
      {[{ id: undefined, displayName: 'None', expirationEnabled: false }, ...types].filter(type => !query || type.displayName.toLocaleLowerCase().includes(query.toLocaleLowerCase())).map(type => <Pressable key={type.id ?? 'base'} accessibilityRole="radio" accessibilityState={{ checked: typeId === type.id, disabled }} accessibilityLabel={type.displayName} disabled={disabled} onPress={() => selectType(type.id)} style={{ minHeight: 48, paddingVertical: spacing.sm, flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
        <View style={{ flex: 1 }}><Text style={{ color: colors.text, fontSize: 17 }}>{type.displayName}</Text>{type.expirationEnabled ? <Text style={{ color: colors.textMuted }}>Tracks expiration dates</Text> : null}</View>
        {typeId === type.id ? <Check size={22} color={colors.action} /> : null}
      </Pressable>)}
    </SelectionRow> : null}
    {selectedType?.expirationEnabled ? <ExpirationField key={`${asset.id}:${typeId}`} initialValue={expiration ?? undefined} initialPickerDate={initialPickerDate} disabled={disabled}
      onChange={(value, valid) => onChange({ ...base, expiration: value ?? null, expirationValid: valid })} />
      : asset.expiration || draft?.expiration !== undefined || draft?.expirationValid === false ? <View>
        <Text style={{ color: colors.textMuted }}>Expiration: {expiration?.date ?? 'Cleared'}. Tracking is disabled for this type.</Text>
        {expiration || draft?.expirationValid === false ? <Pressable accessibilityRole="button" accessibilityLabel="Clear expiration" disabled={disabled} onPress={() => onChange({ ...base, expiration: null, expirationValid: true })} style={{ minHeight: 44, justifyContent: 'center' }}><Text style={{ color: colors.action }}>Clear expiration</Text></Pressable> : null}
      </View> : null}
  </View>;
}
