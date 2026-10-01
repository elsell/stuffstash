import { t } from '../../presentation/localization';
import { ScrollView, Text } from 'react-native';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import { NativeNavigationSearch } from '../components/NativeNavigationSearch';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsChoiceRow, SettingsLoadingRow, SettingsSection, SettingsValueRow, useSettingsListStyles } from './SettingsList';
import type { VoicePlanParentDraft } from './VoicePlanEdits';

export function VoicePlanLocationScreen({ current, proposed, matches, query, loading, error, onQuery, onRetry, onSelect }: {
  readonly current: VoicePlanParentDraft; readonly proposed: readonly VoicePlanParentDraft[];
  readonly matches: readonly ParentLookupResult[]; readonly query: string;
  readonly loading: boolean; readonly error: boolean;
  readonly onQuery: (query: string) => void; readonly onRetry: () => void;
  readonly onSelect: (parent: VoicePlanParentDraft) => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  const selected = (parent: VoicePlanParentDraft) => current.kind === parent.kind &&
    (current.kind === 'root' || (parent.kind !== 'root' && current.id === parent.id));
  return <>
    <NativeNavigationSearch query={query} placeholder={t('mobile.VoicePlanLocationScreen.searchLocations')} onChange={onQuery} onSubmit={onQuery} onClear={() => onQuery('')} />
    <ScrollView style={{ flex: 1, backgroundColor: palette.background }} contentInsetAdjustmentBehavior="automatic"
      automaticallyAdjustKeyboardInsets keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag">
      <SettingsSection><SettingsValueRow label={t('mobile.VoicePlanLocationScreen.currentLocation')} value={current.label} /></SettingsSection>
      <SettingsSection>
        <SettingsChoiceRow label={t('mobile.VoicePlanLocationScreen.inventoryRoot')} accessibilityLabel={t('mobile.VoicePlanLocationScreen.selectInventoryRoot')} selected={current.kind === 'root'}
          onPress={() => onSelect({ kind: 'root', label: t('mobile.VoicePlanLocationScreen.inventoryRoot') })} />
      </SettingsSection>
      {proposed.length ? <SettingsSection title={t('mobile.VoicePlanLocationScreen.createdByThisPlan')}>{proposed.map(parent =>
        <SettingsChoiceRow key={parent.kind === 'root' ? 'root' : parent.id} label={parent.label}
          accessibilityLabel={`Select proposed ${parent.label}`} selected={selected(parent)} onPress={() => onSelect(parent)} />
      )}</SettingsSection> : null}
      <SettingsSection title={t('mobile.VoicePlanLocationScreen.existingLocations')} footer={!loading && !error && !matches.length ? 'No matching locations' : undefined}>
        {loading ? <SettingsLoadingRow label={t('mobile.VoicePlanLocationScreen.loadingLocations')} /> : null}
        {error ? <><Text accessibilityRole="alert" style={styles.errorMessage}>{t('mobile.VoicePlanLocationScreen.couldNotLoadLocations')}</Text>
          <NativeCommandButton label={t('mobile.VoicePlanLocationScreen.retryLocations')} onPress={onRetry} /></> : null}
        {matches.map(match => {
          const parent: VoicePlanParentDraft = { kind: 'asset', id: match.id, label: match.pathLabel };
          const detail = match.disabledReason ?? (match.willPromoteToContainer ? `${match.pathLabel} · Will become a container` : match.pathLabel);
          return <SettingsChoiceRow key={match.id} label={match.title} context={detail}
            accessibilityLabel={`Select ${match.title}, ${detail}`} selected={selected(parent)}
            disabled={match.canSelectAsParent === false} onPress={() => { if (match.canSelectAsParent !== false) onSelect(parent); }} />;
        })}
      </SettingsSection>
    </ScrollView>
  </>;
}
