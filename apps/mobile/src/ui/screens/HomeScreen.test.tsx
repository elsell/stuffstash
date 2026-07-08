import { describe, expect, it, vi } from 'vitest';
import { CheckedOutAssetCard, RecentAssetCard } from './HomeScreen';

vi.mock('expo-router', () => ({
  router: {
    navigate: vi.fn(),
    push: vi.fn()
  },
  useFocusEffect: vi.fn()
}));

vi.mock('lucide-react-native', () => ({
  Settings: 'SettingsIcon'
}));

vi.mock('react-native-safe-area-context', () => ({
  SafeAreaView: 'SafeAreaView'
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Image: 'Image',
  Pressable: 'Pressable',
  RefreshControl: 'RefreshControl',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  TextInput: 'TextInput',
  Modal: 'Modal',
  View: 'View'
}));

describe('RecentAssetCard', () => {
  it('shows assigned tags on recently changed cards', () => {
    const card = RecentAssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: 'Garage drill',
        locationTrailLabel: 'Garage / Tools',
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Item',
        tags: [
          { id: 'tag-tools', label: 'Tools', color: '#2F80ED' },
          { id: 'tag-camping', label: 'Camping', color: '#2E7D32' }
        ]
      },
      onPress: vi.fn()
    });

    expect(collectText(card)).toEqual(expect.arrayContaining(['Tools', 'Camping']));
  });
});

describe('CheckedOutAssetCard', () => {
  it('exposes a one-tap return action without hiding the asset state', () => {
    const onReturn = vi.fn();
    const card = CheckedOutAssetCard({
      asset: {
        id: 'asset-drill',
        title: 'Cordless drill',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: 'Garage drill',
        locationTrailLabel: 'Garage / Tools',
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Item',
        checkedOutLabel: 'Checked out',
        tags: []
      },
      isReturning: false,
      onPress: vi.fn(),
      onReturn
    });

    expect(collectText(card)).toEqual(expect.arrayContaining(['Cordless drill', 'Checked out', 'Return']));

    const returnButton = findPressableWithText(card, 'Return');
    (returnButton?.props?.onPress as (() => void) | undefined)?.();

    expect(onReturn).toHaveBeenCalledTimes(1);
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

function childrenOf(node: ElementNode): readonly unknown[] {
  const children = node.props?.children;
  return Array.isArray(children) ? children : [children];
}

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}

function findPressableWithText(node: unknown, text: string): ElementNode | undefined {
  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.type === 'Pressable' && collectText(node).includes(text)) {
    return node;
  }

  for (const child of childrenOf(node)) {
    const match = findPressableWithText(child, text);
    if (match) {
      return match;
    }
  }

  return undefined;
}
