import { useState } from 'react';
import { Button, Text, View } from 'react-native';
import { SearchHeader } from '../src/ui/screens/BrowseHeader';
import { ExpirationWorkspaceScreen } from '../src/ui/expiration/ExpirationWorkspaceScreen';
import { useAppearanceAwarePalette } from '../src/ui/theme/appearance';
import type { InventoryMapSurface } from '../src/ui/screens/InventoryMapPresentation';
import type { ExpirationMode } from '../src/application/expiration/ExpirationRepository';

/** Production compositions with local, observable commands and no server writes. */
export function AndroidHeaderCompositionFixture() {
  const palette = useAppearanceAwarePalette();
  const [expiration, setExpiration] = useState(false);
  const [surface, setSurface] = useState<InventoryMapSurface>('list');
  const [mode, setMode] = useState<ExpirationMode>('all');
  const [filters, setFilters] = useState(0);
  return <View style={{ flex: 1, backgroundColor: palette.background }}>
    <Button title="Show expiration workspace" onPress={() => setExpiration(true)} />
    <Text style={{ color: palette.text }}>Composition filter activations: {filters}</Text>
    {expiration ? <ExpirationWorkspaceScreen mode={mode} items={[]} loading={false}
      refreshing={false} hasMore={false} onMode={setMode} onSearch={() => {}}
      onFilters={() => setFilters(value => value + 1)} onRefresh={() => {}}
      onMore={() => {}} onOpenAsset={() => {}} />
      : <View style={{ paddingHorizontal: 20 }}><SearchHeader isLoading={false}
        lifecycleState="active" checkoutState="any" palette={palette} resultCount={20}
        scope="all" selectedSurface={surface} selectedTagIds={[]} sort="updated_desc"
        submittedQuery="" onChangeSurface={setSurface} onClearFilters={() => {}}
        onRemoveFilter={() => {}} onToggleFilters={() => setFilters(value => value + 1)} />
        <Text style={{ color: palette.text }}>Selected surface: {surface}</Text>
      </View>}
  </View>;
}
