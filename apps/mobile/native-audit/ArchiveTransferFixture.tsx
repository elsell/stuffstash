import { useEffect, useRef, useState } from 'react';
import { Text, View } from 'react-native';
import { requireNativeModule } from 'expo';
import { Directory, File, Paths } from 'expo-file-system';
import * as Crypto from 'expo-crypto';
import { NativeArchiveUpload, type ArchiveUploadModule } from '../src/adapters/archives/NativeArchiveUpload';
import { NativeCommandButton } from '../src/ui/components/NativeCommandButton';

// Controlled loopback peer; real file and native URLSession/OkHttp transport.
export function ArchiveTransferFixture() {
  const [result, setResult] = useState('Ready for native upload');
  const [busy, setBusy] = useState(false);
  const pending = useRef<AbortController | undefined>(undefined);
  useEffect(() => () => pending.current?.abort(), []);
  async function verify() {
    if (pending.current) return;
    const controller = new AbortController(); pending.current = controller; setBusy(true);
    const directory = new Directory(Paths.cache, `archive-transfer-audit-${Crypto.randomUUID()}`);
    try {
      const base = process.env.EXPO_PUBLIC_STUFF_STASH_ARCHIVE_AUDIT_URL;
      if (!base) throw new Error('Missing audit peer');
      directory.create();
      const file = new File(directory, 'fixture.zip'); file.create();
      file.write('archive-fixture\n'.repeat(65536));
      const transfer = new NativeArchiveUpload(requireNativeModule<ArchiveUploadModule>('StuffStashArchiveTransfer'), Crypto.randomUUID);
      const send = (path: string, signal = controller.signal) => transfer.send(file.uri, new Request(`${base}${path}`, {
        method: 'POST', signal, headers: { Authorization: 'Bearer native-archive-fixture', 'Idempotency-Key': 'native-fixture', 'Content-Type': 'application/zip' }
      }));
      const response = await send('/upload');
      const receipt = await response.json();
      if (response.status !== 200 || receipt.size !== 1048576 || receipt.sha256 !== '62efb6497f972c48f19174ce127609467b06d302059a343ac26c2b26aa7d159c') throw new Error('Upload bytes or authentication differ');
      for (const path of ['/redirect', '/oversized']) {
        let rejected = false;
        try { await send(path); } catch (error) {
          const failure = error as { code?: string; message?: string };
          rejected = path === '/redirect' ? failure.message?.includes('redirects are not permitted') === true : failure.code === 'ERR_ARCHIVE_RESPONSE';
        }
        if (!rejected) throw new Error(`Accepted invalid response: ${path}`);
      }
      const resumed = await send('/upload');
      if (resumed.status !== 200 || (await resumed.json()).sha256 !== receipt.sha256) throw new Error('Transfer did not recover after rejection');
      // The peer holds the response for 15 seconds unless explicitly released.
      const cancel = new AbortController();
      const aborted = send('/stall', cancel.signal);
      setTimeout(() => cancel.abort(), 250);
      let cancelled = false;
      let deadline: ReturnType<typeof setTimeout> | undefined;
      try { await Promise.race([aborted, new Promise((_, reject) => { deadline = setTimeout(() => reject(new Error('Native cancellation timed out')), 3000); })]); }
      catch (error) { cancelled = (error as { name?: string }).name === 'AbortError'; }
      finally { clearTimeout(deadline); await fetch(`${base}/release`); }
      if (!cancelled) throw new Error('Cancellation did not reject');
      if (!controller.signal.aborted) setResult('Native upload verified');
    } catch (error) {
      if (!controller.signal.aborted) setResult(`Native upload failed: ${String(error)}`);
    } finally {
      if (directory.exists) directory.delete();
      pending.current = undefined;
      if (!controller.signal.aborted) setBusy(false);
    }
  }
  return <View style={{ padding: 24, gap: 16 }}><NativeCommandButton label="Verify native upload" disabled={busy} onPress={() => void verify()} /><Text accessibilityLabel={result}>{result}</Text></View>;
}
