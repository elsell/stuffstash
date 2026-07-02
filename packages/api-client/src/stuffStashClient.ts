import createClient, { type Client } from 'openapi-fetch';
import type { components, paths } from './generated/schema';

export type OpenAPIPaths = paths;

export type TokenProvider = () => string | null | Promise<string | null>;

export interface StuffStashClientOptions {
  baseUrl: string;
  tokenProvider: TokenProvider;
  fetch?: typeof fetch;
}

export interface Principal {
  id: string;
  email?: string;
}

export interface AccessSummary {
  relationship: string;
  permissions: string[];
}

export type InventoryAccessRelationship = 'viewer' | 'editor';
export type InvitationStatus = 'pending' | 'accepted' | 'revoked' | 'cancelled' | 'expired';
export type InvitationStatusFilter = InvitationStatus | 'all';

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
  acceptanceToken?: string;
}

export interface InvitationAcceptance {
  grant: InventoryAccessGrant;
  invitation: InventoryAccessInvitation;
}

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

export type AssetKind = 'item' | 'container' | 'location';
export type AssetLifecycleState = 'active' | 'archived';
export type AssetLifecycleFilter = AssetLifecycleState | 'all';
export type AssetListSort = 'id_asc' | 'updated_desc';

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
  customFields: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface Attachment {
  id: string;
  tenantId: string;
  inventoryId: string;
  assetId: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  lifecycleState: AssetLifecycleState;
}

export interface CreateAttachmentInput {
  fileName: string;
  contentType: 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf';
  contentBase64: string;
}

export interface InitiateDirectUploadInput {
  fileName: string;
  contentType: 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf';
  sizeBytes: number;
}

export interface DirectUpload {
  uploadId: string;
  attachmentId: string;
  method: string;
  url: string;
  headers: Record<string, string>;
  formFields: Record<string, string>;
  expiresAt: string;
}

export interface AssetPhotoReference {
  uri: string;
  headers: Record<string, string>;
}

export interface CreateAssetInput {
  kind: AssetKind;
  title: string;
  description?: string;
  parentAssetId?: string | null;
  customAssetTypeId?: string;
  customFields?: Record<string, unknown>;
}

export interface UpdateAssetInput {
  title?: string;
  description?: string;
  parentAssetId?: string | null;
  customFields?: Record<string, unknown>;
}

export interface AssetSearchResult {
  type: 'asset';
  tenantId: string;
  inventory: {
    id: string;
    name: string;
  };
  asset: Asset;
  matches: Array<{
    field: string;
    value: string;
  }>;
}

export type SearchMode = 'fuzzy' | 'exact';

export interface SearchAssetsOptions {
  limit?: number;
  cursor?: string;
  lifecycleState?: AssetLifecycleFilter;
  mode?: SearchMode;
}

export type CustomDefinitionScope = 'tenant' | 'inventory';
export type CustomDefinitionLifecycleState = 'active' | 'archived';
export type CustomFieldType = 'text' | 'number' | 'boolean' | 'date' | 'url' | 'enum';
export type CustomFieldApplicability = 'all_assets' | 'custom_asset_types';

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

export interface CreateCustomAssetTypeInput {
  key: string;
  displayName: string;
  description?: string;
}

export interface UpdateCustomAssetTypeInput {
  displayName?: string;
  description?: string;
}

export interface CreateCustomFieldDefinitionInput {
  key: string;
  displayName: string;
  type: CustomFieldType;
  enumOptions?: string[];
  applicability?: CustomFieldApplicability;
  customAssetTypeIds?: string[];
}

export interface UpdateCustomFieldDefinitionInput {
  displayName?: string;
  enumOptions?: string[];
  applicability?: CustomFieldApplicability;
  customAssetTypeIds?: string[];
}

export interface ProviderProfile {
  id: string;
  tenantId: string;
  capability: string;
  providerKind: string;
  displayName: string;
  endpointUrl: string;
  modelName: string;
  runtimeOptions: Record<string, unknown>;
  capabilityMetadata: Record<string, unknown>;
  promptTemplate?: string;
  credentialStatus: string;
  lifecycleState: string;
  lastTestedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateProviderProfileInput {
  capability: string;
  providerKind: string;
  displayName: string;
  endpointUrl?: string;
  modelName?: string;
  runtimeOptions?: Record<string, unknown>;
  capabilityMetadata?: Record<string, unknown>;
  promptTemplate?: string;
  enable?: boolean;
}

export interface UpdateProviderProfileInput {
  displayName?: string;
  endpointUrl?: string;
  modelName?: string;
  runtimeOptions?: Record<string, unknown>;
  capabilityMetadata?: Record<string, unknown>;
  promptTemplate?: string;
}

export interface ReplaceProviderProfileCredentialInput {
  purpose: string;
  credential?: string;
}

export interface ProviderProfileTestResult {
  providerProfileId: string;
  capability: string;
  providerKind: string;
  status: string;
  message: string;
  testedAt: string;
}

export interface ProviderProfileSummary {
  id: string;
  capability: string;
  providerKind: string;
  displayName: string;
  modelName: string;
  credentialStatus: string;
  credentialPurpose?: string;
  lifecycleState: string;
  lastTestedAt?: string;
}

export interface VoiceProviderSlot {
  capability: string;
  label: string;
  selectedProfileId?: string;
  selectedProfile?: ProviderProfileSummary;
  selectionSource: string;
  readiness: string;
  issues: string[];
  recommendedAction: string;
  duplicateProfiles: ProviderProfileSummary[];
}

export interface VoiceProviderConfiguration {
  tenantId: string;
  readiness: string;
  updatedAt?: string;
  profileIds: {
    speechToText?: string;
    languageInference?: string;
    textToSpeech?: string;
  };
  slots: VoiceProviderSlot[];
}

export interface UpdateVoiceProviderConfigurationInput {
  speechToTextProfileId?: string;
  languageInferenceProfileId?: string;
  textToSpeechProfileId?: string;
}

export interface Pagination {
  limit: number;
  nextCursor: string | null;
  hasMore: boolean;
}

export interface Page<T> {
  items: T[];
  pagination: Pagination;
}

type ErrorEnvelope = components['schemas']['ErrorEnvelope'];
type Meta = components['schemas']['Meta'];
type PrincipalResponse = components['schemas']['PrincipalResponse'];
type TenantResponse = components['schemas']['TenantResponse'];
type InventoryResponse = components['schemas']['InventoryResponse'];
type AssetResponse = components['schemas']['AssetResponse'];
type AttachmentResponse = components['schemas']['AttachmentResponse'];
type DirectUploadResponse = components['schemas']['DirectUploadResponse'];
type GrantResponse = components['schemas']['GrantResponse'];
type InvitationResponse = components['schemas']['InvitationResponse'];
type InvitationAcceptanceResponse = components['schemas']['InvitationAcceptanceResponse'];
type RecordResponse = components['schemas']['RecordResponse'];
type AssetTypeResponse = components['schemas']['AssetTypeResponse'];
type DefinitionResponse = components['schemas']['DefinitionResponse'];
type ProviderProfileResponse = components['schemas']['ProviderProfileResponse'];
type TestProviderProfileResponse = components['schemas']['TestProviderProfileResponse'];
type VoiceProviderConfigurationResponse = components['schemas']['VoiceProviderConfigurationResponse'];

interface SuccessEnvelope<T> {
  data: T;
  meta: Meta;
}

export class StuffStashAPIError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'StuffStashAPIError';
    this.status = status;
    this.code = code;
  }
}

export class StuffStashClient {
  private readonly client: Client<paths>;
  private readonly baseUrl: string;
  private readonly tokenProvider: TokenProvider;

  constructor(options: StuffStashClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, '');
    this.tokenProvider = options.tokenProvider;
    this.client = createClient<paths>({
      baseUrl: this.baseUrl,
      fetch: options.fetch
    });
  }

  async me(): Promise<Principal> {
    const envelope = await this.unwrap(
      this.client.GET('/me', {
        headers: await this.authHeaders()
      })
    );
    return mapPrincipal(envelope.data);
  }

  async createTenant(name: string): Promise<Tenant> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants', {
        headers: await this.authHeaders(),
        body: { name }
      })
    );
    return mapTenant(envelope.data);
  }

  async listMyTenants(limit = 50, cursor?: string): Promise<Page<Tenant>> {
    const envelope = await this.unwrap(
      this.client.GET('/me/tenants', {
        headers: await this.authHeaders(),
        params: {
          query: { limit, cursor }
        }
      })
    );
    return mapPage(envelope, mapTenant);
  }

  async getTenant(tenantId: string): Promise<Tenant> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } }
      })
    );
    return mapTenant(envelope.data);
  }

  async listInventories(tenantId: string, limit = 50, cursor?: string): Promise<Page<Inventory>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId },
          query: { limit, cursor }
        }
      })
    );
    return mapPage(envelope, mapInventory);
  }

  async createInventory(tenantId: string, name: string): Promise<Inventory> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } },
        body: { name }
      })
    );
    return mapInventory(envelope.data);
  }

  async listAssets(
    tenantId: string,
    inventoryId: string,
    limit = 50,
    cursor?: string,
    lifecycleState: AssetLifecycleFilter = 'active',
    sort: AssetListSort = 'id_asc'
  ): Promise<Page<Asset>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/assets', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId },
          query: { limit, cursor, lifecycleState, sort: sort === 'id_asc' ? undefined : sort }
        }
      })
    );
    return mapPage(envelope, mapAsset);
  }

  async createAsset(tenantId: string, inventoryId: string, input: CreateAssetInput): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId } },
        body: { ...input, parentAssetId: input.parentAssetId ?? undefined }
      })
    );
    return mapAsset(envelope.data);
  }

  async getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
    return mapAsset(envelope.data);
  }

  async updateAsset(tenantId: string, inventoryId: string, assetId: string, input: UpdateAssetInput): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } },
        body: input
      })
    );
    return mapAsset(envelope.data);
  }

  async searchAssets(tenantId: string, query: string, options: SearchAssetsOptions = {}): Promise<Page<AssetSearchResult>> {
    const limit = options.limit ?? 20;
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/search/assets', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId },
          query: {
            q: query,
            limit,
            cursor: options.cursor,
            lifecycleState: options.lifecycleState ?? 'active',
            mode: options.mode ?? 'fuzzy'
          }
        }
      })
    );
    return mapPage(envelope, mapAssetSearchResult);
  }

  async listTenantCustomAssetTypes(tenantId: string, limit = 50, cursor?: string): Promise<Page<CustomAssetType>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/custom-asset-types', {
        headers: await this.authHeaders(),
        params: { path: { tenantId }, query: { limit, cursor } }
      })
    );
    return mapPage(envelope, mapCustomAssetType);
  }

  async createTenantCustomAssetType(tenantId: string, input: CreateCustomAssetTypeInput): Promise<CustomAssetType> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/custom-asset-types', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } },
        body: input
      })
    );
    return mapCustomAssetType(envelope.data);
  }

  async updateTenantCustomAssetType(
    tenantId: string,
    customAssetTypeId: string,
    input: UpdateCustomAssetTypeInput
  ): Promise<CustomAssetType> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, customAssetTypeId } },
        body: input
      })
    );
    return mapCustomAssetType(envelope.data);
  }

  async archiveTenantCustomAssetType(tenantId: string, customAssetTypeId: string): Promise<CustomAssetType> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, customAssetTypeId } }
      })
    );
    return mapCustomAssetType(envelope.data);
  }

  async listInventoryCustomAssetTypes(
    tenantId: string,
    inventoryId: string,
    limit = 50,
    cursor?: string
  ): Promise<Page<CustomAssetType>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId }, query: { limit, cursor } }
      })
    );
    return mapPage(envelope, mapCustomAssetType);
  }

  async createInventoryCustomAssetType(
    tenantId: string,
    inventoryId: string,
    input: CreateCustomAssetTypeInput
  ): Promise<CustomAssetType> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId } },
        body: input
      })
    );
    return mapCustomAssetType(envelope.data);
  }

  async updateInventoryCustomAssetType(
    tenantId: string,
    inventoryId: string,
    customAssetTypeId: string,
    input: UpdateCustomAssetTypeInput
  ): Promise<CustomAssetType> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, customAssetTypeId } },
        body: input
      })
    );
    return mapCustomAssetType(envelope.data);
  }

  async archiveInventoryCustomAssetType(
    tenantId: string,
    inventoryId: string,
    customAssetTypeId: string
  ): Promise<CustomAssetType> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, customAssetTypeId } }
      })
    );
    return mapCustomAssetType(envelope.data);
  }

  async listTenantCustomFieldDefinitions(
    tenantId: string,
    limit = 50,
    cursor?: string
  ): Promise<Page<CustomFieldDefinition>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/custom-field-definitions', {
        headers: await this.authHeaders(),
        params: { path: { tenantId }, query: { limit, cursor } }
      })
    );
    return mapPage(envelope, mapCustomFieldDefinition);
  }

  async createTenantCustomFieldDefinition(
    tenantId: string,
    input: CreateCustomFieldDefinitionInput
  ): Promise<CustomFieldDefinition> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/custom-field-definitions', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } },
        body: mapCreateCustomFieldDefinitionInput(input)
      })
    );
    return mapCustomFieldDefinition(envelope.data);
  }

  async updateTenantCustomFieldDefinition(
    tenantId: string,
    definitionId: string,
    input: UpdateCustomFieldDefinitionInput
  ): Promise<CustomFieldDefinition> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/custom-field-definitions/{definitionId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, definitionId } },
        body: input
      })
    );
    return mapCustomFieldDefinition(envelope.data);
  }

  async archiveTenantCustomFieldDefinition(tenantId: string, definitionId: string): Promise<CustomFieldDefinition> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/custom-field-definitions/{definitionId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, definitionId } }
      })
    );
    return mapCustomFieldDefinition(envelope.data);
  }

  async listInventoryCustomFieldDefinitions(
    tenantId: string,
    inventoryId: string,
    limit = 50,
    cursor?: string
  ): Promise<Page<CustomFieldDefinition>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId }, query: { limit, cursor } }
      })
    );
    return mapPage(envelope, mapCustomFieldDefinition);
  }

  async createInventoryCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    input: CreateCustomFieldDefinitionInput
  ): Promise<CustomFieldDefinition> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId } },
        body: mapCreateCustomFieldDefinitionInput(input)
      })
    );
    return mapCustomFieldDefinition(envelope.data);
  }

  async updateInventoryCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    input: UpdateCustomFieldDefinitionInput
  ): Promise<CustomFieldDefinition> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, definitionId } },
        body: input
      })
    );
    return mapCustomFieldDefinition(envelope.data);
  }

  async archiveInventoryCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string
  ): Promise<CustomFieldDefinition> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, definitionId } }
      })
    );
    return mapCustomFieldDefinition(envelope.data);
  }

  async listInventoryAccessGrants(
    tenantId: string,
    inventoryId: string,
    limit = 50,
    cursor?: string
  ): Promise<Page<InventoryAccessGrant>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/access-grants', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId },
          query: { limit, cursor }
        }
      })
    );
    return mapPage(envelope, mapGrant);
  }

  async grantInventoryAccess(
    tenantId: string,
    inventoryId: string,
    input: { principalId: string; relationship: InventoryAccessRelationship }
  ): Promise<InventoryAccessGrant> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/access-grants', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId } },
        body: input
      })
    );
    return mapGrant(envelope.data);
  }

  async revokeInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ): Promise<void> {
    await this.unwrapNoContent(
      this.client.DELETE('/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, principalId, relationship } }
      })
    );
  }

  async listInventoryAccessInvitations(
    tenantId: string,
    inventoryId: string,
    options: { limit?: number; cursor?: string; status?: InvitationStatusFilter } = {}
  ): Promise<Page<InventoryAccessInvitation>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/access-invitations', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId },
          query: { limit: options.limit ?? 50, cursor: options.cursor, status: options.status ?? 'all' }
        }
      })
    );
    return mapPage(envelope, mapInvitation);
  }

  async createInventoryAccessInvitation(
    tenantId: string,
    inventoryId: string,
    input: { email: string; relationship: InventoryAccessRelationship }
  ): Promise<InventoryAccessInvitation> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/access-invitations', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId } },
        body: input
      })
    );
    return mapInvitation(envelope.data);
  }

  async updateInventoryAccessInvitationExpiration(
    tenantId: string,
    inventoryId: string,
    invitationId: string,
    expiresAt: string
  ): Promise<InventoryAccessInvitation> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/expiration', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, invitationId } },
        body: { expiresAt }
      })
    );
    return mapInvitation(envelope.data);
  }

  async cancelInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    await this.unwrapNoContent(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/cancel', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, invitationId } }
      })
    );
  }

  async deleteInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    await this.unwrapNoContent(
      this.client.DELETE('/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, invitationId } }
      })
    );
  }

  async acceptInventoryAccessInvitation(
    tenantId: string,
    inventoryId: string,
    invitationId: string,
    acceptanceToken: string
  ): Promise<InvitationAcceptance> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, invitationId } },
        body: { acceptanceToken }
      })
    );
    return mapInvitationAcceptance(envelope.data);
  }

  async listTenantAuditRecords(
    tenantId: string,
    limit = 50,
    cursor?: string,
    signal?: AbortSignal
  ): Promise<Page<AuditRecord>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/audit-records', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId },
          query: { limit, cursor }
        },
        signal
      })
    );
    return mapPage(envelope, mapAuditRecord);
  }

  async listInventoryAuditRecords(
    tenantId: string,
    inventoryId: string,
    limit = 50,
    cursor?: string,
    signal?: AbortSignal
  ): Promise<Page<AuditRecord>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/audit-records', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId },
          query: { limit, cursor }
        },
        signal
      })
    );
    return mapPage(envelope, mapAuditRecord);
  }

  async listProviderProfiles(tenantId: string): Promise<ProviderProfile[]> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/provider-profiles', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } }
      })
    );
    return (envelope.data ?? []).map(mapProviderProfile);
  }

  async getVoiceProviderConfiguration(tenantId: string): Promise<VoiceProviderConfiguration> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/voice-provider-configuration', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } }
      })
    );
    return mapVoiceProviderConfiguration(envelope.data);
  }

  async updateVoiceProviderConfiguration(
    tenantId: string,
    input: UpdateVoiceProviderConfigurationInput
  ): Promise<VoiceProviderConfiguration> {
    const envelope = await this.unwrap(
      this.client.PUT('/tenants/{tenantId}/voice-provider-configuration', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } },
        body: input
      })
    );
    return mapVoiceProviderConfiguration(envelope.data);
  }

  async createProviderProfile(
    tenantId: string,
    input: CreateProviderProfileInput
  ): Promise<ProviderProfile> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/provider-profiles', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } },
        body: input
      })
    );
    return mapProviderProfile(envelope.data);
  }

  async updateProviderProfile(
    tenantId: string,
    providerProfileId: string,
    input: UpdateProviderProfileInput
  ): Promise<ProviderProfile> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/provider-profiles/{providerProfileId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, providerProfileId } },
        body: input
      })
    );
    return mapProviderProfile(envelope.data);
  }

  async replaceProviderProfileCredential(
    tenantId: string,
    providerProfileId: string,
    input: ReplaceProviderProfileCredentialInput
  ): Promise<ProviderProfile> {
    const envelope = await this.unwrap(
      this.client.PUT('/tenants/{tenantId}/provider-profiles/{providerProfileId}/credential', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, providerProfileId } },
        body: input
      })
    );
    return mapProviderProfile(envelope.data);
  }

  async enableProviderProfile(tenantId: string, providerProfileId: string): Promise<ProviderProfile> {
    return this.changeProviderProfileLifecycle(tenantId, providerProfileId, 'enable');
  }

  async disableProviderProfile(tenantId: string, providerProfileId: string): Promise<ProviderProfile> {
    return this.changeProviderProfileLifecycle(tenantId, providerProfileId, 'disable');
  }

  async archiveProviderProfile(tenantId: string, providerProfileId: string): Promise<ProviderProfile> {
    return this.changeProviderProfileLifecycle(tenantId, providerProfileId, 'archive');
  }

  async testProviderProfile(
    tenantId: string,
    providerProfileId: string
  ): Promise<ProviderProfileTestResult> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/provider-profiles/{providerProfileId}/test', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, providerProfileId } }
      })
    );
    return mapProviderProfileTestResult(envelope.data);
  }

  private async changeProviderProfileLifecycle(
    tenantId: string,
    providerProfileId: string,
    action: 'enable' | 'disable' | 'archive'
  ): Promise<ProviderProfile> {
    const envelope = await this.unwrap(
      this.client.POST(`/tenants/{tenantId}/provider-profiles/{providerProfileId}/${action}`, {
        headers: await this.authHeaders(),
        params: { path: { tenantId, providerProfileId } }
      })
    );
    return mapProviderProfile(envelope.data);
  }

  async archiveAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
    return mapAsset(envelope.data);
  }

  async restoreAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
    return mapAsset(envelope.data);
  }

  async deleteAsset(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    await this.unwrapNoContent(
      this.client.DELETE('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
  }

  async listAssetAttachments(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    limit = 10,
    cursor?: string
  ): Promise<Page<Attachment>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId, assetId },
          query: { limit, cursor }
        }
      })
    );
    return mapPage(envelope, mapAttachment);
  }

  async createAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    input: CreateAttachmentInput
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } },
        body: input
      })
    );
    return mapAttachment(envelope.data);
  }

  async initiateAssetAttachmentDirectUpload(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    input: InitiateDirectUploadInput
  ): Promise<DirectUpload> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } },
        body: input
      })
    );
    return mapDirectUpload(envelope.data);
  }

  async completeAssetAttachmentDirectUpload(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    uploadId: string
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads/{uploadId}/complete', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, uploadId } }
      })
    );
    return mapAttachment(envelope.data);
  }

  async archiveAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, attachmentId } }
      })
    );
    return mapAttachment(envelope.data);
  }

  async restoreAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/restore', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, attachmentId } }
      })
    );
    return mapAttachment(envelope.data);
  }

  async deleteAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<void> {
    await this.unwrapNoContent(
      this.client.DELETE('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, attachmentId } }
      })
    );
  }

  async assetAttachmentThumbnailReference(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetPhotoReference> {
    return {
      uri: [
        this.baseUrl,
        'tenants',
        encodeURIComponent(tenantId),
        'inventories',
        encodeURIComponent(inventoryId),
        'assets',
        encodeURIComponent(assetId),
        'attachments',
        encodeURIComponent(attachmentId),
        'thumbnail?variant=small'
      ].join('/'),
      headers: await this.authHeaders()
    };
  }

  private async authHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {};
    const token = await this.tokenProvider();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    return headers;
  }

  private async unwrap<T>(request: Promise<{ data?: T; error?: ErrorEnvelope; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (!response.ok) {
      throw new StuffStashAPIError(
        response.status,
        error?.error?.code ?? 'request_failed',
        readableAPIErrorMessage(response.status, error)
      );
    }
    if (data === undefined) {
      throw new StuffStashAPIError(response.status, 'invalid_response', 'Invalid API response.');
    }
    return data;
  }

  private async unwrapNoContent(request: Promise<{ error?: ErrorEnvelope; response: Response }>): Promise<void> {
    const { error, response } = await request;
    if (!response.ok) {
      throw new StuffStashAPIError(
        response.status,
        error?.error?.code ?? 'request_failed',
        readableAPIErrorMessage(response.status, error)
      );
    }
  }

}

function readableAPIErrorMessage(status: number, error: ErrorEnvelope | undefined): string {
  const message = error?.error?.message?.trim() || 'Request failed.';
  const detail = error?.error?.details?.find((item) => item?.message?.trim())?.message?.trim();
  const safeValidationStatus = status === 400 || status === 422;
  if (
    detail
    && safeValidationStatus
    && error?.error?.code === 'invalid_request'
    && (message === 'Invalid request.' || message === 'validation failed')
  ) {
    return detail;
  }
  return message;
}

function mapAttachment(response: AttachmentResponse): Attachment {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId,
    assetId: response.assetId,
    fileName: response.fileName,
    contentType: response.contentType,
    sizeBytes: response.sizeBytes,
    lifecycleState: mapAssetLifecycleState(response.lifecycleState)
  };
}

function mapDirectUpload(response: DirectUploadResponse): DirectUpload {
  return {
    uploadId: response.uploadId,
    attachmentId: response.attachmentId,
    method: response.method,
    url: response.url,
    headers: response.headers ?? {},
    formFields: response.formFields ?? {},
    expiresAt: response.expiresAt
  };
}

function mapPrincipal(response: PrincipalResponse): Principal {
  return { id: response.id, email: response.email };
}

function mapTenant(response: TenantResponse): Tenant {
  return { id: response.id, name: response.name, access: mapAccess(response.access) };
}

function mapInventory(response: InventoryResponse): Inventory {
  return {
    id: response.id,
    tenantId: response.tenantId,
    name: response.name,
    access: mapAccess(response.access)
  };
}

function mapAccess(response: components['schemas']['AccessResponse']): AccessSummary {
  return {
    relationship: response.relationship,
    permissions: response.permissions ?? []
  };
}

function mapGrant(response: GrantResponse): InventoryAccessGrant {
  return {
    tenantId: response.tenantId,
    inventoryId: response.inventoryId,
    principalId: response.principalId,
    relationship: mapInventoryAccessRelationship(response.relationship)
  };
}

function mapInvitation(response: InvitationResponse): InventoryAccessInvitation {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId,
    email: response.email,
    relationship: mapInventoryAccessRelationship(response.relationship),
    status: mapInvitationStatus(response.status),
    isExpired: response.isExpired,
    expiresAt: response.expiresAt,
    inviterPrincipalId: response.inviterPrincipalId,
    acceptedPrincipalId: response.acceptedPrincipalId,
    acceptanceToken: response.acceptanceToken
  };
}

function mapInvitationAcceptance(response: InvitationAcceptanceResponse): InvitationAcceptance {
  return {
    grant: mapGrant(response.grant),
    invitation: mapInvitation(response.invitation)
  };
}

function mapAuditRecord(response: RecordResponse): AuditRecord {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId ?? null,
    principalId: response.principalId,
    action: response.action,
    source: response.source,
    targetType: response.targetType,
    targetId: response.targetId,
    occurredAt: response.occurredAt,
    requestId: response.requestId,
    metadata: response.metadata ?? {}
  };
}

function mapAsset(response: AssetResponse): Asset {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId,
    kind: mapAssetKind(response.kind),
    title: response.title,
    description: response.description,
    parentAssetId: response.parentAssetId ?? null,
    lifecycleState: mapAssetLifecycleState(response.lifecycleState),
    customAssetTypeId: response.customAssetTypeId,
    customFields: response.customFields ?? {},
    createdAt: response.createdAt,
    updatedAt: response.updatedAt
  };
}

function mapAssetSearchResult(response: components['schemas']['AssetSearchResultResponse']): AssetSearchResult {
  return {
    type: 'asset',
    tenantId: response.tenantId,
    inventory: response.inventory,
    asset: {
      id: response.asset.id,
      tenantId: response.tenantId,
      inventoryId: response.asset.inventoryId,
      kind: mapAssetKind(response.asset.kind),
      title: response.asset.title,
      description: response.asset.description,
      parentAssetId: response.asset.parentAssetId ?? null,
      lifecycleState: mapAssetLifecycleState(response.asset.lifecycleState),
      customAssetTypeId: response.asset.customAssetTypeId,
      customFields: response.asset.customFields ?? {},
      createdAt: response.asset.createdAt,
      updatedAt: response.asset.updatedAt
    },
    matches: response.matches ?? []
  };
}

function mapCustomAssetType(response: AssetTypeResponse): CustomAssetType {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId ?? null,
    scope: mapCustomDefinitionScope(response.scope),
    key: response.key,
    displayName: response.displayName,
    description: response.description,
    lifecycleState: mapCustomDefinitionLifecycleState(response.lifecycleState)
  };
}

function mapCustomFieldDefinition(response: DefinitionResponse): CustomFieldDefinition {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId ?? null,
    scope: mapCustomDefinitionScope(response.scope),
    key: response.key,
    displayName: response.displayName,
    type: mapCustomFieldType(response.type),
    enumOptions: response.enumOptions ?? [],
    applicability: mapCustomFieldApplicability(response.applicability),
    customAssetTypeIds: response.customAssetTypeIds ?? [],
    lifecycleState: mapCustomDefinitionLifecycleState(response.lifecycleState)
  };
}

function mapProviderProfile(response: ProviderProfileResponse): ProviderProfile {
  return {
    id: response.id,
    tenantId: response.tenantId,
    capability: response.capability,
    providerKind: response.providerKind,
    displayName: response.displayName,
    endpointUrl: response.endpointUrl,
    modelName: response.modelName,
    runtimeOptions: response.runtimeOptions ?? {},
    capabilityMetadata: response.capabilityMetadata ?? {},
    promptTemplate: response.promptTemplate,
    credentialStatus: response.credentialStatus,
    lifecycleState: response.lifecycleState,
    lastTestedAt: response.lastTestedAt,
    createdAt: response.createdAt,
    updatedAt: response.updatedAt
  };
}

function mapProviderProfileTestResult(
  response: TestProviderProfileResponse
): ProviderProfileTestResult {
  return {
    providerProfileId: response.providerProfileId,
    capability: response.capability,
    providerKind: response.providerKind,
    status: response.status,
    message: response.message,
    testedAt: response.testedAt
  };
}

function mapVoiceProviderConfiguration(response: VoiceProviderConfigurationResponse): VoiceProviderConfiguration {
  return {
    tenantId: response.tenantId,
    readiness: response.readiness,
    updatedAt: response.updatedAt,
    profileIds: {
      speechToText: response.profileIds?.speechToText,
      languageInference: response.profileIds?.languageInference,
      textToSpeech: response.profileIds?.textToSpeech
    },
    slots: (response.slots ?? []).map((slot) => ({
      capability: slot.capability,
      label: slot.label,
      selectedProfileId: slot.selectedProfileId,
      selectedProfile: slot.selectedProfile ? mapProviderProfileSummary(slot.selectedProfile) : undefined,
      selectionSource: slot.selectionSource,
      readiness: slot.readiness,
      issues: slot.issues ?? [],
      recommendedAction: slot.recommendedAction,
      duplicateProfiles: (slot.duplicateProfiles ?? []).map(mapProviderProfileSummary)
    }))
  };
}

function mapProviderProfileSummary(response: ProviderProfileSummary): ProviderProfileSummary {
  return {
    id: response.id,
    capability: response.capability,
    providerKind: response.providerKind,
    displayName: response.displayName,
    modelName: response.modelName,
    credentialStatus: response.credentialStatus,
    credentialPurpose: response.credentialPurpose,
    lifecycleState: response.lifecycleState,
    lastTestedAt: response.lastTestedAt
  };
}

function mapCreateCustomFieldDefinitionInput(
  input: CreateCustomFieldDefinitionInput
): components['schemas']['CreateDefinitionBody'] {
  return {
    ...input,
    enumOptions: input.enumOptions,
    customAssetTypeIds: input.customAssetTypeIds
  };
}

function mapAssetKind(kind: string): AssetKind {
  switch (kind) {
    case 'item':
    case 'container':
    case 'location':
      return kind;
    default:
      throw new StuffStashAPIError(200, 'invalid_asset_kind', 'Invalid asset kind.');
  }
}

function mapAssetLifecycleState(lifecycleState: string): AssetLifecycleState {
  switch (lifecycleState) {
    case 'active':
    case 'archived':
      return lifecycleState;
    default:
      throw new StuffStashAPIError(200, 'invalid_asset_lifecycle_state', 'Invalid asset lifecycle state.');
  }
}

function mapCustomDefinitionScope(scope: string): CustomDefinitionScope {
  switch (scope) {
    case 'tenant':
    case 'inventory':
      return scope;
    default:
      throw new StuffStashAPIError(200, 'invalid_custom_definition_scope', 'Invalid custom definition scope.');
  }
}

function mapCustomDefinitionLifecycleState(lifecycleState: string): CustomDefinitionLifecycleState {
  switch (lifecycleState) {
    case 'active':
    case 'archived':
      return lifecycleState;
    default:
      throw new StuffStashAPIError(200, 'invalid_custom_definition_lifecycle_state', 'Invalid custom definition lifecycle state.');
  }
}

function mapCustomFieldType(type: string): CustomFieldType {
  switch (type) {
    case 'text':
    case 'number':
    case 'boolean':
    case 'date':
    case 'url':
    case 'enum':
      return type;
    default:
      throw new StuffStashAPIError(200, 'invalid_custom_field_type', 'Invalid custom field type.');
  }
}

function mapCustomFieldApplicability(applicability: string): CustomFieldApplicability {
  switch (applicability) {
    case 'all_assets':
    case 'custom_asset_types':
      return applicability;
    default:
      throw new StuffStashAPIError(200, 'invalid_custom_field_applicability', 'Invalid custom field applicability.');
  }
}

function mapInventoryAccessRelationship(relationship: string): InventoryAccessRelationship {
  switch (relationship) {
    case 'viewer':
    case 'editor':
      return relationship;
    default:
      throw new StuffStashAPIError(200, 'invalid_inventory_access_relationship', 'Invalid inventory access relationship.');
  }
}

function mapInvitationStatus(status: string): InvitationStatus {
  switch (status) {
    case 'pending':
    case 'accepted':
    case 'revoked':
    case 'cancelled':
    case 'expired':
      return status;
    default:
      throw new StuffStashAPIError(200, 'invalid_invitation_status', 'Invalid invitation status.');
  }
}

function mapPage<TResponse, TItem>(
  envelope: SuccessEnvelope<TResponse[] | null>,
  mapper: (response: TResponse) => TItem
): Page<TItem> {
  return {
    items: (envelope.data ?? []).map(mapper),
    pagination: {
      limit: envelope.meta.pagination?.limit ?? 0,
      nextCursor: envelope.meta.pagination?.nextCursor ?? null,
      hasMore: envelope.meta.pagination?.hasMore ?? false
    }
  };
}
