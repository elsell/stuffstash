import { useEffect, useState } from 'react';
import { createSettingsReadback } from './SettingsReadbackFixture';
import { CustomizationCollectionScreen } from '../src/ui/screens/CustomizationCollectionScreen';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';

/** Shared collection layout with real queries; editor behavior has separate coverage. */
export function CustomizationCollectionJourneyFixture() {
  const [state] = useState(() => createSettingsReadback(Array.from({ length: 40 }, (_, index) => ({
    kind: 'tag' as const, id: `tag-${index}`, key: `tag-${index}`,
    displayName: index === 39 ? 'Tag 40 with a long descriptive household storage name' : `Tag ${String(index + 1).padStart(2, '0')}`
  }))));
  useEffect(() => () => state.client.clear(), [state]);
  const unavailable = () => { throw new Error('Use the connected settings editor fixture for mutation tasks.'); };
  return <MobileServerStateProvider client={state.client} scopeId="settings-readback"
    loadInventoryScope={async () => ({ tenantId: 'readback-household', inventoryId: 'readback-inventory' })}>
    <CustomizationCollectionScreen accessPolicy={state.policy} contextQuery={state.context}
      query={state.query} kind="tag" scope="inventory" onAdd={unavailable} onOpen={unavailable} />
  </MobileServerStateProvider>;
}
