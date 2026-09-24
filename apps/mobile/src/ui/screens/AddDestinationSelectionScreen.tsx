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
  const creation = useFocusedSheetActions({ primaryLabel: 'Create place', secondaryLabel: 'Cancel new place', disabled: props.disabled || props.loading || props.failed || !creationOpen || !props.canCreate,
    secondaryDisabled: props.disabled, onApply: props.onCreate, onBack: () => setCreationOpen(false) });
  const openCreation = useFocusedSheetActions({ primaryLabel: 'New place', secondaryLabel: 'Retry suggestions', disabled: props.disabled || props.loading || props.failed,
    secondaryDisabled: props.disabled, onApply: () => setCreationOpen(true), onBack: props.onRetry });
  const close = useNativeHeaderActionOptions([{ kind: 'close', label: creationOpen ? 'Cancel new place' : 'Cancel location selection',
    disabled: props.disabled, onPress: creationOpen ? creation.onBack : props.onClose }], 'left');
  const command = useNativeHeaderActionOptions(creationOpen
    ? [{ kind: 'save', label: 'Create place', disabled: creation.disabled, onPress: creation.onApply }]
    : [{ kind: 'add', label: 'New place', disabled: openCreation.disabled, onPress: openCreation.onApply }]);
  const headerOptions = useMemo(() => ({ title: creationOpen ? 'New place' : 'Put in', headerBackVisible: false,
    gestureEnabled: !props.disabled, ...close, ...command }), [creationOpen, props.disabled, close, command]);
  const search = { query: props.query, placeholder: 'Search parent', onChange: changeQuery, onSubmit: changeQuery, onClear: () => changeQuery('') };
  const lookupStatus = <>
    {props.loading ? <SettingsSection><SettingsLoadingRow label="Loading suggestions…" /></SettingsSection> : null}
    {props.failed ? <SettingsSection><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.rowContext}>Suggestions could not be loaded.</Text></View><NativeCommandButton label="Retry suggestions" disabled={props.disabled} onPress={openCreation.onBack} /></SettingsSection> : null}
  </>;
  const content = <ScrollView key={creationOpen ? 'creation' : 'selection'} style={{ flex: 1, backgroundColor: palette.background }}
      automaticallyAdjustKeyboardInsets={Platform.OS === 'ios'}
      contentInsetAdjustmentBehavior="automatic"
      keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag"
      contentContainerStyle={{ paddingBottom: 20 + (Platform.OS === 'ios' ? insets.bottom : 0) }}>
        {creationOpen ? <>
          <SettingsSection title="Name" footer="A new place is saved immediately, even if you cancel adding the item later.">
            <View style={styles.navigationRow}>
              <DraftTextField accessibilityLabel="New place name" value={props.query} editable={!props.disabled}
                onChangeText={changeQuery} placeholder="Place name" style={{ minHeight: 48, color: palette.text }} />
            </View>
          </SettingsSection>
          {props.error ? <SettingsSection><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.rowContext}>{props.error}</Text></View></SettingsSection> : null}
          {lookupStatus}
          {props.creating ? <SettingsSection><SettingsLoadingRow label="Creating place…" /></SettingsSection> : null}
        </> : <>
        <SettingsSection footer="Choosing a destination changes this draft only.">
          <View style={styles.navigationRow}><Text style={styles.rowContext}>{`Current: ${props.selected?.pathLabel || props.selected?.title || props.unresolvedSelection || 'Top level in this inventory'}`}</Text></View>
          <SettingsChoiceRow label="Top level" accessibilityLabel="Choose inventory top level" selected={!props.selected && !props.unresolvedSelection} disabled={props.disabled} onPress={() => select()} />
        </SettingsSection>
        {props.error ? <SettingsSection><View style={styles.navigationRow}><Text accessibilityRole="alert" style={styles.rowContext}>{props.error}</Text></View></SettingsSection> : null}
        {lookupStatus}
        <SettingsSection footer={!props.loading && !props.failed && !props.matches.length ? 'No matching destinations' : undefined}>
          {props.matches.map(parent => <SettingsChoiceRow key={parent.id} label={parent.title} context={parent.disabledReason ?? `${parent.selectionHint} · ${parent.pathLabel || parent.subtitle}`}
            accessibilityLabel={`Choose destination ${parent.title}`} selected={props.selected?.id === parent.id} disabled={props.disabled || parent.canSelectAsParent === false} onPress={() => select(parent.id)} />)}
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
