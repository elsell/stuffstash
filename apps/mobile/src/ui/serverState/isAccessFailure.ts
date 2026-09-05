import { CustomizationFailure } from '../../application/customization/CustomizationErrors';
import { SettingsScopeUnavailableError } from '../../application/settings/SettingsQuery';
import { MobileAuthenticationRequiredError } from '../../application/auth/MobileAuthSession';
export function isAccessFailure(error: unknown): boolean {
  if (error instanceof CustomizationFailure && (error.kind === 'permission-denied' || error.kind === 'not-found')) return true;
  if (error instanceof SettingsScopeUnavailableError || error instanceof MobileAuthenticationRequiredError) return true;
  return typeof error === 'object' && error !== null && 'status' in error && [401, 403, 404].includes(Number(error.status));
}
