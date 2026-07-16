import type { InventoryInvitationReference } from './InventoryInvitationRepository';

const maximumLinkLength = 4096;
const maximumIdentifierLength = 200;
const tokenLength = 43;
const identifierPattern = /^[A-Za-z0-9][A-Za-z0-9._:-]*$/;
const tokenPattern = /^[A-Za-z0-9_-]+$/;
const expectedQueryFields = new Set(['tenant', 'inventory', 'invitation']);
const expectedFragmentFields = new Set(['token']);

export class InvalidInventoryInvitationLinkError extends Error {
  constructor() {
    super('This invitation link is invalid.');
    this.name = 'InvalidInventoryInvitationLinkError';
  }
}

export function parseInventoryInvitationLink(
  source: string,
  configuredPublicOrigin: string,
  allowInsecureLocalHTTP = false
): InventoryInvitationReference {
  return parseLink(source, (link) => validateTrustedRoute(link, configuredPublicOrigin, allowInsecureLocalHTTP));
}

export function parseCreatedInventoryInvitationLink(
  source: string,
  trustedInvitationOrigin?: string,
  allowInsecureLocalHTTP = false
): InventoryInvitationReference {
  return parseLink(source, (link) => validateCreatedRoute(link, trustedInvitationOrigin, allowInsecureLocalHTTP));
}

function parseLink(
  source: string,
  validateRoute: (link: URL) => void
): InventoryInvitationReference {
  try {
    if (
      !source ||
      source.length > maximumLinkLength ||
      source.trim() !== source ||
      /[\u0000-\u001F\u007F]/.test(source)
    ) {
      throw new InvalidInventoryInvitationLinkError();
    }
    const link = new URL(source);
    validateRoute(link);
    validateOnlyFields(link.searchParams, expectedQueryFields);

    const fragment = parseFragment(link.hash);
    validateOnlyFields(fragment, expectedFragmentFields);

    return {
      tenantId: requiredIdentifier(link.searchParams, 'tenant'),
      inventoryId: requiredIdentifier(link.searchParams, 'inventory'),
      invitationId: requiredIdentifier(link.searchParams, 'invitation'),
      acceptanceToken: requiredToken(fragment)
    };
  } catch (error) {
    if (error instanceof InvalidInventoryInvitationLinkError) {
      throw error;
    }
    throw new InvalidInventoryInvitationLinkError();
  }
}

function validateCreatedRoute(link: URL, trustedInvitationOrigin?: string, allowInsecureLocalHTTP = false): void {
  if (link.username || link.password || normalizedPath(link.pathname) !== '/invitations/accept') {
    throw new InvalidInventoryInvitationLinkError();
  }
  const secure = link.protocol === 'https:';
  const trustedOrigin = trustedInvitationOrigin === undefined
    ? undefined
    : parseConfiguredOrigin(trustedInvitationOrigin, allowInsecureLocalHTTP);
  const trustedConfiguredOrigin = trustedOrigin !== undefined && link.origin === trustedOrigin.origin;
  const unconfiguredLoopback = trustedOrigin === undefined && isLoopbackHostname(link.hostname);
  if (!(secure && trustedConfiguredOrigin) && !(
    link.protocol === 'http:' &&
    allowInsecureLocalHTTP &&
    (trustedConfiguredOrigin || unconfiguredLoopback)
  )) {
    throw new InvalidInventoryInvitationLinkError();
  }
}

function isLoopbackHostname(hostname: string): boolean {
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '[::1]';
}

function validateTrustedRoute(link: URL, configuredPublicOrigin: string, allowInsecureLocalHTTP = false): void {
  if (link.username || link.password) {
    throw new InvalidInventoryInvitationLinkError();
  }

  if (link.protocol === 'stuffstash:') {
    if (link.host !== 'invitations' || normalizedPath(link.pathname) !== '/accept') {
      throw new InvalidInventoryInvitationLinkError();
    }
    return;
  }

  const configuredOrigin = parseConfiguredOrigin(configuredPublicOrigin, allowInsecureLocalHTTP);
  if (
    (link.protocol !== 'https:' && !(allowInsecureLocalHTTP && isPrivateLocalHTTPOrigin(link))) ||
    link.origin !== configuredOrigin.origin ||
    normalizedPath(link.pathname) !== '/invitations/accept'
  ) {
    throw new InvalidInventoryInvitationLinkError();
  }
}

function parseConfiguredOrigin(value: string, allowInsecureLocalHTTP = false): URL {
  const origin = new URL(value);
  if (
    (origin.protocol !== 'https:' && !(allowInsecureLocalHTTP && isPrivateLocalHTTPOrigin(origin))) ||
    origin.username ||
    origin.password ||
    origin.search ||
    origin.hash ||
    normalizedPath(origin.pathname) !== '/'
  ) {
    throw new InvalidInventoryInvitationLinkError();
  }
  return origin;
}

function isPrivateLocalHTTPOrigin(origin: URL): boolean {
  return origin.protocol === 'http:' && isLocalDevelopmentHostname(origin.hostname);
}

function isLocalDevelopmentHostname(hostname: string): boolean {
  if (isLoopbackHostname(hostname)) return true;
  const parts = hostname.split('.');
  if (parts.length !== 4 || parts.some((part) => !/^\d{1,3}$/.test(part))) return false;
  const octets = parts.map(Number);
  if (octets.some((octet) => octet > 255)) return false;
  return octets[0] === 10 ||
    (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31) ||
    (octets[0] === 192 && octets[1] === 168);
}

function normalizedPath(path: string): string {
  if (path === '/') {
    return path;
  }
  return path.endsWith('/') ? path.slice(0, -1) : path;
}

function parseFragment(hash: string): URLSearchParams {
  if (!hash.startsWith('#') || hash.length === 1) {
    throw new InvalidInventoryInvitationLinkError();
  }
  return new URLSearchParams(hash.slice(1));
}

function validateOnlyFields(values: URLSearchParams, expected: ReadonlySet<string>): void {
  let hasInvalidField = false;
  values.forEach((_value, field) => {
    if (!expected.has(field) || values.getAll(field).length !== 1) {
      hasInvalidField = true;
    }
  });
  if (hasInvalidField) {
    throw new InvalidInventoryInvitationLinkError();
  }
  for (const field of expected) {
    if (values.getAll(field).length !== 1) {
      throw new InvalidInventoryInvitationLinkError();
    }
  }
}

function requiredIdentifier(values: URLSearchParams, field: string): string {
  const value = values.get(field) ?? '';
  if (
    value.length === 0 ||
    value.length > maximumIdentifierLength ||
    !identifierPattern.test(value)
  ) {
    throw new InvalidInventoryInvitationLinkError();
  }
  return value;
}

function requiredToken(values: URLSearchParams): string {
  const value = values.get('token') ?? '';
  if (value.length !== tokenLength || !tokenPattern.test(value)) {
    throw new InvalidInventoryInvitationLinkError();
  }
  return value;
}
