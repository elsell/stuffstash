import type { ImportSourceRequest } from '$lib/domain/inventory';

export type ImportSourceChoice = 'homebox_live' | 'homebox_csv';

export type ImportCSVSelection = {
  name: string;
  size: number;
  lastModified: number;
};

export type ImportRequestDraft = {
  sourceChoice: ImportSourceChoice;
  baseUrl: string;
  username: string;
  password: string;
  includeImages: boolean;
  allowPrivateNetwork: boolean;
  allowInsecureTLS: boolean;
  fileName: string;
  contentBase64: string;
  csvSelection: ImportCSVSelection | null;
};

export function buildImportSourceRequest(draft: ImportRequestDraft): ImportSourceRequest {
  if (draft.sourceChoice === 'homebox_csv') {
    return {
      sourceType: 'legacy_homebox_csv',
      fileName: draft.fileName || undefined,
      contentBase64: draft.contentBase64 || undefined
    };
  }

  return {
    sourceType: 'legacy_homebox',
    baseUrl: normalizedHomeboxURL(draft.baseUrl),
    username: draft.username.trim() || undefined,
    password: draft.password,
    includeImages: draft.includeImages,
    allowPrivateNetwork: draft.allowPrivateNetwork,
    allowInsecureTLS: draft.allowInsecureTLS
  };
}

export function normalizeImportSourceRequest(input: ImportSourceRequest): ImportSourceRequest {
  if (input.sourceType === 'legacy_homebox_csv') {
    return {
      sourceType: 'legacy_homebox_csv',
      fileName: input.fileName || undefined,
      contentBase64: input.contentBase64 || undefined
    };
  }

  return {
    sourceType: 'legacy_homebox',
    baseUrl: normalizedHomeboxURL(input.baseUrl ?? ''),
    username: input.username?.trim() || undefined,
    password: input.password,
    includeImages: input.includeImages,
    allowPrivateNetwork: input.allowPrivateNetwork,
    allowInsecureTLS: input.allowInsecureTLS
  };
}

export function importSourceRequestKey(draft: ImportRequestDraft): string {
  if (draft.sourceChoice === 'homebox_csv') {
    return JSON.stringify({
      sourceType: 'legacy_homebox_csv',
      fileName: draft.fileName,
      fileSize: draft.csvSelection?.size ?? 0,
      fileLastModified: draft.csvSelection?.lastModified ?? 0,
      contentLength: draft.contentBase64.length,
      contentFingerprint: importSourceKeyFingerprint(draft.contentBase64)
    });
  }

  return JSON.stringify({
    sourceType: 'legacy_homebox',
    baseUrl: normalizedHomeboxURL(draft.baseUrl),
    username: draft.username.trim(),
    passwordFingerprint: importSourceKeyFingerprint(draft.password),
    includeImages: draft.includeImages,
    allowPrivateNetwork: draft.allowPrivateNetwork,
    allowInsecureTLS: draft.allowInsecureTLS
  });
}

export function normalizedHomeboxURL(value: string): string | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  if (/^https?:\/\//i.test(trimmed)) return trimmed;
  return `https://${trimmed}`;
}

function importSourceKeyFingerprint(value: string): string {
  let hash = 0x811c9dc5;
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return `${value.length}:${(hash >>> 0).toString(16)}`;
}

export function readableImportActionError(value: unknown, fallback: string): string {
  if (!isErrorLike(value)) return fallback;
  if (isGenericInvalidRequest(value)) return fallback;
  return value.message || fallback;
}

function isGenericInvalidRequest(value: Error & { status?: number; code?: string }): boolean {
  const status = value.status ?? 0;
  const safeValidationStatus = status === 400 || status === 422;
  const message = value.message.trim().toLowerCase();
  return safeValidationStatus && value.code === 'invalid_request' && (message === 'invalid request.' || message === 'validation failed');
}

function isErrorLike(value: unknown): value is Error & { status?: number; code?: string } {
  return value instanceof Error;
}
