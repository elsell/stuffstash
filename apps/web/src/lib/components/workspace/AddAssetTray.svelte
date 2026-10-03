<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { addReturnFocusTarget } from '$lib/application/workspaceAddFocus';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { getContext, onDestroy, tick } from 'svelte';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { printingWorkspaceContext, type PrintingWorkspace } from '$lib/ports/printingRepository';
  import type {PrintScope,LabelSelection} from '$lib/domain/printing';
  import {settingsResourceHref} from '$lib/application/settingsManagementNavigation';
  const printing = getContext<PrintingWorkspace | undefined>(printingWorkspaceContext);
  import X from '@lucide/svelte/icons/x';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Sheet from '$lib/components/ui/sheet/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import {
    addAssetKindCopy,
    addDestinationSummary,
    addFormPresentation,
    addPhotoCountLabel,
    assetKindControlOptions,
    quickParentContainerLabel as buildQuickParentContainerLabel,
    quickParentContainerTrail as buildQuickParentContainerTrail,
    quickParentMissingNameMessage,
    quickParentKindOptions
  } from '$lib/application/workspaceAddPresentation';
  import type {
    AddAssetSubmission,
    AssetExpiration,
    AddAssetSaveResult,
    AssetKind,
    AssetTag,
    AssetTagDraft,
    CustomAssetType,
    CustomFieldDefinition,
    MediaUploadPolicy,
    ParentTargetViewModel,
    SelectedPhoto
  } from '$lib/domain/inventory';
  import { applicableCustomFieldDefinitions } from '$lib/domain/inventory';
  import ExpirationField from './ExpirationField.svelte';
  import AddAssetCustomFieldsSection from './AddAssetCustomFieldsSection.svelte';
  import AddAssetPhotosSection from './AddAssetPhotosSection.svelte';
  import AssetTagSelector from './AssetTagSelector.svelte';
  import BinaryOption from './BinaryOption.svelte';
  import { formatBytes } from './formatBytes';
  import ParentTargetPicker from './ParentTargetPicker.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    open, printScope, pendingPrintDraft,
    initialKind = 'item',
    initialParentAssetId = null,
    closeHref,
    parentTargets,
    mediaPolicy,
    customAssetTypes,
    customFieldDefinitions,
    assetTags = [],
    saving,
    restoreFocusOnClose = true,
    onClose,
    onSave
  }: {
    open: boolean;
    printScope?:PrintScope;
    pendingPrintDraft?:AddAssetSubmission;
    initialKind?: AssetKind;
    initialParentAssetId?: string | null;
    closeHref: string;
    parentTargets: ParentTargetViewModel[];
    mediaPolicy: MediaUploadPolicy;
    customAssetTypes: CustomAssetType[];
    customFieldDefinitions: CustomFieldDefinition[];
    assetTags?: AssetTag[];
    saving: boolean;
    restoreFocusOnClose?: boolean;
    onClose: () => void;
    onSave: (draft: AddAssetSubmission) => Promise<AddAssetSaveResult>;
  } = $props();

  let printChecked = $state(false);
  let printSelection = $state<LabelSelection|null>(null);
  let printLoading = $state(false);
  let printLoadFailed = $state(false);
  let printDefaultResolved=$state(false);
  let printGeneration = 0;
  let lastPrintScope = '';
  const printScopeKey = $derived(JSON.stringify(printScope));
  async function initializePrinting() {
    const generation = ++printGeneration;
    printChecked = false; printSelection = null; printLoadFailed = false; printLoading=false; printDefaultResolved=false;
    if (pendingPrintDraft) {printDefaultResolved=true;printChecked=Boolean(pendingPrintDraft.printLabel);printSelection=pendingPrintDraft.printLabel??null;return;}
    if (!printing || !printScope) {printDefaultResolved=true;return;}
    printLoading = true;
    try {
      const [defaults,printers] = await Promise.all([printing.repository.settings(printScope),printing.repository.printers(printScope)]);
      if (generation!==printGeneration) return;
      const printer=printers.find(p=>p.id===defaults.defaultPrinterId&&!p.retired);
      printDefaultResolved=!defaults.printOnCreateDefault||Boolean(printer);
      if(!printDefaultResolved)printLoadFailed=true;
      if (printer) {printSelection={printerId:printer.id,expectedMediaFingerprint:printer.mediaFingerprint,templateId:defaults.templateId,templateVersion:defaults.templateVersion,showReference:defaults.showReference,copies:1};printChecked=defaults.printOnCreateDefault;}
    } catch {if(generation===printGeneration)printLoadFailed=true;}
    finally {if(generation===printGeneration)printLoading=false;}
  }
  function restorePendingDraft(draft:AddAssetSubmission) {
    kind=draft.kind;title=draft.title;description=draft.description;parentAssetId=draft.parentAssetId??'';
    quickParentEnabled=Boolean(draft.parentQuickCreate);quickParentTitle=draft.parentQuickCreate?.title??'';quickParentKind=draft.parentQuickCreate?.kind??'location';
    customAssetTypeId=draft.customAssetTypeId??'';expiration=draft.expiration;customFieldValues=Object.fromEntries(Object.entries(draft.customFields??{}).map(([key,value])=>[key,String(value)]));
    selectedTagIds=[...(draft.tagIds??[])];newTags=[...(draft.newTags??[])];selectedPhotos=draft.photos.map(photo=>({...photo,previewUrl:URL.createObjectURL(photo.file)}));
  }
  let kind = $state<AssetKind>('item');
  let title = $state('');
  let description = $state('');
  let parentAssetId = $state('');
  let parentSearch = $state('');
  let quickParentEnabled = $state(false);
  let quickParentTitle = $state('');
  let quickParentKind = $state<'location' | 'container'>('location');
  let customAssetTypeId = $state('');
  let expiration = $state<AssetExpiration | undefined>();
  let expirationValid = $state(true);
  let customFieldValues = $state<Record<string, string>>({});
  let selectedTagIds = $state<string[]>([]);
  let newTags = $state<AssetTagDraft[]>([]);
  let selectedPhotos = $state<SelectedPhoto[]>([]);
  let photoError = $state('');
  let fileInputKey = $state(0);
  let lastInitialKind = $state<AssetKind>('item');
  let lastInitialParentAssetId = $state<string | null>(null);
  let wasOpen = $state(false);
  let titleInput = $state<HTMLInputElement | null>(null);
  let returnFocusElement: HTMLElement | null = null;
  const assetKindOptions = assetKindControlOptions();
  let activeCustomAssetTypes = $derived(customAssetTypes.filter((assetType) => assetType.lifecycleState === 'active'));
  let expirationEnabled = $derived(activeCustomAssetTypes.find((type) => type.id === customAssetTypeId)?.expirationEnabled ?? false);
  let applicableFields = $derived(applicableCustomFieldDefinitions(customFieldDefinitions, customAssetTypeId || undefined));
  let quickParentMissingName = $derived(quickParentEnabled && quickParentTitle.trim().length === 0);
  let selectedParent = $derived(parentTargets.find((target) => target.id === parentAssetId) ?? null);
  let parentSearchPicking = $derived(parentSearch.trim().length > 0 && parentSearch.trim() !== (selectedParent?.title ?? ''));
  let kindCopy = $derived(addAssetKindCopy(kind));
  let quickParentContainerLabel = $derived(buildQuickParentContainerLabel(selectedParent));
  let quickParentContainerTrail = $derived(buildQuickParentContainerTrail(selectedParent));
  let parentSummary = $derived(
    addDestinationSummary({
      quickParentEnabled,
      quickParentKind,
      quickParentTitle,
      selectedParent
    })
  );
  let photoSummary = $derived(addPhotoCountLabel(selectedPhotos.length));
  let quickParentNameError = quickParentMissingNameMessage();

  $effect(() => {
    if (open && (!wasOpen || lastPrintScope!==printScopeKey)) {
      returnFocusElement = document.activeElement instanceof HTMLElement ? document.activeElement : null;
      resetDraft(initialKind, validInitialParentId(initialParentAssetId));
      lastPrintScope=printScopeKey;
      if(pendingPrintDraft)restorePendingDraft(pendingPrintDraft);
      void initializePrinting();
      wasOpen = true;
      void tick().then(() => titleInput?.focus());
    } else if (!open && wasOpen) {
      wasOpen = false;
      printGeneration++;
      revokePhotoPreviews(selectedPhotos);
      if (restoreFocusOnClose) addReturnFocusTarget(returnFocusElement)?.focus();
      returnFocusElement = null;
    } else if (open && !pendingPrintDraft && initialKind !== lastInitialKind) {
      kind = initialKind;
      lastInitialKind = initialKind;
    } else if (open && !pendingPrintDraft && initialParentAssetId !== lastInitialParentAssetId) {
      parentAssetId = validInitialParentId(initialParentAssetId) ?? '';
      parentSearch = parentAssetId ? parentTargets.find((target) => target.id === parentAssetId)?.title ?? '' : '';
      lastInitialParentAssetId = initialParentAssetId;
    }
  });

  onDestroy(() => {
    printGeneration++;
    revokePhotoPreviews(selectedPhotos);
  });

  async function save(): Promise<void> {
    if (!title.trim() || photoError || !expirationValid) {
      return;
    }
    const result = await onSave(pendingPrintDraft ?? {
      printLabel: printChecked && printSelection ? {...printSelection} : undefined,
      kind,
      title: title.trim(),
      description: description.trim(),
      parentAssetId: parentAssetId || null,
      parentQuickCreate: quickParentEnabled && quickParentTitle.trim()
        ? { kind: quickParentKind, title: quickParentTitle.trim() }
        : undefined,
      customAssetTypeId: customAssetTypeId || undefined,
      expiration: expirationEnabled ? expiration : undefined,
      customFields: buildCustomFields(),
      tagIds: selectedTagIds,
      newTags,
      photos: selectedPhotos
    });
    if (!result.saved) {
      if (result.createdParentId) {
        parentAssetId = result.createdParentId;
        quickParentEnabled = false;
        quickParentTitle = '';
        quickParentKind = 'location';
      }
      return;
    }
    title = '';
    description = '';
    parentAssetId = '';
    parentSearch = '';
    quickParentEnabled = false;
    quickParentTitle = '';
    quickParentKind = 'location';
    customAssetTypeId = '';
    expiration = undefined;
    expirationValid = true;
    customFieldValues = {};
    selectedTagIds = [];
    newTags = [];
    revokePhotoPreviews(selectedPhotos);
    selectedPhotos = [];
    photoError = '';
    lastInitialKind = kind;
  }

  function resetDraft(nextKind: AssetKind, nextParentAssetId: string | null = null): void {
    revokePhotoPreviews(selectedPhotos);
    kind = nextKind;
    title = '';
    description = '';
    parentAssetId = nextParentAssetId ?? '';
    parentSearch = nextParentAssetId ? parentTargets.find((target) => target.id === nextParentAssetId)?.title ?? '' : '';
    quickParentEnabled = false;
    quickParentTitle = '';
    quickParentKind = 'location';
    customAssetTypeId = '';
    expiration = undefined;
    expirationValid = true;
    customFieldValues = {};
    selectedTagIds = [];
    newTags = [];
    selectedPhotos = [];
    photoError = '';
    fileInputKey += 1;
    lastInitialKind = nextKind;
    lastInitialParentAssetId = initialParentAssetId;
  }

  function closeFromLink(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onClose();
  }

  function captureFiles(files: FileList | undefined): void {
    if (!files) {
      return;
    }
    revokePhotoPreviews(selectedPhotos);
    const nextPhotos: SelectedPhoto[] = [];
    const rejected: string[] = [];
    for (const file of Array.from(files)) {
      if (!mediaPolicy.supportedContentTypes.includes(file.type as SelectedPhoto['contentType'])) {
        rejected.push(t('photos.unsupportedNamed', { name: file.name }));
        continue;
      }
      if (file.size <= 0 || file.size > mediaPolicy.maxBytes) {
        rejected.push(t('photos.oversizedNamed', { name: file.name, size: formatBytes(mediaPolicy.maxBytes) }));
        continue;
      }
      nextPhotos.push({
        id: `${file.name}-${file.lastModified}`,
        name: file.name,
        sizeBytes: file.size,
        contentType: file.type as SelectedPhoto['contentType'],
        previewUrl: URL.createObjectURL(file),
        file
      });
    }
    selectedPhotos = nextPhotos;
    photoError = rejected.join(' ');
    fileInputKey += 1;
  }

  function removePhoto(id: string): void {
    const removed = selectedPhotos.find((photo) => photo.id === id);
    if (removed) {
      URL.revokeObjectURL(removed.previewUrl);
    }
    selectedPhotos = selectedPhotos.filter((photo) => photo.id !== id);
    if (selectedPhotos.length === 0) {
      photoError = '';
    }
  }

  function selectParentTarget(id: string | null): void {
    parentAssetId = id ?? '';
    parentSearch = id ? parentTargets.find((target) => target.id === id)?.title ?? parentSearch : '';
  }

  function validInitialParentId(id: string | null | undefined): string | null {
    return id && parentTargets.some((target) => target.id === id) ? id : null;
  }

  function toggleQuickParent(): void {
    quickParentEnabled = !quickParentEnabled;
    if (!quickParentEnabled) {
      quickParentTitle = '';
      quickParentKind = 'location';
    }
  }

  function setCustomAssetType(nextId: string): void {
    if (nextId === customAssetTypeId) return;
    customAssetTypeId = nextId;
    expiration = undefined;
    expirationValid = true;
    customFieldValues = {};
  }

  function setCustomFieldValue(key: string, value: string): void {
    customFieldValues = { ...customFieldValues, [key]: value };
  }

  function setSelectedTagIds(ids: string[]): void {
    selectedTagIds = ids;
  }

  function setNewTags(tags: AssetTagDraft[]): void {
    newTags = tags;
  }

  function buildCustomFields(): Record<string, unknown> {
    const values: Record<string, unknown> = {};
    for (const field of applicableFields) {
      const value = customFieldValues[field.key] ?? '';
      if (!value) {
        continue;
      }
      values[field.key] = field.type === 'number' ? Number(value) : field.type === 'boolean' ? value === 'true' : value;
    }
    return values;
  }

  function revokePhotoPreviews(photos: SelectedPhoto[]): void {
    for (const photo of photos) {
      URL.revokeObjectURL(photo.previewUrl);
    }
  }
</script>

<Sheet.Root {open} onOpenChange={(nextOpen) => { if (!nextOpen) onClose(); }}>
  <Sheet.Content
    side="right"
    class="add-tray workspace-task-sheet w-full max-w-none gap-0 p-0 sm:max-w-xl"
    style="width: 100%;"
    showCloseButton={false}
    data-parent-search-active={parentSearchPicking ? 'true' : undefined}
    onOpenAutoFocus={(event) => { event.preventDefault(); titleInput?.focus(); }}
    onCloseAutoFocus={(event) => { event.preventDefault(); }}
  >
    <Sheet.Header class="section-heading compact shrink-0 border-b px-5 py-4 pr-16 text-left sm:px-6">
      <Sheet.Title id="add-title">{kindCopy.heading}</Sheet.Title>
      <Button.Root href={closeHref} variant="ghost" size="icon-sm" aria-label={t('web.AddAssetTray.closeAddTray')} onclick={closeFromLink}><X /></Button.Root>
    </Sheet.Header>

    <fieldset class="add-tray-body" disabled={saving || Boolean(pendingPrintDraft)}>
      <div class="add-summary">
        <p class="visually-hidden" aria-live="polite" aria-atomic="true">
          {addFormPresentation.summaryTypeLabel}: {kindCopy.kindLabel}.
          {addFormPresentation.summaryParentLabel}: {parentSummary}.
          {addFormPresentation.summaryPhotosLabel}: {photoSummary}.
        </p>
        <div>
          <small>{addFormPresentation.summaryTypeLabel}</small>
          <strong>{kindCopy.kindLabel}</strong>
        </div>
        <div class="add-summary-destination">
          <small>{addFormPresentation.summaryParentLabel}</small>
          <strong>{parentSummary}</strong>
        </div>
        <div>
          <small>{addFormPresentation.summaryPhotosLabel}</small>
          <strong>{photoSummary}</strong>
        </div>
      </div>

      <fieldset class="selection-field">
        <legend>{addFormPresentation.assetKindLegend}</legend>
        <SegmentedControl
          label={addFormPresentation.assetKindLegend}
          value={kind}
          options={assetKindOptions}
          onSelect={(value) => { kind = value as AssetKind; }}
        />
      </fieldset>

      <div class="field-stack">
        <Label for="asset-title">{kindCopy.nameLabel}</Label>
        <Input id="asset-title" bind:ref={titleInput} bind:value={title} placeholder={kindCopy.namePlaceholder} required aria-required="true" />
      </div>

      <ParentTargetPicker
        legend={addFormPresentation.parentPickerLegend}
        searchId="parent-search"
        groupLabel={addFormPresentation.parentPickerGroupLabel}
        bind:search={parentSearch}
        selectedId={parentAssetId || null}
        targets={parentTargets}
        onSelect={selectParentTarget}
      />

      <fieldset class="selection-field quick-parent-section">
        <legend>{addFormPresentation.quickParentLegend}</legend>
        <BinaryOption
          label={addFormPresentation.quickParentToggleLabel}
          description={addFormPresentation.quickParentToggleDescription}
          checked={quickParentEnabled}
          onToggle={toggleQuickParent}
        />
        {#if quickParentEnabled}
          <div class="quick-parent-fields">
            <div class="quick-parent-context">
              <span>{addFormPresentation.quickParentContextLabel}</span>
              <strong>{quickParentContainerLabel}</strong>
              {#if quickParentContainerTrail}
                <small>{quickParentContainerTrail}</small>
              {/if}
            </div>
            <div class="field-stack">
              <Label for="quick-parent-title">{addFormPresentation.quickParentNameLabel}</Label>
              <Input
                id="quick-parent-title"
                bind:value={quickParentTitle}
                placeholder={addFormPresentation.quickParentNamePlaceholder}
                required={quickParentEnabled}
                aria-required={quickParentEnabled}
                aria-invalid={quickParentMissingName}
                aria-describedby={quickParentMissingName ? 'quick-parent-error' : undefined}
              />
              {#if quickParentMissingName}
                <p id="quick-parent-error" class="denied-note" role="alert">{quickParentNameError}</p>
              {/if}
            </div>
            <SegmentedControl
              label={addFormPresentation.quickParentKindLabel}
              value={quickParentKind}
              options={quickParentKindOptions}
              onSelect={(value) => { quickParentKind = value as 'location' | 'container'; }}
            />
          </div>
        {/if}
      </fieldset>

      <div class="field-stack">
        <Label for="asset-description">{addFormPresentation.descriptionLabel}</Label>
        <Textarea id="asset-description" bind:value={description} placeholder={addFormPresentation.descriptionPlaceholder} />
      </div>

      <AddAssetCustomFieldsSection
        {activeCustomAssetTypes}
        {applicableFields}
        {customAssetTypeId}
        {customFieldValues}
        onCustomAssetTypeSelect={setCustomAssetType}
        onCustomFieldValueChange={setCustomFieldValue}
      />

      {#if expirationEnabled}
        {#key customAssetTypeId}
          <ExpirationField id="asset-expiration" onChange={(value, valid) => { expiration = value; expirationValid = valid; }} />
        {/key}
      {/if}

      <AssetTagSelector
        tags={assetTags}
        selectedIds={selectedTagIds}
        {newTags}
        onSelectedIdsChange={setSelectedTagIds}
        onNewTagsChange={setNewTags}
      />

      <AddAssetPhotosSection
        photos={selectedPhotos}
        summary={photoSummary}
        {mediaPolicy}
        inputKey={fileInputKey}
        error={photoError}
        onFiles={captureFiles}
        onRemove={removePhoto}
      />
      {#if printing && printScope}
        <div class="print-create-choice"><Checkbox id="create-print-label" bind:checked={printChecked} disabled={printLoading||!printSelection}/><Label for="create-print-label">{t('web.Printing.createPrint')}</Label></div>
        {#if printLoadFailed && !printDefaultResolved}<div class="print-fallback"><Button.Root variant="outline" onclick={()=>void initializePrinting()}>{t('web.Printing.retryDefaults')}</Button.Root><Button.Root variant="outline" onclick={()=>{printDefaultResolved=true;printChecked=false;}}>{t('web.Printing.createWithoutLabel')}</Button.Root></div>{/if}
        {#if !printSelection || printLoadFailed}<p>{t(printLoadFailed?'web.Printing.unavailable':'web.Printing.chooseDefault')} <a href={settingsResourceHref({level:'inventory',tenantId:printScope.tenantId,inventoryId:printScope.inventoryId,collection:'printing'})}>{t('web.Printing.title')}</a></p>{/if}
      {/if}
    </fieldset>
    {#if pendingPrintDraft}<p class="pending-print" role="status">{t('web.Printing.pendingCreate')}</p>{/if}

    <Sheet.Footer class="tray-actions shrink-0 border-t px-5 py-4 sm:flex-row sm:justify-end sm:px-6">
      <Button.Root href={closeHref} variant="outline" onclick={closeFromLink}>{t('web.AddAssetTray.cancel')}</Button.Root>
      <Button.Root disabled={saving || printLoading || (Boolean(printing&&printScope)&&!printDefaultResolved) || !expirationValid || title.trim().length === 0 || !!photoError || quickParentMissingName} onclick={() => { void save(); }}>{pendingPrintDraft?t('web.Printing.retryCreate'):kindCopy.saveLabel}</Button.Root>
    </Sheet.Footer>
  </Sheet.Content>
</Sheet.Root>

<style>
.print-create-choice{display:flex;align-items:center;gap:var(--space-3);min-height:var(--space-11)}
.print-create-choice :global(input[type="checkbox"]){min-width:1rem;min-height:1rem;width:1rem;height:1rem}
.print-fallback{display:flex;flex-wrap:wrap;gap:var(--space-3)}
.pending-print{padding-inline:var(--space-5);overflow-wrap:anywhere}
fieldset{min-inline-size:0;border:0}
</style>
