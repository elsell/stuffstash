import { ImagePlus, X } from 'lucide-react-native';
import {
  Alert,
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { SelectedAssetPhoto } from '../../application/add/PhotoSelectionQuery';
import { colors, radius, spacing } from '../theme/tokens';
import { showPhotoSourceChooser } from './PhotoSourceChooser';

export function showVoicePlanPhotoSourceChooser({
  onCamera,
  onLibrary
}: {
  readonly onCamera: () => Promise<void>;
  readonly onLibrary: () => Promise<void>;
}) {
  const run = (action: () => Promise<void>) => {
    action().catch((error: unknown) => {
      Alert.alert('Could not add photos', error instanceof Error ? error.message : 'Photo selection failed.');
    });
  };

  showPhotoSourceChooser({
    onCamera: () => run(onCamera),
    onLibrary: () => run(onLibrary)
  });
}

export function VoicePlanPhotoDraftStrip({
  commandKey,
  onAddPhotos,
  onRemovePhoto,
  photos
}: {
  readonly commandKey: string;
  readonly onAddPhotos: (commandKey: string) => void;
  readonly onRemovePhoto: (commandKey: string, photoId: string) => void;
  readonly photos: readonly SelectedAssetPhoto[];
}) {
  return (
    <View style={styles.planPhotoStrip}>
      <Pressable
        accessibilityLabel="Stage draft photos for this planned item"
        accessibilityRole="button"
        onPress={() => onAddPhotos(commandKey)}
        style={styles.planPhotoAddButton}
      >
        <ImagePlus color={colors.accentStrong} size={17} strokeWidth={2.4} />
        <Text style={styles.planPhotoAddText}>
          {photos.length > 0 ? 'Stage more' : 'Stage photos'}
        </Text>
      </Pressable>
      {photos.length > 0 ? (
        <ScrollView
          horizontal
          contentContainerStyle={styles.planPhotoPreviewList}
          showsHorizontalScrollIndicator={false}
        >
          {photos.map((photo) => (
            <View key={photo.id} style={styles.planPhotoPreviewFrame}>
              <Image
                accessibilityIgnoresInvertColors
                source={{ uri: photo.uri }}
                style={styles.planPhotoPreview}
              />
              <Pressable
                accessibilityLabel="Remove draft photo"
                accessibilityRole="button"
                onPress={() => onRemovePhoto(commandKey, photo.id)}
                style={styles.planPhotoRemoveButton}
              >
                <X color={colors.surface} size={11} strokeWidth={3} />
              </Pressable>
            </View>
          ))}
          <Text style={styles.planPhotoCount}>{photos.length.toString()}</Text>
        </ScrollView>
      ) : null}
      {photos.length > 0 ? (
        <Text style={styles.planPhotoDraftNote}>Attaches after approval.</Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  planPhotoAddButton: {
    alignItems: 'center',
    alignSelf: 'flex-start',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: spacing.xs,
    minHeight: 34,
    paddingHorizontal: spacing.sm
  },
  planPhotoAddText: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900'
  },
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
    height: 34,
    width: 34
  },
  planPhotoPreviewFrame: {
    height: 38,
    justifyContent: 'flex-end',
    width: 38
  },
  planPhotoPreviewList: {
    alignItems: 'center',
    gap: spacing.xs,
    paddingRight: spacing.sm
  },
  planPhotoRemoveButton: {
    alignItems: 'center',
    backgroundColor: colors.text,
    borderColor: colors.surface,
    borderRadius: 9,
    borderWidth: 1,
    height: 18,
    justifyContent: 'center',
    position: 'absolute',
    right: 0,
    top: 0,
    width: 18
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
