import { useEffect, useState } from 'react';
import { Text } from 'react-native';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { CustomizationAccessPolicy } from '../src/application/customization/CustomizationAccess';
import { CustomizationContextQuery } from '../src/application/customization/CustomizationContextQuery';
import { noCustomizationObservability } from '../src/application/customization/CustomizationObservability';
import { CustomizationCollectionQuery } from '../src/application/customization/CustomizationQueries';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { CustomizationCollectionScreen } from '../src/ui/screens/CustomizationCollectionScreen';

const scope = { tenantId: 'audit-household', inventoryId: 'audit-inventory' };
const unavailable = async (): Promise<never> => { throw new Error('This collection fixture does not mutate definitions'); };

export function CustomizationCollectionFixture() {
  const [result, setResult] = useState('');
  const [state] = useState(() => ({
    client: createMobileQueryClient(),
    policy: new CustomizationAccessPolicy(noCustomizationObservability),
    context: new CustomizationContextQuery({ async getSelectedScope() {
      return { tenant: { id: scope.tenantId, name: 'Audit household', permissions: ['configure'] },
        inventory: { id: scope.inventoryId, name: 'Audit inventory', permissions: ['view', 'edit_asset'] } };
    } }),
    query: new CustomizationCollectionQuery({
      async listTags() { return { items: ['Garden', 'Tools', ...Array.from({ length: 20 }, (_, index) => `Storage ${index + 1}`)]
        .map(displayName => ({ kind: 'tag' as const, id: displayName, key: displayName, displayName })) }; },
      async listFields() { return { items: [] }; }, async listAssetTypes() { return { items: [] }; },
      createTag: unavailable, updateTag: unavailable, archiveTag: unavailable,
      createField: unavailable, updateField: unavailable, archiveField: unavailable, restoreField: unavailable, deleteField: unavailable,
      createAssetType: unavailable, updateAssetType: unavailable, archiveAssetType: unavailable, restoreAssetType: unavailable, deleteAssetType: unavailable
    })
  }));
  useEffect(() => () => state.client.clear(), [state]);
  return <MobileServerStateProvider client={state.client} scopeId="audit-customization" loadInventoryScope={async () => scope}>
    {result ? <Text accessibilityLiveRegion="polite">{result}</Text> : null}
    <CustomizationCollectionScreen accessPolicy={state.policy} contextQuery={state.context} query={state.query} kind="tag" scope="inventory"
      onAdd={() => setResult('Add tag requested')} onOpen={row => setResult(`Opened ${row.displayName}`)} />
  </MobileServerStateProvider>;
}
