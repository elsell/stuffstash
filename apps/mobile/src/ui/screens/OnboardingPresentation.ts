import { StyleSheet } from 'react-native';
import { MobileAuthenticationRequiredError } from '../../application/auth/MobileAuthSession';
import { OnboardingRecoveryRequiredError } from '../../application/onboarding/HouseholdSetup';
import { spacing, type MobileColorPalette } from '../theme/tokens';

export const initialInventoryName = 'Home Inventory';
export function onboardingError(error: unknown): string {
  if (error instanceof MobileAuthenticationRequiredError) return 'Sign in again to continue setup.';
  if (error instanceof OnboardingRecoveryRequiredError) return error.message;
  if (error instanceof Error) {
    if (error.message === 'Enter a household name.' || error.message === 'Enter an inventory name.') return error.message;
    if (error.message === 'Enter a Stuff Stash instance URL.') return 'Enter your server address.';
    if (error.message === 'Enter a valid Stuff Stash instance URL.' || error.message === 'Stuff Stash instance URLs must use HTTP or HTTPS.') {
      return 'Enter a valid server address using https:// or http://.';
    }
    if (/cancel(?:led|ed)|dismiss/i.test(error.message)) return 'Sign-in was canceled. Try again when you’re ready.';
    if (error.message.startsWith('No usable ') || error.message.startsWith('Your available households changed.')) {
      return 'No inventory is available to this account. Check your access or sign in with another account.';
    }
  }
  return 'Setup could not finish. Check your connection and try again.';
}

export function onboardingStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    shell: { flex: 1, backgroundColor: colors.background },
    content: { flexGrow: 1, padding: spacing.lg },
    brand: { marginBottom: 54 },
    heading: { color: colors.text, fontSize: 30, lineHeight: 36, fontWeight: '700', marginBottom: 28 },
    field: { marginBottom: 22 },
    label: { color: colors.text, fontSize: 14, fontWeight: '600', marginBottom: 9 },
    input: { backgroundColor: colors.surface, borderColor: colors.border, borderWidth: 1,
      borderRadius: 10, minHeight: 54, paddingHorizontal: 14, paddingVertical: 12, color: colors.text, fontSize: 16 },
    helpAction: { minHeight: 44, justifyContent: 'center', alignSelf: 'flex-start' },
    helpLink: { color: colors.action, fontSize: 14, fontWeight: '500' },
    help: { backgroundColor: colors.surfaceMuted, padding: spacing.md, borderRadius: 10 },
    body: { color: colors.textMuted, fontSize: 14, lineHeight: 21 },
    notice: { backgroundColor: colors.surfaceMuted, borderRadius: 10, padding: spacing.md, marginBottom: spacing.lg },
    error: { color: colors.danger, fontSize: 14, lineHeight: 20, marginTop: spacing.sm },
    footer: { marginTop: 'auto', paddingTop: spacing.xl },
    note: { color: colors.textMuted, fontSize: 13, lineHeight: 19, textAlign: 'center', marginBottom: spacing.md },
    button: { minHeight: 54, borderRadius: 12, paddingVertical: 12, paddingHorizontal: spacing.md,
      alignItems: 'center', justifyContent: 'center', backgroundColor: colors.action },
    buttonPressed: { backgroundColor: colors.actionPressed },
    buttonDisabled: { opacity: 0.7 },
    buttonText: { color: colors.onAction, fontSize: 16, fontWeight: '600', textAlign: 'center' },
    ghost: { backgroundColor: 'transparent', marginTop: spacing.sm },
    ghostPressed: { backgroundColor: colors.surfaceMuted },
    ghostText: { color: colors.action }
  });
}
