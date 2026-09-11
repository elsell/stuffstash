import { expect, it } from 'vitest';
import { expirationRefreshDelay } from './expirationRefreshDelay';
it('refreshes at the earliest inventory calendar boundary, including DST', () => {
 const now = new Date('2026-03-08T06:59:00Z');
 const delay = expirationRefreshDelay({ assets: [{ expirationContext: { timezone: 'America/New_York' } }] }, now);
 expect(typeof delay).toBe('number');
 expect(Number(delay)).toBeGreaterThanOrEqual(21 * 3600_000 + 60_000);
 expect(Number(delay)).toBeLessThan(21 * 3600_000 + 61_000);
 expect(expirationRefreshDelay({ assets: [] }, now)).toBe(false);
});

it('retries stale calendar status after a failed midnight refresh', () => {
 const data = { expirationContext: { timezone: 'America/New_York' } };
 expect(expirationRefreshDelay(data, new Date('2026-09-12T04:01:00Z'), new Date('2026-09-12T03:59:00Z'))).toBe(30_000);
});
