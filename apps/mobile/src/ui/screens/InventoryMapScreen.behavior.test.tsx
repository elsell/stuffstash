import React from 'react';
import { describe, expect, it } from 'vitest';
import { InventoryMapScreen } from './InventoryMapScreen';
import { InventoryMapInfoSheet } from './InventoryMapInfoSheet';
import { dispatchedActions, resetNavigation } from '../../test-support/navigation';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { InventoryMapQuery } from '../../application/assets/InventoryMapQuery';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { AssetContentsQuery } from '../../application/assets/AssetContentsQuery';
import { AssetPhotosQuery } from '../../application/assets/AssetPhotosQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { assetId, type AssetPhoto, type AssetSummary } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
const selectedAsset: AssetSummary = { id: assetId('tent'), title: 'Tent', kind: 'item', lifecycleState: 'active', description: '', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
const mapSnapshot = { sessionScopeId: 'scope', tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), inventoryName: 'Home', permissions: ['view', 'edit_asset'], assets: [selectedAsset] };
const commands = {
  assetCheckoutCommand: { execute: async () => { throw new Error('No checkout requested'); } },
  assetLifecycleCommand: { execute: async () => undefined },
  addAssetPhotosCommand: { execute: async () => ({ attachedCount: 0, failedCount: 0, failedPhotos: [], canRetry: false, message: '' }) },
  deleteAssetPhotoCommand: { execute: async () => ({ message: '' }) },
  photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] })
};

describe('Map server state', () => {
  it('performs one initial traversal and reuses fresh data on remount', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    let calls = 0;
    let inventory = 'inventory';
    const detailRequests: string[] = [];
    const query = new InventoryMapQuery({ listActiveInventoryMapAssets: async () => {
      calls++;
      return inventory === 'inventory' ? mapSnapshot : new Promise(() => undefined);
    } });
    const props: React.ComponentProps<typeof InventoryMapScreen> = {
      ...commands, inventoryMapQuery: query,
      assetCoreQuery: { execute: async (id) => { detailRequests.push(`${inventory}:${id}`); throw new Error('Details unavailable'); } },
      assetContentsQuery: { execute: async () => { throw new Error('No detail selected'); } },
      assetPhotosQuery: { execute: async () => [] },
      canAdd: false, pathStore: { current: new Map() }, selectedSurface: 'map', onAdd: () => undefined, onChangeSurface: () => undefined
    };
    const render = (visible: boolean) => harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider>{visible ? <InventoryMapScreen {...props} /> : null}</AppFeedbackProvider>
    </MobileServerStateProvider>);
    try {
      await render(true); await settle(harness); await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(calls).toBe(1);
      await render(false); await render(true); await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(calls).toBe(1);
      await harness.press(harness.byLabel('Show details for Tent'));
      await settle(harness);
      expect(detailRequests).toEqual(['inventory:tent']);
      await harness.run(() => {
        inventory = 'destination';
        client.setQueryData(mobileQueryKeys.inventoryScope('scope'), { tenantId: 'tenant', inventoryId: inventory });
      });
      await settle(harness); await settle(harness);
      expect(detailRequests).toEqual(['inventory:tent']);
      expect(harness.byType('Modal')?.props.visible).toBe(false);
    } finally { await harness.unmount(); }
  });

  it('shares warm detail core and cancels secondary work when the overlay closes', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    let coreRequests = 0;
    let photoAborted = false;
    const coreQuery = new AssetCoreQuery({ getAssetCore: async () => {
      coreRequests++;
      return { ...mapSnapshot, revision: 'revision', asset: selectedAsset };
    } });
    client.setQueryData(mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'tent'), await coreQuery.execute('tent'));
    const map = await new InventoryMapQuery({ listActiveInventoryMapAssets: async () => mapSnapshot }).execute();
    const props: React.ComponentProps<typeof InventoryMapInfoSheet> = {
      ...commands, asset: map.assets[0], assetCoreQuery: coreQuery,
      assetContentsQuery: new AssetContentsQuery({ getAssetContents: async (core) => ({ asset: core.asset, allAssets: [] }) }),
      assetPhotosQuery: new AssetPhotosQuery({ getAssetPhotos: (_core, request) => new Promise<readonly AssetPhoto[]>((_resolve, reject) => {
        request?.signal?.addEventListener('abort', () => { photoAborted = true; reject(new Error('aborted')); });
      }) }),
      onClose: () => undefined, onMapChanged: () => undefined
    };
    const render = (visible: boolean) => harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider><InventoryMapInfoSheet {...props} asset={visible ? props.asset : undefined} /></AppFeedbackProvider>
    </MobileServerStateProvider>);
    try {
      await render(true); await settle(harness); await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(harness.byLabel('Loading photos')).toBeDefined();
      expect(coreRequests).toBe(1);
      await render(false);
      expect(photoAborted).toBe(true);
    } finally { await harness.unmount(); }
  });
  it('reopens the selected root after navigating inside and closing its overlay', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const root: AssetSummary = { ...selectedAsset, id: assetId('root'), title: 'Root', kind: 'container' };
    const child: AssetSummary = { ...selectedAsset, id: assetId('child'), title: 'Child', parentAssetId: root.id };
    const map = await new InventoryMapQuery({ listActiveInventoryMapAssets: async () => ({ ...mapSnapshot, assets: [root, child] }) }).execute();
    const props: React.ComponentProps<typeof InventoryMapInfoSheet> = {
      ...commands, asset: map.assets.find((asset) => asset.id === 'root'),
      assetCoreQuery: new AssetCoreQuery({ getAssetCore: async (id) => ({ ...mapSnapshot, revision: 'revision', asset: id === root.id ? root : child }) }),
      assetContentsQuery: new AssetContentsQuery({ getAssetContents: async (core) => ({ asset: core.asset, allAssets: [root, child] }) }),
      assetPhotosQuery: new AssetPhotosQuery({ getAssetPhotos: async () => [] }),
      onClose: () => undefined, onMapChanged: () => undefined
    };
    const render = (visible: boolean) => harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider><InventoryMapInfoSheet {...props} asset={visible ? props.asset : undefined} /></AppFeedbackProvider>
    </MobileServerStateProvider>);
    try {
      resetNavigation();
      await render(true); await settle(harness); await settle(harness);
      await harness.press(harness.all().find((node) => typeof node.props.accessibilityLabel === 'string' && node.props.accessibilityLabel.startsWith('Open asset Child')));
      await settle(harness); await settle(harness);
      await render(false); await render(true); await settle(harness);
      await harness.press(harness.byLabel('Edit'));
      expect(dispatchedActions().at(-1)).toMatchObject({ type: 'push', href: '/assets/root/edit' });
    } finally { await harness.unmount(); }
  });

});
