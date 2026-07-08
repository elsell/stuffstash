import type { AssetTagViewModel } from '../../application/assets/AssetViewModels';

export type AssetTagChipPresentation = {
  readonly visibleTags: readonly AssetTagViewModel[];
  readonly hiddenCount: number;
  readonly shouldRender: boolean;
};

export type AssetTagChipLayoutPresentation = {
  readonly compactRow: boolean;
  readonly shrinkVisibleChips: boolean;
};

export type AssetTagChipStylePresentation = {
  readonly colored: boolean;
  readonly backgroundColor?: string;
  readonly borderColor?: string;
};

export function assetTagChipPresentation(
  tags: readonly AssetTagViewModel[] | undefined,
  overflowLimit?: number
): AssetTagChipPresentation {
  const allTags = tags ?? [];
  const visibleLimit = overflowLimit ?? allTags.length;
  const visibleTags = allTags.slice(0, visibleLimit);
  const hiddenCount = Math.max(0, allTags.length - visibleTags.length);
  return {
    visibleTags,
    hiddenCount,
    shouldRender: visibleTags.length > 0 || hiddenCount > 0
  };
}

export function assetTagChipLayoutPresentation(compact = false): AssetTagChipLayoutPresentation {
  return {
    compactRow: compact,
    shrinkVisibleChips: compact
  };
}

export function assetTagChipStylePresentation(tag: Pick<AssetTagViewModel, 'color'>): AssetTagChipStylePresentation {
  if (!tag.color) {
    return { colored: false };
  }
  const rgb = hexToRGB(tag.color);
  if (!rgb) {
    return { colored: false };
  }
  return {
    colored: true,
    backgroundColor: `rgba(${rgb.red}, ${rgb.green}, ${rgb.blue}, 0.14)`,
    borderColor: relativeLuminance(rgb) > 0.78 ? '#D9E1E6' : tag.color
  };
}

function hexToRGB(color: string): { red: number; green: number; blue: number } | null {
  const match = /^#([0-9a-fA-F]{6})$/.exec(color.trim());
  if (!match) {
    return null;
  }
  const value = match[1];
  return {
    red: Number.parseInt(value.slice(0, 2), 16),
    green: Number.parseInt(value.slice(2, 4), 16),
    blue: Number.parseInt(value.slice(4, 6), 16)
  };
}

function relativeLuminance(rgb: { red: number; green: number; blue: number }): number {
  const [red, green, blue] = [rgb.red, rgb.green, rgb.blue].map((channel) => {
    const value = channel / 255;
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}
