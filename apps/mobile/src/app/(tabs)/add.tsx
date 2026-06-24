import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { AddAssetScreen } from '../../ui/screens/AddAssetScreen';

export default function AddRoute() {
  const {
    createAssetCommand,
    homeDashboardQuery,
    locationLookupQuery,
    photoSelectionQuery
  } = useAppServices();

  return (
    <AddAssetScreen
      createAssetCommand={createAssetCommand}
      dashboardQuery={homeDashboardQuery}
      locationLookupQuery={locationLookupQuery}
      photoSelectionQuery={photoSelectionQuery}
    />
  );
}
