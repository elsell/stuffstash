import { useEffect, useState } from 'react';
import { Button, Image, Platform, ScrollView, Text } from 'react-native';
import { createMobilePerformanceSession } from '../src/adapters/observability/MobilePerformanceSession';
import { AssetDetailPhotoGallery } from '../src/ui/components/AssetDetailPhotoGallery';
import { ImagePerformanceProvider } from '../src/ui/components/ImagePerformanceContext';

/** Real native gallery/clock/reporter; only telemetry delivery is a controlled sink. */
export function ImagePerformanceFixture({ onBack }: { readonly onBack: () => void }) {
  const [samples, setSamples] = useState<readonly Record<string, unknown>[]>([]);
  const [missing, setMissing] = useState(false);
  const [session] = useState(() => createMobilePerformanceSession({
    enabled: true,
    platform: Platform.OS === 'ios' ? 'ios' : 'android',
    baseUrl: 'https://native-audit.invalid',
    tokenProvider: () => 'synthetic-audit-session',
    fetch: async input => {
      const request = input as Request;
      const body = await request.json() as { measurements: Record<string, unknown>[] };
      setSamples(previous => [...previous, ...body.measurements]);
      return new Response(JSON.stringify({ data: { accepted: body.measurements.length }, meta: {} }), {
        status: 200, headers: { 'Content-Type': 'application/json' }
      });
    }
  }));
  useEffect(() => () => session.dispose(), [session]);
  const bundled = Image.resolveAssetSource(require('../assets/brand/stuff-stash-glyph.png')).uri;
  return <ImagePerformanceProvider value={session.observer}>
    <ScrollView contentContainerStyle={{ padding: 16, gap: 16 }}>
      <Button title="Back to audit menu" onPress={onBack} />
      <Button title="Load missing image" onPress={() => setMissing(true)} disabled={missing} />
      <Text testID="audit-image-samples">{JSON.stringify(samples)}</Text>
      <AssetDetailPhotoGallery key={missing ? 'missing' : 'bundled'} canAddPhotos={false}
        photos={[{ id: 'audit-image', label: 'Synthetic bundled image', uri: bundled + (missing ? '.missing' : ''), variant: 'large' }]} />
    </ScrollView>
  </ImagePerformanceProvider>;
}
