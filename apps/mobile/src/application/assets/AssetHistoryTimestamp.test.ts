import { expect, it } from 'vitest';
import { formatHistoryTimestamp } from './AssetHistoryTimestamp';

// Construct a local instant so these expectations are independent of runner zone.
const value = new Date(2026, 5, 25, 12, 30, 0).toISOString();
it('uses British date order and a 24-hour clock for checkout history', () => {
  expect(formatHistoryTimestamp(value, 'checkout', 'en-GB')).toBe('25 Jun 2026, 12:30');
});
it('keeps US conventions when that locale is selected', () => {
  expect(formatHistoryTimestamp(value, 'checkout', 'en-US')).toBe('Jun 25, 2026, 12:30 PM');
});
it('keeps activity brief and event detail precise', () => {
  expect(formatHistoryTimestamp(value, 'activity', 'en-GB')).toBe('25 Jun, 12:30');
  expect(formatHistoryTimestamp(value, 'exact', 'en-GB')).toContain('25 June 2026 at 12:30:00');
});
it('does not crash or invent a date for invalid history timestamps', () => {
  for (const detail of ['activity', 'checkout', 'exact'] as const) {
    expect(formatHistoryTimestamp('invalid timestamp', detail, 'en-GB')).toBe('invalid timestamp');
  }
});
it('uses the actual runtime locale when callers omit an override', () => {
  const locale = new Intl.DateTimeFormat().resolvedOptions().locale;
  expect(formatHistoryTimestamp(value, 'checkout')).toBe(formatHistoryTimestamp(value, 'checkout', locale));
});
