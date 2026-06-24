import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiCurrentPrincipalRepository } from '../adapters/identity/ApiCurrentPrincipalRepository';
import { ApiInventorySummaryRepository } from '../adapters/inventories/ApiInventorySummaryRepository';
import { ExpoPhotoSelectionProvider } from '../adapters/photos/ExpoPhotoSelectionProvider';
import { ExpoSettingsDiagnosticsProvider } from '../adapters/settings/ExpoSettingsDiagnosticsProvider';
import { CreateAssetCommand } from '../application/add/CreateAssetCommand';
import { LocationLookupQuery } from '../application/add/LocationLookupQuery';
import { PhotoSelectionQuery } from '../application/add/PhotoSelectionQuery';
import { AssetDetailQuery } from '../application/assets/AssetDetailQuery';
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
  readonly inventoryAssetsQuery: InventoryAssetsQuery;
  readonly createAssetCommand: CreateAssetCommand;
  readonly locationLookupQuery: LocationLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly locationsQuery: LocationsQuery;
  readonly locationAssetsQuery: LocationAssetsQuery;
  readonly settingsQuery: SettingsQuery;
  readonly voiceInteractionPreviewQuery: VoiceInteractionPreviewQuery;
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

  return {
    homeDashboardQuery: new HomeDashboardQuery(inventorySummaries),
    selectInventoryCommand: new SelectInventoryCommand(inventorySummaries),
    searchAssetsQuery: new SearchAssetsQuery(inventorySummaries),
    assetDetailQuery: new AssetDetailQuery(inventorySummaries),
    inventoryAssetsQuery: new InventoryAssetsQuery(inventorySummaries),
    createAssetCommand: new CreateAssetCommand(inventorySummaries),
    locationLookupQuery: new LocationLookupQuery(inventorySummaries),
    photoSelectionQuery: new PhotoSelectionQuery(new ExpoPhotoSelectionProvider()),
    locationsQuery: new LocationsQuery(inventorySummaries),
    locationAssetsQuery: new LocationAssetsQuery(inventorySummaries),
    settingsQuery: new SettingsQuery(
      principals,
      new ExpoSettingsDiagnosticsProvider(config)
    ),
    voiceInteractionPreviewQuery: new VoiceInteractionPreviewQuery(inventorySummaries)
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
