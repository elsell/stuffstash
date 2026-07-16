import { useState } from 'react';
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Check, ExternalLink, MailCheck, Server, ShieldCheck } from 'lucide-react-native';
import { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import {
  OnboardingCommand,
  OnboardingStartState
} from '../../application/onboarding/OnboardingCommand';
import { BrandMark } from '../components/BrandMark';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

type OnboardingScreenProps = {
  readonly command: OnboardingCommand;
  readonly initialApiBaseUrl?: string;
  readonly initialState: OnboardingStartState;
  readonly invitationPending?: boolean;
  readonly onStateChange: (state: OnboardingStartState) => void;
  readonly onComplete: (profile: ConnectionProfile) => void;
};

export function OnboardingScreen({
  command,
  initialApiBaseUrl,
  initialState,
  invitationPending = false,
  onStateChange,
  onComplete
}: OnboardingScreenProps) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const [apiBaseUrl, setApiBaseUrl] = useState(initialState.profile?.apiBaseUrl ?? initialApiBaseUrl ?? '');
  const [tenantName, setTenantName] = useState('');
  const [inventoryName, setInventoryName] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | undefined>();
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function saveInstance(): Promise<void> {
    await submit(async () => {
      const next = await command.saveInstanceUrl({ apiBaseUrl });
      continueWith(next);
    });
  }

  async function createTenant(): Promise<void> {
    const profile = initialState.profile;
    if (!profile) {
      setErrorMessage('Save the instance URL first.');
      return;
    }

    await submit(async () => {
      const next = await command.createTenant({ profile, name: tenantName });
      continueWith(next);
    });
  }

  async function signIn(): Promise<void> {
    const profile = initialState.profile;
    if (!profile) {
      setErrorMessage('Save the instance URL first.');
      return;
    }

    await submit(async () => {
      const next = await command.signIn({ profile });
      continueWith(next);
    });
  }

  async function createInventory(): Promise<void> {
    const profile = initialState.profile;
    if (!profile) {
      setErrorMessage('Create a tenant first.');
      return;
    }

    await submit(async () => {
      const profileAfterInventory = await command.createInventory({
        profile,
        name: inventoryName
      });
      onComplete(profileAfterInventory);
    });
  }

  async function changeInstance(): Promise<void> {
    await submit(async () => {
      await command.reset();
      setApiBaseUrl(initialApiBaseUrl ?? '');
      onStateChange({ step: 'instance' });
    });
  }

  async function submit(action: () => Promise<void>): Promise<void> {
    setIsSubmitting(true);
    setErrorMessage(undefined);
    try {
      await action();
    } catch (error) {
      setErrorMessage(readableError(error));
    } finally {
      setIsSubmitting(false);
    }
  }

  function continueWith(next: OnboardingStartState): void {
    if (next.step === 'complete' && next.profile) {
      onComplete(next.profile);
      return;
    }

    onStateChange(next);
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.keyboardAvoider}
      >
        <ScrollView
          contentContainerStyle={styles.content}
          keyboardShouldPersistTaps="handled"
        >
          <View style={styles.brandRow}>
            <BrandMark showWordmark />
          </View>
          <OnboardingProgress currentStep={initialState.step} />
          {invitationPending ? (
            <View accessibilityRole="summary" style={styles.invitationNotice}>
              <MailCheck color={colors.action} size={20} />
              <Text style={styles.invitationNoticeText}>
                Your invitation is waiting. Finish signing in to review it.
              </Text>
            </View>
          ) : null}
          <Text style={styles.title}>{titleForStep(initialState.step)}</Text>
          <Text style={styles.subtitle}>{subtitleForStep(initialState)}</Text>

          {initialState.step === 'instance' ? (
            <OnboardingTextInput
              autoCapitalize="none"
              keyboardType="url"
              label="Instance URL"
              placeholder="https://stuffstash.example.com"
              value={apiBaseUrl}
              onChangeText={setApiBaseUrl}
              onSubmitEditing={saveInstance}
            />
          ) : null}

          {initialState.step === 'signIn' ? (
            <View style={styles.signInPanel}>
              <View style={styles.signInHeader}>
                <View style={styles.signInIconFrame}>
                  <ShieldCheck color={colors.success} size={24} strokeWidth={2.5} />
                </View>
                <View style={styles.signInHeaderText}>
                  <Text style={styles.signInLabel}>Secure sign-in</Text>
                  <Text style={styles.instanceText} numberOfLines={1}>
                    {initialState.profile?.apiBaseUrl}
                  </Text>
                </View>
              </View>
              <Text style={styles.signInText}>
                Continue with the provider configured by this Stuff Stash instance. You will return here after approval.
              </Text>
              <View style={styles.browserRow}>
                <ExternalLink color={colors.textMuted} size={16} strokeWidth={2.4} />
                <Text style={styles.browserText}>Opens in the system browser</Text>
              </View>
            </View>
          ) : null}

          {initialState.step === 'tenant' ? (
            <OnboardingTextInput
              label="Tenant name"
              placeholder="Ksell Household"
              value={tenantName}
              onChangeText={setTenantName}
              onSubmitEditing={createTenant}
            />
          ) : null}

          {initialState.step === 'inventory' ? (
            <OnboardingTextInput
              label="Inventory name"
              placeholder="Home Inventory"
              value={inventoryName}
              onChangeText={setInventoryName}
              onSubmitEditing={createInventory}
            />
          ) : null}

          {errorMessage ? <Text style={styles.errorText}>{errorMessage}</Text> : null}

          {initialState.step === 'signIn' ? (
            <Pressable
              accessibilityRole="button"
              disabled={isSubmitting}
              onPress={changeInstance}
              style={styles.secondaryButton}
            >
              <Text style={styles.secondaryButtonText}>Change instance</Text>
            </Pressable>
          ) : null}

          <Pressable
            accessibilityRole="button"
            disabled={isSubmitting}
            onPress={
              initialState.step === 'instance'
                ? saveInstance
                : initialState.step === 'signIn'
                  ? signIn
                : initialState.step === 'tenant'
                  ? createTenant
                  : createInventory
            }
            style={[styles.primaryButton, isSubmitting ? styles.primaryButtonDisabled : null]}
          >
            {isSubmitting ? (
              <ActivityIndicator color={colors.onAction} />
            ) : (
              <View style={styles.primaryButtonContent}>
                <Text style={styles.primaryButtonText}>{buttonLabelForStep(initialState.step)}</Text>
                {initialState.step === 'signIn' ? (
                  <View style={styles.primaryButtonIcon}>
                    <ExternalLink color={colors.onAction} size={18} strokeWidth={2.5} />
                  </View>
                ) : null}
              </View>
            )}
          </Pressable>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

function OnboardingProgress({
  currentStep
}: {
  readonly currentStep: OnboardingStartState['step'];
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const steps = onboardingSteps();
  const currentIndex = Math.max(0, steps.findIndex((step) => step.id === currentStep));

  return (
    <View style={styles.progressRow} accessibilityRole="progressbar">
      {steps.map((step, index) => {
        const isComplete = index < currentIndex || currentStep === 'complete';
        const isCurrent = index === currentIndex && currentStep !== 'complete';
        return (
          <View
            key={step.id}
            style={[
              styles.progressStep,
              index < steps.length - 1 ? styles.progressStepWithGap : null
            ]}
          >
            <View
              style={[
                styles.progressDot,
                isComplete ? styles.progressDotComplete : null,
                isCurrent ? styles.progressDotCurrent : null
              ]}
            >
              {isComplete ? (
                <Check color={colors.onAction} size={13} strokeWidth={3} />
              ) : step.id === 'instance' ? (
                <Server color={isCurrent ? colors.onAction : colors.textMuted} size={13} strokeWidth={2.7} />
              ) : null}
            </View>
            <Text
              style={[
                styles.progressLabel,
                isCurrent || isComplete ? styles.progressLabelActive : null
              ]}
              numberOfLines={1}
            >
              {step.label}
            </Text>
          </View>
        );
      })}
    </View>
  );
}

function onboardingSteps(): ReadonlyArray<{
  readonly id: Exclude<OnboardingStartState['step'], 'complete'>;
  readonly label: string;
}> {
  return [
    { id: 'instance', label: 'Instance' },
    { id: 'signIn', label: 'Sign in' },
    { id: 'tenant', label: 'Tenant' },
    { id: 'inventory', label: 'Inventory' }
  ];
}

function OnboardingTextInput({
  label,
  ...inputProps
}: {
  readonly autoCapitalize?: 'none';
  readonly keyboardType?: 'default' | 'url';
  readonly label: string;
  readonly placeholder: string;
  readonly value: string;
  readonly onChangeText: (value: string) => void;
  readonly onSubmitEditing: () => void;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  return (
    <View style={styles.inputGroup}>
      <Text style={styles.inputLabel}>{label}</Text>
      <TextInput
        {...inputProps}
        autoCorrect={false}
        placeholderTextColor={colors.textMuted}
        returnKeyType="next"
        style={styles.input}
      />
    </View>
  );
}

function titleForStep(step: OnboardingStartState['step']): string {
  switch (step) {
    case 'instance':
      return 'Connect Stuff Stash';
    case 'tenant':
      return 'Create tenant';
    case 'signIn':
      return 'Sign in with SSO';
    case 'inventory':
      return 'Create inventory';
    case 'complete':
      return 'Ready';
  }
}

function subtitleForStep(state: OnboardingStartState): string {
  switch (state.step) {
    case 'instance':
      return 'Enter the URL for your Stuff Stash instance.';
    case 'tenant':
      return 'Create the top-level space for this household or organization.';
    case 'signIn':
      return 'Use the sign-in provider configured by this Stuff Stash instance.';
    case 'inventory':
      return `Create the first inventory${state.tenantName ? ` in ${state.tenantName}` : ''}.`;
    case 'complete':
      return 'Stuff Stash is ready.';
  }
}

function buttonLabelForStep(step: OnboardingStartState['step']): string {
  switch (step) {
    case 'instance':
      return 'Continue';
    case 'tenant':
      return 'Create tenant';
    case 'signIn':
      return 'Continue with SSO';
    case 'inventory':
      return 'Create inventory';
    case 'complete':
      return 'Continue';
  }
}

function readableError(error: unknown): string {
  if (!(error instanceof Error)) {
    return 'Setup could not continue. Try again.';
  }

  switch (error.message) {
    case 'Enter a Stuff Stash instance URL.':
    case 'Enter a valid Stuff Stash instance URL.':
    case 'Stuff Stash instance URLs must use HTTP or HTTPS.':
    case 'No usable tenant is available for mobile onboarding.':
      return error.message;
    default:
      return 'Setup could not continue. Check your instance and sign-in settings, then try again.';
  }
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  keyboardAvoider: {
    flex: 1
  },
  content: {
    flexGrow: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  brandRow: {
    marginBottom: spacing.lg
  },
  progressRow: {
    flexDirection: 'row',
    marginBottom: spacing.lg
  },
  progressStep: {
    alignItems: 'center',
    flex: 1,
    minWidth: 0
  },
  progressStepWithGap: {
    marginRight: spacing.xs
  },
  progressDot: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderColor: colors.border,
    borderRadius: 11,
    borderWidth: 1,
    height: 22,
    justifyContent: 'center',
    marginBottom: spacing.xs,
    width: 22
  },
  progressDotComplete: {
    backgroundColor: colors.success,
    borderColor: colors.success
  },
  progressDotCurrent: {
    backgroundColor: colors.action,
    borderColor: colors.action
  },
  progressLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 14,
    maxWidth: '100%'
  },
  progressLabelActive: {
    color: colors.text
  },
  title: {
    color: colors.text,
    fontSize: 32,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 38,
    marginBottom: spacing.sm
  },
  invitationNotice: {
    alignItems: 'center',
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.md,
    flexDirection: 'row',
    marginBottom: spacing.md,
    padding: spacing.md
  },
  invitationNoticeText: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    fontWeight: '700',
    lineHeight: 21,
    marginLeft: spacing.sm
  },
  subtitle: {
    color: colors.textMuted,
    fontSize: 16,
    fontWeight: '600',
    letterSpacing: 0,
    lineHeight: 23,
    marginBottom: spacing.lg
  },
  inputGroup: {
    marginBottom: spacing.md
  },
  inputLabel: {
    color: colors.text,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.xs
  },
  input: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    fontSize: 16,
    minHeight: 48,
    paddingHorizontal: spacing.md
  },
  signInPanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  signInHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    marginBottom: spacing.md
  },
  signInIconFrame: {
    alignItems: 'center',
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.md,
    height: 44,
    justifyContent: 'center',
    marginRight: spacing.sm,
    width: 44
  },
  signInHeaderText: {
    flex: 1,
    minWidth: 0
  },
  signInLabel: {
    color: colors.text,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.xs
  },
  signInText: {
    color: colors.textMuted,
    fontSize: 15,
    fontWeight: '600',
    lineHeight: 21,
    marginBottom: spacing.sm
  },
  browserRow: {
    alignItems: 'center',
    flexDirection: 'row'
  },
  browserText: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    lineHeight: 18,
    marginLeft: spacing.xs
  },
  instanceText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '800',
    lineHeight: 20
  },
  errorText: {
    color: colors.danger,
    fontSize: 14,
    fontWeight: '700',
    lineHeight: 20,
    marginBottom: spacing.md
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 50,
    paddingHorizontal: spacing.md
  },
  primaryButtonContent: {
    alignItems: 'center',
    flexDirection: 'row'
  },
  primaryButtonDisabled: {
    opacity: 0.7
  },
  primaryButtonText: {
    color: colors.onAction,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  primaryButtonIcon: {
    marginLeft: spacing.xs
  },
  secondaryButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    justifyContent: 'center',
    marginBottom: spacing.sm,
    minHeight: 46,
    paddingHorizontal: spacing.md
  },
  secondaryButtonText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  }
  });
}
