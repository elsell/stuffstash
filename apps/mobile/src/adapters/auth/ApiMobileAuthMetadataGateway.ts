import { MobileAuthMetadata, MobileAuthMetadataGateway } from '../../application/auth/MobileAuthSession';

type MobileAuthEnvelope = {
  readonly data?: {
    readonly issuer?: unknown;
    readonly clientId?: unknown;
    readonly redirectUri?: unknown;
    readonly scopes?: unknown;
  };
};

const supportedRedirectUri = 'stuffstash://auth/callback';

export class ApiMobileAuthMetadataGateway implements MobileAuthMetadataGateway {
  constructor(private readonly fetchImpl: typeof fetch = fetch) {}

  async load(apiBaseUrl: string): Promise<MobileAuthMetadata> {
    const baseUrl = apiBaseUrl.replace(/\/+$/, '');
    const response = await this.fetchImpl(`${baseUrl}/.well-known/stuff-stash/mobile-auth`, {
      headers: { Accept: 'application/json' }
    });
    if (!response.ok) {
      throw new Error('Stuff Stash mobile sign-in is not configured for this instance.');
    }

    return parseMobileAuthMetadata((await response.json()) as MobileAuthEnvelope);
  }
}

export function parseMobileAuthMetadata(envelope: MobileAuthEnvelope): MobileAuthMetadata {
  const data = envelope.data;
  const issuer = stringField(data?.issuer, 'issuer').replace(/\/+$/, '');
  const clientId = stringField(data?.clientId, 'clientId');
  const redirectUri = stringField(data?.redirectUri, 'redirectUri');
  if (redirectUri !== supportedRedirectUri) {
    throw new Error('Stuff Stash mobile sign-in metadata contains an unsupported redirect URI.');
  }
  const scopes = arrayOfStrings(data?.scopes, 'scopes');
  if (scopes.length === 0) {
    throw new Error('Stuff Stash mobile sign-in metadata is missing scopes.');
  }

  return {
    issuer,
    clientId,
    redirectUri,
    scopes
  };
}

function stringField(value: unknown, field: string): string {
  if (typeof value !== 'string' || value.trim().length === 0) {
    throw new Error(`Stuff Stash mobile sign-in metadata is missing ${field}.`);
  }
  return value.trim();
}

function arrayOfStrings(value: unknown, field: string): string[] {
  if (!Array.isArray(value)) {
    throw new Error(`Stuff Stash mobile sign-in metadata is missing ${field}.`);
  }

  const result = value
    .filter((item): item is string => typeof item === 'string')
    .map((item) => item.trim())
    .filter((item) => item.length > 0);
  if (result.length !== value.length) {
    throw new Error(`Stuff Stash mobile sign-in metadata contains invalid ${field}.`);
  }
  return [...new Set(result)];
}
