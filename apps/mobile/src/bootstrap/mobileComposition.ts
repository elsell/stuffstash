import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiCurrentPrincipalRepository } from '../adapters/identity/ApiCurrentPrincipalRepository';
import { ApiInventorySummaryRepository } from '../adapters/inventories/ApiInventorySummaryRepository';
import { ApiOnboardingGateway } from '../adapters/onboarding/ApiOnboardingGateway';
import { FileSystemConnectionProfileStore } from '../adapters/onboarding/FileSystemConnectionProfileStore';
import { ExpoPhotoSelectionProvider } from '../adapters/photos/ExpoPhotoSelectionProvider';
import { ApiProviderProfileRepository } from '../adapters/providerProfiles/ApiProviderProfileRepository';
import { ExpoSettingsDiagnosticsProvider } from '../adapters/settings/ExpoSettingsDiagnosticsProvider';
import { ExpoVoiceAudioPlayer, ExpoVoiceAudioRecorder } from '../adapters/voice/ExpoVoiceAudio';
import { WebSocketRealtimeVoiceTransport } from '../adapters/voice/WebSocketRealtimeVoiceTransport';
import { InMemoryAddAssetDraftStore } from '../application/add/AddAssetDraftStore';
import { CreateAssetCommand } from '../application/add/CreateAssetCommand';
import { AddDraftScopeQuery } from '../application/add/AddDraftScopeQuery';
import { ParentLookupQuery } from '../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../application/add/PhotoSelectionQuery';
import { AddAssetPhotosCommand } from '../application/assets/AddAssetPhotosCommand';
import { AssetDetailQuery } from '../application/assets/AssetDetailQuery';
import { AssetLifecycleCommand } from '../application/assets/AssetLifecycleCommand';
import { DeleteAssetPhotoCommand } from '../application/assets/DeleteAssetPhotoCommand';
import { InventoryAssetsQuery } from '../application/assets/InventoryAssetsQuery';
import { MoveAssetCommand } from '../application/assets/MoveAssetCommand';
import { UpdateAssetCommand } from '../application/assets/UpdateAssetCommand';
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
import { ProviderProfileVoiceReadinessCheck } from '../application/providerProfiles/ProviderProfileVoiceReadinessCheck';
import { TestProviderProfileCommand } from '../application/providerProfiles/TestProviderProfileCommand';
import { SearchAssetsQuery } from '../application/search/SearchAssetsQuery';
import { SettingsQuery } from '../application/settings/SettingsQuery';
import { VoiceInteractionPreviewQuery } from '../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController } from '../application/voice/RealtimeVoiceSession';
import { loadMobileRuntimeConfigSeed } from '../config/mobileRuntimeConfig';
import type { MobileRuntimeConfig } from '../config/mobileRuntimeConfigCore';

export type MobileComposition = {
  readonly homeDashboardQuery: HomeDashboardQuery;
  readonly selectInventoryCommand: SelectInventoryCommand;
  readonly searchAssetsQuery: SearchAssetsQuery;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly moveAssetCommand: MoveAssetCommand;
  readonly updateAssetCommand: UpdateAssetCommand;
  readonly inventoryAssetsQuery: InventoryAssetsQuery;
  readonly createAssetCommand: CreateAssetCommand;
  readonly addDraftScopeQuery: AddDraftScopeQuery;
  readonly addAssetDraftStore: InMemoryAddAssetDraftStore;
  readonly parentLookupQuery: ParentLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly locationsQuery: LocationsQuery;
  readonly locationAssetsQuery: LocationAssetsQuery;
  readonly settingsQuery: SettingsQuery;
  readonly providerProfileSettingsQuery: ProviderProfileSettingsQuery;
  readonly manageProviderProfileCommand: ManageProviderProfileCommand;
  readonly testProviderProfileCommand: TestProviderProfileCommand;
  readonly voiceInteractionPreviewQuery: VoiceInteractionPreviewQuery;
  readonly realtimeVoiceSessionController: RealtimeVoiceSessionController;
  readonly voiceDeveloperDiagnosticsEnabled: boolean;
};

const connectionProfiles = new FileSystemConnectionProfileStore();
const runtimeSeed = loadMobileRuntimeConfigSeed();

export function getConnectionProfileStore(): ConnectionProfileStore {
  return connectionProfiles;
}

export function createOnboardingCommand(): OnboardingCommand {
  return new OnboardingCommand(
    connectionProfiles,
    createOnboardingGateway,
    runtimeSeed.devToken
  );
}

export function createSeedConnectionProfile(): ConnectionProfile | undefined {
  if (!runtimeSeed.apiBaseUrl || !runtimeSeed.devToken) {
    return undefined;
  }

  return {
    apiBaseUrl: runtimeSeed.apiBaseUrl,
    devToken: runtimeSeed.devToken,
    tenantId: runtimeSeed.tenantId
  };
}

export function createMobileComposition(profile: ConnectionProfile): MobileComposition {
  const client = createStuffStashClient(profile);
  const inventorySummaries = new ApiInventorySummaryRepository(client, profile.tenantId ?? '');
  const principals = new ApiCurrentPrincipalRepository(client);
  const providerProfiles = new ApiProviderProfileRepository(client, profile.tenantId ?? '');
  const providerProfileSettingsQuery = new ProviderProfileSettingsQuery(providerProfiles);
  const config = toRuntimeConfig(profile);
  const addAssetDraftStore = new InMemoryAddAssetDraftStore(createServiceScopeId());

  return {
    homeDashboardQuery: new HomeDashboardQuery(inventorySummaries),
    selectInventoryCommand: new SelectInventoryCommand(inventorySummaries),
    searchAssetsQuery: new SearchAssetsQuery(inventorySummaries),
    assetDetailQuery: new AssetDetailQuery(inventorySummaries),
    assetLifecycleCommand: new AssetLifecycleCommand(inventorySummaries),
    addAssetPhotosCommand: new AddAssetPhotosCommand(inventorySummaries),
    deleteAssetPhotoCommand: new DeleteAssetPhotoCommand(inventorySummaries),
    moveAssetCommand: new MoveAssetCommand(inventorySummaries),
    updateAssetCommand: new UpdateAssetCommand(inventorySummaries),
    inventoryAssetsQuery: new InventoryAssetsQuery(inventorySummaries),
    createAssetCommand: new CreateAssetCommand(inventorySummaries),
    addDraftScopeQuery: new AddDraftScopeQuery(principals),
    addAssetDraftStore,
    parentLookupQuery: new ParentLookupQuery(inventorySummaries),
    photoSelectionQuery: new PhotoSelectionQuery(new ExpoPhotoSelectionProvider()),
    locationsQuery: new LocationsQuery(inventorySummaries),
    locationAssetsQuery: new LocationAssetsQuery(inventorySummaries),
    settingsQuery: new SettingsQuery(
      principals,
      new ExpoSettingsDiagnosticsProvider(config)
    ),
    providerProfileSettingsQuery,
    manageProviderProfileCommand: new ManageProviderProfileCommand(providerProfiles),
    testProviderProfileCommand: new TestProviderProfileCommand(providerProfiles),
    voiceInteractionPreviewQuery: new VoiceInteractionPreviewQuery(inventorySummaries),
    realtimeVoiceSessionController: new RealtimeVoiceSessionController(
      inventorySummaries,
      new ExpoVoiceAudioRecorder(),
      new WebSocketRealtimeVoiceTransport({
        apiBaseUrl: config?.apiBaseUrl ?? 'http://127.0.0.1:8080',
        tokenProvider: () => config?.devToken ?? '',
        diagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled
      }),
      new ExpoVoiceAudioPlayer(),
      {
        diagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled,
        readinessChecker: new ProviderProfileVoiceReadinessCheck(providerProfileSettingsQuery)
      }
    ),
    voiceDeveloperDiagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled
  };
}

function createOnboardingApiClient(profile: ConnectionProfile): StuffStashClient {
  return createStuffStashClient(profile);
}

function createOnboardingGateway(profile: ConnectionProfile): ApiOnboardingGateway {
  return new ApiOnboardingGateway(createOnboardingApiClient(profile));
}

function createStuffStashClient(profile: ConnectionProfile): StuffStashClient {
  return new StuffStashClient({
    baseUrl: profile.apiBaseUrl,
    fetch: createTimeoutFetch(8000),
    tokenProvider: () => profile.devToken
  });
}

function toRuntimeConfig(profile: ConnectionProfile): MobileRuntimeConfig | undefined {
  if (!profile.tenantId) {
    return undefined;
  }

  return {
    apiBaseUrl: profile.apiBaseUrl,
    tenantId: profile.tenantId,
    devToken: profile.devToken,
    voiceDeveloperDiagnosticsEnabled: runtimeSeed.voiceDeveloperDiagnosticsEnabled
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
