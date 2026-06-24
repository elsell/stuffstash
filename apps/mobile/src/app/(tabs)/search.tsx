import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { SearchScreen } from '../../ui/screens/SearchScreen';

export default function SearchRoute() {
  const { searchAssetsQuery } = useAppServices();

  return <SearchScreen searchAssetsQuery={searchAssetsQuery} />;
}
