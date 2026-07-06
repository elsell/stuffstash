export class AuthenticationRequiredError extends Error {
  readonly status = 401;

  constructor(message = 'Authentication required.') {
    super(message);
    this.name = 'AuthenticationRequiredError';
  }
}

export function isAuthenticationRequiredError(error: unknown): boolean {
  return error instanceof AuthenticationRequiredError;
}
