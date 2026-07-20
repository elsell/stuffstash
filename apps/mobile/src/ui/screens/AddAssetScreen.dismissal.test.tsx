import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import { CreateAssetCommand } from '../../application/add/CreateAssetCommand';
import { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { HomeDashboardQuery } from '../../application/home/HomeDashboardQuery';
import { AddAssetScreen } from './AddAssetScreen';

const testState = vi.hoisted(() => ({ stateIndex: 0 }));

vi.mock('react', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react')>()),
  useEffect: vi.fn(),
  useMemo: <T,>(factory: () => T) => factory(),
  useRef: <T,>(value: T) => ({ current: value }),
  useState: <T,>(initialValue: T) => {
    testState.stateIndex += 1;
    return [initialValue, vi.fn()];
  }
}));

vi.mock('expo-router', () => ({
  router: { push: vi.fn() }
}));

vi.mock('lucide-react-native', () => ({
  Check: 'CheckIcon',
  ChevronDown: 'ChevronDownIcon',
  ChevronUp: 'ChevronUpIcon',
  ImagePlus: 'ImagePlusIcon',
  X: 'XIcon'
}));

vi.mock('react-native-safe-area-context', () => ({
  SafeAreaView: 'SafeAreaView',
  useSafeAreaInsets: () => ({ bottom: 0, left: 0, right: 0, top: 0 })
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Alert: { alert: vi.fn() },
  Image: 'Image',
  Keyboard: {
    addListener: vi.fn(() => ({ remove: vi.fn() })),
    dismiss: vi.fn()
  },
  PanResponder: { create: vi.fn(() => ({ panHandlers: {} })) },
  Platform: { OS: 'ios' },
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: { create: <T,>(styles: T) => styles },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View'
}));

vi.mock('../feedback/AppFeedback', () => ({
  useAppFeedback: () => ({ showDialog: vi.fn(), showNotice: vi.fn() })
}));

vi.mock('../components/FullScreenPhotoViewer', () => ({
  FullScreenPhotoViewer: 'FullScreenPhotoViewer'
}));

vi.mock('../components/FullSpectrumTagColorPicker', () => ({
  FullSpectrumTagColorPicker: 'FullSpectrumTagColorPicker'
}));

vi.mock('../theme/appearance', () => ({
  useAppearanceAwarePalette: () => ({
    accent: '#6B90AA',
    action: '#0066CC',
    background: '#F7FAFB',
    border: '#C5D0D7',
    onAction: '#FFFFFF',
    surface: '#FFFFFF',
    surfaceMuted: '#E8F0F5',
    text: '#243038',
    textMuted: '#52616B'
  })
}));

describe('AddAssetScreen dismissal', () => {
  beforeEach(() => {
    testState.stateIndex = 0;
  });

  it('offers an explicit accessible Close Add control that invokes the dismissal callback', () => {
    const onDismiss = vi.fn();
    const tree = AddAssetScreen({
      addAssetDraftStore: { load: vi.fn(), save: vi.fn() },
      addDraftScopeQuery: new AddDraftScopeQuery(undefined as never),
      createAssetCommand: new CreateAssetCommand(undefined as never),
      dashboardQuery: new HomeDashboardQuery(undefined as never),
      onDismiss,
      parentLookupQuery: new ParentLookupQuery(undefined as never),
      photoSelectionQuery: new PhotoSelectionQuery(undefined as never)
    });
    const close = findByAccessibilityLabel(tree, 'Close Add');

    expect(close?.props?.accessibilityRole).toBe('button');
    expect(styleValue(close?.props?.style, 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(styleValue(close?.props?.style, 'minWidth')).toBeGreaterThanOrEqual(44);

    (close?.props?.onPress as (() => void) | undefined)?.();

    expect(onDismiss).toHaveBeenCalledOnce();
  });
});

type TestNode = {
  readonly props?: Record<string, unknown> & { readonly children?: unknown };
  readonly type?: unknown;
};

function findByAccessibilityLabel(node: unknown, label: string): TestNode | undefined {
  if (!node || typeof node !== 'object') {
    return undefined;
  }
  const candidate = node as TestNode;
  if (candidate.props?.accessibilityLabel === label) {
    return candidate;
  }
  const children = candidate.props?.children;
  const childList = Array.isArray(children) ? children : [children];
  for (const child of childList) {
    const match = findByAccessibilityLabel(child, label);
    if (match) {
      return match;
    }
  }
  return undefined;
}

function styleValue(style: unknown, property: string): number {
  const entries = Array.isArray(style) ? style : [style];
  for (const entry of entries.toReversed()) {
    if (entry && typeof entry === 'object' && property in entry) {
      return (entry as Record<string, number>)[property];
    }
  }
  return 0;
}
