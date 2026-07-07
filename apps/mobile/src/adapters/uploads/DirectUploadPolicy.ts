export type DirectUploadTargetPolicy = {
  readonly allowLocalDevelopmentTargets?: boolean;
};

export function isDirectUploadTargetSupported(value: string, policy: DirectUploadTargetPolicy = {}): boolean {
  if (isSecureDirectUploadURL(value)) {
    return true;
  }
  if (policy.allowLocalDevelopmentTargets !== true) {
    return false;
  }
  return isLocalDirectUploadURL(value) || isPrivateLocalHTTPDirectUploadURL(value);
}

export function isDirectUploadHTTPTransportAllowed(value: string, policy: DirectUploadTargetPolicy = {}): boolean {
  const parsed = parseHTTPURL(value);
  if (!parsed) {
    return false;
  }
  if (parsed.protocol === 'https:') {
    return true;
  }
  return policy.allowLocalDevelopmentTargets === true && isLocalDevelopmentHost(parsed.hostname);
}

export function isLocalDirectUploadURL(value: string): boolean {
  return value.startsWith('stuffstash-local://direct-uploads/');
}

export function directUploadMethod(value: string): 'POST' | 'PUT' | 'PATCH' {
  const method = value.trim().toUpperCase();
  if (method === 'POST' || method === 'PUT' || method === 'PATCH') {
    return method;
  }
  throw new Error('Unsupported direct attachment upload method.');
}

function isSecureDirectUploadURL(value: string): boolean {
  const parsed = parseHTTPURL(value);
  return parsed?.protocol === 'https:';
}

function isPrivateLocalHTTPDirectUploadURL(value: string): boolean {
  const parsed = parseHTTPURL(value);
  return parsed?.protocol === 'http:' && isLocalDevelopmentHost(parsed.hostname);
}

function parseHTTPURL(value: string): URL | undefined {
  try {
    const parsed = new URL(value);
    return parsed.protocol === 'https:' || parsed.protocol === 'http:' ? parsed : undefined;
  } catch {
    return undefined;
  }
}

function isLocalDevelopmentHost(hostname: string): boolean {
  const value = hostname.toLowerCase();
  if (value === 'localhost' || value.endsWith('.local')) {
    return true;
  }
  if (value === '127.0.0.1' || value === '::1' || value === '[::1]') {
    return true;
  }
  const rawOctets = value.split('.');
  if (rawOctets.length !== 4 || rawOctets.some((part) => !/^\d+$/.test(part))) {
    return false;
  }
  const octets = rawOctets.map((part) => Number.parseInt(part, 10));
  if (octets.some((part) => part < 0 || part > 255)) {
    return false;
  }
  const [first, second] = octets;
  return first === 10
    || (first === 172 && second >= 16 && second <= 31)
    || (first === 192 && second === 168);
}
