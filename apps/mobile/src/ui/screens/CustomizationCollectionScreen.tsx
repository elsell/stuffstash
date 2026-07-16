import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Pressable, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native';
import { ChevronRight, Plus, Search } from 'lucide-react-native';
import type { CustomizationContextQuery } from '../../application/customization/CustomizationContextQuery';
import type { CustomizationCollectionQuery } from '../../application/customization/CustomizationQueries';
import type { CustomizationAccessPolicy } from '../../application/customization/CustomizationAccess';
import { CustomizationFailure, safeCustomizationMessage } from '../../application/customization/CustomizationErrors';
import { beginLifecycleTransition, commitLifecycleTransition, rollbackLifecycleTransition, type CustomizationCollectionState } from '../../application/customization/CustomizationCollectionModel';
import type { CustomDefinition, CustomizationKind, CustomizationLifecycle, CustomizationScope } from '../../domain/customization/Customization';
import type { AssetTagDefinition } from '../../domain/customization/Customization';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { SettingsLoadingRow, SettingsSection, SettingsSeparator, useSettingsListStyles } from './SettingsList';
import { DeniedSettingsState } from './ScopedSettingsScreens';
import { tagColorName } from '../components/TagColorPicker';
import { SettingsSegmentedControl } from '../components/SettingsSegmentedControl';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';

type Row = AssetTagDefinition | CustomDefinition;

export function CustomizationCollectionScreen({ accessPolicy, contextQuery, kind, onAdd, onOpen, query, scope }: {
  readonly accessPolicy: CustomizationAccessPolicy;
  readonly contextQuery: CustomizationContextQuery;
  readonly kind: CustomizationKind;
  readonly onAdd: () => void;
  readonly onOpen: (row: Row, inherited: boolean, canManageInherited: boolean) => void;
  readonly query: CustomizationCollectionQuery;
  readonly scope: CustomizationScope;
}) {
  const feedback = useAppFeedback();
  const colors = useAppearancePalette();
  const settings = useSettingsListStyles();
  const styles = createStyles(colors);
  const requestRef = useRef(0);
  const [context, setContext] = useState<Awaited<ReturnType<CustomizationContextQuery['execute']>>>();
  const [collection, setCollection] = useState<CustomizationCollectionState<Row>>({ lifecycle: 'active', rows: [] });
  const [status, setStatus] = useState<'loading' | 'ready' | 'error' | 'denied'>('loading');
  const [refreshing, setRefreshing] = useState(false);
  const [search, setSearch] = useState('');
  const [incomplete, setIncomplete] = useState(false);
  const { lifecycle, pendingLifecycle, rows } = collection;

  const load = useCallback(async (refresh = false, targetLifecycle = lifecycle) => {
    const request = ++requestRef.current;
    if (refresh) setRefreshing(true); else if (rows.length === 0) setStatus('loading');
    try {
      const nextContext = await contextQuery.execute();
      if (!accessPolicy.readOrRecord(nextContext, kind, scope)) {
        if (request === requestRef.current) { setContext(nextContext); setStatus('denied'); }
        return;
      }
      const result = kind === 'tag'
        ? await query.tags(nextContext)
        : kind === 'field'
          ? await query.fields(nextContext, scope, targetLifecycle)
          : await query.assetTypes(nextContext, scope, targetLifecycle);
      if (request !== requestRef.current) return;
      setContext(nextContext); setCollection((current) => commitLifecycleTransition(current, targetLifecycle, result.items)); setIncomplete(!result.complete); setStatus('ready');
      if (!result.complete) feedback.showNotice({ tone: 'warning', title: 'Some settings could not be loaded', message: 'Refresh to try loading the complete list.' });
    } catch (error) {
      if (request !== requestRef.current) return;
      if (error instanceof CustomizationFailure && error.kind === 'permission-denied') {
        let refreshedContext: Awaited<ReturnType<CustomizationContextQuery['execute']>> | undefined;
        try { refreshedContext = await contextQuery.execute(); } catch { /* The API denial remains authoritative. */ }
        if (request !== requestRef.current) return;
        if (refreshedContext) accessPolicy.readOrRecord(refreshedContext, kind, scope);
        setContext(refreshedContext);
        setCollection((current) => ({ ...current, rows: [], pendingLifecycle: undefined }));
        setIncomplete(false);
        setStatus('denied');
        return;
      }
      if (rows.length) { feedback.showNotice({ tone: 'error', title: 'Could not refresh settings', message: safeCustomizationMessage(error, 'Try again.') }); setCollection(rollbackLifecycleTransition); setStatus('ready'); }
      else setStatus('error');
    } finally { if (request === requestRef.current) setRefreshing(false); }
  }, [accessPolicy, contextQuery, feedback, kind, lifecycle, query, rows.length, scope]);

  useEffect(() => { void load(); return () => { requestRef.current += 1; }; }, []);

  const canEdit = context && accessPolicy.canMutate(context, kind, scope);
  const filtered = useMemo(() => rows.filter((row) => row.displayName.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase())), [rows, search]);
  const inherited = scope === 'inventory' && kind !== 'tag' ? filtered.filter((row) => 'scope' in row && row.scope === 'tenant') : [];
  const local = filtered.filter((row) => !inherited.includes(row));

  if (status === 'loading') return <View style={settings.styles.shell}><View style={[styles.loadingGroup, settings.styles.contentBlock]}><SettingsLoadingRow label={`Loading ${plural(kind).toLocaleLowerCase()}…`} /></View></View>;
  if (status === 'error') return <View style={[settings.styles.shell, settings.styles.errorContainer]}><Text accessibilityRole="header" style={settings.styles.errorTitle}>Could not load {plural(kind).toLocaleLowerCase()}</Text><Text style={settings.styles.errorMessage}>Your settings were not changed.</Text><Pressable accessibilityRole="button" onPress={() => void load()} style={settings.styles.retryButton}><Text style={settings.styles.retryText}>Retry</Text></Pressable></View>;
  if (status === 'denied') return <DeniedSettingsState message="You don’t have permission to view these settings." />;
  if (!context) return null;

  return <ScrollView automaticallyAdjustKeyboardInsets contentInsetAdjustmentBehavior="automatic" contentContainerStyle={settings.styles.content} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled" refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => void load(true)} tintColor={colors.action} />} style={settings.styles.shell}>
    <View style={[styles.toolbar, settings.styles.contentBlock]}>
      <View style={styles.searchShell}><Search color={colors.textMuted} size={18} /><AppTextInput accessibilityLabel={`Search ${plural(kind)}`} onChangeText={setSearch} placeholder={`Search ${plural(kind).toLocaleLowerCase()}`} placeholderTextColor={colors.textMuted} style={styles.searchInput} value={search} /></View>
      {canEdit && lifecycle === 'active' ? <Pressable accessibilityLabel={`Add ${singular(kind)}`} accessibilityRole="button" onPress={onAdd} style={styles.addButton}><Plus color={colors.onAction} size={19} /><Text style={styles.addText}>Add</Text></Pressable> : null}
    </View>
    {kind !== 'tag' ? <View style={[styles.lifecycleControl, settings.styles.contentBlock]}><SettingsSegmentedControl disabled={Boolean(pendingLifecycle)} onChange={(value) => { const target = value as CustomizationLifecycle; if (target !== lifecycle && !pendingLifecycle) { setCollection((current) => beginLifecycleTransition(current, target)); void load(false, target); } }} segments={[{ label: 'Active', value: 'active' }, { label: 'Archived', value: 'archived' }]} value={lifecycle} /></View> : null}
    {pendingLifecycle ? <View style={[styles.loadingGroup, settings.styles.contentBlock]}><SettingsLoadingRow label={`Loading ${pendingLifecycle} settings…`} /></View> : null}
    {incomplete ? <View accessibilityLiveRegion="polite" style={[styles.incomplete, settings.styles.contentBlock]}><Text style={styles.incompleteTitle}>Some settings may be missing</Text><Text style={styles.incompleteText}>Pull to refresh and try loading the complete list.</Text></View> : null}
    {search && filtered.length === 0 ? <Empty title="No matches" message={`No ${plural(kind).toLocaleLowerCase()} match “${search}”.`} />
      : filtered.length === 0 ? <Empty title={`No ${lifecycle} ${plural(kind).toLocaleLowerCase()}`} message={canEdit && lifecycle === 'active' ? `Add the first ${singular(kind).toLocaleLowerCase()} here.` : 'There is nothing to show.'} />
      : <>
        {inherited.length ? <ResourceSection name={`From ${context.tenantName}`} rows={inherited} onOpen={(row) => onOpen(row, true, context.tenantPermissions.includes('configure'))} inherited /> : null}
        {local.length ? <ResourceSection name={scope === 'inventory' && kind !== 'tag' ? `Only in ${context.inventoryName}` : undefined} rows={local} onOpen={(row) => onOpen(row, false, false)} /> : null}
      </>}
  </ScrollView>;
}

function ResourceSection({ inherited = false, name, onOpen, rows }: { readonly inherited?: boolean; readonly name?: string; readonly onOpen: (row: Row) => void; readonly rows: readonly Row[] }) {
  const colors = useAppearancePalette(); const styles = createStyles(colors);
  return <SettingsSection title={name}>{rows.map((row, index) => <View key={row.id}>{index ? <SettingsSeparator /> : null}<Pressable accessibilityLabel={`${row.displayName}${row.kind === 'tag' ? `, ${tagColorName(row.color)}` : ''}${inherited ? ', inherited' : ''}`} accessibilityRole="button" onPress={() => onOpen(row)} style={({ pressed }) => [styles.row, row.kind === 'tag' && styles.compactRow, pressed && styles.pressed]}><View style={styles.rowBody}>{row.kind === 'tag' ? <View accessibilityElementsHidden style={[styles.color, row.color ? { backgroundColor: row.color } : styles.noColor]} /> : null}<View style={styles.rowText}><Text style={styles.rowTitle}>{row.displayName}</Text>{row.kind !== 'tag' ? <Text style={styles.rowMeta}>{row.kind === 'field' ? `${fieldType(row.type)} · ${row.applicability === 'all_assets' ? 'All assets' : 'Selected asset types'}` : row.description || row.key}{inherited ? ' · Inherited' : ''}</Text> : null}</View></View><ChevronRight color={colors.textMuted} size={18} /></Pressable></View>)}</SettingsSection>;
}

function Empty({ message, title }: { readonly message: string; readonly title: string }) { const styles = useSettingsListStyles().styles; return <View style={styles.errorContainer}><Text accessibilityRole="header" style={styles.errorTitle}>{title}</Text><Text style={styles.errorMessage}>{message}</Text></View>; }
function singular(kind: CustomizationKind) { return kind === 'tag' ? 'Tag' : kind === 'field' ? 'Custom field' : 'Asset type'; }
function plural(kind: CustomizationKind) { return `${singular(kind)}s`; }
function fieldType(value: string) { return value.charAt(0).toUpperCase() + value.slice(1); }
function createStyles(colors: MobileColorPalette) { return StyleSheet.create({
  toolbar: { alignItems: 'center', flexDirection: 'row', gap: spacing.sm, marginTop: spacing.md }, searchShell: { alignItems: 'center', backgroundColor: colors.surface, borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, flex: 1, flexDirection: 'row', minHeight: 44, paddingHorizontal: spacing.sm }, searchInput: { color: colors.text, flex: 1, fontSize: 16, minHeight: 44, paddingHorizontal: spacing.xs }, addButton: { alignItems: 'center', backgroundColor: colors.action, borderRadius: radius.md, flexDirection: 'row', gap: spacing.xs, minHeight: 44, paddingHorizontal: spacing.md }, addText: { color: colors.onAction, fontSize: 16, fontWeight: '700' }, lifecycleControl: { marginTop: spacing.sm }, loadingGroup: { backgroundColor: colors.surface, borderRadius: radius.md, marginTop: spacing.sm, overflow: 'hidden' }, incomplete: { backgroundColor: colors.warningSurface, borderRadius: radius.md, gap: spacing.xs, marginTop: spacing.sm, padding: spacing.md }, incompleteTitle: { color: colors.warning, fontSize: 14, fontWeight: '700' }, incompleteText: { color: colors.text, fontSize: 13 }, row: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between', minHeight: 58, paddingHorizontal: spacing.md, paddingVertical: spacing.sm }, compactRow: { minHeight: 52, paddingVertical: spacing.xs }, pressed: { backgroundColor: colors.surfaceMuted }, rowBody: { alignItems: 'center', flex: 1, flexDirection: 'row', gap: spacing.sm, minWidth: 0 }, color: { borderColor: colors.border, borderRadius: 12, borderWidth: 1, height: 24, width: 24 }, noColor: { backgroundColor: 'transparent', borderWidth: 2 }, rowText: { flex: 1, gap: 2, minWidth: 0 }, rowTitle: { color: colors.text, fontSize: 16, fontWeight: '600' }, rowMeta: { color: colors.textMuted, fontSize: 13, lineHeight: 18 }
}); }
