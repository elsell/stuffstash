import { useEffect, useState } from 'react';
import { ActivityIndicator, FlatList, Pressable, RefreshControl, StyleSheet, Text, View, useWindowDimensions } from 'react-native';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import type { ExpirationMode } from '../../application/expiration/ExpirationRepository';
import { groupExpirationItems } from '../../application/expiration/ExpirationSections';
import { AppTextInput } from '../components/AppTextInput';
import { AssetCard } from '../components/AssetCard';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { formatAssetExpiration } from '../presentation/ExpirationPresentation';

type Row = { key: string; heading: string } | { key: string; asset: AssetCardViewModel };
export function ExpirationWorkspaceScreen({ mode, items, query = '', filtered = false, loading, refreshing, hasMore, appending = false, error, onMode, onSearch, onFilters, onRefresh, onMore, onOpenAsset }: {
 readonly mode: ExpirationMode; readonly items: readonly AssetCardViewModel[]; readonly query?: string; readonly filtered?: boolean;
 readonly loading: boolean; readonly refreshing: boolean; readonly hasMore: boolean; readonly appending?: boolean; readonly error?: string;
 readonly onMode: (mode: ExpirationMode) => void; readonly onSearch: (query: string) => void;
 readonly onFilters: () => void; readonly onRefresh: () => void; readonly onMore: () => void; readonly onOpenAsset: (id: string) => void;
}) {
 const {fontScale,width} = useWindowDimensions();
 const colors = useAppearanceAwarePalette();
 const [draft, setDraft] = useState(query);
 useEffect(() => setDraft(query), [query]);
 const rows: Row[] = groupExpirationItems(items).flatMap(group => [{ key: group.key, heading: `${group.expired ? 'Expired · ' : ''}${formatAssetExpiration({ date: group.month, precision: 'month' })}` }, ...group.items.map(asset => ({ key: asset.id, asset }))] as Row[]);
 const button = (label: string, action: () => void, disabled = false) => <Pressable accessibilityRole="button" accessibilityLabel={label} disabled={disabled} onPress={action} style={styles.button}><Text style={{ color: colors.action, fontSize: 17 }}>{label}</Text></Pressable>;
 return <View style={[styles.shell, { backgroundColor: colors.background }]}>
  <View style={styles.controls}>
   <Text style={{color:colors.textMuted,fontSize:14}}>Active items · Ordered by expiration date</Text>
   {fontScale > 1.3 || width < 340 ? <NativeActionMenu accessibilityLabel="Expiration status" trigger={{kind:'label',label:mode==='soon'?'Expiring soon':mode==='expired'?'Expired':'All dates'}} groups={[{id:'expiration-mode',items:([{id:'soon',label:'Expiring soon'},{id:'expired',label:'Expired'},{id:'all',label:'All dates'}] as const).map(option=>({...option,isSelected:mode===option.id,onPress:()=>onMode(option.id)}))}]} /> : <NativeSegmentedControl colors={colors} value={mode} onChange={onMode} segments={[{ value: 'soon', label: 'Expiring soon' }, { value: 'expired', label: 'Expired' }, { value: 'all', label: 'All dates' }]} />}
   <AppTextInput accessibilityLabel="Search expiration items" placeholder="Search items" value={draft} onChangeText={setDraft} onSubmitEditing={() => onSearch(draft.trim())} returnKeyType="search" style={[styles.search, { color: colors.text, borderColor: colors.border }]} />
   <View style={styles.actions}>
    {button('Search', () => onSearch(draft.trim()))}
    <Pressable accessibilityRole="button" accessibilityLabel="Filter expiration items" onPress={onFilters} style={styles.button}><Text style={{ color: colors.action, fontSize: 17 }}>Filters{filtered ? ' · Active' : ''}</Text></Pressable>
   </View>
  </View>
  <FlatList<Row> style={styles.shell} contentContainerStyle={styles.content} data={rows} keyExtractor={row => row.key} alwaysBounceVertical keyboardShouldPersistTaps="handled"
   refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor={colors.action} />}
   renderItem={({ item }) => 'heading' in item ? <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>{item.heading}</Text> : <AssetCard asset={item.asset} density="row" palette={colors} onPress={() => onOpenAsset(item.asset.id)} onParentLocationPress={parent => onOpenAsset(parent.id)} />}
   ListHeaderComponent={error ? <View><Text accessibilityRole="alert" style={{ color: colors.text }}>{error}</Text>{button('Retry expiration', onRefresh)}</View> : null}
   ListEmptyComponent={loading ? <ActivityIndicator accessibilityLabel="Loading expiration dates" color={colors.action} /> : !error ? <Text style={[styles.empty, { color: colors.textMuted }]}>{filtered ? 'No matching expiration dates' : mode === 'expired' ? 'No expired items' : mode === 'soon' ? 'No items expiring soon' : 'No expiration dates recorded'}</Text> : null}
   ListFooterComponent={appending ? <ActivityIndicator accessibilityLabel="Loading more expiration dates" color={colors.action} /> : hasMore ? button('Load more', onMore) : null}
  />
 </View>;
}
const styles = StyleSheet.create({ shell: { flex: 1 }, content: { flexGrow: 1, paddingHorizontal: 20, paddingBottom: 32 }, controls: { padding: 20, paddingBottom: 0, gap: 12 }, actions: { flexDirection: 'row', justifyContent: 'space-between', flexWrap: 'wrap' }, button: { minHeight: 44, justifyContent: 'center', paddingHorizontal: 8 }, search: { minHeight: 44, borderWidth: StyleSheet.hairlineWidth, borderRadius: 10, paddingHorizontal: 12, fontSize: 17 }, heading: { fontSize: 20, fontWeight: '600', marginTop: 20, marginBottom: 8 }, empty: { paddingVertical: 32, fontSize: 17 } });
