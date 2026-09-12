import { AlertTriangle, Clock } from 'lucide-react-native';
import { Text, View } from 'react-native';
import type { AssetExpiration, AssetExpirationContext } from '../../domain/assets/AssetSummary';
import { expirationDateLabel } from '../presentation/ExpirationPresentation';
import { useAppearancePalette } from '../theme/AppearanceContext';

export function AssetExpirationStatus({ expiration, context }: { readonly expiration?: AssetExpiration; readonly context?: AssetExpirationContext }) {
  const colors = useAppearancePalette();
  if (!expiration) return null;
  const warning = context?.trackingEnabled && context.state !== 'current';
  const expired = context?.state === 'expired';
  const color = warning ? expired ? colors.danger : colors.warning : colors.textMuted;
  const Icon = expired ? AlertTriangle : Clock;
  const label = expirationDateLabel(expiration, context);
  return <View style={{ flexDirection: 'row', alignItems: 'center', gap: 4 }}>
    {warning ? <Icon size={16} color={color} accessibilityElementsHidden /> : null}
    <Text style={{ color, fontSize: 14, flexShrink: 1 }}>{label}</Text>
  </View>;
}
