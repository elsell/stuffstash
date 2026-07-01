import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';

export type ParentSelection = {
  readonly id: string;
  readonly title: string;
  readonly kind: ParentLookupResult['kind'];
  readonly subtitle: string;
  readonly pathLabel: string;
  readonly selectionHint: string;
  readonly willPromoteToContainer: boolean;
};

export function resolveParentAssetId(
  parentMatches: readonly ParentLookupResult[],
  parentQuery: string,
  parentAssetId: string | undefined
): string | undefined {
  const normalizedQuery = normalizeParentName(parentQuery);
  if (normalizedQuery.length === 0) {
    return parentAssetId;
  }
  if (parentAssetId) {
    return parentAssetId;
  }

  const exactParent = parentMatches.find(
    (parent) => normalizeParentName(parent.title) === normalizedQuery
  );
  if (exactParent) {
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
    (lastParent?.id === parentAssetId ? lastParent : undefined) ??
    {
      id: parentAssetId,
      title: parentQuery,
      kind: 'container',
      subtitle: 'Selected parent',
      pathLabel: parentQuery,
      selectionHint: 'Container',
      willPromoteToContainer: false
    }
  );
}

function normalizeParentName(value: string): string {
  return value.trim().toLocaleLowerCase();
}
