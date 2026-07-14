import { Directory, File, Paths } from 'expo-file-system';
import {
  type AppearancePreferenceTextFile,
  JsonAppearancePreferenceStore
} from './JsonAppearancePreferenceStore';

const settingsDirectory = () => new Directory(Paths.document, 'stuffstash');
const appearanceFile = () => new File(settingsDirectory(), 'appearance-preference.json');

export class FileSystemAppearancePreferenceStore extends JsonAppearancePreferenceStore {
  constructor() {
    super(new ExpoAppearancePreferenceTextFile());
  }
}

class ExpoAppearancePreferenceTextFile implements AppearancePreferenceTextFile {
  async read(): Promise<string | undefined> {
    const file = appearanceFile();
    return file.exists ? file.text() : undefined;
  }

  async write(content: string): Promise<void> {
    const directory = settingsDirectory();
    directory.create({ intermediates: true, idempotent: true });
    appearanceFile().write(content);
  }
}
