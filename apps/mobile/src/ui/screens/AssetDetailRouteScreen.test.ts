import { describe, expect, it } from 'vitest';
import { navigateAfterDeletedAsset } from './AssetDetailNavigation';

describe('navigateAfterDeletedAsset', () => {
  it('uses native back navigation when the asset detail route has history', () => {
    const calls: string[] = [];

    navigateAfterDeletedAsset({
      canGoBack: () => true,
      back: () => {
        calls.push('back');
      },
      replace: (href) => {
        calls.push(`replace:${href.toString()}`);
      }
    });

    expect(calls).toEqual(['back']);
  });

  it('replaces with Home when the deleted asset route has no back stack', () => {
    const calls: string[] = [];

    navigateAfterDeletedAsset({
      canGoBack: () => false,
      back: () => {
        calls.push('back');
      },
      replace: (href) => {
        calls.push(`replace:${href.toString()}`);
      }
    });

    expect(calls).toEqual(['replace:/']);
  });
});
