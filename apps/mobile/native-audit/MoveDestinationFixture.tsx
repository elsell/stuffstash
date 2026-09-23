import { useEffect, useState } from 'react';
import { AssetMoveSheetRouteScreen } from '../src/ui/screens/AssetNativeActionSheetScreens';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { AssetCoreQuery } from '../src/application/assets/AssetCoreQuery';
import type { CreateAssetCommandInput } from '../src/application/add/CreateAssetCommand';
import type { MoveAssetCommandInput } from '../src/application/assets/MoveAssetCommand';
import { assetId } from '../src/domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../src/domain/inventories/InventorySummary';

/** Real Move composition with deterministic recovery; never writes user data. */
export function MoveDestinationFixture() {
  const [fixture] = useState(() => {
    let creates = 0;
    let moves = 0;
    let created = false;
    const client = createMobileQueryClient();
    const defaults = client.getDefaultOptions();
    client.setDefaultOptions({ ...defaults, queries: { ...defaults.queries, retry: false } });
    const existing = { id: 'audit-box', title: 'Camping box', kind: 'container' as const,
      subtitle: 'Container', pathLabel: 'Inventory root', selectionHint: 'Container', willPromoteToContainer: false };
    return { client,
      lookup: { execute: async (query: string) => {
        const options = created ? [existing, { ...existing, id: 'audit-crate', title: 'Audit crate' }] : [existing];
        return options.filter(value => value.title.toLowerCase().includes(query.trim().toLowerCase()));
      } },
      create: { execute: async (input: CreateAssetCommandInput) => {
        if (input.title !== 'Audit crate' || input.kind !== 'container' || input.parentAssetId !== undefined || input.description !== '') {
          throw new Error('Unexpected audit destination payload');
        }
        if (++creates === 1) throw new Error('Audit destination temporarily unavailable');
        created = true;
        return { id: 'audit-crate', title: input.title, message: 'Created Audit crate.' };
      } },
      move: { execute: async (input: MoveAssetCommandInput) => {
        if (!created || input.assetId !== 'audit-tent' || input.parentAssetId !== 'audit-crate') {
          throw new Error('Unexpected audit move payload');
        }
        if (++moves === 1) throw new Error('Audit move temporarily unavailable');
        return { id: input.assetId, title: 'Audit tent', message: 'Moved Audit tent.' };
      } },
      core: new AssetCoreQuery({ getAssetCore: async () => ({
        tenantId: tenantId('audit-tenant'), inventoryId: inventoryId('audit-inventory'), permissions: ['edit_asset'], revision: 'audit',
        asset: { id: assetId('audit-tent'), title: 'Audit tent', kind: 'item', lifecycleState: 'active', description: '',
          locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
      }) })
    };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit" loadInventoryScope={async () => ({ tenantId: 'audit-tenant', inventoryId: 'audit-inventory' })}>
    <AssetMoveSheetRouteScreen assetId="audit-tent" assetCoreQuery={fixture.core} parentLookupQuery={fixture.lookup}
      createAssetCommand={fixture.create} moveAssetCommand={fixture.move} />
  </MobileServerStateProvider>;
}
