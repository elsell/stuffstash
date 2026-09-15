const initialTagChoiceLimit = 12;

/** Retain selections while disclosing naturally ordered inventory tag choices. */
export function tagChoicePresentation<T extends { readonly id: string }>(input: {
  readonly tags: readonly T[];
  readonly selectedIds: readonly string[];
  readonly label: (tag: T) => string;
  readonly expanded: boolean;
  readonly query?: string;
}) {
  const query = (input.query ?? '').trim().toLocaleLowerCase();
  const selected = new Set(input.selectedIds);
  const sorted = [...input.tags].sort((left, right) => input.label(left).localeCompare(input.label(right), undefined, { numeric: true, sensitivity: 'base' }));
  const matches = sorted.filter(tag => input.label(tag).toLocaleLowerCase().includes(query));
  const disclosed = new Set((input.expanded ? matches : matches.slice(0, initialTagChoiceLimit)).map(tag => tag.id));
  return {
    visibleTags: sorted.filter(tag => selected.has(tag.id) || disclosed.has(tag.id)),
    canDisclose: matches.length > initialTagChoiceLimit,
    noMatches: query.length > 0 && matches.length === 0
  };
}
