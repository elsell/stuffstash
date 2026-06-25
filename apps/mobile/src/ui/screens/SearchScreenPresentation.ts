import type { RefObject } from 'react';
import type { TextInput } from 'react-native';
import type { SearchAssetsViewModel } from '../../application/search/SearchAssetsQuery';

type SearchFilterGroupKey = 'status' | 'type' | 'sort';

export type SearchFilterGroupPlacement = {
  readonly key: SearchFilterGroupKey;
  readonly isLast: boolean;
};

export function focusSearchInput(inputRef: RefObject<TextInput | null>): void {
  inputRef.current?.focus();
}

export function buildSearchFilterGroupPlacement(
  resultsMode: SearchAssetsViewModel['mode']
): readonly SearchFilterGroupPlacement[] {
  if (resultsMode === 'browse') {
    return [
      { key: 'status', isLast: false },
      { key: 'type', isLast: false },
      { key: 'sort', isLast: true }
    ];
  }

  return [
    { key: 'status', isLast: false },
    { key: 'type', isLast: true }
  ];
}
