import { describe, expect, it, vi } from 'vitest';
import { AssetDetailRouteErrorState } from './AssetDetailRouteErrorState';

vi.mock('../theme/appearance', () => ({
  useAppearanceAwarePalette: () => ({ action: '#0066CC', text: '#243038', textMuted: '#52616B' })
}));

vi.mock('react-native', () => ({
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: { create: (styles: unknown) => styles },
  Text: 'Text'
}));

describe('AssetDetailRouteErrorState', () => {
  it('scrolls and grows error content instead of clipping accessibility text to the viewport', () => {
    const retry = vi.fn();
    const tree = AssetDetailRouteErrorState({
      canRetry: true,
      message: 'A long error explanation that may wrap across several lines.',
      onRetry: retry,
      title: 'Unable to load this place'
    });

    expect(tree.type).toBe('ScrollView');
    expect(styleValue(tree.props.contentContainerStyle, 'flexGrow')).toBe(1);
    expect(styleValue(tree.props.contentContainerStyle, 'height')).toBeUndefined();
    expect(findFirstTextNode(tree, 'Unable to load this place')?.props?.style?.lineHeight).toBeUndefined();
    expect(findFirstTextNode(tree, 'A long error explanation that may wrap across several lines.')?.props?.style?.lineHeight)
      .toBeUndefined();
    findFirstByProp(tree, 'accessibilityRole', 'button')?.props?.onPress();
    expect(retry).toHaveBeenCalledOnce();
  });
});

type ElementNode = { readonly type: unknown; readonly props: Record<string, any> };

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}

function childrenOf(node: ElementNode): readonly unknown[] {
  const children = node.props.children;
  return Array.isArray(children) ? children : [children];
}

function findFirstTextNode(node: unknown, text: string): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>((found, child) => found ?? findFirstTextNode(child, text), undefined);
  }
  if (!isElementNode(node)) return undefined;
  if (node.type === 'Text' && childrenOf(node).includes(text)) return node;
  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstTextNode(child, text),
    undefined
  );
}

function findFirstByProp(node: unknown, prop: string, value: unknown): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>((found, child) => found ?? findFirstByProp(child, prop, value), undefined);
  }
  if (!isElementNode(node)) return undefined;
  if (node.props[prop] === value) return node;
  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByProp(child, prop, value),
    undefined
  );
}

function styleValue(style: unknown, key: string): unknown {
  if (Array.isArray(style)) {
    return style.reduce<unknown>((found, entry) => found ?? styleValue(entry, key), undefined);
  }
  return style && typeof style === 'object' ? (style as Record<string, unknown>)[key] : undefined;
}
