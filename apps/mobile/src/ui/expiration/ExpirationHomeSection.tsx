import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import type { ExpirationMode } from '../../application/expiration/ExpirationRepository';
import type { ExpirationWorkspaceQuery } from '../../application/expiration/ExpirationWorkspaceQuery';
import { SelectionRow } from '../components/SelectionRow';
import { AssetCard } from '../components/AssetCard';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { createHomeScreenStyles } from '../screens/HomeScreen.styles';

type HomeData = Awaited<ReturnType<ExpirationWorkspaceQuery['home']>>;
export function ExpirationHomeSection({ data, error, onOpen, onOpenAsset, onRetry, canOpen = true }: {
 readonly canOpen?: boolean; readonly data?: HomeData; readonly error?: string;
 readonly onOpen: (mode: ExpirationMode) => void;
 readonly onOpenAsset: (id: string) => void;
 readonly onRetry: () => void;
}) {
 const colors = useAppearanceAwarePalette();
 const styles = createHomeScreenStyles(colors);
 if (data?.counts.all === 0 && !error) return null;
 return <View style={styles.attentionSection}>
  <View style={styles.sectionHeader}>
   <Text accessibilityRole="header" style={styles.sectionTitle}>{t('mobile.ExpirationHomeSection.expiration')}</Text>
   <Pressable accessibilityRole="button" accessibilityLabel={t('mobile.ExpirationHomeSection.viewAllExpirationDates')} disabled={!canOpen} accessibilityState={{ disabled: !canOpen }} style={styles.sectionActionButton} onPress={() => { if (canOpen) onOpen('all'); }}><Text style={styles.sectionAction}>{t('mobile.ExpirationHomeSection.seeAll')}</Text></Pressable>
  </View>
  {error ? <View><Text accessibilityRole="alert" style={styles.stateText}>{error}</Text><NativeCommandButton label={t('mobile.ExpirationHomeSection.retryExpiration')} onPress={onRetry} /></View> : null}
  {!data && !error ? <ActivityIndicator accessibilityLabel={t('mobile.ExpirationHomeSection.loadingExpiration')} color={colors.action} /> : null}
  {data ? <>
   {data.counts.soon + data.counts.expired === 0 ? <Text style={styles.emptyText}>{t('mobile.ExpirationHomeSection.noneExpiringSoon')}</Text> : <View>
    <SelectionRow label={t('mobile.ExpirationHomeSection.expired')} value={String(data.counts.expired)} accessibilityLabel={t('mobile.ExpirationHomeSection.viewExpiredItems', { expired: String(data.counts.expired) })} onPress={() => onOpen('expired')} />
    <SelectionRow label={t('mobile.ExpirationHomeSection.expiringSoon')} value={String(data.counts.soon)} accessibilityLabel={t('mobile.ExpirationHomeSection.viewItemsExpiringSoon', { soon: String(data.counts.soon) })} onPress={() => onOpen('soon')} />
   </View>}
   {data.items.map(item => <AssetCard key={item.id} asset={item} density="row" palette={colors} showTags={false} onPress={() => onOpenAsset(item.id)} onParentLocationPress={parent => onOpenAsset(parent.id)} />)}
  </> : null}
 </View>;
}
