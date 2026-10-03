import { useLocalSearchParams } from 'expo-router';
import { PrintingRoute } from '../../../ui/printing/PrintingRoute';
export default function PrintAssetRoute() { const { assetId, options } = useLocalSearchParams<{ assetId: string; options?: string }>(); return <PrintingRoute assetId={assetId} options={options === '1'} />; }
