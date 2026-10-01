/** Stable transport failure identity; presentation supplies localized recovery copy. */
export class NetworkTimeoutError extends Error {
  readonly name = 'NetworkTimeoutError';
}
