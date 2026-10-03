import { useLocalSearchParams } from 'expo-router';
import { PrintingRoute } from '../../ui/printing/PrintingRoute';
export default function PrintJobRoute() { const { jobId } = useLocalSearchParams<{ jobId: string }>(); return <PrintingRoute jobId={jobId} />; }
