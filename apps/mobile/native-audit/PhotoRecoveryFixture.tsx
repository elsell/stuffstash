import { useEffect, useState } from 'react';
import { Button, Image, Text, View } from 'react-native';
import { AssetPhotoViewerSheet } from '../src/ui/screens/AssetPhotoViewerSheet';
import { assetPhotoViewerModel } from '../src/ui/components/AssetPhotoWorkspacePresentation';
import { useAppFeedback } from '../src/ui/feedback/AppFeedback';

/** Runner-only modal composition; no service, credentials or real photo deletion. */
export function PhotoRecoveryFixture({ onBack, missingImage = false, removeSucceeds = false }: { readonly onBack: () => void; readonly missingImage?: boolean; readonly removeSucceeds?: boolean }) {
  const feedback = useAppFeedback();
  const [selected, setSelected] = useState<string | undefined>('audit-photo');
  const [pending, setPending] = useState(false);
  const [attempts, setAttempts] = useState(0);
  const [photos, setPhotos] = useState(() => [{ id: 'audit-photo', label: 'Audit photo', fileName: 'audit-photo.png',
    uri: Image.resolveAssetSource(require('../assets/brand/stuff-stash-glyph.png')).uri + (missingImage ? '.missing' : '') }]);
  useEffect(() => {
    if (!pending) return;
    const timer = setTimeout(() => {
      setPending(false);
      if (removeSucceeds) { setPhotos([]); setSelected(undefined); return; }
      feedback.showDialog({ title: 'Could not remove photo', message: 'Synthetic removal failed. The photo remains.', primaryAction: { label: 'OK' } });
    }, 2000);
    return () => clearTimeout(timer);
  }, [pending, feedback, removeSucceeds]);
  return <View>
    <Button title="Back to audit menu" onPress={onBack} />
    <Text>{`Removal attempts: ${attempts}`}</Text>
    <Text>{`Photos remaining: ${photos.length}`}</Text>
    <AssetPhotoViewerSheet canRemove={!missingImage} isRemoving={pending} photos={photos}
      model={assetPhotoViewerModel(photos, selected)} onSelectPhoto={setSelected}
      onClose={() => setSelected(undefined)} onRemove={() => { if (!pending) { setAttempts(value => value + 1); setPending(true); } }} />
  </View>;
}
