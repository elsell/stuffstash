import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiCurrentPrincipalRepository } from '../adapters/identity/ApiCurrentPrincipalRepository';
import { ApiInventorySummaryRepository } from '../adapters/inventories/ApiInventorySummaryRepository';
import { ExpoPhotoSelectionProvider } from '../adapters/photos/ExpoPhotoSelectionProvider';
import { ExpoSettingsDiagnosticsProvider } from '../adapters/settings/ExpoSettingsDiagnosticsProvider';
import { DevelopmentVoiceAudioRecorder } from '../adapters/voice/DevelopmentVoiceAudioRecorder';
import { NoopVoiceAudioPlayer } from '../adapters/voice/NoopVoiceAudioPlayer';
import { WebSocketRealtimeVoiceTransport } from '../adapters/voice/WebSocketRealtimeVoiceTransport';
import { InMemoryAddAssetDraftStore } from '../application/add/AddAssetDraftStore';
import { CreateAssetCommand } from '../application/add/CreateAssetCommand';
import { AddDraftScopeQuery } from '../application/add/AddDraftScopeQuery';
import { ParentLookupQuery } from '../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../application/add/PhotoSelectionQuery';
import { AssetDetailQuery } from '../application/assets/AssetDetailQuery';
import { AssetLifecycleCommand } from '../application/assets/AssetLifecycleCommand';
import { InventoryAssetsQuery } from '../application/assets/InventoryAssetsQuery';
import { HomeDashboardQuery } from '../application/home/HomeDashboardQuery';
import {
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../application/home/InventorySummaryRepository';
import { SelectInventoryCommand } from '../application/home/SelectInventoryCommand';
import { LocationAssetsQuery } from '../application/locations/LocationAssetsQuery';
import { LocationsQuery } from '../application/locations/LocationsQuery';
import { SearchAssetsQuery } from '../application/search/SearchAssetsQuery';
import {
  CurrentPrincipalRepository,
  SettingsPrincipal,
  SettingsQuery
} from '../application/settings/SettingsQuery';
import { VoiceInteractionPreviewQuery } from '../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController } from '../application/voice/RealtimeVoiceSession';
import { loadMobileRuntimeConfig } from '../config/mobileRuntimeConfig';
import type { MobileRuntimeConfig } from '../config/mobileRuntimeConfigCore';
import { AssetSummary } from '../domain/assets/AssetSummary';
import { InventorySummary } from '../domain/inventories/InventorySummary';
import { LocationSummary } from '../domain/locations/LocationSummary';

export type MobileComposition = {
  readonly homeDashboardQuery: HomeDashboardQuery;
  readonly selectInventoryCommand: SelectInventoryCommand;
  readonly searchAssetsQuery: SearchAssetsQuery;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly inventoryAssetsQuery: InventoryAssetsQuery;
  readonly createAssetCommand: CreateAssetCommand;
  readonly addDraftScopeQuery: AddDraftScopeQuery;
  readonly addAssetDraftStore: InMemoryAddAssetDraftStore;
  readonly parentLookupQuery: ParentLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly locationsQuery: LocationsQuery;
  readonly locationAssetsQuery: LocationAssetsQuery;
  readonly settingsQuery: SettingsQuery;
  readonly voiceInteractionPreviewQuery: VoiceInteractionPreviewQuery;
  readonly realtimeVoiceSessionController: RealtimeVoiceSessionController;
};

export function createMobileComposition(): MobileComposition {
  let inventorySummaries: InventorySummaryRepository;
  let principals: CurrentPrincipalRepository;
  let config: MobileRuntimeConfig | undefined;

  try {
    const loadedConfig = loadMobileRuntimeConfig();
    config = loadedConfig;
    const client = new StuffStashClient({
      baseUrl: loadedConfig.apiBaseUrl,
      fetch: createTimeoutFetch(8000),
      tokenProvider: () => loadedConfig.devToken
    });
    inventorySummaries = new ApiInventorySummaryRepository(client, loadedConfig.tenantId);
    principals = new ApiCurrentPrincipalRepository(client);
  } catch (error) {
    const unavailableError = toError(error);
    inventorySummaries = new UnavailableInventorySummaryRepository(unavailableError);
    principals = new UnavailableCurrentPrincipalRepository(unavailableError);
  }
  const addAssetDraftStore = new InMemoryAddAssetDraftStore(createServiceScopeId());

  return {
    homeDashboardQuery: new HomeDashboardQuery(inventorySummaries),
    selectInventoryCommand: new SelectInventoryCommand(inventorySummaries),
    searchAssetsQuery: new SearchAssetsQuery(inventorySummaries),
    assetDetailQuery: new AssetDetailQuery(inventorySummaries),
    assetLifecycleCommand: new AssetLifecycleCommand(inventorySummaries),
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
    voiceInteractionPreviewQuery: new VoiceInteractionPreviewQuery(inventorySummaries),
    realtimeVoiceSessionController: new RealtimeVoiceSessionController(
      inventorySummaries,
      new DevelopmentVoiceAudioRecorder(),
      new WebSocketRealtimeVoiceTransport({
        apiBaseUrl: config?.apiBaseUrl ?? 'http://127.0.0.1:8080',
        tokenProvider: () => config?.devToken ?? ''
      }),
      new NoopVoiceAudioPlayer()
    )
  };
}

class UnavailableInventorySummaryRepository implements InventorySummaryRepository {
  constructor(private readonly error: Error) {}

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    throw this.error;
  }

  async getDefaultInventorySummary(): Promise<InventorySummary> {
    throw this.error;
  }

  async selectInventory(): Promise<void> {
    throw this.error;
  }

  async createAsset(): Promise<AssetSummary> {
    throw this.error;
  }

  async addAssetPhoto(): Promise<void> {
    throw this.error;
  }

  async archiveAsset(): Promise<void> {
    throw this.error;
  }

  async restoreAsset(): Promise<void> {
    throw this.error;
  }

  async deleteAsset(): Promise<void> {
    throw this.error;
  }

  async browseAssets(): Promise<never> {
    throw this.error;
  }

  async searchAssets(): Promise<readonly AssetSummary[]> {
    throw this.error;
  }

  async searchLocations(): Promise<readonly LocationSummary[]> {
    throw this.error;
  }
}

class UnavailableCurrentPrincipalRepository implements CurrentPrincipalRepository {
  constructor(private readonly error: Error) {}

  async getCurrentPrincipal(): Promise<SettingsPrincipal> {
    throw this.error;
  }
}

function toError(error: unknown): Error {
  return error instanceof Error ? error : new Error('Mobile API configuration is unavailable.');
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
