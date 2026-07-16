import { describe, expect, it } from 'vitest';
import type { Asset } from '$lib/domain/inventory';
import {
  containableWorkspaceSections,
  containedWorkspaceChildren,
  moveHereCandidatePage,
  moveHereHref
} from './workspaceContainedAssets';

const asset = (
  id: string,
  title: string,
  kind: Asset['kind'],
  parentAssetId: string | null = null,
  lifecycleState: Asset['lifecycleState'] = 'active'
): Asset => ({
  id,
  tenantId: 'tenant-one',
  inventoryId: 'inventory-one',
  title,
  description: '',
  kind,
  parentAssetId,
  lifecycleState
});

describe('container workspace presentation', () => {
  const cabinet = asset('cabinet', 'Garage cabinet', 'container', 'garage');
  const assets = [
    asset('garage', 'Garage', 'location'),
    cabinet,
    asset('drill', 'Drill', 'item', 'cabinet'),
    asset('bin-10', 'Bin 10', 'container', 'cabinet'),
    asset('bin-2', 'bin 2', 'container', 'cabinet'),
    asset('archived', 'Archived', 'item', 'cabinet', 'archived'),
    asset('screws', 'Screws', 'item', 'bin-2'),
    asset('attic', 'Attic', 'location'),
    asset('hammer', 'Hammer', 'item', 'garage'),
    asset('wrench', 'Wrench', 'item'),
    asset('ancestor-bin', 'Ancestor bin', 'container', 'wrench')
  ];

  it('shows immediate active spaces before items with natural stable ordering', () => {
    expect(containedWorkspaceChildren(cabinet, assets).map((child) => child.id)).toEqual([
      'bin-2',
      'bin-10',
      'drill'
    ]);
  });

  it('offers only assets that can move into the target without duplicates or cycles', () => {
    const page = moveHereCandidatePage(cabinet, assets, '', 10);

    expect(page.candidates.map((candidate) => candidate.id)).toEqual([
      'ancestor-bin',
      'attic',
      'hammer',
      'screws',
      'wrench'
    ]);
    expect(page.totalCount).toBe(5);
    expect(page.hasMore).toBe(false);
  });

  it('ranks exact and prefix title matches, remains bounded, and reports overflow', () => {
    const extra = [asset('bin-a', 'Tool bin', 'container'), asset('bin-b', 'Toolbox', 'container')];
    const page = moveHereCandidatePage(cabinet, [...assets, ...extra], 'tool', 1);

    expect(page.candidates.map((candidate) => candidate.title)).toEqual(['Tool bin']);
    expect(page.totalCount).toBe(2);
    expect(page.hasMore).toBe(true);
  });

  it('builds the canonical move-items-here route', () => {
    expect(moveHereHref(cabinet)).toBe(
      '/tenants/tenant-one/inventories/inventory-one/assets/cabinet/move-here'
    );
    expect(moveHereHref(asset('garage-place', 'Garage', 'location'))).toBe(
      '/tenants/tenant-one/inventories/inventory-one/locations/garage-place/move-here'
    );
  });

  it('separates direct location spaces from descendant items with structured relative paths', () => {
    const place = asset('place', 'Garage', 'location');
    const shelf = asset('shelf', 'Shelf / East', 'container', 'place');
    const drawer = asset('drawer', 'Drawer', 'container', 'shelf');
    const directItem = asset('broom', 'Broom', 'item', 'place');
    const nestedItem = asset('screws-nested', 'Screws', 'item', 'drawer');
    const sections = containableWorkspaceSections(place, [place, shelf, drawer, directItem, nestedItem]);

    expect(sections.map((section) => section.key)).toEqual(['spaces', 'items']);
    expect(sections.map((section) => section.countNoun)).toEqual(['space', 'item']);
    expect(sections[0].assets.map((candidate) => candidate.id)).toEqual(['shelf']);
    expect(sections[1].assets.map((candidate) => [candidate.id, candidate.relativePath])).toEqual([
      ['broom', ''],
      ['screws-nested', 'Shelf / East / Drawer']
    ]);
  });

  it('bounds malformed location cycles without duplicating rows', () => {
    const place = asset('place', 'Garage', 'location', 'loop');
    const loop = asset('loop', 'Loop', 'container', 'place');
    const itemInLoop = asset('item-loop', 'Tape', 'item', 'loop');

    const sections = containableWorkspaceSections(place, [place, loop, itemInLoop]);
    expect(sections[0].assets.map((candidate) => candidate.id)).toEqual(['loop']);
    expect(sections[1].assets.map((candidate) => candidate.id)).toEqual(['item-loop']);
  });
});
