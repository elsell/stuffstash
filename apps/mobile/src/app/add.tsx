import { returnToPreviousOrHome } from '../ui/navigation/returnToPreviousOrHome';
import { useMemo } from 'react';
import { router, useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { initialParentFromParams } from '../ui/screens/AddAssetInitialParent';
import { AddAssetScreen } from '../ui/screens/AddAssetScreen';

export default function AddRoute() {
  const params = useLocalSearchParams();
  const initialParent = useMemo(() => initialParentFromParams(params), [
    params.parentAssetId,
    params.parentKind,
    params.parentPathLabel,
    params.parentSelectionHint,
    params.parentSubtitle,
    params.parentTitle,
    params.parentWillPromoteToContainer
  ]);
  const {
    inventoryAssetTypesQuery,
    addAssetDraftStore,
    addAssetContextQuery,
    addDraftScopeQuery,
    createAssetCommand,
    parentLookupQuery,
    photoSelectionQuery
  } = useAppServices();

  return (
    <AddAssetScreen
      inventoryAssetTypesQuery={inventoryAssetTypesQuery}
      addAssetDraftStore={addAssetDraftStore}
      addAssetContextQuery={addAssetContextQuery}
      addDraftScopeQuery={addDraftScopeQuery}
      createAssetCommand={createAssetCommand}
      initialParent={initialParent}
      onDismiss={() => returnToPreviousOrHome(router)}
      parentLookupQuery={parentLookupQuery}
      photoSelectionQuery={photoSelectionQuery}
    />
  );
}
