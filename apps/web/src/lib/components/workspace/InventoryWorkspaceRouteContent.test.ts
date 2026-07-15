import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type {
  Asset,
  AssetKind,
  AssetLifecycleFilter,
  LocationAsset,
  SearchLifecycleFilter,
  SearchMode,
  WorkspaceData,
  WorkspaceMode
} from '$lib/domain/inventory';
import type { InventoryWorkspaceRouteContentProps } from './InventoryWorkspaceRouteContent.svelte';
import InventoryWorkspaceRouteContent from './InventoryWorkspaceRouteContent.svelte';

let component: ReturnType<typeof mount> | null = null;

type RouteContentOverrides = Partial<Omit<InventoryWorkspaceRouteContentProps, 'workspace' | 'status' | 'route' | 'hrefs' | 'handlers'>> & {
  workspace?: Partial<InventoryWorkspaceRouteContentProps['workspace']>;
  status?: Partial<InventoryWorkspaceRouteContentProps['status']>;
  route?: Partial<InventoryWorkspaceRouteContentProps['route']>;
  hrefs?: Partial<InventoryWorkspaceRouteContentProps['hrefs']>;
  handlers?: Partial<InventoryWorkspaceRouteContentProps['handlers']>;
  inventoryPermissions?: InventoryPermissionFixture[];
};

type InventoryPermissionFixture = 'view' | 'create_asset' | 'edit_asset' | 'configure' | 'view_import_job' | 'create_import_job';

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventoryWorkspaceRouteContent', () => {
  it('renders unavailable routes as a named alert with a durable home link', async () => {
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { routeUnavailable: 'That inventory is not available in the current workspace.' },
        hrefs: { homeHref: '/tenants/tenant-home/inventories/inventory-household' }
      })
    });

    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain('Workspace unavailable');
    expect(link('Go home').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household');
  });

  it('renders the home route with recent assets ahead of locations and routes add actions', async () => {
    const openedKinds: AssetKind[] = [];
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        handlers: {
          onOpenAdd: (kind = 'item') => {
            openedKinds.push(kind);
          }
        }
      })
    });

    const mainText = document.body.querySelector('.workspace-main')?.textContent ?? '';
    expect(mainText.indexOf('Recent')).toBeGreaterThanOrEqual(0);
    expect(mainText.indexOf('Places')).toBeGreaterThanOrEqual(0);
    expect(mainText.indexOf('Recent')).toBeLessThan(mainText.indexOf('Places'));

    link('Add location').click();
    await tick();

    expect(openedKinds).toEqual(['location']);
  });

  it('routes empty-inventory add item and location actions separately', async () => {
    const openedKinds: AssetKind[] = [];
    const props = await routeContentProps({
      handlers: {
        onOpenAdd: (kind = 'item') => {
          openedKinds.push(kind);
        }
      }
    });
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: {
        ...props,
        workspace: {
          ...props.workspace,
          assets: [],
          detailAssets: []
        }
      }
    });

    expect(link('Add first location').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');
    expect(link('Add item').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');

    link('Add first location').click();
    await tick();
    link('Add item').click();
    await tick();

    expect(openedKinds).toEqual(['location', 'item']);
  });

  it('keeps search state bindable through the extracted route content', async () => {
    let searched = false;
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'browse' },
        searchQuery: 'pass',
        handlers: {
          onSearch: async () => {
            searched = true;
          }
        }
      })
    });

    const search = document.body.querySelector<HTMLInputElement>('input[aria-label="Search Browse"]');
    if (!search) {
      throw new Error('Missing workspace search input');
    }
    search.value = 'passport';
    search.dispatchEvent(new InputEvent('input', { bubbles: true }));
    document.body.querySelector<HTMLFormElement>('form.browse-search')?.dispatchEvent(new SubmitEvent('submit', { bubbles: true, cancelable: true }));
    await tick();

    expect(searched).toBe(true);
  });

  it('threads active inventory tags into search browsing filters', async () => {
    const searchedTags: string[][] = [];
    const props = await routeContentProps({
      route: { mode: 'browse' },
      handlers: {
        onBrowseStateChange: (state) => {
          searchedTags.push(state.selectedTagIds ?? []);
        }
      }
    });

    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: {
        ...props,
        workspace: {
          ...props.workspace,
          data: {
            ...props.workspace.data,
            context: {
              ...props.workspace.data.context,
              assetTags: [
                { id: 'tag-tools', key: 'tools', displayName: 'Tools', color: '#2F80ED' }
              ]
            }
          }
        }
      }
    });

    const filterTrigger = document.body.querySelector<HTMLButtonElement>('.browse-filter-trigger');
    if (!filterTrigger) throw new Error('Missing Browse filter trigger');
    filterTrigger.click();
    await tick();
    await tick();
    const tagFilter = document.body.querySelector<HTMLElement>('.browse-filter-popover');
    expect(tagFilter?.textContent).toContain('Tools');
    Array.from(document.body.querySelectorAll<HTMLButtonElement>('.browse-filter-popover button')).find((button) => button.textContent?.includes('Tools'))?.click();
    Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="dialog"] button')).find((button) => button.textContent?.includes('Apply filters'))?.click();
    await tick();

    expect(searchedTags).toEqual([['tag-tools']]);
  });

  it('renders the no-inventory branch with starter inventory affordances', async () => {
    const props = await routeContentProps();
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: {
        ...props,
        workspace: {
          ...props.workspace,
          selectedInventory: null,
          data: {
            ...props.workspace.data,
            context: {
              ...props.workspace.data.context,
              inventories: [],
              selectedInventoryId: ''
            }
          }
        }
      }
    });

    expect(document.body.textContent).toContain('No inventory yet');
    expect(document.body.textContent).toContain('Create Household');
  });

  it('renders the no-inventory branch without a starter action when creation is unavailable', async () => {
    const props = await routeContentProps({
      status: { canCreateStarter: false }
    });
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: {
        ...props,
        workspace: {
          ...props.workspace,
          selectedInventory: null,
          data: {
            ...props.workspace.data,
            context: {
              ...props.workspace.data.context,
              inventories: [],
              selectedInventoryId: ''
            }
          }
        }
      }
    });

    expect(document.body.textContent).toContain('You can view this tenant, but you cannot create inventories in it.');
    expect(document.body.textContent).not.toContain('Create Household');
  });

  it('renders the location branch with contained assets', async () => {
    const props = await routeContentProps();
    const location = requiredAsset(props.workspace.data, 'location-garage') as LocationAsset;
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'location' },
        workspace: { selectedLocation: location }
      })
    });

    expect(document.body.textContent).toContain('Garage');
    expect(document.body.textContent).toContain('Storage bin');
  });

  it('renders the asset detail branch with action route state', async () => {
    const props = await routeContentProps();
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'asset', assetAction: 'edit' },
        workspace: { selectedAsset: requiredAsset(props.workspace.data, 'asset-passport') }
      })
    });
    await tick();

    expect(document.body.textContent).toContain('Passport');
    expect(document.body.textContent).toContain('Edit asset');
  });

  it('opens a location child from a container through location navigation', async () => {
    const base = await routeContentProps();
    const container = asset('container-one', 'Tool cabinet', 'container');
    const nestedLocation = asset('location-nested', 'Nested place', 'location', container.id);
    const assets = [...base.workspace.data.assets, container, nestedLocation];
    const openedLocations: string[] = [];
    const openedAssets: string[] = [];
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'asset' },
        workspace: {
          data: { ...base.workspace.data, assets },
          assets,
          detailAssets: assets,
          selectedAsset: container
        },
        handlers: {
          onOpenLocation: (candidate) => openedLocations.push(candidate.id),
          onOpenAsset: async (candidate) => { openedAssets.push(candidate.id); }
        }
      })
    });

    link('Nested place').click();
    await tick();

    expect(openedLocations).toEqual(['location-nested']);
    expect(openedAssets).toEqual([]);
  });

  it('renders an accessible busy status while asset details load', async () => {
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({ route: { mode: 'asset', assetDetailLoading: true } })
    });

    const status = document.body.querySelector('[role="status"]');
    expect(status?.textContent).toContain('Loading asset details');
    expect(status?.getAttribute('aria-live')).toBe('polite');
    expect(status?.closest('[aria-busy]')?.getAttribute('aria-busy')).toBe('true');
    expect(document.body.textContent).not.toContain('Recently changed');
  });

  it('renders import and settings branches with their routed sections', async () => {
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'import' }
      })
    });

    expect(document.body.textContent).toContain('Imports');
    expect(document.body.textContent).toContain('No import runs yet');

    unmount(component);
    component = null;
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'settings', settingsSection: 'fields' }
      })
    });

    expect(document.body.textContent).toContain('Fields');
  });

});

async function routeContentProps(overrides: RouteContentOverrides = {}): Promise<InventoryWorkspaceRouteContentProps> {
  const repository = new SeededInventoryRepository({
    principal: { id: 'principal-one', email: 'owner@example.test' },
    tenants: [{ id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } }],
    inventories: [
      {
        id: 'inventory-household',
        tenantId: 'tenant-home',
        name: 'Household',
        access: {
          relationship: 'owner',
          permissions: overrides.inventoryPermissions ?? [
            'view',
            'create_asset',
            'edit_asset',
            'configure',
            'view_import_job',
            'create_import_job'
          ]
        }
      }
    ],
    customAssetTypes: [],
    customFieldDefinitions: [],
    assets: [
      asset('asset-passport', 'Passport', 'item'),
      asset('location-garage', 'Garage', 'location'),
      asset('asset-bin', 'Storage bin', 'container', 'location-garage')
    ]
  });
  const data = await repository.loadWorkspace();
  const selectedTenant = data.context.tenants[0] ?? null;
  const selectedInventory = data.context.inventories[0] ?? null;

  const props: InventoryWorkspaceRouteContentProps = {
    workspace: {
      data,
      repository,
      selectedTenant,
      selectedInventory,
      selectedLocation: null,
      selectedAsset: null,
	      assets: data.assets,
	      detailAssets: data.assets,
	      selectedAssetAttachments: [],
	      selectedAssetCheckoutHistory: []
    },
    status: {
      busy: false,
      canCreateStarter: true,
      createAssetAllowed: true,
      editAssetAllowed: true
    },
    route: {
      routeUnavailable: '',
      assetDetailLoading: false,
      mode: 'home' as WorkspaceMode,
      searchResults: [],
      searchSuggestions: [],
      searchSubmitted: false,
      searchError: '',
      browseSurface: 'list',
      browseScope: 'all',
      browseSort: 'updated_desc',
      browseTagIds: [],
      browseAssets: data.assets,
      browseInventoryEmpty: false,
      browseHasMore: false,
      browseLoadingMore: false,
      browseBusy: false,
      browseErrorPhase: null,
      assetAction: null,
      attachmentId: null,
      attachmentAction: null,
      settingsSection: 'overview',
      invitationStatus: 'all',
      accessInvitationAction: null,
      accessInvitationId: null,
      auditScope: 'inventory',
      customizationAction: null,
      customAssetTypeId: null,
      customFieldDefinitionId: null,
      importSource: null,
      importJobId: null,
      importTab: null
    },
    hrefs: {
      homeHref: '/tenants/tenant-home/inventories/inventory-household',
      assetDetailBackHref: '/tenants/tenant-home/inventories/inventory-household'
    },
    searchQuery: '',
	    searchLifecycleState: 'active' as SearchLifecycleFilter,
	    searchMode: 'fuzzy' as SearchMode,
	    searchCheckoutState: 'any',
    handlers: {
      onHome: () => {},
      onCreateStarterInventory: async () => {},
      onOpenLocation: () => {},
      onBrowseStateChange: () => {},
      onBrowseLoadMore: async () => {},
      onBrowseRetry: async () => {},
      onEditLocation: () => {},
      onOpenAsset: async () => {},
      onOpenAdd: () => {},
      onCloseLocation: () => {},
      onCloseAssetDetail: () => {},
      onAssetActionOpen: () => {},
      onAssetActionClose: () => {},
      onAssetSave: async () => {},
	    onMoveAssetHere: async () => {},
	      onAssetArchive: async () => {},
	      onAssetRestore: async () => {},
	      onAssetDelete: async () => {},
	      onAssetCheckout: async () => {},
	      onAssetReturn: async () => {},
	      onHomeAssetReturn: async () => {},
	      onAssetUploadAttachment: async () => {},
      onAssetArchiveAttachment: async () => {},
      onAttachmentDeleteOpen: () => {},
      onAttachmentDeleteClose: () => {},
      onAssetDeleteAttachment: async () => {},
      onSearch: async () => {},
      onOpenSearchAsset: () => {},
      onImportSourceChange: () => {},
      onImportJobSelectionChange: () => {},
      onImportJobTabChange: () => {},
      onImportJobInventoryChanged: async () => {},
      onOpenImportedAssetId: async () => {},
      onOpenInventoryAuditHistory: () => {},
      onSettingsSectionChange: () => {},
      onInvitationStatusChange: () => {},
      onAccessInvitationActionOpen: () => {},
      onAccessInvitationActionClose: () => {},
      onAuditScopeChange: () => {},
      onCustomizationArchiveOpen: () => {},
      onCustomizationArchiveClose: () => {},
      onCustomizationChange: () => {},
      onSelectLifecycle: async (_lifecycleState: WorkspaceData['context']['assetLifecycleState'] | AssetLifecycleFilter) => {}
    }
  };

  return {
    ...props,
    ...overrides,
    workspace: { ...props.workspace, ...overrides.workspace },
    status: { ...props.status, ...overrides.status },
    route: { ...props.route, ...overrides.route },
    hrefs: { ...props.hrefs, ...overrides.hrefs },
    handlers: { ...props.handlers, ...overrides.handlers }
  };
}

function asset(id: string, title: string, kind: Asset['kind'], parentAssetId: string | null = null): Asset {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind,
    title,
    description: '',
    parentAssetId,
    lifecycleState: 'active'
  };
}

function requiredAsset(data: WorkspaceData, id: string): Asset {
  const found = data.assets.find((candidate) => candidate.id === id);
  if (!found) {
    throw new Error(`Missing asset ${id}`);
  }
  return found;
}

function link(text: string): HTMLAnchorElement {
  const target = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent?.includes(text));
  if (!target) {
    throw new Error(`Missing link ${text}`);
  }
  return target;
}
