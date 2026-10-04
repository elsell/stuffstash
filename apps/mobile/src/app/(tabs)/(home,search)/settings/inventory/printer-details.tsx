import { useLocalSearchParams } from 'expo-router';
import { PrintingRoute } from '../../../../../ui/printing/PrintingRoute';
export default function Route() { const { printerId } = useLocalSearchParams<{ printerId: string }>(); return <PrintingRoute settingsView="printer" printerId={printerId} />; }
