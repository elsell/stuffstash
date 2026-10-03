import { LabelTaskHeader } from '../ui/labels/LabelTaskHeader';
import { useRouter } from 'expo-router';
import { useAppConnectionActions, useAppServices } from '../ui/navigation/AppServicesContext';
import { usePendingLabel } from '../ui/labels/LabelLinkContext';
import { LabelScanScreen } from '../ui/labels/LabelScanScreen';
import { assetDetailHref } from '../ui/screens/AssetDetailNavigation';
export default function ScanLabelRoute() {
  const services = useAppServices(); const pending = usePendingLabel(); const router = useRouter();
  const { changeServer, signOut } = useAppConnectionActions();
  return <><LabelTaskHeader onCancel={pending.clear} /><LabelScanScreen key={services.serviceScopeId} open={services.openLabel} pending={pending.reference} invalid={pending.invalid}
    parse={source => { const reference = services.labels.parse(source); pending.capture(source); return reference; }}
    onResolved={assetId => { pending.clear(); router.replace(assetDetailHref(assetId)); }}
    onServer={() => { void changeServer(); }} onAccount={() => { void signOut(); }} /></>;
}
