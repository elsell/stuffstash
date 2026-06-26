import {
  ConnectionProfileStore,
  normalizeInstanceUrl,
  SavedConnectionProfile
} from '../../application/onboarding/ConnectionProfile';

export interface ConnectionProfileTextFile {
  read(): Promise<string | undefined>;
  write(content: string): Promise<void>;
  delete(): Promise<void>;
}

export class JsonConnectionProfileStore implements ConnectionProfileStore {
  constructor(private readonly file: ConnectionProfileTextFile) {}

  async load(): Promise<SavedConnectionProfile | undefined> {
    const content = await this.file.read();
    if (!content) {
      return undefined;
    }

    let parsed: Partial<SavedConnectionProfile>;
    try {
      parsed = JSON.parse(content) as Partial<SavedConnectionProfile>;
    } catch {
      return undefined;
    }

    if (!parsed.apiBaseUrl) {
      return undefined;
    }

    return {
      apiBaseUrl: normalizeInstanceUrl(parsed.apiBaseUrl),
      tenantId: parsed.tenantId?.trim() || undefined
    };
  }

  async save(profile: SavedConnectionProfile): Promise<void> {
    await this.file.write(
      JSON.stringify({
        apiBaseUrl: normalizeInstanceUrl(profile.apiBaseUrl),
        tenantId: profile.tenantId?.trim() || undefined
      })
    );
  }

  async clear(): Promise<void> {
    await this.file.delete();
  }
}
