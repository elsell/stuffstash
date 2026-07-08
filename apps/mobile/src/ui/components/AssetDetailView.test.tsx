import { describe, expect, it, vi } from 'vitest';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetDetailView } from './AssetDetailView';

vi.mock('lucide-react-native', () => ({
  Camera: 'CameraIcon',
  CheckCircle2: 'CheckCircle2Icon',
  ChevronRight: 'ChevronRightIcon',
  MoreHorizontal: 'MoreHorizontalIcon',
  MoveRight: 'MoveRightIcon',
  Pencil: 'PencilIcon',
  Plus: 'PlusIcon'
}));

vi.mock('react-native', () => ({
  Image: 'Image',
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  View: 'View'
}));

describe('AssetDetailView', () => {
  it('makes detail tag chips search actions', () => {
    const searchedTags: string[] = [];
    const tree = AssetDetailView({
      asset: assetDetail(),
      onTagPress: (tag) => {
        searchedTags.push(tag.label);
      }
    });

    const tagChip = findFirstByProp(tree, 'accessibilityLabel', 'Search for tag Camping');
    expect(tagChip?.props?.accessibilityRole).toBe('button');
    press(tagChip);

    expect(searchedTags).toEqual(['Camping']);
  });
});

function assetDetail(): AssetDetailViewModel {
  return {
    id: 'asset-tent',
    title: 'Family tent',
    kind: 'item',
    kindLabel: 'Item',
    description: 'Sleeps four.',
    locationTrailLabel: 'Garage / Camp bin',
    parentLocationTrailLabel: 'Garage / Camp bin',
    lifecycleLabel: 'Active',
    isActive: true,
    canEdit: true,
    canMove: true,
    canAddPhotos: true,
    canArchive: true,
    canRestore: false,
    canDeletePermanently: false,
    isCheckedOut: false,
    checkoutLabel: 'Available',
    tags: [{ id: 'tag-camping', label: 'Camping', color: '#2F80ED' }],
    canCheckout: true,
    canReturn: false,
    containedAssets: [],
    containedAssetsLabel: '0 things inside',
    canContainAssets: false,
    canAddContainedAssets: false,
    updatedAtLabel: 'Updated today',
    photoLabel: 'Needs photo',
    imagePlaceholderLabel: 'Item',
    photos: []
  };
}

function press(node: ElementNode | undefined): void {
  const onPress = node?.props?.onPress;
  if (typeof onPress !== 'function') {
    throw new Error('Missing press handler');
  }
  onPress();
}

function findFirstByProp(node: unknown, prop: string, value: unknown): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstByProp(child, prop, value),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.props?.[prop] === value) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstByProp(node.type(node.props), prop, value);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByProp(child, prop, value),
    undefined
  );
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
