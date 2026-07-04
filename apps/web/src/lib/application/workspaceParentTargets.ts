import type { ParentTargetViewModel } from '$lib/domain/inventory';

export interface ParentTargetSearchResult {
  matchingTargets: ParentTargetViewModel[];
  visibleTargets: ParentTargetViewModel[];
  locationResults: ParentTargetViewModel[];
  containerResults: ParentTargetViewModel[];
}

export type ParentTargetPickerStatusKind = 'none' | 'no-matches' | 'overflow' | 'no-targets';

export interface ParentTargetPickerStatus {
  kind: ParentTargetPickerStatusKind;
  message: string;
}

export interface ParentTargetPickerPresentation {
  resultCountLabel: string;
  destinationCountLabel: string;
  suggestedCountLabel: string;
  status: ParentTargetPickerStatus;
}

export function parentTargetSuggestions(targets: ParentTargetViewModel[], selectedId: string | null, limit: number): ParentTargetViewModel[] {
  const sorted = [...targets].sort(compareParentTargets);
  const chosen: ParentTargetViewModel[] = [];
  for (const target of sorted) {
    if (chosen.length >= limit) {
      break;
    }
    if (target.id === selectedId || chosen.some((candidate) => candidate.id === target.id)) {
      continue;
    }
    chosen.push(target);
  }
  return chosen;
}

export function searchParentTargets(targets: ParentTargetViewModel[], query: string, limit: number): ParentTargetSearchResult {
  const normalizedQuery = normalizeParentTargetQuery(query);
  const matchingTargets = targets
    .filter((target) => parentTargetMatches(target, normalizedQuery))
    .sort((left, right) => compareParentTargetsForSearch(left, right, normalizedQuery));
  const visibleTargets = matchingTargets.slice(0, limit);

  return {
    matchingTargets,
    visibleTargets,
    locationResults: visibleTargets.filter((target) => target.kind === 'location'),
    containerResults: visibleTargets.filter((target) => target.kind === 'container')
  };
}

export function normalizeParentTargetQuery(query: string): string {
  return query.trim().toLowerCase();
}

export function parentTargetPickerPresentation(input: {
  hasSearch: boolean;
  matchingCount: number;
  visibleCount: number;
  targetCount: number;
  suggestedCount: number;
}): ParentTargetPickerPresentation {
  return {
    resultCountLabel: input.hasSearch ? `${input.matchingCount} ${input.matchingCount === 1 ? 'match' : 'matches'}` : '',
    destinationCountLabel: `${input.targetCount} possible ${input.targetCount === 1 ? 'destination' : 'destinations'}`,
    suggestedCountLabel: `Showing ${input.suggestedCount} suggested ${input.suggestedCount === 1 ? 'destination' : 'destinations'}.`,
    status: parentTargetPickerStatus(input)
  };
}

function parentTargetPickerStatus(input: {
  hasSearch: boolean;
  matchingCount: number;
  visibleCount: number;
  targetCount: number;
}): ParentTargetPickerStatus {
  if (input.hasSearch && input.visibleCount === 0) {
    return { kind: 'no-matches', message: 'No matching locations or containers.' };
  }
  if (input.hasSearch && input.matchingCount > input.visibleCount) {
    return { kind: 'overflow', message: `Showing the first ${input.visibleCount} of ${input.matchingCount} matches.` };
  }
  if (!input.hasSearch && input.targetCount === 0) {
    return { kind: 'no-targets', message: 'No locations or containers yet.' };
  }
  return { kind: 'none', message: '' };
}

function parentTargetMatches(target: ParentTargetViewModel, query: string): boolean {
  if (!query) {
    return true;
  }
  return target.title.toLowerCase().includes(query) || target.containmentTrail.toLowerCase().includes(query);
}

function compareParentTargets(left: ParentTargetViewModel, right: ParentTargetViewModel): number {
  const kindRank = parentTargetKindRank(left.kind) - parentTargetKindRank(right.kind);
  if (kindRank !== 0) {
    return kindRank;
  }
  return left.title.localeCompare(right.title);
}

function compareParentTargetsForSearch(left: ParentTargetViewModel, right: ParentTargetViewModel, query: string): number {
  if (!query) {
    return compareParentTargets(left, right);
  }
  const kindRank = parentTargetKindRank(left.kind) - parentTargetKindRank(right.kind);
  if (kindRank !== 0) {
    return kindRank;
  }
  const relevanceRank = parentTargetSearchRank(left, query) - parentTargetSearchRank(right, query);
  if (relevanceRank !== 0) {
    return relevanceRank;
  }
  return left.title.localeCompare(right.title);
}

function parentTargetSearchRank(target: ParentTargetViewModel, query: string): number {
  const title = target.title.toLowerCase();
  const trail = target.containmentTrail.toLowerCase();
  const trailSegments = trail.split('/').map((segment) => segment.trim());
  if (title === query) {
    return 0;
  }
  if (title.startsWith(query)) {
    return 1;
  }
  if (title.includes(query)) {
    return 2;
  }
  if (trailSegments.includes(query)) {
    return 3;
  }
  if (trail.includes(query)) {
    return 4;
  }
  return 5;
}

function parentTargetKindRank(kind: ParentTargetViewModel['kind']): number {
  return kind === 'location' ? 0 : 1;
}
