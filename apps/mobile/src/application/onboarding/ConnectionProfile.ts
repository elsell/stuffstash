export type ConnectionProfile = {
  readonly apiBaseUrl: string;
  readonly tenantId?: string;
};

export type SavedConnectionProfile = {
  readonly apiBaseUrl: string;
  readonly tenantId?: string;
};

export interface ConnectionProfileStore {
  load(): Promise<SavedConnectionProfile | undefined>;
  save(profile: SavedConnectionProfile): Promise<void>;
  clear(): Promise<void>;
}

export function normalizeInstanceUrl(value: string): string {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    throw new Error('Enter a Stuff Stash instance URL.');
  }

  const withScheme = /^[a-z][a-z0-9+.-]*:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`;
  let parsed: URL;
  try {
    parsed = new URL(withScheme);
  } catch {
    throw new Error('Enter a valid Stuff Stash instance URL.');
  }

  if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
    throw new Error('Stuff Stash instance URLs must use HTTP or HTTPS.');
  }

  parsed.pathname = parsed.pathname.replace(/\/+$/, '');
  parsed.search = '';
  parsed.hash = '';

  return parsed.toString().replace(/\/+$/, '');
}
