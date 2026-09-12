import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import type { ExpirationMode } from '../../application/expiration/ExpirationRepository';
import type { ExpirationWorkspaceQuery } from '../../application/expiration/ExpirationWorkspaceQuery';
import { SelectionRow } from '../components/SelectionRow';
import { AssetCard } from '../components/AssetCard';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { createHomeScreenStyles } from '../screens/HomeScreen.styles';

type HomeData = Awaited<ReturnType<ExpirationWorkspaceQuery['home']>>;
export function ExpirationHomeSection({ data, error, onOpen, onOpenAsset, onRetry }: {
 readonly data?: HomeData; readonly error?: string;
 readonly onOpen: (mode: ExpirationMode) => void;
 readonly onOpenAsset: (id: string) => void;
 readonly onRetry: () => void;
}) {
 const colors = useAppearanceAwarePalette();
 const styles = createHomeScreenStyles(colors);
 if (data?.counts.all === 0 && !error) return null;
 return <View style={styles.attentionSection}>
  <View style={styles.sectionHeader}>
   <Text accessibilityRole="header" style={styles.sectionTitle}>Expiration</Text>
   <Pressable accessibilityRole="button" accessibilityLabel="View all expiration dates" style={styles.sectionActionButton} onPress={() => onOpen('all')}><Text style={styles.sectionAction}>See all</Text></Pressable>
  </View>
  {error ? <View><Text accessibilityRole="alert" style={styles.stateText}>{error}</Text><Pressable accessibilityRole="button" accessibilityLabel="Retry expiration" onPress={onRetry} style={styles.sectionActionButton}><Text style={styles.sectionAction}>Retry</Text></Pressable></View> : null}
  {!data && !error ? <ActivityIndicator accessibilityLabel="Loading expiration" color={colors.action} /> : null}
  {data ? <>
   {data.counts.soon + data.counts.expired === 0 ? <Text style={styles.emptyText}>None expiring soon</Text> : <View>
    <SelectionRow label="Expired" value={String(data.counts.expired)} accessibilityLabel={`View ${data.counts.expired} expired items`} onPress={() => onOpen('expired')} />
    <SelectionRow label="Expiring soon" value={String(data.counts.soon)} accessibilityLabel={`View ${data.counts.soon} items expiring soon`} onPress={() => onOpen('soon')} />
   </View>}
   {data.items.map(item => <AssetCard key={item.id} asset={item} density="row" palette={colors} showTags={false} onPress={() => onOpenAsset(item.id)} onParentLocationPress={parent => onOpenAsset(parent.id)} />)}
  </> : null}
 </View>;
}
