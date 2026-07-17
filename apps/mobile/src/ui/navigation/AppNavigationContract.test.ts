import { describe, expect, it } from 'vitest';

// @ts-expect-error Vitest's Vite transform provides raw source imports to structural tests.
import homeScreenSource from '../screens/HomeScreen.tsx?raw';
// @ts-expect-error Vitest's Vite transform provides raw source imports to structural tests.
import browseScreenSource from '../screens/SearchScreen.tsx?raw';
// @ts-expect-error Vitest's Vite transform provides raw source imports to structural tests.
import voiceScreenSource from '../screens/VoiceSessionSheetScreen.tsx?raw';

// @ts-expect-error Vitest's Vite transform provides the app route manifest to structural tests.
const appSources = import.meta.glob('../../app/**/*.tsx', {
  eager: true,
  import: 'default',
  query: '?raw'
}) as Record<string, string>;

const tabLayoutSource = appSources['../../app/(tabs)/_layout.tsx'];
const rootLayoutSource = appSources['../../app/_layout.tsx'];

const nativeTabTriggerNames = (source: string): string[] =>
  [...source.matchAll(/<NativeTabs\.Trigger\s+name=["']([^"']+)["']/g)].map(
    ([, name]) => name
  );

describe('mobile navigation contract', () => {
  it('uses native tabs for Home and Browse navigation only', () => {
    expect(nativeTabTriggerNames(tabLayoutSource)).toEqual(['index', 'search']);
    expect(tabLayoutSource).not.toContain('name="add"');
    expect(tabLayoutSource).not.toContain('name="settings"');
  });

  it('keeps Voice attached as the native tab bottom accessory', () => {
    expect(tabLayoutSource).toContain('<NativeTabs.BottomAccessory>');
    expect(tabLayoutSource).toContain('<VoiceBottomAccessory />');
    expect(tabLayoutSource).toContain('</NativeTabs.BottomAccessory>');
  });

  it('owns Add as a non-tab stack route', () => {
    expect(appSources).toHaveProperty('../../app/add.tsx');
    expect(appSources).not.toHaveProperty('../../app/(tabs)/add.tsx');
    expect(rootLayoutSource).toMatch(/<Stack\.Screen\s+name=["']add["']/);
  });

  it('keeps Settings as a non-tab stack route', () => {
    expect(nativeTabTriggerNames(tabLayoutSource)).not.toContain('settings');
    expect(rootLayoutSource).toMatch(/<Stack\.Screen\s+name=["']settings\/index["']/);
    expect(appSources).toHaveProperty('../../app/settings/account.tsx');
    expect(appSources).toHaveProperty('../../app/settings/appearance.tsx');
    expect(appSources).toHaveProperty('../../app/settings/connection.tsx');
    expect(appSources).toHaveProperty('../../app/settings/voice/index.tsx');
    expect(appSources).toHaveProperty('../../app/settings/voice/profiles/index.tsx');
    expect(appSources).toHaveProperty('../../app/settings/voice/profiles/add.tsx');
    expect(appSources).not.toHaveProperty('../../app/settings.tsx');
    expect(voiceScreenSource).toContain("router.push('/settings/voice')");
  });

  it('opens grounded voice response entities through the asset detail route', () => {
    expect(voiceScreenSource).toContain("import { assetDetailHref } from './AssetDetailNavigation'");
    expect(voiceScreenSource).toContain('router.push(assetDetailHref(artifact.assetId))');
    expect(voiceScreenSource).not.toMatch(/artifact\.(?:href|url|route)/);
  });

  it('keeps invitation acceptance outside the tab hierarchy', () => {
    expect(appSources).toHaveProperty('../../app/invitations/accept.tsx');
    expect(rootLayoutSource).toMatch(/<Stack\.Screen\s+name=["']invitations\/accept["']/);
    expect(nativeTabTriggerNames(tabLayoutSource)).not.toContain('invitations');
  });

  it('makes Add reachable from both primary browsing destinations', () => {
    expect(homeScreenSource).toMatch(/router\.(?:push|navigate)\(["']\/add["']\)/);
    expect(browseScreenSource).toMatch(/router\.(?:push|navigate)\(["']\/add["']\)/);
    expect(browseScreenSource).toMatch(/<InventoryMapScreen[\s\S]*?canAdd=\{inventoryContext\?\.canAdd \?\? false\}/);
  });
});
