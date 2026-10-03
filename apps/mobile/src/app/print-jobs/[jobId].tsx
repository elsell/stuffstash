import { useLocalSearchParams } from 'expo-router';
import { PrintingRoute } from '../../ui/printing/PrintingRoute';
export default function PrintJobRoute() { const { jobId, action } = useLocalSearchParams<{ jobId: string; action?: string }>(); return <PrintingRoute jobId={jobId} reprint={action === 'reprint'} />; }
