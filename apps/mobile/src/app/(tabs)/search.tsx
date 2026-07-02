import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { SearchScreen } from '../../ui/screens/SearchScreen';
import { parseBrowseScope } from '../../ui/screens/SearchScreenPresentation';

export default function SearchRoute() {
  const { locationsQuery, searchAssetsQuery } = useAppServices();
  const params = useLocalSearchParams<{ readonly scope?: string }>();

  return (
    <SearchScreen
      initialScope={parseBrowseScope(params.scope)}
      locationsQuery={locationsQuery}
      searchAssetsQuery={searchAssetsQuery}
    />
  );
}
