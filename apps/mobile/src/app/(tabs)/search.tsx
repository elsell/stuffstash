import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { parseBrowseRouteParams } from '../../ui/screens/BrowseRouteParams';
import { SearchScreen } from '../../ui/screens/SearchScreen';

export default function SearchRoute() {
  const {
    addAssetPhotosCommand,
    assetCheckoutCommand,
    assetDetailQuery,
    assetLifecycleCommand,
    deleteAssetPhotoCommand,
    inventoryMapQuery,
    inventoryAssetTagsQuery,
    locationsQuery,
    photoSelectionQuery,
    searchAssetsQuery
  } = useAppServices();
  const params = useLocalSearchParams();
  const initialBrowseState = parseBrowseRouteParams(params);

  return (
    <SearchScreen
      {...initialBrowseState}
      addAssetPhotosCommand={addAssetPhotosCommand}
      assetCheckoutCommand={assetCheckoutCommand}
      assetDetailQuery={assetDetailQuery}
      assetLifecycleCommand={assetLifecycleCommand}
      deleteAssetPhotoCommand={deleteAssetPhotoCommand}
      inventoryMapQuery={inventoryMapQuery}
      inventoryAssetTagsQuery={inventoryAssetTagsQuery}
      locationsQuery={locationsQuery}
      photoSelectionQuery={photoSelectionQuery}
      searchAssetsQuery={searchAssetsQuery}
    />
  );
}
