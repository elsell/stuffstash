import { describe, expect, it } from 'vitest';
import { NativeInvitationLinkActions } from './NativeInvitationLinkActions';

// @ts-expect-error Vitest provides raw adapter sources to structural boundary tests.
const expoAdapterSources = import.meta.glob('./ExpoInvitationLinkActions.ts', {
  eager: true,
  import: 'default',
  query: '?raw'
}) as Record<string, string>;

describe('NativeInvitationLinkActions', () => {
  it('copies only after an explicit action and uses the native share sheet', async () => {
    const copied: string[] = [];
    const shared: unknown[] = [];
    const actions = new NativeInvitationLinkActions(
      { setStringAsync: async (value) => { copied.push(value); return true; } },
      { share: async (content) => { shared.push(content); return { action: 'sharedAction' }; } }
    );
    const link = 'https://stash.example/invitations/accept#token=secret';
    expect(copied).toEqual([]);
    await actions.copy(link);
    await actions.share({ link, inventoryName: 'Household' });
    expect(copied).toEqual([link]);
    expect(shared).toEqual([{
      message: `You’re invited to Household in Stuff Stash.\n\n${link}`,
      title: 'Share Stuff Stash invitation'
    }]);
  });

  it('wires the production adapter to Expo Clipboard and the native Share sheet', () => {
    const source = Object.values(expoAdapterSources)[0] ?? '';
    expect(source).toContain("import('expo-clipboard')");
    expect(source).toContain('clipboard.setStringAsync(link)');
    expect(source).toContain('Share.share');
  });
});
