import type {
  AssetBrowseCheckoutFilter,
  AssetBrowseLifecycleFilter,
  AssetBrowseSort
} from '../../application/home/InventorySummaryRepository';
import type { BrowseScope } from './SearchScreenPresentation';

type RouteParamValue = string | readonly string[] | undefined;

export type BrowseRouteParams = {
  readonly scope?: RouteParamValue;
  readonly query?: RouteParamValue;
  readonly tagId?: RouteParamValue;
  readonly lifecycleState?: RouteParamValue;
  readonly checkoutState?: RouteParamValue;
  readonly sort?: RouteParamValue;
};

export type InitialBrowseState = {
  readonly initialScope: BrowseScope;
  readonly initialQuery: string;
  readonly initialTagIds: readonly string[];
  readonly initialLifecycleState: AssetBrowseLifecycleFilter;
  readonly initialCheckoutState: AssetBrowseCheckoutFilter;
  readonly initialSort: AssetBrowseSort;
};

export function parseBrowseRouteParams(params: BrowseRouteParams): InitialBrowseState {
  return {
    initialScope: parseScope(firstValue(params.scope)),
    initialQuery: firstValue(params.query) ?? '',
    initialTagIds: parseTagIds(params.tagId),
    initialLifecycleState: parseLifecycleState(firstValue(params.lifecycleState)),
    initialCheckoutState: parseCheckoutState(firstValue(params.checkoutState)),
    initialSort: parseSort(firstValue(params.sort))
  };
}

function firstValue(value: RouteParamValue): string | undefined {
  return typeof value === 'string' ? value : value?.[0];
}

function parseTagIds(value: RouteParamValue): readonly string[] {
  const values = typeof value === 'string' ? [value] : value ?? [];
  const normalized = values.map((tagId) => tagId.trim()).filter((tagId) => tagId.length > 0);
  return [...new Set(normalized)];
}

function parseScope(value: string | undefined): BrowseScope {
  switch (value) {
    case 'places':
    case 'containers':
    case 'items':
    case 'all':
      return value;
    default:
      return 'all';
  }
}

function parseLifecycleState(value: string | undefined): AssetBrowseLifecycleFilter {
  switch (value) {
    case 'active':
    case 'archived':
    case 'all':
      return value;
    default:
      return 'active';
  }
}

function parseCheckoutState(value: string | undefined): AssetBrowseCheckoutFilter {
  switch (value) {
    case 'any':
    case 'checked_out':
    case 'available':
      return value;
    default:
      return 'any';
  }
}

function parseSort(value: string | undefined): AssetBrowseSort {
  switch (value) {
    case 'updated_desc':
    case 'id_asc':
      return value;
    default:
      return 'updated_desc';
  }
}
