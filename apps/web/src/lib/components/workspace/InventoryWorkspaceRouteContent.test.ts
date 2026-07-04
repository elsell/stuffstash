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
};

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
    expect(mainText.indexOf('Locations')).toBeGreaterThanOrEqual(0);
    expect(mainText.indexOf('Recent')).toBeLessThan(mainText.indexOf('Locations'));

    link('Add location').click();
    await tick();

    expect(openedKinds).toEqual(['location']);
  });

  it('keeps search state bindable through the extracted route content', async () => {
    let searched = false;
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'search' },
        searchQuery: 'pass',
        handlers: {
          onSearch: async () => {
            searched = true;
          }
        }
      })
    });

    const search = document.body.querySelector<HTMLInputElement>('#search-page-query');
    if (!search) {
      throw new Error('Missing workspace search input');
    }
    search.value = 'passport';
    search.dispatchEvent(new InputEvent('input', { bubbles: true }));
    document.body.querySelector<HTMLFormElement>('form.search-panel')?.dispatchEvent(new SubmitEvent('submit', { bubbles: true, cancelable: true }));
    await tick();

    expect(searched).toBe(true);
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

  it('renders the top-level locations route without the recent rail', async () => {
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'locations' }
      })
    });

    expect(document.body.querySelector('#home-title')?.textContent).toBe('Locations');
    expect(document.body.textContent).toContain('The places where your things live.');
    expect(document.body.textContent).not.toContain('Recently added');
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

  it('renders import and settings branches with their routed sections', async () => {
    component = mount(InventoryWorkspaceRouteContent, {
      target: document.body,
      props: await routeContentProps({
        route: { mode: 'import', importSourceType: 'legacy_homebox' }
      })
    });

    expect(document.body.textContent).toContain('Homebox');

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
        access: { relationship: 'owner', permissions: ['view', 'create_asset', 'edit_asset'] }
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
      selectedAssetAttachments: []
    },
    status: {
      busy: false,
      canCreateStarter: true,
      createAssetAllowed: true,
      editAssetAllowed: true
    },
    route: {
      routeUnavailable: '',
      mode: 'home' as WorkspaceMode,
      searchResults: [],
      searchSuggestions: [],
      searchSubmitted: false,
      searchError: '',
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
      importSourceType: 'legacy_homebox'
    },
    hrefs: {
      homeHref: '/tenants/tenant-home/inventories/inventory-household',
      assetDetailBackHref: '/tenants/tenant-home/inventories/inventory-household'
    },
    searchQuery: '',
    searchLifecycleState: 'active' as SearchLifecycleFilter,
    searchMode: 'fuzzy' as SearchMode,
    handlers: {
      onHome: () => {},
      onCreateStarterInventory: async () => {},
      onOpenLocation: () => {},
      onEditLocation: () => {},
      onOpenAsset: async () => {},
      onOpenAdd: () => {},
      onCloseLocation: () => {},
      onCloseAssetDetail: () => {},
      onAssetActionOpen: () => {},
      onAssetActionClose: () => {},
      onAssetSave: async () => {},
      onAssetArchive: async () => {},
      onAssetRestore: async () => {},
      onAssetDelete: async () => {},
      onAssetUploadAttachment: async () => {},
      onAssetArchiveAttachment: async () => {},
      onAttachmentDeleteOpen: () => {},
      onAttachmentDeleteClose: () => {},
      onAssetDeleteAttachment: async () => {},
      onSearch: async () => {},
      onOpenSearchAsset: () => {},
      onImportSourceChange: () => {},
      onImported: async () => {},
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
