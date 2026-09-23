import { expect, it } from 'vitest';
import { createBrowseFilterJourney } from './BrowseFilterJourneyFixture';

it('filters real fixture data and preserves query semantics across Browse and expiration', async () => {
  const journey = createBrowseFilterJourney();
  const browse = await journey.search.execute({ query: 'Camping', kind: 'all', lifecycleState: 'active', checkoutState: 'available', sort: 'updated_desc', tagIds: [] });
  expect(browse.assets.length).toBeGreaterThan(12);
  expect(browse.assets.every(asset => !asset.checkedOutLabel)).toBe(true);
  const missing = await journey.search.execute({ query: 'No match', kind: 'all', lifecycleState: 'active', checkoutState: 'any', sort: 'updated_desc' });
  expect(missing.assets).toEqual([]);
  const expired = await journey.expiration.list('filter-tenant', 'filter-inventory', { mode: 'expired', query: 'Camping', checkoutState: 'available' });
  expect(expired.items.length).toBeGreaterThan(0);
  expect(expired.items.some(asset => asset.title === 'Kitchen item')).toBe(false);
  const unsearched = await journey.expiration.list('filter-tenant', 'filter-inventory', { mode: 'expired', checkoutState: 'available' });
  expect(unsearched.items.some(asset => asset.title === 'Kitchen item')).toBe(true);
  expect(expired.items.length).toBeLessThan(browse.assets.length);
  expect(expired.items.every(asset => asset.expiration?.date === '2026-01-01')).toBe(true);
});

it('rejects foreign inventory reads and unknown asset details', async () => {
  const journey = createBrowseFilterJourney();
  await expect(journey.expiration.list('another-tenant', 'filter-inventory', { mode: 'all' })).rejects.toThrow();
  await expect(journey.expiration.list('filter-tenant', 'another-inventory', { mode: 'all' })).rejects.toThrow();
  await expect(journey.core.execute('not-in-fixture')).rejects.toThrow();
});
