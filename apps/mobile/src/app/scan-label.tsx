import { LabelTaskHeader } from '../ui/labels/LabelTaskHeader';
import { ActivityIndicator, View } from 'react-native';
import { useRouter, useLocalSearchParams } from 'expo-router';
import { useAppConnectionActions, useAppServices } from '../ui/navigation/AppServicesContext';
import { usePendingLabel } from '../ui/labels/LabelLinkContext';
import { LabelScanScreen } from '../ui/labels/LabelScanScreen';
import { openScannedAsset } from '../ui/labels/openScannedAsset';
export default function ScanLabelRoute() {
  const { external } = useLocalSearchParams<{ external?: string }>();
  const services = useAppServices(); const pending = usePendingLabel(); const router = useRouter();
  const { changeServer, signOut } = useAppConnectionActions();
  return <><LabelTaskHeader onCancel={pending.clear} />{external === '1' && !pending.initialized ? <View style={{ flex: 1, justifyContent: 'center' }}><ActivityIndicator /></View> : <LabelScanScreen key={services.serviceScopeId} open={services.openLabel} pending={pending.reference} revision={pending.revision} retain={pending.retain} invalid={pending.invalid || external === '1' && !pending.reference}
    parse={source => services.labels.parse(source)}
    onResolved={assetId => { pending.clear(); openScannedAsset(router, assetId); }}
    onServer={() => { void changeServer(); }} onAccount={() => { void signOut(); }} />}</>;
}
