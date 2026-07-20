import { describe, expect, it, vi } from 'vitest';
import {
  AssetBreadcrumbTrail,
  AssetCard,
  shouldShowAssetCardSupportingDetails
} from './AssetCard';
import { lightHighContrastPalette } from '../theme/tokens';

vi.mock('react-native', () => ({
  DynamicColorIOS: ({ light }: { readonly light: string }) => light,
  Image: 'Image',
  Platform: { OS: 'android' },
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles,
    hairlineWidth: 0.5
  },
  Text: 'Text',
  View: 'View'
}));

vi.mock('./AssetTagChips', () => ({
  AssetTagChips: 'AssetTagChips'
}));

vi.mock('../theme/AppearanceContext', () => ({
  useAppearancePalette: () => ({
    surface: '#ffffff',
    border: '#D9E1E6',
    controlBorder: '#6F7E88',
    surfaceMuted: '#E8F0F5',
    selected: '#E8F0F5',
    text: '#243038',
    textMuted: '#52616B',
    accentStrong: '#303A41',
    warningSurface: '#FFF3DF',
    warning: '#8A4F00',
    action: '#0066CC',
    actionPressed: '#004F9F',
    onAction: '#ffffff'
  })
}));

describe('AssetCard', () => {
  it('does not reserve a blank supporting-details row for assets without details', () => {
    const asset = {
      id: 'asset-empty',
      title: 'Unlabeled box',
      kindLabel: 'Container',
      description: '  ',
      locationTrailLabel: 'Garage',
      parentLocationTrail: [],
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      imagePlaceholderLabel: 'Box'
    };

    expect(shouldShowAssetCardSupportingDetails(asset, 'standard')).toBe(false);
    expect(shouldShowAssetCardSupportingDetails({ ...asset, searchMatchLabels: ['title'] }, 'standard')).toBe(true);
    expect(shouldShowAssetCardSupportingDetails({ ...asset, description: 'Seasonal storage' }, 'standard')).toBe(true);
    expect(shouldShowAssetCardSupportingDetails({ ...asset, description: 'Seasonal storage' }, 'compact')).toBe(false);

    const searchCard = AssetCard({
      asset: { ...asset, searchMatchLabels: ['title'] },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });
    expect(collectText(searchCard).join('')).toContain('Matched title');
    expect(findFirstByStyleValue(searchCard, 'minHeight', 36)).toBeUndefined();
  });

  it('shows the asset name above placement breadcrumbs instead of asset type chips', () => {
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Bottom drawer',
        kindLabel: 'Item',
        customTypeLabel: 'Tool',
        description: 'Garage drill',
        locationTrailLabel: 'Garage / Holiday / seasonal bin / Bottom drawer',
        parentLocationTrail: [
          { id: 'asset-garage', title: 'Garage', isImmediateParent: false },
          { id: 'asset-holiday-bin', title: 'Holiday / seasonal bin', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Item'
      },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });
    const text = collectText(card);
    const breadcrumbScroller = findFirstByType(card, 'ScrollView');
    const breadcrumbText = collectText(breadcrumbScroller);
    const longBreadcrumb = findFirstByText(card, 'Holiday / seasonal bin');

    expect(text).toContain('Bottom drawer');
    expect(breadcrumbText).toEqual(expect.arrayContaining(['Garage', 'Holiday / seasonal bin']));
    expect(breadcrumbText).not.toContain('Bottom drawer');
    expect(styleValue(longBreadcrumb?.props?.style, 'maxWidth')).toBeGreaterThanOrEqual(232);
    expect(text.indexOf('Bottom drawer')).toBeLessThan(text.indexOf('Garage'));
    expect(text).not.toContain('Tool');
    expect(text).not.toContain('Updated today');
    expect(text).not.toContain('Needs photo');
    expect(findFirstByAccessibilityLabel(card, 'Open asset Bottom drawer')?.props?.accessibilityRole).toBe('button');
    expect(findFirstByStyleValue(card, 'width', '100%')?.props?.accessible).toBe(false);
    expect(findFirstByStyleValue(card, 'width', '100%')?.props?.importantForAccessibility).toBeUndefined();
  });

  it('overlays checked-out status on the asset image and omits updated and photo-readiness text', () => {
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: 'Garage drill',
        locationTrailLabel: 'Garage / Tool chest',
        parentLocationTrail: [
          { id: 'asset-garage', title: 'Garage', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        checkedOutLabel: 'Checked out',
        imagePlaceholderLabel: 'Item'
      },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });
    const text = collectText(card);
    const breadcrumbScroller = findFirstByType(card, 'ScrollView');
    const imageFrame = findFirstByStyleValue(card, 'aspectRatio', 1);

    expect(text).toContain('Checked out');
    expect(text).not.toContain('Updated today');
    expect(text).not.toContain('Needs photo');
    expect(collectText(breadcrumbScroller)).not.toContain('Checked out');
    expect(collectText(imageFrame)).toContain('Checked out');
  });

  it('keeps the card surface light when device appearance is not consulted', () => {
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        description: 'Garage drill',
        locationTrailLabel: 'Garage',
        parentLocationTrail: [],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Photo ready',
        imagePlaceholderLabel: 'Item'
      },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });

    expect(styleValue(card.props?.style, 'backgroundColor')).toBe('#ffffff');
    expect(styleValue(card.props?.style, 'borderColor')).toBe('#D9E1E6');
    expect(styleValue(card.props?.style, 'borderWidth')).toBe(0.5);
    expect(styleValue(findFirstByText(card, 'Cordless drill')?.props?.style, 'color')).toBe('#243038');
  });

  it('accepts an increased-contrast semantic palette without changing the card structure', () => {
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        description: 'Garage drill',
        locationTrailLabel: 'Garage',
        parentLocationTrail: [{ id: 'asset-garage', title: 'Garage', isImmediateParent: true }],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Photo ready',
        imagePlaceholderLabel: 'Item'
      },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn(),
      palette: lightHighContrastPalette
    });

    expect(styleValue(card.props?.style, 'backgroundColor')).toBe(lightHighContrastPalette.surface);
    expect(styleValue(card.props?.style, 'borderColor')).toBe(lightHighContrastPalette.border);
    expect(styleValue(findFirstByText(card, 'Cordless drill')?.props?.style, 'color')).toBe(lightHighContrastPalette.text);
  });

  it('uses the shared compact layout for Home cards while preserving tags and breadcrumbs', () => {
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: 'Garage drill',
        locationTrailLabel: 'Garage / Tools',
        parentLocationTrail: [
          { id: 'asset-garage', title: 'Garage', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Item',
        tags: [
          { id: 'tag-tools', label: 'Tools', color: '#2F80ED' },
          { id: 'tag-camping', label: 'Camping', color: '#2E7D32' }
        ]
      },
      density: 'compact',
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });

    expect(styleValue(card.props?.style, 'width')).toBe(164);
    expect(collectText(card)).toEqual(expect.arrayContaining(['Cordless drill', 'Garage']));
    expect(findFirstByType(card, 'AssetTagChips')?.props?.tags).toEqual([
      { id: 'tag-tools', label: 'Tools', color: '#2F80ED' },
      { id: 'tag-camping', label: 'Camping', color: '#2E7D32' }
    ]);
    expect(collectText(card)).not.toContain('Garage drill');
  });

  it('offers a full-width compact row without turning the thumbnail into the dominant surface', () => {
    const card = renderHomeRow({
      asset: makeHomeRowAsset(),
      density: 'row',
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });
    const thumbnail = findFirstByStyleNumberRange(card, 'width', 56, 80);

    expect(styleValue(card.props?.style, 'width')).toBe('100%');
    expect(styleValue(card.props?.style, 'flexDirection')).toBe('row');
    expect(styleValue(card.props?.style, 'backgroundColor')).toBe('transparent');
    expect(styleValue(card.props?.style, 'minHeight')).toBeLessThanOrEqual(112);
    expect(styleValue(thumbnail?.props?.style, 'width')).toBeGreaterThanOrEqual(56);
    expect(styleValue(thumbnail?.props?.style, 'width')).toBeLessThanOrEqual(80);
    expect(styleValue(thumbnail?.props?.style, 'height')).toBe(
      styleValue(thumbnail?.props?.style, 'width')
    );
  });

  it('keeps Home row typography restrained and reveals recency only when explicitly requested', () => {
    const hiddenRecencyCard = renderHomeRow({
      asset: makeHomeRowAsset(),
      density: 'row',
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });
    const visibleRecencyCard = renderHomeRow({
      asset: makeHomeRowAsset(),
      density: 'row',
      onParentLocationPress: vi.fn(),
      onPress: vi.fn(),
      showUpdatedAt: true
    });
    const title = findFirstByText(visibleRecencyCard, 'Cordless drill');
    const recency = findFirstByText(visibleRecencyCard, 'Updated 2 hours ago');

    expect(collectText(hiddenRecencyCard)).not.toContain('Updated 2 hours ago');
    expect(collectText(visibleRecencyCard)).toContain('Updated 2 hours ago');
    expect(styleValue(title?.props?.style, 'fontWeight')).toBe('600');
    expect(styleValue(title?.props?.style, 'lineHeight')).toBeUndefined();
    expect(title?.props?.numberOfLines).toBeUndefined();
    expect(styleValue(recency?.props?.style, 'fontWeight')).not.toBe('700');
    expect(styleValue(recency?.props?.style, 'lineHeight')).toBeUndefined();
    expect(styleValue(recency?.props?.style, 'fontSize')).toBeGreaterThanOrEqual(12);
    expect(styleValue(recency?.props?.style, 'fontSize')).toBeLessThanOrEqual(15);
  });

  it('preserves actionable placement and checkout affordances in a restrained Home row', () => {
    const onLocationPress = vi.fn();
    const onReturn = vi.fn();
    const card = renderHomeRow({
      asset: {
        ...makeHomeRowAsset(),
        checkedOutLabel: 'Checked out to Taylor'
      },
      density: 'row',
      footerAction: {
        label: 'Return',
        onPress: onReturn
      },
      onParentLocationPress: onLocationPress,
      onPress: vi.fn()
    });
    const garage = findFirstByAccessibilityLabel(card, 'Open location Garage');
    const returnButton = findPressableWithText(card, 'Return');

    expect(collectText(card)).toEqual(expect.arrayContaining([
      'Cordless drill',
      'Garage',
      'Checked out to Taylor',
      'Return'
    ]));
    expect(garage?.props?.accessibilityRole).toBe('button');
    (garage?.props?.onPress as (() => void) | undefined)?.();
    expect(onLocationPress).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'asset-garage', title: 'Garage' })
    );
    expect(styleValue(resolvePressableStyle(returnButton?.props?.style, false), 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(styleValue(resolvePressableStyle(returnButton?.props?.style, false), 'minWidth')).toBeGreaterThanOrEqual(44);
    expect(styleValue(card.props?.style, 'borderWidth')).toBe(0);
    expect(Number(styleValue(card.props?.style, 'borderBottomWidth') ?? 0)).toBeLessThanOrEqual(1);
  });

  it('keeps the standard card flexible for an ordinary-phone two-column grid', () => {
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        description: 'Garage drill',
        locationTrailLabel: 'Garage',
        parentLocationTrail: [],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Photo ready',
        imagePlaceholderLabel: 'Item'
      },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });

    expect(styleValue(card.props?.style, 'flex')).toBe(1);
    expect(styleValue(card.props?.style, 'width')).toBeUndefined();
  });

  it('owns the compact footer action used by checked-out Home cards', () => {
    const onReturn = vi.fn();
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: 'Garage drill',
        locationTrailLabel: 'Garage / Tools',
        parentLocationTrail: [
          { id: 'asset-garage', title: 'Garage', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Item',
        checkedOutLabel: 'Checked out',
        tags: [{ id: 'tag-tools', label: 'Tools', color: '#2F80ED' }]
      },
      density: 'compact',
      footerAction: {
        disabled: false,
        label: 'Return',
        onPress: onReturn
      },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn(),
      showTags: false
    });

    expect(collectText(card)).toEqual(expect.arrayContaining(['Cordless drill', 'Checked out', 'Return']));
    expect(collectText(card)).not.toContain('Tools');

    const returnButton = findPressableWithText(card, 'Return');
    expect(styleValue(resolvePressableStyle(returnButton?.props?.style, false), 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(resolvePressableStyle(returnButton?.props?.style, true)).not.toEqual(
      resolvePressableStyle(returnButton?.props?.style, false)
    );
    (returnButton?.props?.onPress as (() => void) | undefined)?.();

    expect(onReturn).toHaveBeenCalledTimes(1);
  });
});

type ExperimentalHomeRowProps = Omit<Parameters<typeof AssetCard>[0], 'density'> & {
  readonly density: 'row';
  readonly showUpdatedAt?: boolean;
};

function renderHomeRow(props: ExperimentalHomeRowProps): ReturnType<typeof AssetCard> {
  return AssetCard(props as unknown as Parameters<typeof AssetCard>[0]);
}

function makeHomeRowAsset() {
  return {
    id: 'asset-drill',
    title: 'Cordless drill',
    kindLabel: 'Item',
    description: 'Garage drill',
    locationTrailLabel: 'Garage',
    parentLocationTrail: [
      { id: 'asset-garage', title: 'Garage', isImmediateParent: true }
    ],
    updatedAtLabel: 'Updated 2 hours ago',
    photoLabel: 'Photo ready',
    imagePlaceholderLabel: 'Item'
  };
}

describe('AssetBreadcrumbTrail', () => {
  it('supports a more prominent detail-page treatment without changing card density', () => {
    const trail = AssetBreadcrumbTrail({
      prominence: 'detail',
      segments: [{ id: 'asset-garage', title: 'Garage', isImmediateParent: true }],
      onSegmentPress: vi.fn()
    });

    expect(findFirstByText(trail, 'Garage')?.props?.style).toEqual(expect.arrayContaining([
      expect.objectContaining({ fontSize: 15 })
    ]));
  });

  it('renders the trail in a horizontal scroller and defaults to the most specific parent', () => {
    const trail = AssetBreadcrumbTrail({
      segments: [
        { id: 'asset-garage', title: 'Garage', isImmediateParent: false },
        { id: 'asset-holiday-bin', title: 'Holiday / seasonal bin', isImmediateParent: true }
      ],
      onSegmentPress: vi.fn()
    });
    const scroller = findFirstByType(trail, 'ScrollView');
    const fakeScroller = { scrollToEnd: vi.fn() };

    expect(scroller?.props?.horizontal).toBe(true);
    expect(collectText(scroller)).toEqual(expect.arrayContaining(['Garage', 'Holiday / seasonal bin']));
    expect(collectText(scroller)).not.toContain('Bottom drawer');
    attachScroller(scroller, fakeScroller);
    triggerContentSizeChange(scroller);

    expect(fakeScroller.scrollToEnd).toHaveBeenCalledWith({ animated: false });
  });

  it('opens the selected parent location and gives ancestors lower accessible visual weight', () => {
    const onSegmentPress = vi.fn();
    const trail = AssetBreadcrumbTrail({
      segments: [
        { id: 'asset-garage', title: 'Garage', isImmediateParent: false },
        { id: 'asset-holiday-bin', title: 'Holiday / seasonal bin', isImmediateParent: true }
      ],
      onSegmentPress
    });
    const ancestorButton = findFirstByAccessibilityLabel(trail, 'Open location Garage');
    const immediateButton = findFirstByAccessibilityLabel(trail, 'Open location Holiday / seasonal bin');
    const ancestorText = findFirstByText(ancestorButton, 'Garage');
    const immediateText = findFirstByText(immediateButton, 'Holiday / seasonal bin');

    expect(ancestorButton?.props?.accessibilityRole).toBe('button');
    expect(immediateButton?.props?.accessibilityRole).toBe('button');
    expect(ancestorButton?.props?.hitSlop).toBeUndefined();
    expect(styleValue(resolvePressableStyle(ancestorButton?.props?.style, false), 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(styleValue(resolvePressableStyle(ancestorButton?.props?.style, false), 'minWidth')).toBeGreaterThanOrEqual(44);
    expect(resolvePressableStyle(ancestorButton?.props?.style, true)).not.toEqual(
      resolvePressableStyle(ancestorButton?.props?.style, false)
    );
    expect(styleValue(resolvePressableStyle(ancestorButton?.props?.style, true), 'backgroundColor')).toBeUndefined();
    expect(styleValue(resolvePressableStyle(ancestorButton?.props?.style, true), 'opacity')).toBeLessThan(1);
    expect(styleValue(ancestorText?.props?.style, 'fontWeight')).toBe('600');
    expect(styleValue(immediateText?.props?.style, 'fontWeight')).toBe('700');
    expect(styleValue(ancestorText?.props?.style, 'opacity')).toBeUndefined();

    (ancestorButton?.props?.onPress as (() => void) | undefined)?.();

    expect(onSegmentPress).toHaveBeenCalledWith({
      id: 'asset-garage',
      title: 'Garage',
      isImmediateParent: false
    });
  });

  it('uses opacity-only card feedback without introducing heavy borders or fills', () => {
    const card = AssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        description: 'Garage drill',
        locationTrailLabel: 'Garage',
        parentLocationTrail: [],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Photo ready',
        imagePlaceholderLabel: 'Item'
      },
      onParentLocationPress: vi.fn(),
      onPress: vi.fn()
    });
    const openAsset = findFirstByAccessibilityLabel(card, 'Open asset Cordless drill');
    const mediaRegion = findFirstPressableByStyleValue(card, 'width', '100%');

    const idleText = resolvePressableStyle(openAsset?.props?.style, false);
    const pressedText = resolvePressableStyle(openAsset?.props?.style, true);
    const idleMedia = resolvePressableStyle(mediaRegion?.props?.style, false);
    const pressedMedia = resolvePressableStyle(mediaRegion?.props?.style, true);

    expect(styleValue(pressedText, 'opacity')).toBeLessThan(1);
    expect(styleValue(pressedMedia, 'opacity')).toBeLessThan(1);
    for (const key of ['backgroundColor', 'borderColor', 'borderWidth']) {
      expect(styleValue(pressedText, key)).toBe(styleValue(idleText, key));
      expect(styleValue(pressedMedia, key)).toBe(styleValue(idleMedia, key));
    }
  });

  it('does not render breadcrumbs when no parent trail exists', () => {
    const trail = AssetBreadcrumbTrail({ segments: [], onSegmentPress: vi.fn() });

    expect(trail).toBeNull();
  });
});

type ElementNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};

function collectText(node: unknown): readonly string[] {
  if (typeof node === 'string') {
    return [node];
  }

  if (Array.isArray(node)) {
    return node.flatMap(collectText);
  }

  if (!isElementNode(node)) {
    return [];
  }

  if (typeof node.type === 'function') {
    return collectText(node.type(node.props));
  }

  return childrenOf(node).flatMap(collectText);
}

function findFirstByType(node: unknown, type: unknown): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>((found, child) => found ?? findFirstByType(child, type), undefined);
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.type === type) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstByType(node.type(node.props), type);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByType(child, type),
    undefined
  );
}

function findFirstByText(node: unknown, text: string): ElementNode | undefined {
  if (!isElementNode(node)) {
    return undefined;
  }

  if (childrenOf(node).includes(text)) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstByText(node.type(node.props), text);
  }

  for (const child of childrenOf(node)) {
    const match = findFirstByText(child, text);
    if (match) {
      return match;
    }
  }

  return undefined;
}

function styleValue(style: unknown, key: string): unknown {
  if (Array.isArray(style)) {
    return style.reduce<unknown>((found, entry) => styleValue(entry, key) ?? found, undefined);
  }
  return style && typeof style === 'object' ? (style as Record<string, unknown>)[key] : undefined;
}

function resolvePressableStyle(style: unknown, pressed: boolean): unknown {
  return typeof style === 'function' ? style({ pressed }) : style;
}

function findFirstByStyleValue(node: unknown, styleKey: string, value: unknown): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstByStyleValue(child, styleKey, value),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (hasStyleValue(node.props?.style, styleKey, value)) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstByStyleValue(node.type(node.props), styleKey, value);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByStyleValue(child, styleKey, value),
    undefined
  );
}

function findFirstByStyleNumberRange(
  node: unknown,
  styleKey: string,
  minimum: number,
  maximum: number
): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstByStyleNumberRange(child, styleKey, minimum, maximum),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  const value = styleValue(node.props?.style, styleKey);
  if (typeof value === 'number' && value >= minimum && value <= maximum) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstByStyleNumberRange(node.type(node.props), styleKey, minimum, maximum);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByStyleNumberRange(child, styleKey, minimum, maximum),
    undefined
  );
}

function findFirstPressableByStyleValue(node: unknown, styleKey: string, value: unknown): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstPressableByStyleValue(child, styleKey, value),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.type === 'Pressable' && hasStyleValue(node.props?.style, styleKey, value)) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstPressableByStyleValue(node.type(node.props), styleKey, value);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstPressableByStyleValue(child, styleKey, value),
    undefined
  );
}

function hasStyleValue(style: unknown, styleKey: string, value: unknown): boolean {
  if (typeof style === 'function') {
    return hasStyleValue(resolvePressableStyle(style, false), styleKey, value);
  }
  if (Array.isArray(style)) {
    return style.some((entry) => hasStyleValue(entry, styleKey, value));
  }
  return Boolean(style && typeof style === 'object' && (style as Record<string, unknown>)[styleKey] === value);
}

function childrenOf(node: ElementNode): readonly unknown[] {
  const children = node.props?.children;
  return Array.isArray(children) ? children : [children];
}

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}

function findFirstByAccessibilityLabel(node: unknown, label: string): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstByAccessibilityLabel(child, label),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.props?.accessibilityLabel === label) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstByAccessibilityLabel(node.type(node.props), label);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByAccessibilityLabel(child, label),
    undefined
  );
}

function findPressableWithText(node: unknown, value: string): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findPressableWithText(child, value),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.type === 'Pressable' && collectText(node).includes(value)) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findPressableWithText(node.type(node.props), value);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findPressableWithText(child, value),
    undefined
  );
}

function attachScroller(node: ElementNode | undefined, scroller: { readonly scrollToEnd: () => void }): void {
  const ref = node?.props?.ref;
  if (typeof ref !== 'function') {
    throw new Error('Missing ScrollView ref callback');
  }
  ref(scroller);
}

function triggerContentSizeChange(node: ElementNode | undefined): void {
  const onContentSizeChange = node?.props?.onContentSizeChange;
  if (typeof onContentSizeChange !== 'function') {
    throw new Error('Missing content size handler');
  }
  onContentSizeChange();
}
