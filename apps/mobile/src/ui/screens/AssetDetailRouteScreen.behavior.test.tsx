import React from 'react';
import { describe, expect, it } from 'vitest';
import { AssetDetailRouteScreen } from './AssetDetailRouteScreen';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { AssetCoreQuery, type AssetCoreSnapshot } from '../../application/assets/AssetCoreQuery';
import { AssetContentsQuery, type AssetContentsSnapshot } from '../../application/assets/AssetContentsQuery';
import { AssetPhotosQuery } from '../../application/assets/AssetPhotosQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { assetId, type AssetPhoto } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((finish) => { resolve = finish; });
  return { promise, resolve };
}
const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));

function snapshot(parent = 'garage'): AssetCoreSnapshot {
  return {
    tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['view', 'edit_asset'], revision: 'revision',
    asset: {
      id: assetId('tent'), title: 'Family tent', kind: 'item', lifecycleState: 'active', parentAssetId: assetId(parent),
      description: '', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: 'Updated now', hasPhoto: false
    }
  };
}

function setup() {
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  let core = snapshot();
  let coreRequests = 0;
  let photoRequests = 0;
  const contents = deferred<AssetContentsSnapshot>();
  const photos = deferred<readonly AssetPhoto[]>();
  const props: React.ComponentProps<typeof AssetDetailRouteScreen> = {
    assetId: 'tent',
    assetCoreQuery: new AssetCoreQuery({ getAssetCore: async () => { coreRequests++; return core; } }),
    assetContentsQuery: new AssetContentsQuery({ getAssetContents: () => contents.promise }),
    assetPhotosQuery: new AssetPhotosQuery({ getAssetPhotos: () => { photoRequests++; return photos.promise; } }),
    assetCheckoutCommand: { execute: async () => { throw new Error('No checkout configured'); } },
    assetLifecycleCommand: { execute: async () => undefined },
    undoAssetEditCommand: { execute: async () => undefined },
    deleteAssetPhotoCommand: { execute: async () => ({ message: 'Removed' }) },
    addAssetPhotosCommand: { execute: async () => ({ attachedCount: 0, failedCount: 0, failedPhotos: [], message: '', canRetry: false }) },
    photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })
  };
  const render = () => harness.render(
    <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider><AssetDetailRouteScreen {...props} /></AppFeedbackProvider>
    </MobileServerStateProvider>
  );
  return { client, harness, contents, photos, render, core: () => core, move: () => { core = snapshot('attic'); }, counts: () => ({ coreRequests, photoRequests }) };
}

describe('progressive asset detail route', () => {
  it('renders core and independent photo actions before delayed contents', async () => {
    const test = setup();
    try {
      await test.render();
      await settle(test.harness);
      await settle(test.harness);
      expect(test.harness.allText()).toContain('Family tent');
      expect(test.harness.byLabel('Loading photos')).toBeDefined();
      expect(test.harness.byLabel('Loading location and contents')).toBeDefined();
      await test.harness.run(() => test.photos.resolve([]));
      await settle(test.harness);
      expect(test.harness.byLabel('Loading photos')).toBeUndefined();
      expect(test.harness.byLabel('Add photos')).toBeDefined();
      expect(test.harness.byLabel('Loading location and contents')).toBeDefined();
      await test.harness.run(() => test.contents.resolve({ asset: test.core().asset, allAssets: [] }));
      await settle(test.harness);
      expect(test.harness.byLabel('Loading location and contents')).toBeUndefined();
      expect(test.counts()).toEqual({ coreRequests: 1, photoRequests: 1 });
    } finally { await test.harness.unmount(); }
  });

  it('refreshes photos even when core refresh discovers a new parent', async () => {
    const test = setup();
    try {
      test.contents.resolve({ asset: test.core().asset, allAssets: [] });
      test.photos.resolve([]);
      await test.render();
      await settle(test.harness);
      await settle(test.harness);
      test.move();
      await test.harness.run(() => test.harness.byType('RefreshControl')?.props.onRefresh());
      await settle(test.harness);
      expect(test.counts()).toEqual({ coreRequests: 2, photoRequests: 2 });
      expect(test.client.getQueryData(mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'tent')))
        .toMatchObject({ snapshot: { asset: { parentAssetId: 'attic' } } });
    } finally { await test.harness.unmount(); }
  });
});
