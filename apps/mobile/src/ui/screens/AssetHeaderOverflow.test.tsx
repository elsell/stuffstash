import { describe, expect, it, vi } from 'vitest';
import { AssetOverflowMenu } from './AssetOverflowMenu';
import { assetHeaderOverflowScreenOptions } from './AssetHeaderOverflow';

const props = {
  asset: { title: 'Drill', canArchive: true, canRestore: false, canDeletePermanently: false },
  disabled: false,
  onCheckoutHistory: vi.fn(),
  onHistory: vi.fn(),
  onLifecycleAction: vi.fn()
};

describe('AssetHeaderOverflow platform contract', () => {
  it('keeps the shared anchored menu in Android headerRight', () => {
    const options = assetHeaderOverflowScreenOptions(props);
    const menu = options.headerRight();

    expect(menu.type).toBe(AssetOverflowMenu);
    expect(menu.props).toEqual(props);
  });
});
