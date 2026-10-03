import { t } from '../../presentation/localization';

/** Explicitly presentable recovery owned by the client catalog, never transport text. */
export class CatalogRecoveryError extends Error {
  constructor(readonly key: Parameters<typeof t>[0]) {
    super(t(key));
    this.name = 'CatalogRecoveryError';
  }
}

export function catalogRecoveryMessage(error: unknown, fallback: string): string {
  return error instanceof CatalogRecoveryError ? t(error.key) : fallback;
}
