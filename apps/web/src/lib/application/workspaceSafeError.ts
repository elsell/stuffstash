export function safeWorkspaceErrorMessage(caught: unknown, fallback: string): string {
  const safeForUser = typeof caught === 'object' && caught !== null &&
    (caught as { safeForUser?: unknown }).safeForUser === true;
  if (isGenericSafeValidation(caught)) {
    return fallback;
  }
  if (safeForUser && caught instanceof Error && caught.message.trim()) {
    return caught.message.trim();
  }
  return fallback;
}

function isGenericSafeValidation(caught: unknown): boolean {
  if (!(caught instanceof Error)) return false;
  const adapterError = caught as Error & { status?: number; code?: string };
  if ((adapterError.status !== 400 && adapterError.status !== 422) || adapterError.code !== 'invalid_request') {
    return false;
  }
  const message = adapterError.message.trim().toLowerCase();
  return message === 'invalid request.' || message === 'validation failed';
}
