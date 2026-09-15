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
import { QueryClientInventoryMutationObserver } from '../../adapters/serverState/QueryClientInventoryMutationObserver';
import { latestAlert, latestActionSheetCallback } from '../../test-support/react-native';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((finish, fail) => { resolve = finish; reject = fail; });
  return { promise, resolve, reject };
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

function setup(overrides: Partial<React.ComponentProps<typeof AssetDetailRouteScreen>> = {}) {
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
    assetPhotosQuery: new AssetPhotosQuery({ getAssetPhotos: () => { photoRequests++; return photos.promise.then(items => [...items]); } }),
    assetCheckoutCommand: { execute: async () => { throw new Error('No checkout configured'); } },
    assetLifecycleCommand: { execute: async () => undefined },
    undoAssetEditCommand: { execute: async () => undefined },
    deleteAssetPhotoCommand: { execute: async () => ({ message: 'Removed' }) },
    addAssetPhotosCommand: { execute: async () => ({ attachedCount: 0, failedCount: 0, failedPhotos: [], message: '', canRetry: false }) },
    photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] }),
    ...overrides
  };
  let routeVisible = true;
  let routeAssetId = 'tent';
  const render = () => harness.render(
    <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider>{routeVisible ? <AssetDetailRouteScreen {...props} assetId={routeAssetId} /> : null}</AppFeedbackProvider>
    </MobileServerStateProvider>
  );
  return { client, harness, contents, photos, render, hide: () => { routeVisible = false; }, changeAsset: (id: string) => {
    routeAssetId = id;
    core = { ...snapshot(), asset: { ...snapshot().asset, id: assetId(id), title: 'Other item' } };
  }, core: () => core, move: () => { core = snapshot('attic'); }, counts: () => ({ coreRequests, photoRequests }) };
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


describe('pending photo removal', () => {
  it('rejects duplicate confirmation and preserves a different photo selected during removal', async () => {
    const removal = deferred<{ message: string }>();
    const calls: string[] = [];
    const test = setup({ deleteAssetPhotoCommand: { execute: async input => {
      calls.push(input.photoId);
      const result = await removal.promise;
      new QueryClientInventoryMutationObserver(test.client, 'scope').onInventoryMutation({
        kind: 'asset_photo_changed', tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), assetId: assetId('tent')
      });
      return result;
    } } });
    try {
      const remainingPhotos = [{ id: 'first', uri: 'https://example.invalid/first' }, { id: 'second', uri: 'https://example.invalid/second' }];
      test.photos.resolve(remainingPhotos);
      test.contents.resolve({ asset: test.core().asset, allAssets: [] });
      await test.render(); await settle(test.harness); await settle(test.harness);
      await test.harness.press(test.harness.byLabel('Open photo 1 of 2'));
      await test.harness.press(test.harness.byLabel('Remove photo'));
      const confirm = latestAlert()?.buttons.find(button => button.text === 'Remove')?.onPress;
      expect(confirm).toBeDefined();
      await test.harness.run(() => { confirm?.(); confirm?.(); });
      expect(calls).toEqual(['first']);
      expect(test.harness.byLabel('Remove photo')?.props.disabled).toBe(true);
      expect(test.harness.allText()).toContain('Removing photo…');
      await test.harness.press(test.harness.byLabel('Next photo'));
      remainingPhotos.shift();
      await test.harness.run(() => removal.resolve({ message: 'Removed' }));
      await settle(test.harness);
      expect(test.harness.byType('ImageViewing')?.props.visible).toBe(true);
      expect(test.harness.byType('ImageViewing')?.props.imageIndex).toBe(0);
      expect(test.harness.byType('ImageViewing')?.props.images).toHaveLength(1);
      expect(test.harness.byLabel('Remove photo')?.props.disabled).toBe(false);
    } finally { await test.harness.unmount(); }
  });
});


it.each([false, true])('handles photo-removal failure while mounted or after route teardown', async leaveRoute => {
  const firstRemoval = deferred<{ message: string }>(); let calls = 0;
  const test = setup({ deleteAssetPhotoCommand: { execute: async () => {
    calls++; return calls === 1 ? firstRemoval.promise : { message: 'Removed' };
  } } });
  try {
    test.photos.resolve([{ id: 'first', uri: 'https://example.invalid/first' }]);
    test.contents.resolve({ asset: test.core().asset, allAssets: [] });
    await test.render(); await settle(test.harness); await settle(test.harness);
    await test.harness.press(test.harness.byLabel('Open photo 1 of 1'));
    await test.harness.press(test.harness.byLabel('Remove photo'));
    await test.harness.run(() => { latestAlert()?.buttons.find(button => button.text === 'Remove')?.onPress?.(); });
    if (leaveRoute) { test.hide(); await test.render(); }
    await test.harness.run(() => firstRemoval.reject(new Error('Connection failed')));
    await settle(test.harness);
    if (leaveRoute) {
      expect(latestAlert()?.title).not.toBe('Could not remove photo');
    } else {
      expect(latestAlert()?.title).toBe('Could not remove photo');
      expect(latestAlert()?.message).toBe('Connection failed');
      expect(latestAlert()?.buttons.map(button => button.text)).toEqual(['OK']);
      await test.harness.run(() => { latestAlert()?.buttons.find(button => button.text === 'OK')?.onPress?.(); });
      expect(calls).toBe(1);
      expect(test.harness.byType('ImageViewing')?.props.visible).toBe(true);
      expect(test.harness.byLabel('Remove photo')?.props.disabled).toBe(false);
      await test.harness.press(test.harness.byLabel('Remove photo'));
      await test.harness.run(() => { latestAlert()?.buttons.find(button => button.text === 'Remove')?.onPress?.(); });
      await settle(test.harness);
      expect(calls).toBe(2);
      expect(test.harness.byType('ImageViewing')).toBeUndefined();
    }
  } finally { await test.harness.unmount(); }
});


it('does not let a previous asset removal settle the new asset operation', async () => {
  const oldRemoval = deferred<{ message: string }>(); const newRemoval = deferred<{ message: string }>();
  const calls: string[] = [];
  const test = setup({ deleteAssetPhotoCommand: { execute: input => {
    calls.push(input.assetId); return input.assetId === 'tent' ? oldRemoval.promise : newRemoval.promise;
  } } });
  try {
    test.photos.resolve([{ id: 'first', uri: 'https://example.invalid/first' }]);
    test.contents.resolve({ asset: test.core().asset, allAssets: [] });
    await test.render(); await settle(test.harness); await settle(test.harness);
    await test.harness.press(test.harness.byLabel('Open photo 1 of 1'));
    await test.harness.press(test.harness.byLabel('Remove photo'));
    const staleConfirmation = latestAlert()?.buttons.find(button => button.text === 'Remove')?.onPress;
    await test.harness.run(() => { staleConfirmation?.(); });
    test.changeAsset('other'); await test.render(); await settle(test.harness); await settle(test.harness);
    await test.harness.run(() => { staleConfirmation?.(); });
    expect(calls).toEqual(['tent']);
    await test.harness.press(test.harness.byLabel('Open photo 1 of 1'));
    await test.harness.press(test.harness.byLabel('Remove photo'));
    await test.harness.run(() => { latestAlert()?.buttons.find(button => button.text === 'Remove')?.onPress?.(); });
    expect(calls).toEqual(['tent', 'other']);
    await test.harness.run(() => oldRemoval.resolve({ message: 'Old photo removed' }));
    await settle(test.harness);
    expect(test.harness.byType('ImageViewing')?.props.visible).toBe(true);
    expect(test.harness.byLabel('Remove photo')?.props.disabled).toBe(true);
    expect(test.harness.allText()).not.toContain('Old photo removed');
    await test.harness.run(() => newRemoval.resolve({ message: 'New photo removed' }));
    await settle(test.harness);
    expect(test.harness.byType('ImageViewing')).toBeUndefined();
  } finally { await test.harness.unmount(); }
});


const selectedPhoto = {
  id: 'selection', uri: 'file:///photo.jpg', fileName: 'photo.jpg',
  contentType: 'image/jpeg' as const, contentBase64: 'ZmFrZQ==', sizeBytes: 4
};

it.each(['asset change', 'route teardown'])('discards pending selection after %s', async destination => {
  const selection = deferred<readonly typeof selectedPhoto[]>();
  const uploads: string[] = [];
  const test = setup({
    photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: () => selection.promise, captureFromCamera: async () => [] }),
    addAssetPhotosCommand: { execute: async input => {
      uploads.push(input.assetId);
      return { attachedCount: 1, failedCount: 0, failedPhotos: [], message: 'Uploaded', canRetry: false };
    } }
  });
  try {
    test.photos.resolve([]); test.contents.resolve({ asset: test.core().asset, allAssets: [] });
    await test.render(); await settle(test.harness); await settle(test.harness);
    await test.harness.press(test.harness.byLabel('Add photos'));
    await test.harness.run(() => latestActionSheetCallback()?.(1));
    if (destination === 'asset change') test.changeAsset('other'); else test.hide();
    await test.render();
    await test.harness.run(() => selection.resolve([selectedPhoto]));
    await settle(test.harness);
    expect(uploads).toEqual([]);
  } finally { await test.harness.unmount(); }
});

it('rejects duplicate source callbacks and hides old upload failures after changing asset', async () => {
  const upload = deferred<never>(); let calls = 0;
  const test = setup({
    photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: async () => [selectedPhoto], captureFromCamera: async () => [] }),
    addAssetPhotosCommand: { execute: () => { calls++; return upload.promise; } }
  });
  try {
    test.photos.resolve([]); test.contents.resolve({ asset: test.core().asset, allAssets: [] });
    await test.render(); await settle(test.harness); await settle(test.harness);
    await test.harness.press(test.harness.byLabel('Add photos'));
    const choose = latestActionSheetCallback();
    await test.harness.run(() => { choose?.(1); choose?.(1); });
    expect(calls).toBe(1);
    test.changeAsset('other'); await test.render(); await settle(test.harness);
    await test.harness.run(() => upload.reject(new Error('Old upload failed')));
    await settle(test.harness);
    expect(test.harness.allText()).not.toContain('Old upload failed');
    expect(test.harness.byLabel('Add photos')?.props.disabled).not.toBe(true);
  } finally { await test.harness.unmount(); }
});


it('keeps the new asset upload pending when the previous upload finishes', async () => {
  type UploadResult = Awaited<ReturnType<React.ComponentProps<typeof AssetDetailRouteScreen>['addAssetPhotosCommand']['execute']>>;
  const oldUpload = deferred<UploadResult>(); const newUpload = deferred<UploadResult>();
  const calls: string[] = [];
  const test = setup({
    photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: async () => [selectedPhoto], captureFromCamera: async () => [] }),
    addAssetPhotosCommand: { execute: async input => {
      calls.push(input.assetId);
      const result = await (input.assetId === 'tent' ? oldUpload.promise : newUpload.promise);
      input.onPhotoProgress?.({ index: 0, fileName: 'old-name.jpg', status: 'attached' });
      return result;
    } }
  });
  try {
    test.photos.resolve([]); test.contents.resolve({ asset: test.core().asset, allAssets: [] });
    await test.render(); await settle(test.harness); await settle(test.harness);
    await test.harness.press(test.harness.byLabel('Add photos'));
    await test.harness.run(() => latestActionSheetCallback()?.(1));
    test.changeAsset('other'); await test.render(); await settle(test.harness); await settle(test.harness);
    await test.harness.press(test.harness.byLabel('Add photos'));
    await test.harness.run(() => latestActionSheetCallback()?.(1));
    expect(calls).toEqual(['tent', 'other']);
    await test.harness.run(() => oldUpload.resolve({ attachedCount: 1, failedCount: 0, failedPhotos: [], message: 'Old upload complete', canRetry: false }));
    await settle(test.harness);
    expect(test.harness.allText()).not.toContain('Old upload complete');
    expect(test.harness.allText()).not.toContain('old-name.jpg');
    expect(test.harness.allText()).toContain('Updating photos...');
    await test.harness.run(() => newUpload.resolve({ attachedCount: 1, failedCount: 0, failedPhotos: [], message: 'New upload complete', canRetry: false }));
    await settle(test.harness);
    expect(test.harness.allText()).toContain('New upload complete');
    expect(test.harness.byLabel('Add photos')?.props.disabled).not.toBe(true);
  } finally { await test.harness.unmount(); }
});
