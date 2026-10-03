import { useEffect, useLayoutEffect, useRef } from 'react';
import { CameraView, useCameraPermissions } from 'expo-camera';

/** Native camera adapter: only mounted during a focused foreground scan task. */
export function ExpoLabelCamera({ onCode, onUnavailable }: { readonly onCode: (value: string) => void; readonly onUnavailable: () => void }) {
  const [permission, requestPermission] = useCameraPermissions();
  const current = useRef({ onCode, onUnavailable });
  const mounted = useRef(false);
  useLayoutEffect(() => { current.current = { onCode, onUnavailable }; mounted.current = true; return () => { mounted.current = false; }; }, [onCode, onUnavailable]);
  useEffect(() => {
    let active = true;
    void requestPermission().then(result => { if (active && !result.granted) current.current.onUnavailable(); })
      .catch(() => { if (active) current.current.onUnavailable(); });
    return () => { active = false; };
  }, []);
  if (!permission?.granted) return null;
  return <CameraView style={{ flex: 1 }} barcodeScannerSettings={{ barcodeTypes: ['qr'] }}
    onBarcodeScanned={result => { if (mounted.current) current.current.onCode(result.data); }} onMountError={() => { if (mounted.current) current.current.onUnavailable(); }} />;
}
