import { useLocalSearchParams } from 'expo-router';
import { PrintingRoute } from '../../../ui/printing/PrintingRoute';
export default function PrintAssetRoute() { const { assetId } = useLocalSearchParams<{ assetId: string }>(); return <PrintingRoute assetId={assetId} />; }
