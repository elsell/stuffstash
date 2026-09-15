import { NativeCommandButton } from '../components/NativeCommandButton';
import {
  Alert,
  Image,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { SelectedAssetPhoto } from '../../application/add/PhotoSelectionQuery';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { showPhotoSourceChooser } from './PhotoSourceChooser';

export function showVoicePlanPhotoSourceChooser({
  isCurrent,
  onCamera,
  onLibrary
}: {
  readonly isCurrent: () => boolean;
  readonly onCamera: () => Promise<void>;
  readonly onLibrary: () => Promise<void>;
}) {
  if (!isCurrent()) return;
  const run = (action: () => Promise<void>) => {
    if (!isCurrent()) return;
    action().catch((error: unknown) => {
      if (!isCurrent()) return;
      Alert.alert('Could not add photos', error instanceof Error ? error.message : 'Photo selection failed.');
    });
  };

  showPhotoSourceChooser({
    isCurrent,
    onCamera: () => run(onCamera),
    onLibrary: () => run(onLibrary)
  });
}

export function VoicePlanPhotoDraftStrip({
  readOnly = false,
  commandKey,
  onAddPhotos,
  onRemovePhoto,
  photos
}: {
  readonly readOnly?: boolean;
  readonly commandKey: string;
  readonly onAddPhotos: (commandKey: string) => void;
  readonly onRemovePhoto: (commandKey: string, photoId: string) => void;
  readonly photos: readonly SelectedAssetPhoto[];
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <View style={styles.planPhotoStrip}>
      {!readOnly ? <NativeCommandButton label="Add photos" onPress={() => onAddPhotos(commandKey)} /> : null}
      {photos.length > 0 ? (
        <ScrollView
          testID="voice-plan-photo-previews"
          horizontal
          contentContainerStyle={styles.planPhotoPreviewList}
          showsHorizontalScrollIndicator={false}
        >
          {photos.map((photo, index) => (
            <View key={photo.id} style={styles.planPhotoPreviewFrame}>
              <Image
                accessibilityIgnoresInvertColors
                source={{ uri: photo.uri }}
                style={styles.planPhotoPreview}
              />
              {!readOnly ? <NativeCommandButton label={`Remove photo ${index + 1}`}
                onPress={() => onRemovePhoto(commandKey, photo.id)} /> : null}
            </View>
          ))}
          <Text style={styles.planPhotoCount}>{photos.length.toString()}</Text>
        </ScrollView>
      ) : null}
      {photos.length > 0 ? (
        <Text style={styles.planPhotoDraftNote}>{readOnly ? 'Draft kept on this device.' : 'Attaches after approval.'}</Text>
      ) : null}
    </View>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  planPhotoCount: {
    alignSelf: 'center',
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    minWidth: 18,
    textAlign: 'center'
  },
  planPhotoDraftNote: {
    color: colors.textMuted,
    flexBasis: '100%',
    fontSize: 11,
    fontWeight: '700',
    lineHeight: 15,
    marginLeft: 1
  },
  planPhotoPreview: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    height: 72,
    width: 96
  },
  planPhotoPreviewFrame: {
    alignItems: 'center',
    width: 120
  },
  planPhotoPreviewList: {
    alignItems: 'center',
    gap: spacing.xs,
    paddingRight: spacing.sm
  },
  planPhotoStrip: {
    alignItems: 'center',
    flexWrap: 'wrap',
    flexDirection: 'row',
    gap: spacing.sm,
    marginLeft: 28 + spacing.md,
    minHeight: 38
  }
  });
}
