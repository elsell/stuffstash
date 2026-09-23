import { useCallback, useLayoutEffect, useRef, useState } from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, Text } from 'react-native';
import { Stack, useFocusEffect } from 'expo-router';
import { useHeaderHeight } from '@react-navigation/elements';
import { SafeAreaView } from 'react-native-safe-area-context';
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
  const [creationOpen, setCreationOpen] = useState(false);
  const current = useRef<AddDestinationSelectionProps | undefined>(undefined);
  const focused = useRef(false);
  useLayoutEffect(() => { current.current = props; return () => { current.current = undefined; }; }, [props]);
  useFocusEffect(useCallback(() => { focused.current = true; return () => { focused.current = false; }; }, []));
  const select = useCallback((id?: string) => {
    const latest = current.current;
    if (!focused.current || !latest || latest.disabled) return;
    const parent = id === undefined ? undefined : latest.matches.find(option => option.id === id);
    if (id !== undefined && (!parent || parent.canSelectAsParent === false)) return;
    latest.onSelect(parent);
  }, []);
  const changeQuery = useCallback((query: string) => {
    const latest = current.current;
    if (focused.current && latest && !latest.disabled) latest.onQuery(query);
  }, []);
  const creation = useFocusedSheetActions({ primaryLabel: 'Create place', secondaryLabel: 'Cancel new place', disabled: props.disabled || props.loading || props.failed || !creationOpen || !props.canCreate,
    secondaryDisabled: props.disabled, onApply: props.onCreate, onBack: () => setCreationOpen(false) });
  const openCreation = useFocusedSheetActions({ primaryLabel: 'New place', secondaryLabel: 'Retry suggestions', disabled: props.disabled || props.loading || props.failed,
    secondaryDisabled: props.disabled, onApply: () => setCreationOpen(true), onBack: props.onRetry });
  const close = useNativeHeaderActionOptions([{ kind: 'close', label: 'Cancel location selection', disabled: props.disabled, onPress: props.onClose }], 'left');
  const search = { query: props.query, placeholder: 'Search parent', onChange: changeQuery, onSubmit: changeQuery, onClear: () => changeQuery('') };
  return <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === 'android' ? 'height' : undefined} keyboardVerticalOffset={headerHeight}>
    <Stack.Screen options={{ title: 'Put in', headerBackVisible: false, gestureEnabled: !props.disabled, ...close }} />
    {Platform.OS === 'ios' ? <NativeNavigationSearch {...search} enabled={!creationOpen && !props.disabled} /> : !creationOpen && !props.disabled ? <NativeFilterSearch {...search} /> : null}
    <SafeAreaView edges={['bottom']} style={{ flex: 1, backgroundColor: palette.background }}>
      <ScrollView automaticallyAdjustKeyboardInsets={Platform.OS === 'ios'} contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag" contentContainerStyle={{ paddingBottom: 20 }}>
        <SettingsSection footer="Choosing a destination changes this draft only.">
          <Text style={styles.rowContext}>{`Current: ${props.selected?.pathLabel || props.selected?.title || props.unresolvedSelection || 'Top level in this inventory'}`}</Text>
          <SettingsChoiceRow label="Top level" accessibilityLabel="Choose inventory top level" selected={!props.selected && !props.unresolvedSelection} disabled={props.disabled} onPress={() => select()} />
        </SettingsSection>
        {props.error ? <SettingsSection><Text accessibilityRole="alert" style={styles.rowContext}>{props.error}</Text></SettingsSection> : null}
        {props.loading ? <SettingsSection><SettingsLoadingRow label="Loading suggestions…" /></SettingsSection> : null}
        {props.failed ? <SettingsSection><Text accessibilityRole="alert" style={styles.rowContext}>Suggestions could not be loaded.</Text><NativeCommandButton label="Retry suggestions" disabled={props.disabled} onPress={openCreation.onBack} /></SettingsSection> : null}
        <SettingsSection footer={!props.loading && !props.failed && !props.matches.length ? 'No matching destinations' : undefined}>
          {props.matches.map(parent => <SettingsChoiceRow key={parent.id} label={parent.title} context={parent.disabledReason ?? `${parent.selectionHint} · ${parent.pathLabel || parent.subtitle}`}
            accessibilityLabel={`Choose destination ${parent.title}`} selected={props.selected?.id === parent.id} disabled={props.disabled || parent.canSelectAsParent === false} onPress={() => select(parent.id)} />)}
        </SettingsSection>
        <SettingsSection footer="A new place is saved immediately, even if you cancel adding the item later.">
          {creationOpen ? <>
            <DraftTextField accessibilityLabel="New place name" value={props.query} editable={!props.disabled} onChangeText={changeQuery} placeholder="Place name" style={{ minHeight: 48, padding: 12, color: palette.text }} />
            <NativeCommandButton label={props.creating ? 'Creating place…' : 'Create place'} disabled={creation.disabled} onPress={creation.onApply} />
            <NativeCommandButton label="Cancel new place" disabled={props.disabled} onPress={creation.onBack} />
          </> : <NativeCommandButton label="New place" disabled={openCreation.disabled} onPress={openCreation.onApply} />}
        </SettingsSection>
      </ScrollView>
    </SafeAreaView>
  </KeyboardAvoidingView>;
}
