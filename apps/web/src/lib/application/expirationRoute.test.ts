import { expect, it } from 'vitest';
import { parseWorkspaceRoute, workspaceRouteHref } from './workspaceRoute';
it('round trips scoped expiration modes and combined filters in reloadable links', () => {
 const href = workspaceRouteHref({ mode: 'expiration', expirationFilter: { mode: 'expired', query: 'ear drops', typeId: 'medicine', tagIds: ['cold', 'kids'], locationId: 'cabinet', fromDate: '2026-01-01', throughDate: '2026-09-30' } }, 'tenant', 'inventory');
 expect(href).toContain('/expiration?');
 const route = parseWorkspaceRoute(new URL(href, 'https://stash.test'));
 expect(route.mode).toBe('expiration'); expect(route.expirationFilter).toEqual({ mode: 'expired', query: 'ear drops', typeId: 'medicine', tagIds: ['cold', 'kids'], locationId: 'cabinet', fromDate: '2026-01-01', throughDate: '2026-09-30' });
});
