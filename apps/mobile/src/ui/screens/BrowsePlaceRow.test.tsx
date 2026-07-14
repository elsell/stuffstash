import { describe, expect, it, vi } from 'vitest';
import { BrowsePlaceRow } from './BrowsePlaceRow';
import { lightPalette } from '../theme/tokens';

vi.mock('react-native', () => ({
  DynamicColorIOS: undefined,
  Image: 'Image',
  Platform: { OS: 'android' },
  Pressable: 'Pressable',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  useColorScheme: () => 'light',
  View: 'View'
}));

describe('BrowsePlaceRow', () => {
  it('keeps useful place context without exposing photo-maintenance metadata', () => {
    const onPress = vi.fn();
    const row = BrowsePlaceRow({
      location: {
        id: 'location-garage',
        title: 'Garage',
        description: 'Tools and seasonal storage',
        containedAssetCountLabel: '8 assets',
        recentAssetLabel: 'Cordless drill, socket set'
      },
      palette: lightPalette,
      onPress
    });

    expect(collectText(row)).toEqual(expect.arrayContaining([
      'Place',
      'Garage',
      'Tools and seasonal storage',
      '8 assets',
      'Cordless drill, socket set'
    ]));
    expect(collectText(row)).not.toContain('Needs photo');
    expect(row.props?.accessibilityLabel).toBe('Open place Garage, 8 assets');
    expect(row.props?.accessibilityHint).toBe(
      'Tools and seasonal storage. Cordless drill, socket set'
    );
    expect(row.props?.accessibilityRole).toBe('button');
    expect(styleValue(row.props?.style({ pressed: false }), 'minHeight')).toBeGreaterThanOrEqual(112);

    row.props?.onPress?.();
    expect(onPress).toHaveBeenCalledTimes(1);
  });

  it('renders the place photo and exposes a visible pressed state', () => {
    const row = BrowsePlaceRow({
      location: {
        id: 'location-kitchen',
        title: 'Kitchen',
        description: '',
        containedAssetCountLabel: '12 assets',
        recentAssetLabel: 'Travel mug',
        photo: {
          uri: 'https://photos.example/kitchen.jpg',
          headers: { Authorization: 'Bearer test' }
        }
      },
      palette: lightPalette,
      onPress: vi.fn()
    });

    expect(findFirstByType(row, 'Image')?.props?.source).toEqual({
      uri: 'https://photos.example/kitchen.jpg',
      headers: { Authorization: 'Bearer test' }
    });
    expect(collectText(row)).not.toContain('Photo ready');
    expect(styleValue(row.props?.style({ pressed: true }), 'backgroundColor')).not.toBe(
      styleValue(row.props?.style({ pressed: false }), 'backgroundColor')
    );
  });
});

type TestNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};

function collectText(node: unknown): string[] {
  if (typeof node === 'string') {
    return [node];
  }
  if (Array.isArray(node)) {
    return node.flatMap(collectText);
  }
  if (!node || typeof node !== 'object') {
    return [];
  }
  return collectText((node as TestNode).props?.children);
}

function findFirstByType(node: unknown, type: unknown): TestNode | undefined {
  if (!node || typeof node !== 'object') {
    return undefined;
  }
  if (Array.isArray(node)) {
    for (const child of node) {
      const match = findFirstByType(child, type);
      if (match) return match;
    }
    return undefined;
  }
  const element = node as TestNode;
  if (element.type === type) {
    return element;
  }
  return findFirstByType(element.props?.children, type);
}

function styleValue(style: unknown, key: string): unknown {
  const styles = Array.isArray(style) ? style : [style];
  return styles.reduce<unknown>((value, entry) => {
    if (entry && typeof entry === 'object' && key in entry) {
      return (entry as Record<string, unknown>)[key];
    }
    return value;
  }, undefined);
}
