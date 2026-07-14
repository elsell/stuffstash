import { useEffect, useState } from 'react';
import { router } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  Image,
  Pressable,
  RefreshControl,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  LocationBrowserItemViewModel,
  LocationsQuery,
  LocationsViewModel
} from '../../application/locations/LocationsQuery';
import { IdentityLabel } from '../components/IdentityIcon';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

type LocationsScreenProps = {
  readonly locationsQuery: LocationsQuery;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly locations: LocationsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function LocationsScreen({ locationsQuery }: LocationsScreenProps) {
  const styles = createStyles(useAppearanceAwarePalette());
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    let isCurrent = true;

    locationsQuery
      .execute()
      .then((locations) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', locations });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Stuff Stash could not load locations.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [locationsQuery]);

  async function refreshLocations(): Promise<void> {
    setIsRefreshing(true);

    try {
      const locations = await locationsQuery.execute();
      setScreenState({ status: 'ready', locations });
    } catch (error) {
      setScreenState({
        status: 'error',
        message: readableError(error, 'Stuff Stash could not refresh locations.')
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <LocationsList
          isRefreshing={isRefreshing}
          locations={screenState.locations}
          onRefresh={refreshLocations}
        />
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading locations</Text>
    </View>
  );
}

function ErrorState({ message }: { readonly message: string }) {
  const styles = createStyles(useAppearanceAwarePalette());
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load</Text>
      <Text style={styles.stateText}>{message}</Text>
    </View>
  );
}

function LocationsList({
  isRefreshing,
  locations,
  onRefresh
}: {
  readonly isRefreshing: boolean;
  readonly locations: LocationsViewModel;
  readonly onRefresh: () => void;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  return (
    <FlatList
      data={locations.locations}
      keyExtractor={(location) => location.id}
      contentContainerStyle={styles.content}
      refreshControl={(
        <RefreshControl
          onRefresh={onRefresh}
          refreshing={isRefreshing}
          tintColor={colors.action}
        />
      )}
      ListHeaderComponent={
        <View>
          <Text style={styles.title}>Locations</Text>
          <View style={styles.contextLine}>
            <IdentityLabel
              iconSize="xs"
              kind="inventory"
              label={locations.inventoryName}
              textStyle={styles.contextText}
            />
            <IdentityLabel
              iconSize="xs"
              kind="tenant"
              label={locations.tenantName}
              textStyle={styles.contextText}
            />
          </View>
        </View>
      }
      ListEmptyComponent={<Text style={styles.emptyText}>No locations yet.</Text>}
      renderItem={({ item }) => <LocationRow location={item} />}
    />
  );
}

function LocationRow({ location }: { readonly location: LocationBrowserItemViewModel }) {
  const styles = createStyles(useAppearanceAwarePalette());
  return (
    <Pressable
      accessibilityRole="button"
      onPress={() => router.push(`/locations/${location.id}`)}
      style={styles.locationCard}
    >
      <View style={styles.locationImageFrame}>
        {location.photo ? (
          <Image
            accessibilityIgnoresInvertColors
            source={{ uri: location.photo.uri, headers: location.photo.headers }}
            style={styles.locationImage}
          />
        ) : (
          <Text style={styles.locationImageLabel}>Place</Text>
        )}
      </View>
      <View style={styles.locationBody}>
        <View style={styles.locationHeader}>
          <View style={styles.locationText}>
            <Text style={styles.locationTitle}>{location.title}</Text>
            <Text style={styles.locationDescription}>{location.description}</Text>
          </View>
          <Text style={location.photoLabel === 'Photo ready' ? styles.photoReady : styles.photoNeeded}>
            {location.photoLabel}
          </Text>
        </View>
        <View style={styles.locationFooter}>
          <Text style={styles.locationCount}>{location.containedAssetCountLabel}</Text>
          <Text style={styles.recentAssetLabel}>{location.recentAssetLabel}</Text>
        </View>
      </View>
    </Pressable>
  );
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xl
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  stateText: {
    color: colors.textMuted,
    fontSize: 16,
    lineHeight: 23,
    marginTop: spacing.md,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 24,
    fontWeight: '800',
    letterSpacing: 0
  },
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36
  },
  contextLine: {
    alignItems: 'center',
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
    marginBottom: spacing.md,
    marginTop: spacing.xs
  },
  contextText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0
  },
  emptyText: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22
  },
  locationCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    overflow: 'hidden'
  },
  locationImageFrame: {
    alignItems: 'center',
    aspectRatio: 16 / 9,
    backgroundColor: colors.surfaceMuted,
    justifyContent: 'center'
  },
  locationImageLabel: {
    color: colors.accentStrong,
    fontSize: 28,
    fontWeight: '900',
    letterSpacing: 0
  },
  locationImage: {
    height: '100%',
    width: '100%'
  },
  locationBody: {
    padding: spacing.md
  },
  locationHeader: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.md,
    justifyContent: 'space-between'
  },
  locationText: {
    flex: 1,
    minWidth: 0
  },
  locationTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  locationDescription: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20,
    marginTop: spacing.xs
  },
  locationFooter: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    gap: spacing.xs,
    marginTop: spacing.md,
    paddingTop: spacing.md
  },
  locationCount: {
    color: colors.accentStrong,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0
  },
  recentAssetLabel: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 17
  },
  photoReady: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  photoNeeded: {
    backgroundColor: colors.warningSurface,
    borderRadius: radius.sm,
    color: colors.warning,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  }
  });
}
