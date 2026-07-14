import {
  type AppearancePreference,
  type AppearancePreferenceStore,
  isAppearancePreference
} from '../../application/settings/AppearancePreference';

export interface AppearancePreferenceTextFile {
  read(): Promise<string | undefined>;
  write(content: string): Promise<void>;
}

export class JsonAppearancePreferenceStore implements AppearancePreferenceStore {
  constructor(private readonly file: AppearancePreferenceTextFile) {}

  async load(): Promise<AppearancePreference | undefined> {
    const content = await this.file.read();
    if (!content) {
      return undefined;
    }

    try {
      const parsed = JSON.parse(content) as { readonly preference?: unknown };
      return isAppearancePreference(parsed.preference) ? parsed.preference : undefined;
    } catch {
      return undefined;
    }
  }

  async save(preference: AppearancePreference): Promise<void> {
    await this.file.write(JSON.stringify({ preference }));
  }
}
