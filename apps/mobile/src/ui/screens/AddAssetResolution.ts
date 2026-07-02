import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';

export type ParentSelection = {
  readonly id: string;
  readonly title: string;
  readonly kind: ParentLookupResult['kind'];
  readonly subtitle: string;
  readonly pathLabel: string;
  readonly selectionHint: string;
  readonly willPromoteToContainer: boolean;
  readonly canSelectAsParent?: boolean;
  readonly disabledReason?: string;
};

export function resolveParentAssetId(
  parentMatches: readonly ParentLookupResult[],
  parentQuery: string,
  parentAssetId: string | undefined
): string | undefined {
  const normalizedQuery = normalizeParentName(parentQuery);
  if (parentAssetId) {
    const selectedMatch = parentMatches.find((parent) => parent.id === parentAssetId);
    assertSelectableParent(selectedMatch);
    return parentAssetId;
  }
  if (normalizedQuery.length === 0) {
    return undefined;
  }

  const exactParent = parentMatches.find(
    (parent) => normalizeParentName(parent.title) === normalizedQuery
  );
  if (exactParent) {
    assertSelectableParent(exactParent);
    return exactParent.id;
  }

  throw new Error('Create this parent or clear the Put in field.');
}

export function resolveSelectedParent(
  parentMatches: readonly ParentLookupResult[],
  parentAssetId: string | undefined,
  parentQuery: string,
  lastParent: ParentSelection | undefined
): ParentSelection | undefined {
  if (!parentAssetId) {
    return undefined;
  }

  return (
    parentMatches.find((parent) => parent.id === parentAssetId) ??
    (lastParent?.id === parentAssetId ? lastParent : undefined)
  );
}

export function assertSelectableParent(parent: ParentSelection | ParentLookupResult | undefined): void {
  if (parent?.canSelectAsParent === false) {
    throw new Error(parent.disabledReason ?? 'Choose a place or container for Put in.');
  }
}

function normalizeParentName(value: string): string {
  return value.trim().toLocaleLowerCase();
}
