import type { InvitationLinkMaterial } from '$lib/domain/invitation';

const invitationPath = '/invitations/accept';
const tokenLength = 43;
const maximumIdentifierLength = 128;
const maximumURLLength = 4096;

export function parseInvitationLink(value: string, expectedOrigin: string): InvitationLinkMaterial | null {
  if (value.length === 0 || value.length > maximumURLLength) return null;
  let url: URL;
  try {
    url = new URL(value, expectedOrigin);
  } catch {
    return null;
  }
  if (url.origin !== expectedOrigin || url.pathname !== invitationPath || url.username || url.password) return null;
  if (!hasOnlyFields(url.searchParams, new Set(['tenant', 'inventory', 'invitation']))) return null;

  const tenantId = exactlyOne(url.searchParams, 'tenant');
  const inventoryId = exactlyOne(url.searchParams, 'inventory');
  const invitationId = exactlyOne(url.searchParams, 'invitation');
  const fragment = new URLSearchParams(url.hash.slice(1));
  if (!hasOnlyFields(fragment, new Set(['token']))) return null;
  const token = exactlyOne(fragment, 'token');
  if (!validIdentifier(tenantId) || !validIdentifier(inventoryId) || !validIdentifier(invitationId) || !validToken(token)) {
    return null;
  }
  return { tenantId, inventoryId, invitationId, token };
}

export function invitationReturnPath(material: InvitationLinkMaterial): string {
  const query = new URLSearchParams({
    tenant: material.tenantId,
    inventory: material.inventoryId,
    invitation: material.invitationId
  });
  const fragment = new URLSearchParams({ token: material.token });
  return `${invitationPath}?${query.toString()}#${fragment.toString()}`;
}

function hasOnlyFields(params: URLSearchParams, allowed: ReadonlySet<string>): boolean {
  let valid = true;
  params.forEach((_value, key) => {
    if (!allowed.has(key)) valid = false;
  });
  return valid;
}

function exactlyOne(params: URLSearchParams, key: string): string | null {
  const values = params.getAll(key);
  return values.length === 1 ? values[0] : null;
}

function validIdentifier(value: string | null): value is string {
  return value !== null && value.length > 0 && value.length <= maximumIdentifierLength && !/[\\\u0000-\u001f\u007f]/.test(value);
}

function validToken(value: string | null): value is string {
  return value !== null && value.length === tokenLength && /^[A-Za-z0-9_-]+$/.test(value);
}
