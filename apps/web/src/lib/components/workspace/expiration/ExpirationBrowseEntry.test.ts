import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import ExpirationBrowseEntry from './ExpirationBrowseEntry.svelte';
import { expirationWorkspaceContext } from '$lib/ports/expirationRepository';
import { t } from '$lib/presentation/localization';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); document.body.innerHTML = ''; });
it('catalogs each expiration link while retaining its browse context', () => {
  component = mount(ExpirationBrowseEntry, {
    target: document.body,
    context: new Map([[expirationWorkspaceContext, {}]]),
    props: { tenantId: 'tenant', inventoryId: 'inventory', query: 'flashlight', tagIds: [], scope: 'items', checkoutState: 'any' },
  });
  const links = [...document.querySelectorAll('a')];
  expect(links.map(link => link.textContent?.trim())).toEqual([
    t('web.ExpirationBrowseEntry.expiringSoon'), t('web.ExpirationBrowseEntry.expired'), t('web.ExpirationBrowseEntry.allDates'),
  ]);
  expect(links).toHaveLength(3);
  for (const [index, link] of links.entries()) {
    const target = new URL(link.href);
    expect(target.searchParams.get('q')).toBe('flashlight');
    expect(target.searchParams.get('kind')).toBe('item');
    expect(target.searchParams.get('expiration')).toBe(['soon', 'expired', 'all'][index]);
  }
});
