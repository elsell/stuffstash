import { describe, expect, it, vi } from 'vitest';
import { VoiceResponseEntityText } from './VoiceResponseEntityText';

vi.mock('react-native', () => ({
  Pressable: 'Pressable',
  StyleSheet: { create: (styles: unknown) => styles, hairlineWidth: 0.5 },
  Text: 'Text',
  View: 'View'
}));

vi.mock('../theme/AppearanceContext', () => ({
  useAppearancePalette: () => ({
    accentStrong: '#123456',
    border: '#cccccc',
    surface: '#ffffff',
    text: '#111111'
  })
}));

describe('VoiceResponseEntityText', () => {
  it('renders distinguishable duplicate controls and opens the intended assets', () => {
    const onOpen = vi.fn();
    const tree = VoiceResponseEntityText({
      enabled: true,
      onOpen,
      references: [
        { type: 'asset_reference', assetId: 'garage-drill', title: 'Drill', assetKind: 'item', context: 'Garage toolbox' },
        { type: 'asset_reference', assetId: 'basement-drill', title: 'Drill', assetKind: 'item', context: 'Basement cabinet' }
      ],
      text: 'Did you mean Drill?'
    });

    const buttons = findNodes(tree, 'Pressable');
    expect(buttons.map((button) => button.props?.accessibilityLabel)).toEqual([
      'Open Drill in Garage toolbox',
      'Open Drill in Basement cabinet'
    ]);
    buttons[0].props?.onPress?.();
    buttons[1].props?.onPress?.();
    expect(onOpen.mock.calls.map(([artifact]) => artifact.assetId)).toEqual(['garage-drill', 'basement-drill']);
  });

  it('does not announce disabled inline text as an Open link', () => {
    const tree = VoiceResponseEntityText({
      enabled: false,
      onOpen: vi.fn(),
      references: [{ type: 'asset_reference', assetId: 'drill', title: 'Drill', assetKind: 'item' }],
      text: 'The Drill is here.'
    });

    const linkedTitle = findNodes(tree, 'Text').find((node) => node.props?.children === 'Drill');
    expect(linkedTitle?.props?.accessibilityLabel).toBeUndefined();
    expect(linkedTitle?.props?.accessibilityRole).toBeUndefined();
    expect(linkedTitle?.props?.onPress).toBeUndefined();
  });
});

type TestNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly accessibilityLabel?: string;
    readonly accessibilityRole?: string;
    readonly onPress?: () => void;
  };
};

function findNodes(node: unknown, type: string): TestNode[] {
  if (Array.isArray(node)) {
    return node.flatMap((child) => findNodes(child, type));
  }
  if (!node || typeof node !== 'object') {
    return [];
  }
  const current = node as TestNode;
  return [
    ...(current.type === type ? [current] : []),
    ...findNodes(current.props?.children, type)
  ];
}
