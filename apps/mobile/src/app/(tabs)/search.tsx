import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { parseBrowseRouteParams } from '../../ui/screens/BrowseRouteParams';
import { SearchScreen } from '../../ui/screens/SearchScreen';

export default function SearchRoute() {
  const {
    inventoryMapQuery,
    inventoryContextQuery,
    inventoryAssetTagsQuery,
    locationsQuery,
    searchAssetsQuery
  } = useAppServices();
  const params = useLocalSearchParams();
  const initialBrowseState = parseBrowseRouteParams(params);

  return (
    <SearchScreen
      {...initialBrowseState}
      inventoryMapQuery={inventoryMapQuery}
      inventoryContextQuery={inventoryContextQuery}
      inventoryAssetTagsQuery={inventoryAssetTagsQuery}
      locationsQuery={locationsQuery}
      searchAssetsQuery={searchAssetsQuery}
    />
  );
}
