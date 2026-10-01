import { t } from '../../presentation/localization';
import { useCallback, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, Text, View } from 'react-native';
import { Stack, useFocusEffect } from 'expo-router';
import { useHeaderHeight } from '@react-navigation/elements';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import type { ParentSelection } from './AddAssetResolution';
import { SettingsChoiceRow, SettingsLoadingRow, SettingsSection, useSettingsListStyles } from './SettingsList';
import { NativeNavigationSearch } from '../components/NativeNavigationSearch';
import { NativeFilterSearch } from '../components/NativeFilterSearch.android';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { DraftTextField } from '../components/DraftTextField';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { useFocusedSheetActions } from '../components/useFocusedSheetActions';

export type AddDestinationSelectionProps = {
  readonly query: string;
  readonly selected?: ParentSelection;
  readonly unresolvedSelection?: string;
  readonly matches: readonly ParentSelection[];
  readonly disabled: boolean;
  readonly loading: boolean;
  readonly failed: boolean;
  readonly creating: boolean;
  readonly canCreate: boolean;
  readonly error?: string;
  readonly onQuery: (query: string) => void;
  readonly onRetry: () => void;
  readonly onSelect: (parent: ParentSelection | undefined) => void;
  readonly onCreate: () => void;
  readonly onClose: () => void;
};
export function AddDestinationSelectionScreen(props: AddDestinationSelectionProps) {
  const { palette, styles } = useSettingsListStyles();
  const headerHeight = useHeaderHeight();
  const insets = useSafeAreaInsets();
  const [creationOpen, setCreationOpen] = useState(false);
  const current = useRef<(AddDestinationSelectionProps & { readonly creationOpen: boolean }) | undefined>(undefined);
  const focused = useRef(false);
  useLayoutEffect(() => { current.current = { ...props, creationOpen }; return () => { current.current = undefined; }; }, [props, creationOpen]);
  useFocusEffect(useCallback(() => { focused.current = true; return () => { focused.current = false; }; }, []));
  const select = useCallback((id?: string) => {
    const latest = current.current;
    if (!focused.current || !latest || latest.disabled || latest.creationOpen) return;
    const parent = id === undefined ? undefined : latest.matches.find(option => option.id === id);
    if (id !== undefined && (!parent || parent.canSelectAsParent === false)) return;
    latest.onSelect(parent);
  }, []);
  const changeQuery = useCallback((query: string) => {
    const latest = current.current;
    if (focused.current && latest && !latest.disabled) latest.onQuery(query);
  }, []);
  const creation = useFocusedSheetActions({ primaryLabel: t('mobile.AddDestinationSelectionScreen.createPlace'), secondaryLabel: t('mobile.AddDestinationSelectionScreen.cancelNewPlace'), disabled: props.disabled || props.loading || props.failed || !creationOpen || !props.canCreate,
    secondaryDisabled: props.disabled, onApply: props.onCreate, onBack: () => setCreationOpen(false) });
  const openCreation = useFocusedSheetActions({ primaryLabel: t('mobile.AddDestinationSelectionScreen.newPlace'), secondaryLabel: t('mobile.AddDestinationSelectionScreen.retrySuggestions'), disabled: props.disabled || props.loading || props.failed,
    secondaryDisabled: props.disabled, onApply: () => setCreationOpen(true), onBack: props.onRetry });
  const close = useNativeHeaderActionOptions([{ kind: 'close', label: creationOpen ? t('mobile.AddDestinationSelectionScreen.cancelNewPlace') : t('mobile.AddDestinationSelectionScreen.cancelLocationSelection'),
    disabled: props.disabled, onPress: creationOpen ? creation.onBack : props.onClose }], 'left');
  const command = useNativeHeaderActionOptions(creationOpen
    ? [{ kind: 'save', label: t('mobile.AddDestinationSelectionScreen.createPlace'), disabled: creation.disabled, onPress: creation.onApply }]
    : [{ kind: 'add', label: t('mobile.AddDestinationSelectionScreen.newPlace'), disabled: openCreation.disabled, onPress: openCreation.onApply }]);
  const headerOptions = useMemo(() => ({ title: creationOpen ? t('mobile.AddDestinationSelectionScreen.newPlace') : t('mobile.AddDestinationSelectionScreen.putIn'), headerBackVisible: false,
    gestureEnabled: !props.disabled, ...close, ...command }), [creationOpen, props.disabled, close, command]);
  const search = { query: props.query, placeholder: t('mobile.AddDestinationSelectionScreen.searchParent'), onChange: changeQuery, onSubmit: changeQuery, onClear: () => changeQuery('') };
  const lookupStatus = <>
    {props.loading ? <SettingsSection><SettingsLoadingRow label={t('mobile.AddDestinationSelectionScreen.loadingSuggestions')} /></SettingsSection> : null}
    {props.failed ? <SettingsSection><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.rowContext}>{t('mobile.AddDestinationSelectionScreen.suggestionsCouldNotBeLoaded')}</Text></View><NativeCommandButton label={t('mobile.AddDestinationSelectionScreen.retrySuggestions')} disabled={props.disabled} onPress={openCreation.onBack} /></SettingsSection> : null}
  </>;
  const content = <ScrollView key={creationOpen ? 'creation' : 'selection'} style={{ flex: 1, backgroundColor: palette.background }}
      automaticallyAdjustKeyboardInsets={Platform.OS === 'ios'}
      contentInsetAdjustmentBehavior="automatic"
      keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag"
      contentContainerStyle={{ paddingBottom: 20 + (Platform.OS === 'ios' ? insets.bottom : 0) }}>
        {creationOpen ? <>
          <SettingsSection title={t('mobile.AddDestinationSelectionScreen.name')} footer="A new place is saved immediately, even if you cancel adding the item later.">
            <View style={styles.navigationRow}>
              <DraftTextField accessibilityLabel={t('mobile.AddDestinationSelectionScreen.newPlaceName')} value={props.query} editable={!props.disabled}
                onChangeText={changeQuery} placeholder={t('mobile.AddDestinationSelectionScreen.placeName')} style={{ minHeight: 48, color: palette.text }} />
            </View>
          </SettingsSection>
          {props.error ? <SettingsSection><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.rowContext}>{props.error}</Text></View></SettingsSection> : null}
          {lookupStatus}
          {props.creating ? <SettingsSection><SettingsLoadingRow label={t('mobile.AddDestinationSelectionScreen.creatingPlace')} /></SettingsSection> : null}
        </> : <>
        <SettingsSection footer="Choosing a destination changes this draft only.">
          <View style={styles.navigationRow}><Text style={styles.rowContext}>{t('mobile.AddDestinationSelectionScreen.current', { value: String(props.selected?.pathLabel || props.selected?.title || props.unresolvedSelection || 'Top level in this inventory') })}</Text></View>
          <SettingsChoiceRow label={t('mobile.AddDestinationSelectionScreen.topLevel')} accessibilityLabel={t('mobile.AddDestinationSelectionScreen.chooseInventoryTopLevel')} selected={!props.selected && !props.unresolvedSelection} disabled={props.disabled} onPress={() => select()} />
        </SettingsSection>
        {props.error ? <SettingsSection><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.rowContext}>{props.error}</Text></View></SettingsSection> : null}
        {lookupStatus}
        <SettingsSection footer={!props.loading && !props.failed && !props.matches.length ? t('mobile.AddDestinationSelectionScreen.noMatchingDestinations') : undefined}>
          {props.matches.map(parent => <SettingsChoiceRow key={parent.id} label={parent.title} context={parent.disabledReason ?? `${parent.selectionHint} · ${parent.pathLabel || parent.subtitle}`}
            accessibilityLabel={t('mobile.AddDestinationSelectionScreen.chooseDestination', { title: String(parent.title) })} selected={props.selected?.id === parent.id} disabled={props.disabled || parent.canSelectAsParent === false} onPress={() => select(parent.id)} />)}
        </SettingsSection>
        </>}
      </ScrollView>;
  return <>
    <Stack.Screen options={headerOptions} />
    {Platform.OS === 'ios' ? <NativeNavigationSearch {...search} placement="stacked" enabled={!creationOpen && !props.disabled} /> : !creationOpen && !props.disabled ? <NativeFilterSearch {...search} /> : null}
    {Platform.OS === 'ios' ? content : <KeyboardAvoidingView style={{ flex: 1 }} behavior="height" keyboardVerticalOffset={headerHeight}>
      <SafeAreaView edges={['bottom']} style={{ flex: 1, backgroundColor: palette.background }}>{content}</SafeAreaView>
    </KeyboardAvoidingView>}
  </>;
}
