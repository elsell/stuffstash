import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { SearchScreen } from '../../ui/screens/SearchScreen';
import { parseBrowseScope } from '../../ui/screens/SearchScreenPresentation';

export default function SearchRoute() {
  const {
    addAssetPhotosCommand,
    assetCheckoutCommand,
    assetDetailQuery,
    assetLifecycleCommand,
    deleteAssetPhotoCommand,
    inventoryMapQuery,
    locationsQuery,
    photoSelectionQuery,
    searchAssetsQuery
  } = useAppServices();
  const params = useLocalSearchParams<{ readonly query?: string; readonly scope?: string }>();

  return (
    <SearchScreen
      initialScope={parseBrowseScope(params.scope)}
      initialQuery={Array.isArray(params.query) ? params.query[0] ?? '' : params.query ?? ''}
      addAssetPhotosCommand={addAssetPhotosCommand}
      assetCheckoutCommand={assetCheckoutCommand}
      assetDetailQuery={assetDetailQuery}
      assetLifecycleCommand={assetLifecycleCommand}
      deleteAssetPhotoCommand={deleteAssetPhotoCommand}
      inventoryMapQuery={inventoryMapQuery}
      locationsQuery={locationsQuery}
      photoSelectionQuery={photoSelectionQuery}
      searchAssetsQuery={searchAssetsQuery}
    />
  );
}
