import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { parseBrowseRouteParams } from '../../ui/screens/BrowseRouteParams';
import { SearchScreen } from '../../ui/screens/SearchScreen';

export default function SearchRoute() {
  const {
    addAssetPhotosCommand,
    assetCheckoutCommand,
    assetCoreQuery,
    assetContentsQuery,
    assetPhotosQuery,
    assetLifecycleCommand,
    deleteAssetPhotoCommand,
    inventoryMapQuery,
    inventoryContextQuery,
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
      assetCoreQuery={assetCoreQuery}
      assetContentsQuery={assetContentsQuery}
      assetPhotosQuery={assetPhotosQuery}
      assetLifecycleCommand={assetLifecycleCommand}
      deleteAssetPhotoCommand={deleteAssetPhotoCommand}
      inventoryMapQuery={inventoryMapQuery}
      inventoryContextQuery={inventoryContextQuery}
      inventoryAssetTagsQuery={inventoryAssetTagsQuery}
      locationsQuery={locationsQuery}
      photoSelectionQuery={photoSelectionQuery}
      searchAssetsQuery={searchAssetsQuery}
    />
  );
}
