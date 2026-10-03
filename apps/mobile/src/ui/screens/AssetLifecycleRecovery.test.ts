import { expect, it } from 'vitest';
import { t } from '../../presentation/localization';
import { assetLifecycleFailurePresentation } from './AssetLifecyclePresentation';

it.each([
  ['archive', 'private diagnostic: active children', 'mobile.AssetLifecyclePresentation.moveOrArchiveActiveThingsInsideThisAssetThen'],
  ['restore', 'private diagnostic: parent is archived', 'mobile.AssetLifecyclePresentation.checkThatItsParentIsActiveThenTryAgain'],
  ['delete', 'private diagnostic: active children', 'mobile.AssetLifecyclePresentation.permanentDeleteWillNotContinueWhileActiveThingsAre'],
  ['archive', 'private transport diagnostic', 'mobile.AssetDetailRouteScreen.lifecycleActionFailed']
] as const)('localizes %s recovery without echoing diagnostics', (action, diagnostic, key) => {
  const result = assetLifecycleFailurePresentation(action, { title: 'Tool box', canContainAssets: true }, new Error(diagnostic));
  expect(result.message).toBe(t(key));
  expect(result.message).not.toContain('private');
  expect(result.title).toContain('Tool box');
});
