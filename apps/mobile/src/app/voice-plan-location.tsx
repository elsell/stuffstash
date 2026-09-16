import { useLocalSearchParams } from 'expo-router';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { VoicePlanLocationRouteScreen } from '../ui/screens/VoicePlanLocationRouteScreen';

export default function VoicePlanLocationRoute() {
  const params = useLocalSearchParams<{ planId: string; commandId: string; scope: string }>();
  const { parentLookupQuery } = useAppServices();
  return <VoicePlanLocationRouteScreen params={params} parentLookupQuery={parentLookupQuery} />;
}
