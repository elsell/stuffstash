export type AssetKind = 'item' | 'container' | 'location';
export type AssetLifecycleState = 'active' | 'archived';
export type AssetLifecycleFilter = AssetLifecycleState;
export type SearchLifecycleFilter = AssetLifecycleFilter | 'all';
export type AttachmentContentType = 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf';
export type SearchMode = 'fuzzy' | 'exact';
export type SearchCheckoutFilter = 'any' | 'checked_out' | 'available';
export type BrowseSurface = 'list' | 'map';
export type BrowseScope = 'all' | 'places' | 'containers' | 'items';
export type BrowseSort = 'updated_desc' | 'id_asc';
export type WorkspaceMode = 'home' | 'browse' | 'locations' | 'location' | 'asset' | 'search' | 'import' | 'settings';
export type Capability = 'editor' | 'viewer';
export const inventoryAccessRelationships = ['viewer', 'editor'] as const;
export type InventoryAccessRelationship = (typeof inventoryAccessRelationships)[number];
export type InvitationStatus = 'pending' | 'accepted' | 'revoked' | 'cancelled' | 'expired';
export type InvitationStatusFilter = InvitationStatus | 'all';
export const customDefinitionScopes = ['inventory', 'tenant'] as const;
export type CustomDefinitionScope = (typeof customDefinitionScopes)[number];
export type CustomDefinitionLifecycleState = 'active' | 'archived';
export const customFieldTypes = ['text', 'number', 'boolean', 'date', 'url', 'enum'] as const;
export type CustomFieldType = (typeof customFieldTypes)[number];
export const customFieldApplicabilities = ['all_assets', 'custom_asset_types'] as const;
export type CustomFieldApplicability = (typeof customFieldApplicabilities)[number];

export interface Principal {
  id: string;
  email?: string;
}

export interface AccessSummary {
  relationship: string;
  permissions: string[];
}

export interface Tenant {
  id: string;
  name: string;
  access: AccessSummary;
}

export interface Inventory {
  id: string;
  tenantId: string;
  name: string;
  access: AccessSummary;
}

export interface AssetPhoto {
  id: string;
  assetId: string;
  url: string;
  alt: string;
}

export interface AssetTag {
  id: string;
  key: string;
  displayName: string;
  color?: string;
}

export interface AssetTagDraft {
  displayName: string;
  color?: string;
}

export interface ImportJobPreviewTag {
  key: string;
  displayName: string;
  color?: string;
}

export const assetTagDisplayNameMaxLength = 80;

export function assetTagDisplayNameByteLength(value: string): number {
  return new TextEncoder().encode(value.trim()).length;
}

export function assetTagKeyFromDisplayName(value: string): string {
  let key = '';
  let lastHyphen = false;
  for (const character of value.trim().toLowerCase()) {
    if ((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9')) {
      key += character;
      lastHyphen = false;
      continue;
    }
    if (key.length > 0 && !lastHyphen) {
      key += '-';
      lastHyphen = true;
    }
  }
  key = key.replace(/^-+|-+$/g, '');
  if (key.length > 80) {
    key = key.slice(0, 80).replace(/-+$/g, '');
  }
  return key;
}

export interface AssetAttachment {
  id: string;
  tenantId: string;
  inventoryId: string;
  assetId: string;
  fileName: string;
  contentType: AttachmentContentType;
  sizeBytes: number;
  lifecycleState: AssetLifecycleState;
  thumbnailUrl?: string;
  thumbnailHeaders?: Record<string, string>;
}

export interface InventoryAccessGrant {
  tenantId: string;
  inventoryId: string;
  principalId: string;
  relationship: InventoryAccessRelationship;
}

export interface InventoryAccessInvitation {
  id: string;
  tenantId: string;
  inventoryId: string;
  email: string;
  relationship: InventoryAccessRelationship;
  status: InvitationStatus;
  isExpired: boolean;
  expiresAt: string;
  inviterPrincipalId: string;
  acceptedPrincipalId?: string;
}

export interface CreatedInventoryAccessInvitation {
  invitation: InventoryAccessInvitation;
  inviteUrl: string;
}

export type AuditScope = 'inventory' | 'tenant';

export interface AuditRecord {
  id: string;
  tenantId: string;
  inventoryId: string | null;
  principalId: string;
  action: string;
  source: string;
  targetType: string;
  targetId: string;
  occurredAt: string;
  requestId?: string;
  metadata: Record<string, string>;
}

export interface CustomAssetType {
  id: string;
  tenantId: string;
  inventoryId: string | null;
  scope: CustomDefinitionScope;
  key: string;
  displayName: string;
  description: string;
  lifecycleState: CustomDefinitionLifecycleState;
}

export interface CustomFieldDefinition {
  id: string;
  tenantId: string;
  inventoryId: string | null;
  scope: CustomDefinitionScope;
  key: string;
  displayName: string;
  type: CustomFieldType;
  enumOptions: string[];
  applicability: CustomFieldApplicability;
  customAssetTypeIds: string[];
  lifecycleState: CustomDefinitionLifecycleState;
}

export interface MediaUploadPolicy {
  supportedContentTypes: AttachmentContentType[];
  maxBytes: number;
}

export interface Asset {
  id: string;
  tenantId: string;
  inventoryId: string;
  kind: AssetKind;
  title: string;
  description: string;
  parentAssetId: string | null;
  lifecycleState: AssetLifecycleState;
  customAssetTypeId?: string;
  customFields?: Record<string, unknown>;
  customAssetTypeLabel?: string;
  tags?: AssetTag[];
  photo?: AssetPhoto;
  photoUnavailable?: boolean;
  currentCheckout?: CurrentCheckout;
  updatedAt?: string;
}

export type LocationAsset = Asset & { kind: 'location' };

export type AssetCheckoutState = 'open' | 'returned' | 'undone';

export interface CurrentCheckout {
  id: string;
  state: AssetCheckoutState;
  checkedOutAt: string;
  checkedOutByPrincipalId: string;
}

export interface AssetCheckout extends CurrentCheckout {
  tenantId: string;
  inventoryId: string;
  assetId: string;
  checkoutDetails?: string;
  returnedAt?: string;
  returnedByPrincipalId?: string;
  returnDetails?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CheckedOutAsset {
  asset: Asset;
  checkout: CurrentCheckout;
}

export interface AssetCheckoutDraft {
  details?: string;
}

export interface SearchResult {
  type: 'asset';
  asset: Asset;
  inventory: Pick<Inventory, 'id' | 'name'>;
  matches: Array<{
    field: string;
    value: string;
  }>;
}

export interface SearchRequest {
  tenantId: string;
  inventoryId: string;
  query: string;
  tagIds?: string[];
  lifecycleState: SearchLifecycleFilter;
  mode: SearchMode;
  checkoutState?: SearchCheckoutFilter;
}

export type ImportSourceType = 'legacy_homebox' | 'legacy_homebox_csv';

export interface ImportSourceRequest {
  sourceType: ImportSourceType;
  baseUrl?: string;
  username?: string;
  password?: string;
  includeImages?: boolean;
  allowInsecureTLS?: boolean;
  allowPrivateNetwork?: boolean;
  fileName?: string;
  contentBase64?: string;
}

export interface ImportCounts {
  fields: number;
  tags?: number;
  locations: number;
  assets: number;
  attachments: number;
  warnings: number;
  errors: number;
}

export interface ImportMessage {
  code: string;
  severity: string;
  summary: string;
  detail?: string;
  sourceId?: string;
  sourceName?: string;
}

export type ImportJobStatus =
  | 'previewed'
  | 'running'
  | 'succeeded'
  | 'failed'
  | 'cancel_requested'
  | 'cancelled_kept'
  | 'cancelled_discarded'
  | 'discard_failed';

export type ImportJobCancellationMode = 'keep_partial_progress' | 'discard_partial_progress';

export interface ImportJobPreview {
  fields: Array<{
    key: string;
    displayName: string;
    type: string;
  }>;
  tags?: ImportJobPreviewTag[];
  locations: Array<{
    sourceId?: string;
    kind: string;
    title: string;
    parentSourceId?: string;
    archived: boolean;
  }>;
  assets: Array<{
    sourceId?: string;
    kind: string;
    title: string;
    parentSourceId?: string;
    archived: boolean;
  }>;
  attachments: Array<{
    sourceId?: string;
    assetSourceId?: string;
    fileName: string;
    contentType: string;
    sizeBytes: number;
    primary: boolean;
  }>;
  messages: ImportMessage[];
  fieldsTruncated: boolean;
  locationsTruncated: boolean;
  assetsTruncated: boolean;
  attachmentsTruncated: boolean;
  messagesTruncated: boolean;
  tagsTruncated?: boolean;
}

export interface ImportJobResourceSummary {
  resourceType: string;
  resourceId: string;
  displayName?: string;
  resourceOwnerId?: string;
  sourceEntityType: string;
  sourceEntityId: string;
  createdAt: string;
}

export interface ImportJobProgress {
  phase: string;
  done: number;
  total: number;
  message?: string;
  updatedAt?: string;
}

export interface ImportJob {
  id: string;
  status: ImportJobStatus;
  actorId?: string;
  actor?: Principal;
  source: {
    type: string;
    name: string;
    baseUrl?: string;
    version?: string;
    imageImport: string;
    allowPrivateNetwork?: boolean;
    allowInsecureTLS?: boolean;
    fingerprint?: string;
  };
  counts: ImportCounts & {
    fieldsCreated: number;
    fieldsExisting: number;
    tagsCreated?: number;
    tagsExisting?: number;
    locationsCreated: number;
    assetsCreated: number;
    assetsSkipped: number;
    attachmentsCreated: number;
    attachmentsSkipped: number;
    recordsDiscarded: number;
    sourceLinksDiscarded: number;
  };
  preview: ImportJobPreview;
  progress: ImportJobProgress;
  progressHistory: ImportJobProgress[];
  cancellationMode?: ImportJobCancellationMode;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt: string;
  resources: ImportJobResourceSummary[];
  messages: ImportMessage[];
}

export interface AddAssetDraft {
  kind: AssetKind;
  title: string;
  description: string;
  parentAssetId: string | null;
  customAssetTypeId?: string;
  customFields?: Record<string, unknown>;
  tagIds?: string[];
  newTags?: AssetTagDraft[];
  photos: SelectedPhoto[];
}

export interface AddAssetSubmission extends AddAssetDraft {
  parentQuickCreate?: {
    kind: 'location' | 'container';
    title: string;
  };
}

export type AddAssetSaveResult =
  | { saved: true }
  | {
      saved: false;
      createdParentId?: string;
    };

export interface UpdateAssetDraft {
  title: string;
  description: string;
  parentAssetId: string | null;
  customFields?: Record<string, unknown>;
  tagIds?: string[];
  newTags?: AssetTagDraft[];
}

export interface SelectedPhoto {
  id: string;
  name: string;
  sizeBytes: number;
  contentType: AttachmentContentType;
  previewUrl: string;
  file: File;
}

export interface SelectedAttachment {
  id: string;
  name: string;
  sizeBytes: number;
  contentType: AttachmentContentType;
  previewUrl?: string;
  file: File;
}

export interface WorkspaceContext {
  principal: Principal;
  tenants: Tenant[];
  inventories: Inventory[];
  selectedTenantId: string;
  selectedInventoryId: string;
  assetLifecycleState: AssetLifecycleFilter;
  mediaUploadPolicy: MediaUploadPolicy;
  customAssetTypes: CustomAssetType[];
  customFieldDefinitions: CustomFieldDefinition[];
  assetTags?: AssetTag[];
  capability: Capability;
}

export interface WorkspaceData {
  context: WorkspaceContext;
  assets: Asset[];
  checkedOutAssets: CheckedOutAsset[];
}

export interface LocationSummary {
  location: LocationAsset;
  assetCount: number;
}

export interface AssetViewModel extends Asset {
  containmentTrail: string;
}

export type ParentTargetViewModel = AssetViewModel & { kind: 'location' | 'container'; lifecycleState: 'active' };

export const assetKinds: AssetKind[] = ['item', 'container', 'location'];
export const defaultMediaUploadPolicy: MediaUploadPolicy = {
  supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'],
  maxBytes: 5 * 1024 * 1024
};

export function assetKindLabel(kind: AssetKind): string {
  switch (kind) {
    case 'item':
      return 'Item';
    case 'container':
      return 'Container';
    case 'location':
      return 'Location';
  }
}

export function hasAccessPermission(access: AccessSummary | null | undefined, permission: string): boolean {
  return access?.permissions.includes(permission) ?? false;
}

export function canCreateInventory(tenant: Tenant | null | undefined): boolean {
  return hasAccessPermission(tenant?.access, 'create_inventory');
}

export function canCreateAsset(inventory: Inventory | null | undefined): boolean {
  return hasAccessPermission(inventory?.access, 'create_asset');
}

export function canEditInventory(inventory: Inventory | null | undefined): boolean {
  return canCreateAsset(inventory) || hasAccessPermission(inventory?.access, 'edit_asset');
}

export function canEditAsset(inventory: Inventory | null | undefined): boolean {
  return hasAccessPermission(inventory?.access, 'edit_asset');
}

export function canViewImportJobs(inventory: Inventory | null | undefined): boolean {
  return hasAccessPermission(inventory?.access, 'view_import_job');
}

export function canCreateImportJob(inventory: Inventory | null | undefined): boolean {
  return hasAccessPermission(inventory?.access, 'create_import_job');
}

export function applicableCustomFieldDefinitions(
  definitions: CustomFieldDefinition[],
  customAssetTypeId: string | undefined
): CustomFieldDefinition[] {
  return definitions.filter(
    (definition) =>
      definition.lifecycleState === 'active' &&
      (definition.applicability === 'all_assets' ||
        (!!customAssetTypeId && definition.customAssetTypeIds.includes(customAssetTypeId)))
  );
}
