import { useEffect, useState } from 'react';
import { useLocalSearchParams } from 'expo-router';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AssetActivityQuery, type AssetActivityEntry } from '../src/application/assets/AssetActivityQuery';
import { RevertAssetChangeCommand } from '../src/application/assets/RevertAssetChangeCommand';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { AssetHistoryRouteScreen } from '../src/ui/screens/AssetHistoryRouteScreen';
import { AssetHistoryDetailRouteScreen } from '../src/ui/screens/AssetHistoryDetailRouteScreen';

const scope = { tenantId: 'history-tenant', inventoryId: 'history-inventory', assetId: 'history-item', assetTitle: 'Camping equipment' };
const entries: readonly AssetActivityEntry[] = Array.from({ length: 24 }, (_, index) => ({
  id: `history-${index}`, principalId: 'history-member',
  principal: { id: 'history-member', email: 'household.member@example.invalid' },
  action: 'asset.updated', category: 'change', source: 'mobile',
  occurredAt: new Date(Date.UTC(2026, 8, 24, 12, 0, -index * 60)).toISOString(),
  changes: [
    { field: 'title', previousValue: `Camping equipment ${index}`, currentValue: `Updated camping equipment ${index}` },
    { field: 'description', previousValue: 'Stored with the outdoor equipment.', currentValue: 'Checked and packed for our next camping trip. Keep the spare parts together in the garage.' },
    { field: 'parent', previousValue: 'Garage', currentValue: 'Wooden shelf' },
    { field: 'tags', previousValue: 'camping', currentValue: 'camping, outdoors' }
  ],
  requestId: `history-request-${index}`, technical: {}
}));

export function HistoryListFixture() { return <HistoryJourneyFixture page="list" />; }
export function HistoryDetailFixture() { return <HistoryJourneyFixture page="detail" />; }
function HistoryJourneyFixture({ page }: { readonly page: 'list' | 'detail' }) {
  const { activityId = 'history-0' } = useLocalSearchParams<{ activityId?: string }>();
  const [fixture] = useState(() => ({
    client: createMobileQueryClient(),
    query: new AssetActivityQuery({ listAssetActivity: async ({ cursor, limit }) => {
      const offset = Number(cursor ?? 0);
      const page = entries.slice(offset, offset + limit);
      const hasMore = offset + page.length < entries.length;
      return { entries: page, hasMore, ...(hasMore ? { nextCursor: String(offset + page.length) } : {}) };
    } }),
    revert: new RevertAssetChangeCommand({ reverseAssetOperation: async () => { throw new Error('Revert is outside this layout fixture.'); } })
  }));
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="history-layout"
    loadInventoryScope={async () => ({ tenantId: scope.tenantId, inventoryId: scope.inventoryId })}>
    {page === 'list' ? <AssetHistoryRouteScreen {...scope} assetActivityQuery={fixture.query} />
      : <AssetHistoryDetailRouteScreen {...scope} activityId={activityId} assetActivityQuery={fixture.query} revertAssetChangeCommand={fixture.revert} />}
  </MobileServerStateProvider>;
}
