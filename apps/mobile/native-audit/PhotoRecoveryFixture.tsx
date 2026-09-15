import { useEffect, useState } from 'react';
import { Button, Image, Text, View } from 'react-native';
import { AssetPhotoViewerSheet } from '../src/ui/screens/AssetPhotoViewerSheet';
import { assetPhotoViewerModel } from '../src/ui/components/AssetPhotoWorkspacePresentation';
import { useAppFeedback } from '../src/ui/feedback/AppFeedback';

/** Runner-only modal composition; no service, credentials or real photo deletion. */
export function PhotoRecoveryFixture({ onBack, missingImage = false }: { readonly onBack: () => void; readonly missingImage?: boolean }) {
  const feedback = useAppFeedback();
  const [selected, setSelected] = useState<string | undefined>('audit-photo');
  const [pending, setPending] = useState(false);
  const [attempts, setAttempts] = useState(0);
  const [photos] = useState(() => [{ id: 'audit-photo', label: 'Audit photo', fileName: 'audit-photo.png',
    uri: Image.resolveAssetSource(require('../assets/brand/stuff-stash-glyph.png')).uri + (missingImage ? '.missing' : '') }]);
  useEffect(() => {
    if (!pending) return;
    const timer = setTimeout(() => {
      setPending(false);
      feedback.showDialog({ title: 'Could not remove photo', message: 'Synthetic removal failed. The photo remains.', primaryAction: { label: 'OK' } });
    }, 2000);
    return () => clearTimeout(timer);
  }, [pending, feedback]);
  return <View>
    <Button title="Back to audit menu" onPress={onBack} />
    <Text>{`Removal attempts: ${attempts}`}</Text>
    <Text>{`Photos remaining: ${photos.length}`}</Text>
    <AssetPhotoViewerSheet canRemove={!missingImage} isRemoving={pending} photos={photos}
      model={assetPhotoViewerModel(photos, selected)} onSelectPhoto={setSelected}
      onClose={() => setSelected(undefined)} onRemove={() => { if (!pending) { setAttempts(value => value + 1); setPending(true); } }} />
  </View>;
}
