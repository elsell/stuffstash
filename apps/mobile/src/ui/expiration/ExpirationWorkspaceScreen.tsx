import { Stack } from 'expo-router';
import { expirationFilterHeaderOptions } from './ExpirationFilterHeader';
import { useExpirationSearch } from './useExpirationSearch';
import { ActivityIndicator, FlatList, Pressable, RefreshControl, StyleSheet, Text, View, useWindowDimensions } from 'react-native';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import type { ExpirationMode } from '../../application/expiration/ExpirationRepository';
import { groupExpirationItems } from '../../application/expiration/ExpirationSections';
import { AssetCard } from '../components/AssetCard';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { formatAssetExpiration } from '../presentation/ExpirationPresentation';

type Row = { key: string; heading: string } | { key: string; asset: AssetCardViewModel };
export function ExpirationWorkspaceScreen({ mode, items, query = '', filtered = false, refinementsActive = false, loading, refreshing, hasMore, appending = false, error, onMode, onSearch, onFilters, onRefresh, onMore, onOpenAsset }: {
 readonly mode: ExpirationMode; readonly items: readonly AssetCardViewModel[]; readonly query?: string; readonly filtered?: boolean; readonly refinementsActive?: boolean;
 readonly loading: boolean; readonly refreshing: boolean; readonly hasMore: boolean; readonly appending?: boolean; readonly error?: string;
 readonly onMode: (mode: ExpirationMode) => void; readonly onSearch: (query: string) => void;
 readonly onFilters: (query: string) => void; readonly onRefresh: () => void; readonly onMore: () => void; readonly onOpenAsset: (id: string) => void;
}) {
 const {fontScale,width} = useWindowDimensions();
 const colors = useAppearanceAwarePalette();
 const search = useExpirationSearch(query, onSearch);
 const rows: Row[] = groupExpirationItems(items).flatMap(group => [{ key: group.key, heading: `${group.expired ? 'Expired · ' : ''}${formatAssetExpiration({ date: group.month, precision: 'month' })}` }, ...group.items.map(asset => ({ key: asset.id, asset }))] as Row[]);
 const button = (label: string, action: () => void, disabled = false) => <Pressable accessibilityRole="button" accessibilityLabel={label} disabled={disabled} onPress={action} style={styles.button}><Text style={{ color: colors.action, fontSize: 17 }}>{label}</Text></Pressable>;
 const controls = <View style={styles.controls}>
   {fontScale > 1.3 || width < 340 ? <NativeActionMenu accessibilityLabel="Expiration status" trigger={{kind:'label',label:mode==='soon'?'Expiring soon':mode==='expired'?'Expired':'All dates'}} groups={[{id:'expiration-mode',items:([{id:'soon',label:'Expiring soon'},{id:'expired',label:'Expired'},{id:'all',label:'All dates'}] as const).map(option=>({...option,isSelected:mode===option.id,onPress:()=>onMode(option.id)}))}]} /> : <NativeSegmentedControl colors={colors} value={mode} onChange={onMode} segments={[{ value: 'soon', label: 'Expiring soon' }, { value: 'expired', label: 'Expired' }, { value: 'all', label: 'All dates' }]} />}
  </View>;
 return <View style={[styles.shell, { backgroundColor: colors.background }]}>
  <Stack.Screen options={{ ...expirationFilterHeaderOptions({active: refinementsActive, onPress: () => { const value = search.flush(); search.ref.current?.blur(); onFilters(value); }}),
   headerSearchBarOptions: {
    ref: search.ref, placeholder: 'Search items', placement: 'stacked', hideWhenScrolling: false,
    hideNavigationBar: false, obscureBackground: false, autoCapitalize: 'none',
    onChangeText: event => search.change(event.nativeEvent.text),
    onSearchButtonPress: event => { search.submit(event.nativeEvent.text); search.ref.current?.blur(); },
    onCancelButtonPress: search.clear, onClose: search.clear,
   },
  }} />
  <FlatList<Row> style={styles.shell} contentContainerStyle={styles.content} data={rows} keyExtractor={row => row.key} alwaysBounceVertical keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag" contentInsetAdjustmentBehavior="automatic"
   refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={colors.action} />}
   renderItem={({ item }) => 'heading' in item ? <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>{item.heading}</Text> : <AssetCard asset={item.asset} density="row" palette={colors} onPress={() => { search.flush(); onOpenAsset(item.asset.id); }} onParentLocationPress={parent => { search.flush(); onOpenAsset(parent.id); }} />}
   ListHeaderComponent={<>{controls}{error ? <View><Text accessibilityRole="alert" style={{ color: colors.text }}>{error}</Text>{button('Retry expiration', onRefresh)}</View> : null}</>}
   ListEmptyComponent={loading ? <ActivityIndicator accessibilityLabel="Loading expiration dates" color={colors.action} /> : !error ? <Text style={[styles.empty, { color: colors.textMuted }]}>{filtered ? 'No matching expiration dates' : mode === 'expired' ? 'No expired items' : mode === 'soon' ? 'No items expiring soon' : 'No expiration dates recorded'}</Text> : null}
   ListFooterComponent={appending ? <ActivityIndicator accessibilityLabel="Loading more expiration dates" color={colors.action} /> : hasMore ? button('Load more', onMore) : null}
  />
 </View>;
}
const styles = StyleSheet.create({ shell: { flex: 1 }, content: { flexGrow: 1, paddingHorizontal: 20, paddingBottom: 32 }, controls: { paddingTop: 12, paddingBottom: 4 }, button: { minHeight: 44, justifyContent: 'center', paddingHorizontal: 8 }, heading: { fontSize: 20, fontWeight: '600', marginTop: 20, marginBottom: 8 }, empty: { paddingVertical: 32, fontSize: 17 } });
