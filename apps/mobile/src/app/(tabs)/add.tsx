import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { AddAssetScreen } from '../../ui/screens/AddAssetScreen';

export default function AddRoute() {
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
      parentLookupQuery={parentLookupQuery}
      photoSelectionQuery={photoSelectionQuery}
    />
  );
}
