import type { VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';

export type VoiceResponseEntityLinkSegment = {
  readonly text: string;
  readonly reference?: VoiceResponseArtifact;
};

export type VoiceResponseEntityLinkPresentation = {
  readonly segments: readonly VoiceResponseEntityLinkSegment[];
  readonly fallbackReferences: readonly VoiceResponseArtifact[];
};

type EntityMatch = {
  readonly start: number;
  readonly end: number;
  readonly reference: VoiceResponseArtifact;
};

export function buildVoiceResponseEntityLinks(
  text: string,
  references: readonly VoiceResponseArtifact[]
): VoiceResponseEntityLinkPresentation {
  const byTitle = new Map<string, VoiceResponseArtifact[]>();
  for (const reference of references) {
    const key = reference.title;
    byTitle.set(key, [...(byTitle.get(key) ?? []), reference]);
  }

  const matches: EntityMatch[] = [];
  for (const group of byTitle.values()) {
    if (group.length !== 1) {
      continue;
    }
    const reference = group[0];
    matches.push(...exactTitleMatches(text, reference));
  }
  matches.sort((left, right) => {
    const lengthDifference = (right.end - right.start) - (left.end - left.start);
    return lengthDifference || left.start - right.start;
  });

  const selected: EntityMatch[] = [];
  for (const match of matches) {
    if (selected.some((existing) => match.start < existing.end && match.end > existing.start)) {
      continue;
    }
    selected.push(match);
  }
  selected.sort((left, right) => left.start - right.start);

  const segments: VoiceResponseEntityLinkSegment[] = [];
  let cursor = 0;
  for (const match of selected) {
    if (match.start > cursor) {
      segments.push({ text: text.slice(cursor, match.start) });
    }
    segments.push({ text: text.slice(match.start, match.end), reference: match.reference });
    cursor = match.end;
  }
  if (cursor < text.length || segments.length === 0) {
    segments.push({ text: text.slice(cursor) });
  }

  const placed = new Set(selected.map((match) => match.reference.assetId));
  return {
    segments,
    fallbackReferences: references.filter((reference) => !placed.has(reference.assetId))
  };
}

function exactTitleMatches(text: string, reference: VoiceResponseArtifact): EntityMatch[] {
  const matches: EntityMatch[] = [];
  if (!reference.title) {
    return matches;
  }
  const pattern = new RegExp(`(^|[^\\p{L}\\p{N}])(${escapeRegularExpression(reference.title)})(?=$|[^\\p{L}\\p{N}])`, 'gu');
  for (const match of text.matchAll(pattern)) {
    const prefix = match[1] ?? '';
    const title = match[2] ?? '';
    const start = (match.index ?? 0) + prefix.length;
    matches.push({ start, end: start + title.length, reference });
  }
  return matches;
}

function escapeRegularExpression(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export function voiceResponseEntityOpenLabel(
  reference: VoiceResponseArtifact,
  references: readonly VoiceResponseArtifact[]
): string {
  const context = reference.context?.trim();
  const base = `Open ${reference.title}${context ? ` in ${context}` : ''}`;
  const indistinguishable = references.filter((candidate) =>
    candidate.title === reference.title && (candidate.context?.trim() ?? '') === (context ?? '')
  );
  if (indistinguishable.length < 2) {
    return base;
  }
  const index = indistinguishable.findIndex((candidate) => candidate.assetId === reference.assetId);
  return `${base} (${index + 1} of ${indistinguishable.length})`;
}
