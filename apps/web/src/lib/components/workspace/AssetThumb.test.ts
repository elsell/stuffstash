import { readFileSync } from 'node:fs';
import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { Asset } from '$lib/domain/inventory';
import AssetThumb from './AssetThumb.svelte';
import AssetThumbHarness from './AssetThumbHarness.test.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
});

describe('AssetThumb', () => {
  it.each([
    { size: 'sm' as const, pixels: 36 },
    { size: 'md' as const, pixels: 48 },
    { size: 'lg' as const, pixels: 96 }
  ])('renders and styles the $size thumbnail contract', ({ size, pixels }) => {
    component = mount(AssetThumb, {
      target: document.body,
      props: { asset: asset('asset-one', 'photo-one', 'Tape measure'), size }
    });
    expect(document.body.querySelector('.asset-thumb')?.classList.contains(`asset-thumb-${size}`)).toBe(true);

    const styles = readFileSync('src/styles.css', 'utf8');
    expect(styles).toMatch(new RegExp(`\\.asset-thumb-${size}\\s*\\{[^}]*width:\\s*${pixels}px;[^}]*height:\\s*${pixels}px;`));
  });

  it('hydrates only its own declared primary photo through the injected loader', async () => {
    const asset: Asset = {
      id: 'asset-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      kind: 'item',
      title: 'Socket set',
      description: '',
      parentAssetId: null,
      lifecycleState: 'active',
      primaryPhotoId: 'photo-one'
    };
    const requested: string[] = [];
    component = mount(AssetThumbHarness, {
      target: document.body,
      props: {
        asset,
        loader: {
          async loadAssetThumbnail(candidate) {
            requested.push(candidate.id);
            return { id: 'photo-one', assetId: candidate.id, url: 'blob:photo-one', alt: candidate.title };
          }
        }
      }
    });

    await tick();
    await tick();

    expect(requested).toEqual(['asset-one']);
    expect(document.body.querySelector('img')).toMatchObject({ src: 'blob:photo-one', alt: 'Socket set' });
  });

  it('restarts photo loading when a reused component receives a different asset', async () => {
    const requested: string[] = [];
    const first = asset('asset-one', 'photo-one', 'Socket set');
    component = mount(AssetThumbHarness, {
      target: document.body,
      props: {
        asset: first,
        loader: {
          async loadAssetThumbnail(candidate) {
            requested.push(candidate.id);
            return { id: candidate.primaryPhotoId!, assetId: candidate.id, url: `blob:${candidate.id}`, alt: candidate.title };
          }
        }
      }
    });
    await tick();
    await tick();

    (component as unknown as { replaceAsset(next: Asset): void }).replaceAsset(asset('asset-two', 'photo-two', 'Tent'));
    await tick();
    await tick();

    expect(requested).toEqual(['asset-one', 'asset-two']);
    expect(document.body.querySelector('img')).toMatchObject({ src: 'blob:asset-two', alt: 'Tent' });
    expect(document.body.textContent).not.toContain('Photo unavailable');
  });

  it('keeps ordinary fallback assets synchronous and does not retain a prior load failure', async () => {
    const requested: string[] = [];
    component = mount(AssetThumbHarness, {
      target: document.body,
      props: {
        asset: asset('asset-one', 'photo-one', 'Socket set'),
        loader: {
          async loadAssetThumbnail(candidate) {
            requested.push(candidate.id);
            return null;
          }
        }
      }
    });
    await tick();
    await tick();

    expect(document.body.textContent).toContain('Photo unavailable');
    (component as unknown as { replaceAsset(next: Asset): void }).replaceAsset({
      ...asset('asset-two', 'unused', 'Tent'),
      primaryPhotoId: undefined
    });
    await tick();

    expect(requested).toEqual(['asset-one']);
    expect(document.body.querySelector('img')).toBeNull();
    expect(document.body.textContent).not.toContain('Photo unavailable');
  });

  it.each([
    ['unavailable', async () => null],
    ['rejected', async () => Promise.reject(new Error('transport failed'))]
  ])('shows an accessible unavailable state when loading is %s', async (_state, loadAssetThumbnail) => {
    component = mount(AssetThumbHarness, {
      target: document.body,
      props: { asset: asset('asset-one', 'photo-one', 'Socket set'), loader: { loadAssetThumbnail } }
    });

    await tick();
    await tick();
    await Promise.resolve();
    await new Promise((resolve) => setTimeout(resolve, 0));
    await tick();

    expect(document.body.querySelector('img')).toBeNull();
    expect(document.body.textContent).toContain('Photo unavailable');
  });
});

function asset(id: string, primaryPhotoId: string, title: string): Asset {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'item',
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    primaryPhotoId
  };
}

it('connects the rendered thumbnail lifecycle to its session observer', async () => {
 const records: {context: unknown; outcome?: string}[] = [];
 component = mount(AssetThumbHarness, {target: document.body, props: {
  asset: asset('asset-one', 'photo-one', 'Socket set'),
  loader: {async loadAssetThumbnail(candidate) {return {id: 'photo-one', assetId: candidate.id, url: 'blob:private-photo', alt: candidate.title, variant: 'small'};}},
  observer: {start(context) {const record: {context: unknown; outcome?: string} = {context}; records.push(record); return outcome => {record.outcome = outcome;};}}
 }});
 await tick(); await tick();
 const image = document.body.querySelector('img')!;
 image.dispatchEvent(new Event('load'));
 expect(records).toEqual([{context: {operation: 'image', surface: 'list', variant: 'small'}, outcome: 'success'}]);
});

it('cancels an in-flight source when Svelte replaces the source on a reused image node', async () => {
 const outcomes: string[][] = [];
 const first = {...asset('asset-one', 'photo-one', 'Socket set'), photo: {id: 'photo-one', assetId: 'asset-one', url: 'blob:first', alt: 'Socket set', variant: 'small' as const}};
 component = mount(AssetThumbHarness, {target: document.body, props: {
  asset: first,
  loader: {async loadAssetThumbnail() {return null;}},
  observer: {start() {const values: string[] = []; outcomes.push(values); return outcome => values.push(outcome);}}
 }});
 await tick();
 const image = document.body.querySelector('img')!;
 (component as unknown as {replaceAsset(next: Asset): void}).replaceAsset({...first, photo: {...first.photo, url: 'blob:second'}});
 await tick();
 expect(document.body.querySelector('img')).toBe(image);
 expect(image.src).toBe('blob:second');
 image.dispatchEvent(new Event('load'));
 await unmount(component); component = null;
 expect(outcomes).toEqual([['cancelled'], ['success']]);
});
