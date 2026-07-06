import {
  MobileAuthSession,
  MobileAuthSessionStore
} from '../../application/auth/MobileAuthSession';

export type SecureStoreModule = {
  readonly getItemAsync: (key: string, options?: any) => Promise<string | null>;
  readonly setItemAsync: (key: string, value: string, options?: any) => Promise<void>;
  readonly deleteItemAsync: (key: string, options?: any) => Promise<void>;
  readonly WHEN_UNLOCKED_THIS_DEVICE_ONLY?: number;
};

const sessionKey = 'stuffstash.mobile.auth.session';

export class ExpoSecureAuthSessionStore implements MobileAuthSessionStore {
  constructor(private readonly secureStore: SecureStoreModule) {}

  async load(): Promise<MobileAuthSession | undefined> {
    const value = await this.secureStore.getItemAsync(sessionKey, secureStoreOptions(this.secureStore));
    if (!value) {
      return undefined;
    }

    return parseStoredSession(value);
  }

  async save(session: MobileAuthSession): Promise<void> {
    await this.secureStore.setItemAsync(
      sessionKey,
      JSON.stringify(session),
      secureStoreOptions(this.secureStore)
    );
  }

  async clear(): Promise<void> {
    await this.secureStore.deleteItemAsync(sessionKey, secureStoreOptions(this.secureStore));
  }
}

export function parseStoredSession(value: string): MobileAuthSession | undefined {
  try {
    const parsed = JSON.parse(value) as Partial<MobileAuthSession>;
    if (
      typeof parsed.apiBaseUrl !== 'string' ||
      typeof parsed.issuer !== 'string' ||
      typeof parsed.clientId !== 'string' ||
      typeof parsed.idToken !== 'string' ||
      typeof parsed.expiresAt !== 'number' ||
      !Number.isFinite(parsed.expiresAt) ||
      parsed.expiresAt <= 0
    ) {
      return undefined;
    }
    return {
      apiBaseUrl: parsed.apiBaseUrl,
      issuer: parsed.issuer,
      clientId: parsed.clientId,
      idToken: parsed.idToken,
      accessToken: typeof parsed.accessToken === 'string' ? parsed.accessToken : undefined,
      refreshToken: typeof parsed.refreshToken === 'string' ? parsed.refreshToken : undefined,
      expiresAt: parsed.expiresAt
    };
  } catch {
    return undefined;
  }
}

function secureStoreOptions(secureStore: SecureStoreModule): Record<string, unknown> {
  return secureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY
    ? {
        keychainService: 'stuffstash.mobile.auth',
        keychainAccessible: secureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY
      }
    : { keychainService: 'stuffstash.mobile.auth' };
}
