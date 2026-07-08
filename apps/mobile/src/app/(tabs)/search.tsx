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
    inventoryAssetTagsQuery,
    locationsQuery,
    photoSelectionQuery,
    searchAssetsQuery
  } = useAppServices();
  const params = useLocalSearchParams();
  const initialTagId = Array.isArray(params.tagId) ? params.tagId[0] : params.tagId;
  const initialQuery = Array.isArray(params.query) ? params.query[0] ?? '' : params.query ?? '';

  return (
    <SearchScreen
      initialScope={parseBrowseScope(params.scope)}
      initialQuery={initialQuery}
      initialTagIds={initialTagId ? [initialTagId] : []}
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
