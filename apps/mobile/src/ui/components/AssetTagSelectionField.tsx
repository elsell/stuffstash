import type { CreateAssetTagDraft } from '../../application/assets/AssetTagDraftResolution';
import { useAssetTagSelectionVisit } from '../navigation/AssetTagSelectionTask';
import type { AssetTagSelectionOption } from '../screens/AssetTagSelectionScreen';
import { SettingsNavigationRow } from '../screens/SettingsList';

export function AssetTagSelectionField({ scope, tags, selectedIds, newTags, disabled, onChange }: {
  readonly scope: string;
  readonly tags: readonly AssetTagSelectionOption[];
  readonly selectedIds: readonly string[];
  readonly newTags?: readonly CreateAssetTagDraft[];
  readonly disabled: boolean;
  readonly onChange: (ids: readonly string[], newTags?: readonly CreateAssetTagDraft[]) => void;
}) {
  const open = useAssetTagSelectionVisit({ scope, tags, selectedIds, newTags, disabled, onChange });
  const labels = [...tags.filter(tag => selectedIds.includes(tag.id)).map(tag => tag.label), ...(newTags ?? []).map(tag => tag.displayName)];
  return <SettingsNavigationRow label="Tags" accessibilityLabel="Choose tags" disabled={disabled} onPress={open}
    value={`${selectedIds.length + (newTags?.length ?? 0)}`} context={labels.length ? labels.join(', ') : 'None selected'} />;
}
