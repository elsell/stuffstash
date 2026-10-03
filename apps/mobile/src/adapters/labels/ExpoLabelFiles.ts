import { Directory, File, Paths } from 'expo-file-system';
import * as Crypto from 'expo-crypto';
import * as Sharing from 'expo-sharing';
import * as Print from 'expo-print';
import { Platform } from 'react-native';
import type { LabelFile, LabelFiles } from '../../application/labels/LabelWorkspace';
import { assertReadActive } from '../../application/shared/ReadRequest';

const retentionMs = 24 * 60 * 60 * 1000;
/** Private cache only; never pass authenticated API URLs or credentials to another app. */
export class ExpoLabelFiles implements LabelFiles {
  private readonly active = new Set<string>();
  constructor(private readonly cleanupFailure: () => void, private readonly now: () => number = Date.now) {}
  private remove(directory: Directory) {
    this.active.delete(directory.uri);
    try { if (directory.exists) directory.delete(); } catch { this.cleanupFailure(); }
  }
  private write(file: LabelFile) {
    const root = new Directory(Paths.cache, 'asset-labels');
    if (root.exists) for (const entry of root.list()) {
      if (entry instanceof Directory && !this.active.has(entry.uri) && this.now() - Number(entry.name.split('-')[0]) >= retentionMs) this.remove(entry);
    }
    const directory = new Directory(root, `${this.now()}-${Crypto.randomUUID()}`);
    directory.create({ intermediates: true }); this.active.add(directory.uri);
    try {
      const target = new File(directory, `stuff-stash-label.${file.format}`);
      target.write(file.bytes);
      return { uri: target.uri, directory };
    } catch (error) { this.remove(directory); throw error; }
  }
  async preview(file: LabelFile, signal: AbortSignal) {
    assertReadActive(signal);
    const target = this.write(file);
    return { uri: target.uri, release: () => this.remove(target.directory) };
  }
  async deliver(file: LabelFile, action: 'share' | 'print', signal: AbortSignal) {
    assertReadActive(signal);
    if (action === 'share' && !await Sharing.isAvailableAsync()) throw new Error('Sharing unavailable.');
    assertReadActive(signal);
    const target = this.write(file);
    let success = false;
    try {
      assertReadActive(signal);
      if (action === 'print') await Print.printAsync({ uri: target.uri });
      else await Sharing.shareAsync(target.uri, { mimeType: file.format === 'pdf' ? 'application/pdf' : 'image/png', UTI: file.format === 'pdf' ? 'com.adobe.pdf' : 'public.png' });
      success = true;
    } finally {
      // Android receivers may keep reading after the chooser closes. Sweep on a future write.
      if (success && Platform.OS === 'android') this.active.delete(target.directory.uri);
      else this.remove(target.directory);
    }
  }
}
