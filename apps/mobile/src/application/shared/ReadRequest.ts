export type ReadRequest = {
  readonly signal?: AbortSignal;
};

/** Uses the AbortSignal API shared by React Native and browser runtimes. */
export function assertReadActive(signal?: Pick<AbortSignal, 'aborted'> & { readonly reason?: unknown }): void {
  if (!signal?.aborted) return;
  if (signal.reason !== undefined) throw signal.reason;
  const error = new Error('Read cancelled.');
  error.name = 'AbortError';
  throw error;
}
