import { StuffStashClient } from '@stuff-stash/api-client';
import * as SecureStore from 'expo-secure-store';
import { ApiMobileAuthMetadataGateway } from '../adapters/auth/ApiMobileAuthMetadataGateway';
import { ExpoOidcNativeClient } from '../adapters/auth/ExpoOidcNativeClient';
import { ExpoSecureAuthSessionStore } from '../adapters/auth/ExpoSecureAuthSessionStore';
import { ApiAssetActivityRepository } from '../adapters/audit/ApiAssetActivityRepository';
import { ApiAssetOperationReversalRepository } from '../adapters/assets/ApiAssetOperationReversalRepository';
import { ApiAssetCheckoutHistoryRepository } from '../adapters/audit/ApiAssetCheckoutHistoryRepository';
import { ApiCurrentPrincipalRepository } from '../adapters/identity/ApiCurrentPrincipalRepository';
import { ApiCustomizationRepository } from '../adapters/customization/ApiCustomizationRepository';
import { ApiInventorySummaryRepository } from '../adapters/inventories/ApiInventorySummaryRepository';
import { ApiInventoryInvitationRepository } from '../adapters/invitations/ApiInventoryInvitationRepository';
import { ApiOnboardingGateway } from '../adapters/onboarding/ApiOnboardingGateway';
import { FileSystemConnectionProfileStore } from '../adapters/onboarding/FileSystemConnectionProfileStore';
import { ExpoPhotoSelectionProvider } from '../adapters/photos/ExpoPhotoSelectionProvider';
import { ApiProviderProfileRepository } from '../adapters/providerProfiles/ApiProviderProfileRepository';
import { ExpoSettingsDiagnosticsProvider } from '../adapters/settings/ExpoSettingsDiagnosticsProvider';
import { FileSystemAppearancePreferenceStore } from '../adapters/settings/FileSystemAppearancePreferenceStore';
import { ApiSettingsScopeRepository } from '../adapters/settings/ApiSettingsScopeRepository';
import { ApiInventoryInvitationManagementRepository } from '../adapters/sharing/ApiInventoryInvitationManagementRepository';
import { ExpoInvitationLinkActions } from '../adapters/sharing/ExpoInvitationLinkActions';
import { ExpoVoiceAudioPlayer, ExpoVoiceAudioRecorder } from '../adapters/voice/ExpoVoiceAudio';
import { WebSocketRealtimeVoiceTransport } from '../adapters/voice/WebSocketRealtimeVoiceTransport';
import { InMemoryAddAssetDraftStore } from '../application/add/AddAssetDraftStore';
import { CreateAssetCommand } from '../application/add/CreateAssetCommand';
import { AddDraftScopeQuery } from '../application/add/AddDraftScopeQuery';
import { ParentLookupQuery } from '../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../application/add/PhotoSelectionQuery';
import { AddAssetPhotosCommand } from '../application/assets/AddAssetPhotosCommand';
import { AssetActivityQuery } from '../application/assets/AssetActivityQuery';
import { AssetCheckoutCommand } from '../application/assets/AssetCheckoutCommand';
import { AssetCheckoutHistoryQuery } from '../application/assets/AssetCheckoutHistoryQuery';
import { AssetDetailQuery } from '../application/assets/AssetDetailQuery';
import { AssetLifecycleCommand } from '../application/assets/AssetLifecycleCommand';
import { DeleteAssetPhotoCommand } from '../application/assets/DeleteAssetPhotoCommand';
import { InventoryAssetsQuery } from '../application/assets/InventoryAssetsQuery';
import { InventoryAssetTagsQuery } from '../application/assets/InventoryAssetTagsQuery';
import { InventoryMapQuery } from '../application/assets/InventoryMapQuery';
import { MoveAssetCommand } from '../application/assets/MoveAssetCommand';
import { UpdateAssetCommand } from '../application/assets/UpdateAssetCommand';
import { UndoAssetEditCommand } from '../application/assets/UndoAssetEditCommand';
import { RevertAssetChangeCommand } from '../application/assets/RevertAssetChangeCommand';
import { HomeDashboardQuery } from '../application/home/HomeDashboardQuery';
import { SelectInventoryCommand } from '../application/home/SelectInventoryCommand';
import { LocationAssetsQuery } from '../application/locations/LocationAssetsQuery';
import { LocationsQuery } from '../application/locations/LocationsQuery';
import {
  ConnectionProfile,
  ConnectionProfileStore
} from '../application/onboarding/ConnectionProfile';
import { OnboardingCommand } from '../application/onboarding/OnboardingCommand';
import { ManageProviderProfileCommand } from '../application/providerProfiles/ManageProviderProfileCommand';
import { ProviderProfileSettingsQuery } from '../application/providerProfiles/ProviderProfileSettingsQuery';
import { TestProviderProfileCommand } from '../application/providerProfiles/TestProviderProfileCommand';
import { SearchAssetsQuery } from '../application/search/SearchAssetsQuery';
import { AcceptInventoryInvitationCommand } from '../application/invitations/AcceptInventoryInvitationCommand';
import { PreviewInventoryInvitationQuery } from '../application/invitations/PreviewInventoryInvitationQuery';
import { SettingsQuery } from '../application/settings/SettingsQuery';
import { CustomizationContextQuery } from '../application/customization/CustomizationContextQuery';
import { CustomizationCollectionQuery } from '../application/customization/CustomizationQueries';
import { ManageTags } from '../application/customization/ManageTags';
import { ManageCustomFields } from '../application/customization/ManageCustomFields';
import { ManageCustomAssetTypes } from '../application/customization/ManageCustomAssetTypes';
import { CustomizationAccessPolicy } from '../application/customization/CustomizationAccess';
import { BufferedCustomizationObservability, type CustomizationEvent } from '../application/customization/CustomizationObservability';
import { AppearancePreferenceController } from '../application/settings/AppearancePreference';
import {
  CancelInventoryInvitationCommand,
  CreateInventoryInvitationCommand,
  ListInventoryInvitationsQuery,
  type InvitationLinkActions
} from '../application/sharing/InventorySharing';
import { VoiceInteractionPreviewQuery } from '../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController } from '../application/voice/RealtimeVoiceSession';
import {
  MobileAuthenticationRequiredError,
  MobileAuthSessionController
} from '../application/auth/MobileAuthSession';
import { loadMobileRuntimeConfigSeed } from '../config/mobileRuntimeConfig';
import type { MobileRuntimeConfig } from '../config/mobileRuntimeConfigCore';

export type MobileComposition = {
  readonly homeDashboardQuery: HomeDashboardQuery;
  readonly selectInventoryCommand: SelectInventoryCommand;
  readonly searchAssetsQuery: SearchAssetsQuery;
  readonly assetActivityQuery: AssetActivityQuery;
  readonly assetCheckoutHistoryQuery: AssetCheckoutHistoryQuery;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly moveAssetCommand: MoveAssetCommand;
  readonly updateAssetCommand: UpdateAssetCommand;
  readonly undoAssetEditCommand: UndoAssetEditCommand;
  readonly revertAssetChangeCommand: RevertAssetChangeCommand;
  readonly inventoryAssetsQuery: InventoryAssetsQuery;
  readonly inventoryAssetTagsQuery: InventoryAssetTagsQuery;
  readonly inventoryMapQuery: InventoryMapQuery;
  readonly createAssetCommand: CreateAssetCommand;
  readonly addDraftScopeQuery: AddDraftScopeQuery;
  readonly addAssetDraftStore: InMemoryAddAssetDraftStore;
  readonly parentLookupQuery: ParentLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly locationsQuery: LocationsQuery;
  readonly locationAssetsQuery: LocationAssetsQuery;
  readonly previewInventoryInvitationQuery: PreviewInventoryInvitationQuery;
  readonly acceptInventoryInvitationCommand: AcceptInventoryInvitationCommand;
  readonly settingsQuery: SettingsQuery;
  readonly customizationContextQuery: CustomizationContextQuery;
  readonly customizationCollectionQuery: CustomizationCollectionQuery;
  readonly manageTags: ManageTags;
  readonly manageCustomFields: ManageCustomFields;
  readonly manageCustomAssetTypes: ManageCustomAssetTypes;
  readonly customizationAccessPolicy: CustomizationAccessPolicy;
  readonly customizationObservability: BufferedCustomizationObservability;
  readonly listInventoryInvitationsQuery: ListInventoryInvitationsQuery;
  readonly createInventoryInvitationCommand: CreateInventoryInvitationCommand;
  readonly cancelInventoryInvitationCommand: CancelInventoryInvitationCommand;
  readonly invitationLinkActions: InvitationLinkActions;
  readonly providerProfileSettingsQuery: ProviderProfileSettingsQuery;
  readonly manageProviderProfileCommand: ManageProviderProfileCommand;
  readonly testProviderProfileCommand: TestProviderProfileCommand;
  readonly voiceInteractionPreviewQuery: VoiceInteractionPreviewQuery;
  readonly realtimeVoiceSessionController: RealtimeVoiceSessionController;
  readonly authSessionController: MobileAuthSessionController;
  readonly voiceDeveloperDiagnosticsEnabled: boolean;
};

export type MobileCompositionOptions = {
  readonly onAuthenticationRequired?: () => void;
  readonly onCustomizationEvent?: (event: CustomizationEvent) => void;
};

const connectionProfiles = new FileSystemConnectionProfileStore();
const appearancePreferences = new AppearancePreferenceController(
  new FileSystemAppearancePreferenceStore()
);
const runtimeSeed = loadMobileRuntimeConfigSeed();
const authSessionController = new MobileAuthSessionController(
  new ExpoSecureAuthSessionStore(SecureStore),
  new ApiMobileAuthMetadataGateway(),
  new ExpoOidcNativeClient()
);

export function getConnectionProfileStore(): ConnectionProfileStore {
  return connectionProfiles;
}

export function getAppearancePreferenceController(): AppearancePreferenceController {
  return appearancePreferences;
}

export function createOnboardingCommand(): OnboardingCommand {
  return new OnboardingCommand(
    connectionProfiles,
    createOnboardingGateway,
    authSessionController
  );
}

export function createSeedConnectionProfile(): ConnectionProfile | undefined {
  if (!runtimeSeed.apiBaseUrl) {
    return undefined;
  }

  return {
    apiBaseUrl: runtimeSeed.apiBaseUrl,
    tenantId: runtimeSeed.tenantId
  };
}

export function createMobileComposition(
  profile: ConnectionProfile,
  options: MobileCompositionOptions = {}
): MobileComposition {
  const client = createStuffStashClient(profile, options);
  const serviceScopeId = createServiceScopeId();
  const config = toRuntimeConfig(profile);
  const directUploadPolicy = {
    allowLocalDevelopmentTargets: runtimeSeed.directUploadLocalDevelopmentTargetsEnabled
  };
  const inventorySummaries = new ApiInventorySummaryRepository(
    client,
    profile.tenantId ?? '',
    undefined,
    serviceScopeId,
    directUploadPolicy
  );
  const inventoryInvitations = new ApiInventoryInvitationRepository(client);
  const managedInvitations = new ApiInventoryInvitationManagementRepository(
    client,
    runtimeSeed.invitationOrigin,
    runtimeSeed.invitationAllowInsecureLocalHTTP
  );
  const assetActivity = new ApiAssetActivityRepository(client);
  const assetChangeReversal = new ApiAssetOperationReversalRepository(client);
  const assetCheckoutHistory = new ApiAssetCheckoutHistoryRepository(client, inventorySummaries);
  const principals = new ApiCurrentPrincipalRepository(client);
  const providerProfiles = new ApiProviderProfileRepository(client, inventorySummaries);
  const providerProfileSettingsQuery = new ProviderProfileSettingsQuery(providerProfiles);
  const addAssetDraftStore = new InMemoryAddAssetDraftStore(serviceScopeId);
  const settingsQuery = new SettingsQuery(
    principals,
    new ExpoSettingsDiagnosticsProvider(config),
    new ApiSettingsScopeRepository(client, inventorySummaries)
  );
  const customization = new ApiCustomizationRepository(client);
  const customizationObservability = new BufferedCustomizationObservability(100, options.onCustomizationEvent);
  const customizationAccessPolicy = new CustomizationAccessPolicy(customizationObservability);

  return {
    homeDashboardQuery: new HomeDashboardQuery(inventorySummaries),
    selectInventoryCommand: new SelectInventoryCommand(inventorySummaries),
    searchAssetsQuery: new SearchAssetsQuery(inventorySummaries),
    assetActivityQuery: new AssetActivityQuery(assetActivity),
    assetCheckoutHistoryQuery: new AssetCheckoutHistoryQuery(assetCheckoutHistory),
    assetDetailQuery: new AssetDetailQuery(inventorySummaries, inventorySummaries, inventorySummaries),
    assetCheckoutCommand: new AssetCheckoutCommand(inventorySummaries),
    assetLifecycleCommand: new AssetLifecycleCommand(inventorySummaries),
    addAssetPhotosCommand: new AddAssetPhotosCommand(inventorySummaries),
    deleteAssetPhotoCommand: new DeleteAssetPhotoCommand(inventorySummaries),
    moveAssetCommand: new MoveAssetCommand(inventorySummaries),
    updateAssetCommand: new UpdateAssetCommand(inventorySummaries),
    undoAssetEditCommand: new UndoAssetEditCommand(assetChangeReversal),
    revertAssetChangeCommand: new RevertAssetChangeCommand(assetChangeReversal),
    inventoryAssetsQuery: new InventoryAssetsQuery(inventorySummaries),
    inventoryAssetTagsQuery: new InventoryAssetTagsQuery(inventorySummaries),
    inventoryMapQuery: new InventoryMapQuery(inventorySummaries),
    createAssetCommand: new CreateAssetCommand(inventorySummaries),
    addDraftScopeQuery: new AddDraftScopeQuery(principals),
    addAssetDraftStore,
    parentLookupQuery: new ParentLookupQuery(inventorySummaries),
    photoSelectionQuery: new PhotoSelectionQuery(new ExpoPhotoSelectionProvider()),
    locationsQuery: new LocationsQuery(inventorySummaries),
    locationAssetsQuery: new LocationAssetsQuery(inventorySummaries),
    previewInventoryInvitationQuery: new PreviewInventoryInvitationQuery(inventoryInvitations),
    acceptInventoryInvitationCommand: new AcceptInventoryInvitationCommand(inventoryInvitations),
    settingsQuery,
    customizationContextQuery: new CustomizationContextQuery(settingsQuery),
    customizationCollectionQuery: new CustomizationCollectionQuery(customization, customizationObservability),
    manageTags: new ManageTags(customization, customizationObservability),
    manageCustomFields: new ManageCustomFields(customization, customizationObservability),
    manageCustomAssetTypes: new ManageCustomAssetTypes(customization, customizationObservability),
    customizationAccessPolicy,
    customizationObservability,
    listInventoryInvitationsQuery: new ListInventoryInvitationsQuery(managedInvitations),
    createInventoryInvitationCommand: new CreateInventoryInvitationCommand(managedInvitations),
    cancelInventoryInvitationCommand: new CancelInventoryInvitationCommand(managedInvitations),
    invitationLinkActions: new ExpoInvitationLinkActions(),
    providerProfileSettingsQuery,
    manageProviderProfileCommand: new ManageProviderProfileCommand(providerProfiles),
    testProviderProfileCommand: new TestProviderProfileCommand(providerProfiles),
    voiceInteractionPreviewQuery: new VoiceInteractionPreviewQuery(inventorySummaries),
    realtimeVoiceSessionController: new RealtimeVoiceSessionController(
      inventorySummaries,
      new ExpoVoiceAudioRecorder(),
      new WebSocketRealtimeVoiceTransport({
        apiBaseUrl: config?.apiBaseUrl ?? 'http://127.0.0.1:8080',
        tokenProvider: () => validIdTokenForProfile(profile, options),
        diagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled,
        directUploadPolicy
      }),
      new ExpoVoiceAudioPlayer(),
      {
        diagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled
      }
    ),
    authSessionController,
    voiceDeveloperDiagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled
  };
}

function createOnboardingApiClient(profile: ConnectionProfile): StuffStashClient {
  return createStuffStashClient(profile);
}

function createOnboardingGateway(profile: ConnectionProfile): ApiOnboardingGateway {
  return new ApiOnboardingGateway(createOnboardingApiClient(profile));
}

function createStuffStashClient(
  profile: ConnectionProfile,
  options: MobileCompositionOptions = {}
): StuffStashClient {
  return new StuffStashClient({
    baseUrl: profile.apiBaseUrl,
    fetch: createAuthenticatedFetch(createTimeoutFetch(8000), options),
    tokenProvider: () => validIdTokenForProfile(profile, options)
  });
}

async function validIdTokenForProfile(
  profile: ConnectionProfile,
  options: MobileCompositionOptions
): Promise<string> {
  try {
    return await authSessionController.validIdToken(profile.apiBaseUrl);
  } catch (error) {
    if (error instanceof MobileAuthenticationRequiredError) {
      options.onAuthenticationRequired?.();
    }
    throw error;
  }
}

function createAuthenticatedFetch(
  fetchImpl: typeof fetch,
  options: MobileCompositionOptions
): typeof fetch {
  return async (input, init) => {
    const response = await fetchImpl(input, init);
    if (response.status === 401) {
      options.onAuthenticationRequired?.();
    }
    return response;
  };
}

function toRuntimeConfig(profile: ConnectionProfile): MobileRuntimeConfig | undefined {
  if (!profile.tenantId) {
    return undefined;
  }

  return {
    apiBaseUrl: profile.apiBaseUrl,
    tenantId: profile.tenantId,
    voiceDeveloperDiagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled,
    directUploadLocalDevelopmentTargetsEnabled: runtimeSeed.directUploadLocalDevelopmentTargetsEnabled,
    invitationOrigin: runtimeSeed.invitationOrigin,
    invitationAllowInsecureLocalHTTP: runtimeSeed.invitationAllowInsecureLocalHTTP
  };
}

function createServiceScopeId(): string {
  return `mobile-composition-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

function createTimeoutFetch(timeoutMs: number): typeof fetch {
  return async (input, init) => {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), timeoutMs);

    try {
      return await fetch(input, {
        ...init,
        signal: init?.signal ?? controller.signal
      });
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error('Network request timed out. Check that the API is reachable from this phone.');
      }

      throw error;
    } finally {
      clearTimeout(timeout);
    }
  };
}
