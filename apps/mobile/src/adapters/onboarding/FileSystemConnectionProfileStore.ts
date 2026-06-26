import { Directory, File, Paths } from 'expo-file-system';
import {
  ConnectionProfileTextFile,
  JsonConnectionProfileStore
} from './JsonConnectionProfileStore';

const profileDirectory = () => new Directory(Paths.document, 'stuffstash');
const profileFile = () => new File(profileDirectory(), 'connection-profile.json');

export class FileSystemConnectionProfileStore extends JsonConnectionProfileStore {
  constructor() {
    super(new ExpoConnectionProfileTextFile());
  }
}

class ExpoConnectionProfileTextFile implements ConnectionProfileTextFile {
  async read(): Promise<string | undefined> {
    const file = profileFile();
    if (!file.exists) {
      return undefined;
    }

    return file.text();
  }

  async write(content: string): Promise<void> {
    const directory = profileDirectory();
    directory.create({ intermediates: true, idempotent: true });

    profileFile().write(content);
  }

  async delete(): Promise<void> {
    const file = profileFile();
    if (file.exists) {
      file.delete();
    }
  }
}
