import { expect, it } from 'vitest';

it('localizes voice navigation and move notices without changing user titles or duplicate identity', async () => {
  const previous = process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
  process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = 'en-XA';
  try {
    const { createTimeoutFetch } = await import('../../adapters/network/TimeoutFetch');
    const pendingFetch: typeof fetch = async (_input, init) => new Promise((_resolve, reject) => {
      init?.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')), { once: true });
    });
    const { shouldRetryMobileQuery } = await import('../../adapters/serverState/MobileQueryClient');
    const timeout = await createTimeoutFetch(1, pendingFetch)('https://api.example.test').catch(error => error);
    expect(timeout.message).toMatch(/^\[/);
    expect(shouldRetryMobileQuery(0, timeout)).toBe(true);
    expect(shouldRetryMobileQuery(1, timeout)).toBe(false);
    expect(shouldRetryMobileQuery(0, new DOMException('Aborted', 'AbortError'))).toBe(false);
    const { voiceResponseEntityOpenLabel } = await import('../../ui/screens/VoiceResponseEntityLinks');
    const { MoveAssetCommand } = await import('../assets/MoveAssetCommand');
    const references = [
      { type: 'asset_reference' as const, assetId: 'one', title: 'İSTANBUL {box}', assetKind: 'item' as const, context: 'Garage' },
      { type: 'asset_reference' as const, assetId: 'two', title: 'İSTANBUL {box}', assetKind: 'item' as const, context: 'Garage' }
    ];
    const labels = references.map(reference => voiceResponseEntityOpenLabel(reference, references));
    expect(labels[0]).not.toBe(labels[1]);
    for (const label of labels) {
      expect(label).toContain('İSTANBUL {box}');
      expect(label).toContain('Garage');
      expect(label).toMatch(/^\[/);
      expect(label).not.toContain('Open ');
    }
    const { expirationDateLabel } = await import('../../ui/presentation/ExpirationPresentation');
    const expiration = expirationDateLabel({ date: '2028-02', precision: 'month' });
    expect(expiration).not.toContain('end of month');
    expect(expiration).toMatch(/^\[/);
    const requests: unknown[] = [];
    const command = new MoveAssetCommand({ async updateAsset(input) {
      requests.push(input);
      return { id: input.assetId, title: 'İSTANBUL {box}', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], description: '', updatedAtLabel: '', hasPhoto: false };
    } });
    const moved = await command.execute({ assetId: 'one' });
    expect(requests).toEqual([{ assetId: 'one', parentAssetId: null }]);
    expect(moved.title).toBe('İSTANBUL {box}');
    expect(moved.message).toContain('İSTANBUL {box}');
    expect(moved.message).toMatch(/^\[/);
    expect(moved.message).not.toContain('No parent');
  } finally {
    if (previous === undefined) delete process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE;
    else process.env.EXPO_PUBLIC_STUFF_STASH_UI_LOCALE = previous;
  }
});
