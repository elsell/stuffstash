import { NativeCommandButton } from '../components/NativeCommandButton';
import { AppSwitchField } from '../components/AppSwitchField';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { useCustomizationReads } from '../serverState/useCustomizationReads';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useFocusEffect, useNavigation } from 'expo-router';
import { usePreventRemove } from '@react-navigation/native';
import { AccessibilityInfo, Alert, findNodeHandle, KeyboardAvoidingView, Platform, Pressable, ScrollView, StyleSheet, Text, View, type TextInput } from 'react-native';
import { ChevronDown } from 'lucide-react-native';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import type { CustomizationContextQuery } from '../../application/customization/CustomizationContextQuery';
import type { CustomizationCollectionQuery } from '../../application/customization/CustomizationQueries';
import type { ManageTags } from '../../application/customization/ManageTags';
import type { ManageCustomFields } from '../../application/customization/ManageCustomFields';
import type { ManageCustomAssetTypes } from '../../application/customization/ManageCustomAssetTypes';
import type { CustomizationAccessPolicy } from '../../application/customization/CustomizationAccess';
import { CustomizationFailure, safeCustomizationMessage } from '../../application/customization/CustomizationErrors';
import { runCustomizationLifecycleIntent, saveCustomizationEditor } from '../../application/customization/CustomizationEditorCommands';
import { customizationEditorIsDirty, customizationEditorIsValid, customizationEditorSnapshot, customizationEditorValidation, effectiveInheritedOwnership, withEditorName, withManualEditorKey, type CustomizationEditorDraft } from '../../application/customization/CustomizationEditorModel';
import { CustomizationEditorWorkflow } from '../../application/customization/CustomizationEditorWorkflow';
import type { AssetTagDefinition, CustomAssetTypeDefinition, CustomDefinition, CustomFieldApplicability, CustomFieldDefinition, CustomFieldType, CustomizationKind, CustomizationLifecycle, CustomizationScope } from '../../domain/customization/Customization';
import { TagColorPicker, tagColorName } from '../components/TagColorPicker';
import { CustomizationFieldControls, CustomizationLabeledInput, CustomizationReadOnlyValue } from '../components/CustomizationEditorFields';
import { CustomizationLifecycleSection } from '../components/CustomizationLifecycleSection';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { SettingsLoadingRow, SettingsSection, SettingsSeparator, SettingsValueRow, useSettingsListStyles } from './SettingsList';
import { DeniedSettingsState } from './ScopedSettingsScreens';
import { appKeyboardDismissMode } from '../components/AppTextInput';

type EditorRecord = AssetTagDefinition | CustomDefinition;

export function CustomizationEditorScreen({ accessPolicy, contextQuery: sourceContext, inherited = false, kind, lifecycle = 'active', manageAssetTypes, manageFields, manageTags, mode, onDone, onManageInherited, query: sourceQuery, resourceId, scope }: {
  readonly accessPolicy: CustomizationAccessPolicy; readonly contextQuery: CustomizationContextQuery; readonly inherited?: boolean; readonly kind: CustomizationKind; readonly lifecycle?: CustomizationLifecycle;
  readonly manageAssetTypes: ManageCustomAssetTypes; readonly manageFields: ManageCustomFields; readonly manageTags: ManageTags;
  readonly mode: 'create' | 'edit'; readonly onDone: () => void; readonly onManageInherited?: () => void; readonly query: CustomizationCollectionQuery; readonly resourceId?: string; readonly scope: CustomizationScope;
}) {
  const feedback = useAppFeedback(); const colors = useAppearancePalette(); const settings = useSettingsListStyles(); const styles = createStyles(colors);
  const workflowRef = useRef(new CustomizationEditorWorkflow());
  const errorSummaryRef = useRef<View>(null);
  const keyInputRef = useRef<TextInput>(null);
  const keyFocusRequestedRef = useRef(false);
  const navigation = useNavigation();
  const backOptions = useNativeHeaderActionOptions([{ kind: 'back', label: 'Back to settings collection', onPress: onDone }], 'left');
  const [status, setStatus] = useState<'loading' | 'ready' | 'error' | 'denied'>('loading');
  const [context, setContext] = useState<Awaited<ReturnType<CustomizationContextQuery['execute']>>>();
  const [record, setRecord] = useState<EditorRecord>(); const [eligibleTypes, setEligibleTypes] = useState<readonly CustomAssetTypeDefinition[]>([]);
  const [expirationEnabled, setExpirationEnabled] = useState(false); const [name, setName] = useState(''); const [key, setKey] = useState(''); const [description, setDescription] = useState(''); const [color, setColor] = useState('');
  const [fieldType, setFieldType] = useState<CustomFieldType>('text'); const [applicability, setApplicability] = useState<CustomFieldApplicability>('all_assets');
  const [enumOptions, setEnumOptions] = useState<readonly string[]>([]); const [newOption, setNewOption] = useState(''); const [targetIds, setTargetIds] = useState<readonly string[]>([]);
  const [advanced, setAdvanced] = useState(false); const [saving, setSaving] = useState(false); const [error, setError] = useState<string>();
  const [errorTitle, setErrorTitle] = useState('Could not save');
  const [deniedMessage, setDeniedMessage] = useState("You don’t have permission to change this setting.");
  const [initialSnapshot, setInitialSnapshot] = useState('');
  const [draftDenied, setDraftDenied] = useState(false);
  const [completion, setCompletion] = useState<'Saved' | 'Archived' | 'Restored' | 'Deleted'>();
  const completed = completion !== undefined;
  const [lifecycleBusy, setLifecycleBusy] = useState(false);
  const [nameTouched, setNameTouched] = useState(false);
  const [keyManuallyEdited, setKeyManuallyEdited] = useState(false);
  const [exitAuthorized, setExitAuthorized] = useState(false);

  const reads = useCustomizationReads(sourceContext, sourceQuery, kind, scope, lifecycle, accessPolicy, mode === 'edit', true);
  const { contextQuery, query } = reads;
  const focusOwner = useRef<object | undefined>(undefined);
  const resourceOwner = useRef<object | undefined>(undefined);
  useFocusEffect(useCallback(() => {
    focusOwner.current = {};
    return () => { focusOwner.current = undefined; };
  }, [resourceId, lifecycle, scope, kind, reads.ownerKey]));

  useEffect(() => {
    if (isAccessFailure(reads.contextError) || isAccessFailure(reads.resource.error)) {
      if (status === 'ready') setDraftDenied(true);
      else setStatus('denied');
    }
    if (reads.context && status === 'ready') {
      setContext(reads.context);
      if (!accessPolicy.canRead(reads.context, scope) || !accessPolicy.canMutate(reads.context, kind, scope, inherited)) setDraftDenied(true);
    }
  }, [reads.contextError, reads.resource.error, reads.context?.tenantPermissions.join(','), reads.context?.inventoryPermissions.join(',')]);

  useEffect(() => {
    if (reads.types.data) setEligibleTypes(reads.types.data.items.filter((item): item is CustomAssetTypeDefinition => item.kind === 'asset-type' && (scope === 'inventory' || item.scope === 'tenant')));
  }, [reads.types.data, scope]);

  async function load() {
    const workflow = workflowRef.current;
    const request = workflow.beginLoad();
    setStatus('loading');
    clearLoadedRecord();
    setKeyManuallyEdited(false);
    try {
      const nextContext = await contextQuery.execute();
      if (!workflow.isCurrentLoad(request)) return;
      setContext(nextContext);
      const canRead = accessPolicy.readOrRecord(nextContext, kind, scope);
      const mutationRoute = mode === 'create';
      if (!canRead || (mutationRoute && !accessPolicy.mutationOrRecord(nextContext, kind, scope, inherited))) {
        setStatus('denied');
        return;
      }
      const typeResult = kind === 'field' ? await query.assetTypes(nextContext, scope, 'active') : undefined;
      if (!workflow.isCurrentLoad(request)) return;
      if (typeResult) setEligibleTypes(typeResult.items.filter((item) => scope === 'inventory' || item.scope === 'tenant'));
      if (mode === 'edit') {
        const result = kind === 'tag' ? await query.tags(nextContext) : kind === 'field' ? await query.fields(nextContext, scope, lifecycle) : await query.assetTypes(nextContext, scope, lifecycle);
        if (!workflow.isCurrentLoad(request)) return;
        const found = result.items.find((item) => item.id === resourceId);
        if (!found) { setStatus('error'); return; }
        setRecord(found); setName(found.displayName); setKey(found.key); setKeyManuallyEdited(false);
        if (found.kind === 'tag') setColor(found.color ?? '');
        if (found.kind === 'asset-type') { setDescription(found.description); setExpirationEnabled(found.expirationEnabled ?? false); }
        if (found.kind === 'field') { setFieldType(found.type); setApplicability(found.applicability); setEnumOptions(found.enumOptions); setTargetIds(found.customAssetTypeIds); }
        setInitialSnapshot(snapshotFor(found));
      }
      setStatus('ready');
    } catch (cause) {
      if (!workflow.isCurrentLoad(request)) return;
      if (cause instanceof CustomizationFailure && cause.kind === 'permission-denied') {
        let refreshedContext: Awaited<ReturnType<CustomizationContextQuery['execute']>> | undefined;
        try { refreshedContext = await reads.refreshContext(); } catch { /* The API denial remains authoritative. */ }
        if (!workflow.isCurrentLoad(request)) return;
        if (refreshedContext) {
          setContext(refreshedContext);
          accessPolicy.readOrRecord(refreshedContext, kind, scope);
        }
        clearLoadedRecord();
        setDeniedMessage('Your access changed. This setting can’t be shown.');
        setStatus('denied');
        return;
      }
      setStatus('error');
    }
  }
  useEffect(() => {
    const workflow = new CustomizationEditorWorkflow();
    workflowRef.current = workflow; resourceOwner.current = {};
    setSaving(false); setLifecycleBusy(false); void load();
    return () => { resourceOwner.current = undefined; workflow.invalidateLoads(); };
  }, [resourceId, lifecycle, scope, kind, reads.ownerKey]);

  const effectiveInherited = effectiveInheritedOwnership({ routeHint: inherited, recordScope: record && 'scope' in record ? record.scope : undefined, screenScope: scope });
  const canMutate = Boolean(context && !draftDenied && !isAccessFailure(reads.contextError) && !isAccessFailure(reads.resource.error) && accessPolicy.canMutate(context, kind, scope, effectiveInherited));
  const editorDraft: CustomizationEditorDraft = { expirationEnabled, name, key, keyManuallyEdited, description, color, fieldType, applicability, enumOptions, pendingEnumOption: newOption, targetIds };
  const current = customizationEditorSnapshot(editorDraft);
  const dirty = customizationEditorIsDirty(editorDraft, initialSnapshot, mode, completed);
  const validation = customizationEditorValidation(editorDraft, kind, mode);
  const colorValid = validation.colorValid;
  const valid = customizationEditorIsValid(editorDraft, kind, mode);
  const editorMutable = canMutate && lifecycle === 'active';
  const draftEditable = editorMutable && !saving && !lifecycleBusy;

  useEffect(() => {
    const keyBlocksCreate = kind !== 'tag' && mode === 'create' && nameTouched && validation.nameValid && !validation.keyValid;
    if (!keyBlocksCreate) { keyFocusRequestedRef.current = false; return; }
    if (!advanced) { setAdvanced(true); return; }
    if (!keyFocusRequestedRef.current) {
      keyFocusRequestedRef.current = true;
      keyInputRef.current?.focus();
    }
  }, [advanced, kind, mode, nameTouched, validation.keyValid, validation.nameValid]);

  useEffect(() => {
    if (!error) return;
    const target = findNodeHandle(errorSummaryRef.current);
    if (target) AccessibilityInfo.setAccessibilityFocus(target);
  }, [error]);

  usePreventRemove(dirty && !saving && !exitAuthorized, ({ data }: { data: { action: unknown } }) => {
    const focus = focusOwner.current; const resource = resourceOwner.current; const workflow = workflowRef.current;
    if (!focus || !resource) return;
    Alert.alert('Discard changes?', 'Your unsaved changes will be lost.', [
      { text: 'Keep Editing', style: 'cancel' },
      { text: 'Discard', style: 'destructive', onPress: () => {
        if (focusOwner.current !== focus || resourceOwner.current !== resource) return;
        if (!workflow.authorizeExit(data.action)) return;
        setExitAuthorized(true);
      } }
    ]);
  });

  useEffect(() => {
    if (!exitAuthorized) return;
    const action = workflowRef.current.takeAuthorizedExit();
    if (!action) return;
    try {
      navigation.dispatch(action as never);
    } catch (cause) {
      workflowRef.current.resetExit();
      setExitAuthorized(false);
      throw cause;
    }
  }, [exitAuthorized, navigation]);

  useEffect(() => {
    navigation.setOptions({
      gestureEnabled: !dirty,
      headerBackVisible: false,
      ...backOptions
    });
    return () => navigation.setOptions({ gestureEnabled: true, headerBackVisible: true, headerLeft: undefined, unstable_headerLeftItems: undefined });
  }, [dirty, navigation, backOptions]);

  if (status === 'ready' && (!reads.context && !isAccessFailure(reads.contextError) || reads.context && (context?.tenantId !== reads.context.tenantId || context?.inventoryId !== reads.context.inventoryId))) return <SettingsLoadingRow label="Loading setting…" />;
  if (status === 'loading') return <View style={settings.styles.shell}><View style={styles.loadingGroup}><SettingsLoadingRow label={`Loading ${label(kind).toLocaleLowerCase()}…`} /></View></View>;
  if (status === 'error') return <ScrollView style={settings.styles.shell} contentContainerStyle={[settings.styles.errorContainer, { flexGrow: 1 }]} contentInsetAdjustmentBehavior="automatic"><Text accessibilityRole="header" style={settings.styles.errorTitle}>Setting unavailable</Text><Text style={settings.styles.errorMessage}>It may have been archived, deleted, or you may no longer have access.</Text><NativeCommandButton label="Retry" onPress={() => void load()} /></ScrollView>;
  if (status === 'denied') return <DeniedSettingsState message={deniedMessage} />;
  if (!context) return null;
  if (mode === 'create' && !canMutate && !draftDenied) return <DeniedSettingsState message="You don’t have permission to add this setting." />;

  if (completion) return <ScrollView style={settings.styles.shell} contentContainerStyle={settings.styles.content} contentInsetAdjustmentBehavior="automatic">
    <SettingsSection title={completion} footer="This change is complete. Return to the collection to continue.">
      <NativeCommandButton label="Return to collection" onPress={onDone} />
    </SettingsSection>
  </ScrollView>;

  async function save() {
    const owner = focusOwner.current; const resource = resourceOwner.current; const workflow = workflowRef.current;
    if (completed || !resource || !owner || !context || !valid || (mode === 'edit' && !dirty) || !workflow.beginSave()) return;
    setSaving(true); setError(undefined); setErrorTitle('Could not save');
    try {
      await saveCustomizationEditor({ context, draft: editorDraft, kind, managers: { assetTypes: manageAssetTypes, fields: manageFields, tags: manageTags }, mode, record: record?.kind === 'field' ? record : undefined, resourceId, scope });
      if (resourceOwner.current !== resource) return;
      setInitialSnapshot(current); setCompletion('Saved');
      if (focusOwner.current === owner) { feedback.showNotice({ tone: 'success', title: `${label(kind)} saved` }); onDone(); }
    } catch (cause) { if (resourceOwner.current === resource) await handleFailure(cause, `${label(kind)} was not saved.`); }
    finally { workflow.finishSave(); if (resourceOwner.current === resource) setSaving(false); }
  }

  function lifecycleAction(action: 'archive' | 'restore' | 'delete') {
    const owner = focusOwner.current; const resource = resourceOwner.current; const workflow = workflowRef.current;
    if (completed || !resource || !owner || !context || !resourceId || !workflow.beginLifecycleConfirmation()) return;
    setLifecycleBusy(true);
    const destructive = action !== 'restore';
    const title = action === 'delete' ? `Delete ${name} permanently?` : `${capitalize(action)} ${name}?`;
    const message = lifecycleMessage(kind, action);
    Alert.alert(title, message, [{ text: 'Cancel', style: 'cancel', onPress: () => { if (workflow.cancelLifecycleConfirmation() && resourceOwner.current === resource) setLifecycleBusy(false); } }, { text: action === 'delete' ? 'Delete Permanently' : capitalize(action), style: destructive ? 'destructive' : 'default', onPress: async () => {
      if (focusOwner.current !== owner || resourceOwner.current !== resource) {
        if (workflow.cancelLifecycleConfirmation() && resourceOwner.current === resource) setLifecycleBusy(false);
        return;
      }
      if (!workflow.beginLifecycleMutation()) return;
      try {
        await runCustomizationLifecycleIntent({ action, context, kind, managers: { assetTypes: manageAssetTypes, fields: manageFields, tags: manageTags }, resourceId, scope });
        workflow.finishLifecycle();
        if (resourceOwner.current !== resource) return;
        setLifecycleBusy(false); setCompletion(action === 'archive' ? 'Archived' : action === 'restore' ? 'Restored' : 'Deleted');
        if (focusOwner.current === owner) { feedback.showNotice({ tone: 'success', title: `${label(kind)} ${action === 'archive' ? 'archived' : action === 'restore' ? 'restored' : 'deleted'}` }); onDone(); }
      } catch (cause) {
        if (resourceOwner.current === resource) await handleFailure(cause, `${label(kind)} was not changed.`, `Could not ${action}`);
        workflow.finishLifecycle(); if (resourceOwner.current === resource) setLifecycleBusy(false);
      }
    }}], { onDismiss: () => { if (workflow.cancelLifecycleConfirmation() && resourceOwner.current === resource) setLifecycleBusy(false); } });
  }

  return <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={settings.styles.shell}><ScrollView automaticallyAdjustKeyboardInsets contentInsetAdjustmentBehavior="automatic" contentContainerStyle={settings.styles.content} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
    {draftDenied ? <View accessibilityLiveRegion="assertive" style={styles.errorSummary}><Text accessibilityRole="header" style={styles.errorTitle}>Access changed</Text><Text style={styles.errorText}>Your change was not saved. Your draft is shown below and is read-only.</Text><Text style={styles.errorText}>Ask a household manager to restore your access, then refresh access.</Text><NativeCommandButton label="Refresh access" onPress={() => void refreshDraftAccess()} /></View> : error ? <View accessibilityLiveRegion="assertive" ref={errorSummaryRef} style={styles.errorSummary}><Text accessibilityRole="header" style={styles.errorTitle}>{errorTitle}</Text><Text style={styles.errorText}>{error}</Text></View> : null}
    <SettingsSection title={mode === 'create' ? `New ${label(kind)}` : label(kind)}>
      {editorMutable ? <CustomizationLabeledInput error={nameTouched && !name.trim() ? 'Name is required.' : undefined} label="Name" onChangeText={(value) => { setNameTouched(true); const next = withEditorName(editorDraft, value); setName(next.name); setKey(next.key); }} required value={name} editable={draftEditable} /> : <CustomizationReadOnlyValue label="Name" value={name} />}
      {kind === 'tag' ? editorMutable ? <View style={styles.formRow}><Text style={styles.label}>Color</Text><TagColorPicker disabled={!draftEditable} value={color} onChange={setColor} />{!colorValid ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>Enter a six-digit hex color such as #2F80ED.</Text> : null}</View> : <CustomizationReadOnlyValue label="Color" value={tagColorName(color)} /> : null}
      {kind === 'asset-type' ? editorMutable
        ? <AppSwitchField label="Track expiration dates" description="Each asset can have its own optional date. Turning this off keeps existing dates." value={expirationEnabled} onValueChange={setExpirationEnabled} disabled={saving || lifecycleBusy} />
        : <CustomizationReadOnlyValue label="Track expiration dates" value={expirationEnabled ? 'Enabled' : 'Disabled'} />
        : null}
      {kind === 'asset-type' ? editorMutable ? <CustomizationLabeledInput label="Description" multiline onChangeText={setDescription} value={description} editable={draftEditable} /> : <CustomizationReadOnlyValue label="Description" value={description || 'No description'} /> : null}
      {kind === 'field' ? <CustomizationFieldControls persistedApplicability={record?.kind === 'field' ? record.applicability : undefined} applicability={applicability} canMutate={editorMutable} busy={saving || lifecycleBusy} eligibleTypes={eligibleTypes} enumOptions={enumOptions} fieldType={fieldType} mode={mode} newOption={newOption} onApplicability={setApplicability} onEnumOptions={setEnumOptions} onFieldType={setFieldType} onNewOption={setNewOption} onTargets={setTargetIds} persistedEnumOptions={record?.kind === 'field' ? record.enumOptions : []} persistedTargetIds={record?.kind === 'field' ? record.customAssetTypeIds : []} targetIds={targetIds} /> : null}
    </SettingsSection>
    {kind !== 'tag' ? <SettingsSection title="Details"><Pressable accessibilityRole="button" onPress={() => setAdvanced((value) => !value)} style={styles.disclosure}><Text style={styles.disclosureText}>{advanced ? 'Hide technical details' : 'Show technical details'}</Text><ChevronDown color={colors.textMuted} size={18} style={{ transform: [{ rotate: advanced ? '180deg' : '0deg' }] }} /></Pressable>{advanced ? <>{mode === 'create' ? <CustomizationLabeledInput editable={draftEditable} error={!validation.keyValid ? validation.keyMessage : undefined} inputRef={keyInputRef} label="Stable key" onChangeText={(value) => { const next = withManualEditorKey(editorDraft, value); setKey(next.key); setKeyManuallyEdited(next.keyManuallyEdited); }} value={key} /> : <><SettingsSeparator /><SettingsValueRow label="Key" value={key} /></>}<SettingsSeparator /><SettingsValueRow label="Scope" value={scope === 'tenant' || effectiveInherited ? context.tenantName : context.inventoryName} /></> : null}</SettingsSection> : null}
    {effectiveInherited ? <Text style={styles.readOnly}>{`Inherited from ${context.tenantName}. Manage it from household settings.`}</Text> : null}
    {effectiveInherited && context.tenantPermissions.includes('configure') && onManageInherited ? <SettingsSection><NativeCommandButton label={`Manage in ${context.tenantName}`} onPress={onManageInherited} /></SettingsSection> : null}
    {canMutate && lifecycle === 'active' ? <View style={[settings.styles.contentBlock, { marginTop: spacing.lg }]}><NativeCommandButton prominence="primary" label={saving ? 'Saving…' : 'Save'} disabled={!valid || saving || lifecycleBusy || (mode === 'edit' && !dirty)} onPress={() => void save()} /></View> : null}
    {mode === 'edit' && canMutate ? <CustomizationLifecycleSection busy={lifecycleBusy || saving} kind={kind} lifecycle={lifecycle} onAction={lifecycleAction} /> : null}
    <SettingsRefreshNotice visible={reads.resource.isRefetchError || reads.types.isRefetchError} onRetry={reads.refreshDefinitions} />
    {dirty ? <Text style={styles.unsaved}>Unsaved changes</Text> : null}
  </ScrollView></KeyboardAvoidingView>;

  async function handleFailure(cause: unknown, fallback: string, title = 'Could not save'): Promise<void> {
    const resource = resourceOwner.current;
    setErrorTitle(title);
    setError(safeCustomizationMessage(cause, fallback));
    if (!(cause instanceof CustomizationFailure) || cause.kind !== 'permission-denied') return;
    setDraftDenied(true);
    setError(undefined);
    try {
      const refreshed = await reads.refreshContext();
      if (resourceOwner.current !== resource) return;
      setContext(refreshed);
      accessPolicy.mutationOrRecord(refreshed, kind, scope, effectiveInherited);
    } catch { /* Keep the populated draft fail-closed until access can be refreshed. */ }
  }

  async function refreshDraftAccess(): Promise<void> {
    const resource = resourceOwner.current; const focus = focusOwner.current;
    try {
      const refreshed = await reads.refreshContext();
      if (resourceOwner.current !== resource) return;
      setContext(refreshed);
      if (accessPolicy.mutationOrRecord(refreshed, kind, scope, effectiveInherited)) {
        setDraftDenied(false);
        setError(undefined);
        return;
      }
      if (focusOwner.current !== focus) return;
      feedback.showNotice({ tone: 'warning', title: 'Access is still unavailable', message: 'Your draft remains read-only.' });
    } catch {
      if (resourceOwner.current !== resource || focusOwner.current !== focus) return;
      feedback.showNotice({ tone: 'error', title: 'Could not refresh access', message: 'Your draft remains read-only. Try again.' });
    }
  }

  function clearLoadedRecord(): void {
    setExpirationEnabled(false); setRecord(undefined); setEligibleTypes([]); setName(''); setKey(''); setDescription(''); setColor('');
    setFieldType('text'); setApplicability('all_assets'); setEnumOptions([]); setNewOption(''); setTargetIds([]); setInitialSnapshot(''); setKeyManuallyEdited(false);
    setAdvanced(false); setNameTouched(false); setDraftDenied(false); setCompletion(undefined); setError(undefined);
    keyFocusRequestedRef.current = false; workflowRef.current.resetExit(); setExitAuthorized(false);
  }
}

function label(kind: CustomizationKind) { return kind === 'tag' ? 'Tag' : kind === 'field' ? 'Custom field' : 'Asset type'; }
function capitalize(value: string) { return value.charAt(0).toUpperCase() + value.slice(1).replaceAll('_', ' '); }
function lifecycleMessage(kind: CustomizationKind, action: 'archive' | 'restore' | 'delete') { if (action === 'delete') return kind === 'field' ? 'This cannot be undone. Deletion is blocked while active assets store a value for this field.' : 'This cannot be undone. Deletion is blocked while active assets or custom fields reference this type.'; if (action === 'restore') return 'This will make it available for normal use again.'; return kind === 'tag' ? 'This tag will no longer be available for new assignment or normal filtering.' : kind === 'field' ? 'Existing values remain, but this field will be hidden from normal editing and validation.' : 'Existing references remain, but this type will no longer be available for new assignments.'; }
function snapshotFor(record: EditorRecord): string { return customizationEditorSnapshot({ expirationEnabled: record.kind === 'asset-type' ? record.expirationEnabled ?? false : false, name: record.displayName, key: record.key, keyManuallyEdited: false, description: record.kind === 'asset-type' ? record.description : '', color: record.kind === 'tag' ? record.color ?? '' : '', fieldType: record.kind === 'field' ? record.type : 'text', applicability: record.kind === 'field' ? record.applicability : 'all_assets', enumOptions: record.kind === 'field' ? record.enumOptions : [], targetIds: record.kind === 'field' ? record.customAssetTypeIds : [] }); }

function createStyles(colors: MobileColorPalette) { return StyleSheet.create({
  formRow: { gap: spacing.sm, padding: spacing.md }, labelRow: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between' }, label: { color: colors.text, fontSize: 15, fontWeight: '700' }, required: { color: colors.textMuted, fontSize: 13 }, input: { backgroundColor: colors.surface, borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, color: colors.text, fontSize: 16, minHeight: 44, paddingHorizontal: spacing.sm, paddingVertical: spacing.sm }, multiline: { minHeight: 100, textAlignVertical: 'top' }, loadingGroup: { backgroundColor: colors.surface, borderRadius: radius.md, marginHorizontal: spacing.md, marginTop: spacing.md, overflow: 'hidden' }, errorSummary: { backgroundColor: colors.dangerSurface, borderRadius: radius.md, gap: spacing.xs, marginHorizontal: spacing.md, marginTop: spacing.md, padding: spacing.md }, errorTitle: { color: colors.danger, fontSize: 16, fontWeight: '800' }, errorText: { color: colors.text, fontSize: 14 }, validationText: { color: colors.danger, fontSize: 13 }, disclosure: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between', minHeight: 44, paddingHorizontal: spacing.md }, disclosureText: { color: colors.action, fontSize: 15, fontWeight: '700' }, readOnly: { color: colors.textMuted, fontSize: 14, lineHeight: 20, marginHorizontal: spacing.md, marginTop: spacing.md }, readOnlyValue: { color: colors.text, fontSize: 17, lineHeight: 23 }, unsaved: { color: colors.warning, fontSize: 13, fontWeight: '700', marginHorizontal: spacing.md, marginTop: spacing.sm, textAlign: 'center' }, choiceGrid: { gap: spacing.xs }, choice: { alignItems: 'center', borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, flexDirection: 'row', gap: spacing.sm, minHeight: 44, paddingHorizontal: spacing.sm }, choiceSelected: { borderColor: colors.action, borderWidth: 2 }, choiceText: { color: colors.text, flex: 1, fontSize: 15 }, lockedValue: { color: colors.text, fontSize: 15, minHeight: 30 }, choiceDisclosure: { alignItems: 'center', borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, flexDirection: 'row', justifyContent: 'space-between', minHeight: 44, paddingHorizontal: spacing.sm }, choiceDisclosureText: { color: colors.text, fontSize: 16 }, choiceModal: { backgroundColor: colors.background, flex: 1, padding: spacing.md }, choiceModalHeader: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between', minHeight: 52 }, choiceModalTitle: { color: colors.text, flex: 1, fontSize: 22, fontWeight: '800' }, choiceModalDone: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 54 }, choiceModalGroup: { backgroundColor: colors.surface, borderRadius: radius.md, overflow: 'hidden' }, choiceModalRow: { alignItems: 'center', flexDirection: 'row', gap: spacing.sm, minHeight: 52, paddingHorizontal: spacing.md }, choiceCheckSpace: { height: 18, width: 18 }, inline: { alignItems: 'center', flexDirection: 'row', gap: spacing.sm }, inlineAction: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 54 }, inlineActionText: { color: colors.action, fontSize: 15, fontWeight: '700' }
}); }
