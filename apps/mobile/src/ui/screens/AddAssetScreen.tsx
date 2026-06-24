import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Image,
  Keyboard,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { ImagePlus, X } from 'lucide-react-native';
import { CreateAssetCommand } from '../../application/add/CreateAssetCommand';
import {
  LocationLookupQuery,
  LocationLookupResult
} from '../../application/add/LocationLookupQuery';
import {
  PhotoSelectionQuery,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';
import {
  HomeDashboardLocationViewModel,
  HomeDashboardQuery,
  HomeDashboardViewModel
} from '../../application/home/HomeDashboardQuery';
import type { AssetKind } from '../../domain/assets/AssetSummary';
import { IdentityIcon, IdentityLabel } from '../components/IdentityIcon';
import { colors, radius, spacing } from '../theme/tokens';

type AddAssetScreenProps = {
  readonly createAssetCommand: CreateAssetCommand;
  readonly dashboardQuery: HomeDashboardQuery;
  readonly locationLookupQuery: LocationLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
};

type LoadState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly dashboard: HomeDashboardViewModel }
  | { readonly status: 'error'; readonly message: string };

type SaveState =
  | { readonly status: 'idle' }
  | { readonly status: 'saving' }
  | { readonly status: 'saved'; readonly message: string }
  | { readonly status: 'error'; readonly message: string };

const assetKinds: Array<{ readonly kind: AssetKind; readonly label: string }> = [
  { kind: 'item', label: 'Item' },
  { kind: 'container', label: 'Container' },
  { kind: 'location', label: 'Location' }
];

export function AddAssetScreen({
  createAssetCommand,
  dashboardQuery,
  locationLookupQuery,
  photoSelectionQuery
}: AddAssetScreenProps) {
  const [loadState, setLoadState] = useState<LoadState>({ status: 'loading' });
  const [kind, setKind] = useState<AssetKind>('item');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [parentAssetId, setParentAssetId] = useState<string | undefined>();
  const [locationQuery, setLocationQuery] = useState('');
  const [locationMatches, setLocationMatches] = useState<readonly LocationLookupResult[]>([]);
  const [isCreatingLocation, setIsCreatingLocation] = useState(false);
  const [selectedPhotos, setSelectedPhotos] = useState<readonly SelectedAssetPhoto[]>([]);
  const [saveState, setSaveState] = useState<SaveState>({ status: 'idle' });
  const [keyboardBar, setKeyboardBar] = useState({ isVisible: false, keyboardHeight: 0 });

  useEffect(() => {
    let isCurrent = true;

    dashboardQuery
      .execute()
      .then((dashboard) => {
        if (isCurrent) {
          setLoadState({ status: 'ready', dashboard });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setLoadState({
            status: 'error',
            message: readableError(error, 'Could not load inventory context.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [dashboardQuery]);

  useEffect(() => {
    let isCurrent = true;
    const query = locationQuery.trim();

    if (query.length === 0) {
      setLocationMatches([]);
      return () => {
        isCurrent = false;
      };
    }

    locationLookupQuery
      .execute(query)
      .then((matches) => {
        if (isCurrent) {
          setLocationMatches(matches);
        }
      })
      .catch(() => {
        if (isCurrent) {
          setLocationMatches([]);
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [locationLookupQuery, locationQuery]);

  useEffect(() => {
    if (Platform.OS !== 'ios') {
      return undefined;
    }

    const showSubscription = Keyboard.addListener('keyboardWillShow', (event) => {
      setKeyboardBar({
        isVisible: true,
        keyboardHeight: event.endCoordinates.height
      });
    });
    const changeSubscription = Keyboard.addListener('keyboardWillChangeFrame', (event) => {
      setKeyboardBar({
        isVisible: true,
        keyboardHeight: event.endCoordinates.height
      });
    });
    const hideSubscription = Keyboard.addListener('keyboardWillHide', () => {
      setKeyboardBar({ isVisible: false, keyboardHeight: 0 });
    });

    return () => {
      showSubscription.remove();
      changeSubscription.remove();
      hideSubscription.remove();
    };
  }, []);

  async function saveAsset(): Promise<void> {
    setSaveState({ status: 'saving' });

    try {
      const resolvedParentAssetId = resolveParentAssetId(
        loadState,
        locationMatches,
        locationQuery,
        parentAssetId
      );
      const result = await createAssetCommand.execute({
        kind,
        title,
        description,
        parentAssetId: resolvedParentAssetId,
        photos: selectedPhotos.map((photo) => ({
          fileName: photo.fileName,
          contentType: photo.contentType,
          contentBase64: photo.contentBase64
        }))
      });
      setTitle('');
      setDescription('');
      setParentAssetId(undefined);
      setLocationQuery('');
      setSelectedPhotos([]);
      setSaveState({ status: 'saved', message: result.message });
    } catch (error) {
      setSaveState({ status: 'error', message: readableError(error, 'Could not save asset.') });
    }
  }

  async function createLocation(): Promise<void> {
    const locationName = locationQuery.trim();
    if (loadState.status !== 'ready' || locationName.length === 0) {
      return;
    }

    setIsCreatingLocation(true);
    setSaveState({ status: 'idle' });
    try {
      const result = await createAssetCommand.execute({
        kind: 'location',
        title: locationName,
        description: ''
      });
      const dashboard = await dashboardQuery.execute();
      setLoadState({ status: 'ready', dashboard });
      setParentAssetId(result.id);
      setLocationQuery(result.title);
    } catch (error) {
      setSaveState({ status: 'error', message: readableError(error, 'Could not create location.') });
    } finally {
      setIsCreatingLocation(false);
    }
  }

  async function pickPhotos(): Promise<void> {
    try {
      const photos = await photoSelectionQuery.execute(selectedPhotos.length);
      if (photos.length === 0) {
        return;
      }

      setSelectedPhotos((current) => [...current, ...photos]);
      setSaveState({ status: 'idle' });
    } catch (error) {
      setSaveState({
        status: 'error',
        message: readableError(error, 'Could not select photos.')
      });
    }
  }

  function removePhoto(photoId: string): void {
    setSelectedPhotos((current) => current.filter((photo) => photo.id !== photoId));
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      <ScrollView
        contentContainerStyle={styles.content}
        keyboardDismissMode="interactive"
        keyboardShouldPersistTaps="handled"
      >
        {loadState.status === 'loading' ? (
          <View style={styles.centerState}>
            <ActivityIndicator color={colors.accent} />
            <Text style={styles.stateText}>Loading inventory</Text>
          </View>
        ) : null}
        {loadState.status === 'error' ? (
          <View style={styles.centerState}>
            <Text style={styles.errorTitle}>Could not load</Text>
            <Text style={styles.stateText}>{loadState.message}</Text>
          </View>
        ) : null}
        {loadState.status === 'ready' ? (
          <View>
            <Text style={styles.title}>Add</Text>
            <View style={styles.contextLine}>
              <IdentityLabel
                iconSize="xs"
                kind="inventory"
                label={loadState.dashboard.inventoryName}
                textStyle={styles.contextText}
              />
              <IdentityLabel
                iconSize="xs"
                kind="tenant"
                label={loadState.dashboard.tenantName}
                textStyle={styles.contextText}
              />
            </View>

            {!loadState.dashboard.canAdd ? (
              <View style={styles.unavailablePanel}>
                <Text style={styles.unavailableTitle}>Add is unavailable</Text>
                <Text style={styles.unavailableText}>
                  This inventory does not allow you to create assets.
                </Text>
              </View>
            ) : (
              <>
            <View style={styles.segmentedControl}>
              {assetKinds.map((option) => (
                <Pressable
                  accessibilityRole="button"
                  accessibilityState={{ selected: option.kind === kind }}
                  key={option.kind}
                  onPress={() => setKind(option.kind)}
                  style={[styles.segment, option.kind === kind ? styles.segmentSelected : null]}
                >
                  <Text
                    style={[
                      styles.segmentText,
                      option.kind === kind ? styles.segmentTextSelected : null
                    ]}
                  >
                    {option.label}
                  </Text>
                </Pressable>
              ))}
            </View>

            <TextInput
              accessibilityLabel="Asset name"
              onChangeText={setTitle}
              placeholder="Name"
              placeholderTextColor={colors.textMuted}
              style={styles.input}
              value={title}
            />
            <TextInput
              accessibilityLabel="Asset description"
              multiline
              onChangeText={setDescription}
              placeholder="Description"
              placeholderTextColor={colors.textMuted}
              style={[styles.input, styles.textArea]}
              value={description}
            />

            <LocationPicker
              dashboard={loadState.dashboard}
              isCreatingLocation={isCreatingLocation}
              locationMatches={locationMatches}
              locationQuery={locationQuery}
              onChangeLocationQuery={(value) => {
                setLocationQuery(value);
                setParentAssetId(undefined);
              }}
              onCreateLocation={createLocation}
              onSelectLocation={(location) => {
                setParentAssetId(location?.id);
                setLocationQuery(location?.title ?? '');
              }}
              parentAssetId={parentAssetId}
            />

            <Text style={styles.sectionTitle}>Photos</Text>
            <View style={styles.photoActions}>
              <Pressable accessibilityRole="button" onPress={pickPhotos} style={styles.addPhotoButton}>
                <ImagePlus color={colors.action} size={20} strokeWidth={2.2} />
                <Text style={styles.addPhotoText}>Add photos</Text>
              </Pressable>
              {selectedPhotos.length > 0 ? (
                <Text style={styles.photoCount}>{selectedPhotos.length.toString()}</Text>
              ) : null}
            </View>
            {selectedPhotos.length > 0 ? (
              <View style={styles.photoGrid}>
                {selectedPhotos.map((photo) => (
                  <View key={photo.id} style={styles.photoPreview}>
                    <Image
                      accessibilityIgnoresInvertColors
                      source={{ uri: photo.uri }}
                      style={styles.photoPreviewImage}
                    />
                    <Pressable
                      accessibilityLabel={`Remove ${photo.fileName}`}
                      accessibilityRole="button"
                      onPress={() => removePhoto(photo.id)}
                      style={styles.removePhotoButton}
                    >
                      <X color={colors.text} size={16} strokeWidth={2.4} />
                    </Pressable>
                  </View>
                ))}
              </View>
            ) : null}

            {saveState.status === 'saved' ? (
              <Text style={styles.savedText}>{saveState.message}</Text>
            ) : null}
            {saveState.status === 'error' ? (
              <Text style={styles.errorText}>{saveState.message}</Text>
            ) : null}

            <Pressable
              accessibilityRole="button"
              disabled={saveState.status === 'saving'}
              onPress={saveAsset}
              style={[styles.saveButton, saveState.status === 'saving' ? styles.disabledButton : null]}
            >
              {saveState.status === 'saving' ? (
                <ActivityIndicator color={colors.onAction} />
              ) : (
                <Text style={styles.saveButtonText}>Save</Text>
              )}
            </Pressable>
              </>
            )}
          </View>
        ) : null}
      </ScrollView>
      <KeyboardDismissBar
        keyboardHeight={keyboardBar.keyboardHeight}
        visible={keyboardBar.isVisible}
      />
    </SafeAreaView>
  );
}

function LocationPicker({
  dashboard,
  isCreatingLocation,
  locationMatches,
  locationQuery,
  onChangeLocationQuery,
  onCreateLocation,
  onSelectLocation,
  parentAssetId
}: {
  readonly dashboard: HomeDashboardViewModel;
  readonly isCreatingLocation: boolean;
  readonly locationMatches: readonly LocationLookupResult[];
  readonly locationQuery: string;
  readonly onChangeLocationQuery: (value: string) => void;
  readonly onCreateLocation: () => void;
  readonly onSelectLocation: (location: LocationOption | undefined) => void;
  readonly parentAssetId: string | undefined;
}) {
  const normalizedQuery = normalizeLocationName(locationQuery);
  const locations = normalizedQuery.length === 0
    ? dashboard.locations.slice(0, 3)
    : mergeLocationOptions(locationMatches, dashboard.locations);
  const exactLocation = locations.find(
    (location) => normalizeLocationName(location.title) === normalizedQuery
  );
  const canCreateLocation = normalizedQuery.length > 0 && !exactLocation;

  return (
    <View>
      <Text style={styles.sectionTitle}>Location</Text>
      <TextInput
        accessibilityLabel="Location"
        onChangeText={onChangeLocationQuery}
        placeholder="Type a location"
        placeholderTextColor={colors.textMuted}
        style={styles.input}
        value={locationQuery}
      />
      <ParentOption
        identityKind="inventory"
        isSelected={parentAssetId === undefined && locationQuery.trim().length === 0}
        label="Inventory root"
        onPress={() => onSelectLocation(undefined)}
      />
      {locations.map((location) => (
        <ParentOption
          isSelected={parentAssetId === location.id}
          key={location.id}
          label={location.title}
          location={location}
          onPress={() => onSelectLocation(location)}
        />
      ))}
      {canCreateLocation ? (
        <Pressable
          accessibilityRole="button"
          disabled={isCreatingLocation}
          onPress={onCreateLocation}
          style={[styles.createLocationButton, isCreatingLocation ? styles.disabledButton : null]}
        >
          {isCreatingLocation ? (
            <ActivityIndicator color={colors.action} />
          ) : (
            <Text style={styles.createLocationText}>Create {locationQuery.trim()}</Text>
          )}
        </Pressable>
      ) : null}
    </View>
  );
}

function KeyboardDismissBar({
  keyboardHeight,
  visible
}: {
  readonly keyboardHeight: number;
  readonly visible: boolean;
}) {
  if (Platform.OS !== 'ios' || !visible) {
    return null;
  }

  return (
    <View style={[styles.keyboardDismissBar, { bottom: keyboardHeight }]}>
      <Pressable
        accessibilityLabel="Dismiss keyboard"
        accessibilityRole="button"
        hitSlop={8}
        onPress={Keyboard.dismiss}
        style={styles.keyboardDoneButton}
      >
        <Text style={styles.keyboardDoneText}>Done</Text>
      </Pressable>
    </View>
  );
}

function ParentOption({
  identityKind,
  isSelected,
  label,
  location,
  onPress
}: {
  readonly identityKind?: 'inventory';
  readonly isSelected: boolean;
  readonly label: string;
  readonly location?: LocationOption;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ selected: isSelected }}
      onPress={onPress}
      style={[styles.parentOption, isSelected ? styles.parentOptionSelected : null]}
    >
      <Text style={styles.parentCheck}>{isSelected ? '✓' : ''}</Text>
      {identityKind ? <IdentityIcon kind={identityKind} size="sm" /> : null}
      <View style={styles.parentText}>
        <Text style={styles.parentTitle}>{label}</Text>
        {location ? <Text style={styles.parentMeta}>{location.containedAssetCountLabel} assets</Text> : null}
      </View>
    </Pressable>
  );
}

function resolveParentAssetId(
  loadState: LoadState,
  locationMatches: readonly LocationLookupResult[],
  locationQuery: string,
  parentAssetId: string | undefined
): string | undefined {
  const normalizedQuery = normalizeLocationName(locationQuery);
  if (normalizedQuery.length === 0) {
    return parentAssetId;
  }
  if (parentAssetId) {
    return parentAssetId;
  }
  if (loadState.status !== 'ready') {
    return undefined;
  }

  const exactLocation = mergeLocationOptions(locationMatches, loadState.dashboard.locations).find(
    (location) => normalizeLocationName(location.title) === normalizedQuery
  );
  if (exactLocation) {
    return exactLocation.id;
  }

  throw new Error('Create location or clear the location field.');
}

function normalizeLocationName(value: string): string {
  return value.trim().toLocaleLowerCase();
}

type LocationOption = {
  readonly id: string;
  readonly title: string;
  readonly containedAssetCountLabel: string;
};

function mergeLocationOptions(
  primary: readonly LocationLookupResult[],
  fallback: readonly HomeDashboardLocationViewModel[]
): readonly LocationOption[] {
  const locationsById = new Map<string, LocationOption>();
  for (const location of [...primary, ...fallback]) {
    if (!locationsById.has(location.id)) {
      locationsById.set(location.id, location);
    }
  }
  return [...locationsById.values()].slice(0, 4);
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

const styles = StyleSheet.create({
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
  unavailablePanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  unavailableTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0
  },
  unavailableText: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20,
    marginTop: spacing.xs
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
  segmentedControl: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: spacing.xs,
    marginBottom: spacing.md,
    padding: spacing.xs
  },
  segment: {
    alignItems: 'center',
    borderRadius: radius.sm,
    flex: 1,
    minHeight: 40,
    justifyContent: 'center'
  },
  segmentSelected: {
    backgroundColor: colors.surface
  },
  segmentText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0
  },
  segmentTextSelected: {
    color: colors.accentStrong
  },
  input: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    fontSize: 16,
    marginBottom: spacing.sm,
    minHeight: 48,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  textArea: {
    minHeight: 96,
    textAlignVertical: 'top'
  },
  keyboardDismissBar: {
    alignItems: 'flex-end',
    backgroundColor: colors.surface,
    borderTopColor: colors.border,
    borderTopWidth: 1,
    justifyContent: 'center',
    left: 0,
    minHeight: 44,
    paddingHorizontal: spacing.md,
    position: 'absolute',
    right: 0
  },
  keyboardDoneButton: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 36,
    minWidth: 56
  },
  keyboardDoneText: {
    color: colors.action,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.sm,
    marginTop: spacing.sm
  },
  parentOption: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.sm,
    minHeight: 56,
    padding: spacing.md
  },
  parentOptionSelected: {
    borderColor: colors.action
  },
  createLocationButton: {
    alignItems: 'center',
    borderColor: colors.action,
    borderRadius: radius.md,
    borderWidth: 1,
    justifyContent: 'center',
    marginBottom: spacing.sm,
    minHeight: 48,
    paddingHorizontal: spacing.md
  },
  createLocationText: {
    color: colors.action,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  parentCheck: {
    color: colors.action,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0,
    width: 20
  },
  parentText: {
    flex: 1,
    minWidth: 0
  },
  parentTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  parentMeta: {
    color: colors.textMuted,
    fontSize: 12,
    letterSpacing: 0,
    marginTop: 2
  },
  photoActions: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.sm
  },
  addPhotoButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 48,
    paddingHorizontal: spacing.md
  },
  addPhotoText: {
    color: colors.action,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  photoCount: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  photoGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
    marginBottom: spacing.md
  },
  photoPreview: {
    aspectRatio: 1,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    overflow: 'hidden',
    width: 92
  },
  photoPreviewImage: {
    height: '100%',
    width: '100%'
  },
  removePhotoButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: 999,
    height: 28,
    justifyContent: 'center',
    position: 'absolute',
    right: 6,
    top: 6,
    width: 28
  },
  savedText: {
    color: colors.accentStrong,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0,
    marginBottom: spacing.sm,
    marginTop: spacing.sm
  },
  errorText: {
    color: colors.warning,
    fontSize: 14,
    lineHeight: 20,
    marginBottom: spacing.sm,
    marginTop: spacing.sm
  },
  saveButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 48,
    marginTop: spacing.sm
  },
  disabledButton: {
    opacity: 0.65
  },
  saveButtonText: {
    color: colors.onAction,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  }
});
