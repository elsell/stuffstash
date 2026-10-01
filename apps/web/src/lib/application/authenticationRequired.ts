import { t } from '$lib/presentation/localization';
export class AuthenticationRequiredError extends Error {
  readonly status = 401;

  constructor(message = t('web.authenticationRequired.authenticationRequired')) {
    super(message);
    this.name = 'AuthenticationRequiredError';
  }
}

export function isAuthenticationRequiredError(error: unknown): boolean {
  return error instanceof AuthenticationRequiredError;
}
