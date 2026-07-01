import { useMemo } from 'react';
import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { initialParentFromParams } from '../../ui/screens/AddAssetInitialParent';
import { AddAssetScreen } from '../../ui/screens/AddAssetScreen';

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
    addAssetDraftStore,
    addDraftScopeQuery,
    createAssetCommand,
    homeDashboardQuery,
    parentLookupQuery,
    photoSelectionQuery
  } = useAppServices();

  return (
    <AddAssetScreen
      addAssetDraftStore={addAssetDraftStore}
      addDraftScopeQuery={addDraftScopeQuery}
      createAssetCommand={createAssetCommand}
      dashboardQuery={homeDashboardQuery}
      initialParent={initialParent}
      parentLookupQuery={parentLookupQuery}
      photoSelectionQuery={photoSelectionQuery}
    />
  );
}
