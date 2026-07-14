import { describe, expect, it, vi } from 'vitest';
import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';
import {
  AssetDetailPhotoGallery,
  assetDetailPhotoPages,
  assetDetailPhotoWidth
} from './AssetDetailPhotoGallery';
import { lightHighContrastPalette } from '../theme/tokens';

vi.mock('lucide-react-native', () => ({
  Camera: 'CameraIcon'
}));

vi.mock('../theme/appearance', () => ({
  useAppearanceAwarePalette: () => ({
    action: '#0066CC',
    border: '#64727C',
    onScrim: '#FFFFFF',
    surface: '#FFFFFF',
    surfaceMuted: '#E8F0F5',
    textMuted: '#52616B'
  })
}));

vi.mock('react-native', () => ({
  DynamicColorIOS: ({ light }: { light: string }) => light,
  Image: 'Image',
  Platform: { OS: 'ios' },
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  View: 'View',
  useColorScheme: () => 'light',
  useWindowDimensions: () => ({ width: 390, height: 844, scale: 3, fontScale: 1 })
}));

describe('AssetDetailPhotoGallery presentation', () => {
  it('derives position labels without exposing photo file names', () => {
    expect(assetDetailPhotoPages(photos())).toEqual([
      { accessibilityLabel: 'Open photo 1 of 2', positionLabel: '1 of 2' },
      { accessibilityLabel: 'Open photo 2 of 2', positionLabel: '2 of 2' }
    ]);
  });

  it('uses the viewport content width without a fixed minimum', () => {
    expect(assetDetailPhotoWidth(390)).toBe(342);
    expect(assetDetailPhotoWidth(320)).toBe(272);
    expect(assetDetailPhotoWidth(40)).toBe(0);
  });
});

describe('AssetDetailPhotoGallery', () => {
  it('renders equal paged media widths and opens the selected photo', () => {
    const opened: string[] = [];
    const tree = AssetDetailPhotoGallery({
      canAddPhotos: true,
      imagePlaceholderLabel: 'Item',
      onAddPhotos: () => undefined,
      onPhotoPress: (id) => opened.push(id),
      photos: photos()
    });

    const scroller = findFirstByType(tree, 'ScrollView');
    expect(scroller?.props?.horizontal).toBe(true);
    expect(scroller?.props?.snapToInterval).toBe(352);
    expect(scroller?.props?.decelerationRate).toBe('fast');

    const first = findFirstByProp(tree, 'accessibilityLabel', 'Open photo 1 of 2');
    const second = findFirstByProp(tree, 'accessibilityLabel', 'Open photo 2 of 2');
    expect(first?.props?.style).toEqual(expect.arrayContaining([expect.objectContaining({ width: 342 })]));
    expect(second?.props?.style).toEqual(expect.arrayContaining([expect.objectContaining({ width: 342 })]));
    press(second);
    expect(opened).toEqual(['photo-two']);

    expect(findText(tree, 'IMG_0042.JPG')).toBe(false);
    expect(findText(tree, 'First photo')).toBe(false);
  });

  it('keeps empty media presentational with one quiet Add photos affordance below it', () => {
    let addCount = 0;
    const tree = AssetDetailPhotoGallery({
      canAddPhotos: true,
      imagePlaceholderLabel: 'Item',
      onAddPhotos: () => {
        addCount += 1;
      },
      palette: lightHighContrastPalette,
      photos: []
    });

    const addActions = findAllByProp(tree, 'accessibilityLabel', 'Add photos');
    expect(addActions).toHaveLength(1);
    expect(addActions[0]?.props?.accessibilityRole).toBe('button');
    expect(resolvePressableStyle(addActions[0])).toEqual(expect.arrayContaining([
      expect.objectContaining({ minHeight: 44, borderWidth: 1 }),
      expect.objectContaining({
        backgroundColor: lightHighContrastPalette.elevatedSurface,
        borderColor: lightHighContrastPalette.controlBorder
      })
    ]));
    expect(findFirstByProp(tree, 'accessibilityLabel', 'No photos')?.props?.style).toEqual(expect.arrayContaining([
      expect.objectContaining({ borderWidth: 1 }),
      expect.objectContaining({
        backgroundColor: lightHighContrastPalette.elevatedSurface,
        borderColor: lightHighContrastPalette.border
      })
    ]));
    expect(findFirstByProp(tree, 'accessibilityLabel', 'No photos')?.type).toBe('View');
    press(addActions[0]);
    expect(addCount).toBe(1);
  });

  it('exposes a single separate Add photos action after populated media', () => {
    const tree = AssetDetailPhotoGallery({
      canAddPhotos: true,
      imagePlaceholderLabel: 'Item',
      onAddPhotos: () => undefined,
      photos: photos()
    });

    expect(findAllByProp(tree, 'accessibilityLabel', 'Add photos')).toHaveLength(1);
  });

  it('accepts the asset-detail appearance palette', () => {
    const tree = AssetDetailPhotoGallery({
      canAddPhotos: false,
      imagePlaceholderLabel: 'Item',
      palette: lightHighContrastPalette,
      photos: photos()
    });

    const first = findFirstByProp(tree, 'accessibilityLabel', 'Open photo 1 of 2');
    expect(first?.props?.style).toEqual(expect.arrayContaining([
      expect.objectContaining({ backgroundColor: lightHighContrastPalette.surfaceMuted })
    ]));
  });
});

function photos(): readonly AssetPhotoViewModel[] {
  return [
    {
      id: 'photo-one',
      fileName: 'IMG_0042.JPG',
      label: 'IMG_0042.JPG',
      uri: 'https://example.test/photo-one-thumb',
      heroUri: 'https://example.test/photo-one-hero'
    },
    {
      id: 'photo-two',
      fileName: 'garage-bin.png',
      label: 'garage-bin.png',
      uri: 'https://example.test/photo-two'
    }
  ];
}

function press(node: ElementNode | undefined): void {
  const onPress = node?.props?.onPress;
  if (typeof onPress !== 'function') {
    throw new Error('Missing press handler');
  }
  onPress();
}

function resolvePressableStyle(node: ElementNode | undefined): unknown {
  const style = node?.props?.style;
  return typeof style === 'function' ? style({ pressed: false }) : style;
}

function findFirstByProp(node: unknown, prop: string, value: unknown): ElementNode | undefined {
  return findAllByProp(node, prop, value)[0];
}

function findAllByProp(node: unknown, prop: string, value: unknown): ElementNode[] {
  if (Array.isArray(node)) {
    return node.flatMap((child) => findAllByProp(child, prop, value));
  }
  if (!isElementNode(node)) {
    return [];
  }
  if (typeof node.type === 'function') {
    return findAllByProp(node.type(node.props), prop, value);
  }
  const matches = node.props?.[prop] === value ? [node] : [];
  return [...matches, ...childrenOf(node).flatMap((child) => findAllByProp(child, prop, value))];
}

function findFirstByType(node: unknown, type: unknown): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstByType(child, type),
      undefined
    );
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

function findText(node: unknown, text: string): boolean {
  if (node === text) {
    return true;
  }
  if (Array.isArray(node)) {
    return node.some((child) => findText(child, text));
  }
  return isElementNode(node) && childrenOf(node).some((child) => findText(child, text));
}

function childrenOf(node: ElementNode): readonly unknown[] {
  const children = node.props?.children;
  return Array.isArray(children) ? children : [children];
}

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}

type ElementNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};
