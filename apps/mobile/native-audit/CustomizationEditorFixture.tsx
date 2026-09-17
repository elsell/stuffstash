import { useEffect, useState } from 'react';
import { router, type Href } from 'expo-router';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { CustomizationAccessPolicy } from '../src/application/customization/CustomizationAccess';
import { CustomizationContextQuery } from '../src/application/customization/CustomizationContextQuery';
import { noCustomizationObservability } from '../src/application/customization/CustomizationObservability';
import { CustomizationCollectionQuery } from '../src/application/customization/CustomizationQueries';
import type { CustomizationRepository } from '../src/application/customization/CustomizationRepository';
import { ManageTags } from '../src/application/customization/ManageTags';
import { ManageCustomFields } from '../src/application/customization/ManageCustomFields';
import { ManageCustomAssetTypes } from '../src/application/customization/ManageCustomAssetTypes';
import type { AssetTagDefinition } from '../src/domain/customization/Customization';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { CustomizationEditorScreen } from '../src/ui/screens/CustomizationEditorScreen';

const scope = { tenantId: 'audit-household', inventoryId: 'audit-inventory' };
const unavailable = async (): Promise<never> => { throw new Error('Only tag editing is supported in this fixture'); };

export function CustomizationEditorFixture() {
  const [state] = useState(() => {
    let tag: AssetTagDefinition | undefined = { kind: 'tag', id: 'tools', key: 'tools', displayName: 'Tools' };
    const repository: CustomizationRepository = {
      async listTags() { return { items: tag ? [tag] : [] }; },
      async updateTag(_context, id, input) {
        if (!tag || tag.id !== id) throw new Error('Unknown fixture tag');
        tag = { ...tag, ...input }; return tag;
      },
      async archiveTag(_context, id) {
        if (!tag || tag.id !== id) throw new Error('Unknown fixture tag');
        tag = undefined;
      },
      createTag: unavailable,
      async listFields() { return { items: [] }; }, async listAssetTypes() { return { items: [] }; },
      createField: unavailable, updateField: unavailable, archiveField: unavailable, restoreField: unavailable, deleteField: unavailable,
      createAssetType: unavailable, updateAssetType: unavailable, archiveAssetType: unavailable, restoreAssetType: unavailable, deleteAssetType: unavailable
    };
    return {
      client: createMobileQueryClient(),
      policy: new CustomizationAccessPolicy(noCustomizationObservability),
      context: new CustomizationContextQuery({ async getSelectedScope() {
        return { tenant: { id: scope.tenantId, name: 'Audit household', permissions: ['configure'] },
          inventory: { id: scope.inventoryId, name: 'Audit inventory', permissions: ['view', 'edit_asset'] } };
      } }),
      query: new CustomizationCollectionQuery(repository),
      tags: new ManageTags(repository, noCustomizationObservability),
      fields: new ManageCustomFields(repository, noCustomizationObservability),
      types: new ManageCustomAssetTypes(repository, noCustomizationObservability)
    };
  });
  useEffect(() => () => state.client.clear(), [state]);
  return <MobileServerStateProvider client={state.client} scopeId="audit-customization-editor" loadInventoryScope={async () => scope}>
    <CustomizationEditorScreen accessPolicy={state.policy} contextQuery={state.context} query={state.query}
      kind="tag" scope="inventory" mode="edit" resourceId="tools"
      manageTags={state.tags} manageFields={state.fields} manageAssetTypes={state.types}
      onDone={() => router.replace('/audit-customization' as Href)} />
  </MobileServerStateProvider>;
}
