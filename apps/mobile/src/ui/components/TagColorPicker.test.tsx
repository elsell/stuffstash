import { describe, expect, it, vi } from 'vitest';
import { TagColorPicker } from './TagColorPicker';

vi.mock('lucide-react-native', () => ({
  Check: 'CheckIcon',
  X: 'XIcon'
}));

vi.mock('react-native', () => ({
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  View: 'View'
}));

type ElementNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};

describe('TagColorPicker', () => {
  it('offers swatches and clears the optional tag color', () => {
    const changes: string[] = [];
    const picker = TagColorPicker({
      value: '#2f80ed',
      onChange: (value) => {
        changes.push(value);
      }
    });

    const selectedBlue = findFirstByProp(picker, 'accessibilityLabel', 'Choose tag color #2F80ED');
    expect(selectedBlue?.props?.accessibilityState).toMatchObject({ selected: true });
    const green = findFirstByProp(picker, 'accessibilityLabel', 'Choose tag color #2E7D32');
    const clear = findFirstByProp(picker, 'accessibilityLabel', 'No tag color');

    press(green);
    press(clear);

    expect(changes).toEqual(['#2E7D32', '']);
    expect(collectText(picker)).toContain('Or type a hex color');
  });

  it('distinguishes invalid typed colors from no color', () => {
    const picker = TagColorPicker({
      value: 'blue',
      onChange: () => {}
    });

    const clear = findFirstByProp(picker, 'accessibilityLabel', 'No tag color');
    expect(clear?.props?.accessibilityState).toMatchObject({ selected: false });
    expect(clear?.props?.style).not.toContainEqual(expect.objectContaining({ borderWidth: 2 }));
    expect(collectText(picker)).toContain('Enter a #RRGGBB color');
  });
});

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
