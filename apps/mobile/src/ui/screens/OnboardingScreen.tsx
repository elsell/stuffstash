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
import { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import {
  OnboardingCommand,
  OnboardingStartState
} from '../../application/onboarding/OnboardingCommand';
import { BrandMark } from '../components/BrandMark';
import { colors, radius, spacing } from '../theme/tokens';

type OnboardingScreenProps = {
  readonly command: OnboardingCommand;
  readonly initialApiBaseUrl?: string;
  readonly initialState: OnboardingStartState;
  readonly onStateChange: (state: OnboardingStartState) => void;
  readonly onComplete: (profile: ConnectionProfile) => void;
};

export function OnboardingScreen({
  command,
  initialApiBaseUrl,
  initialState,
  onStateChange,
  onComplete
}: OnboardingScreenProps) {
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

          <Pressable
            accessibilityRole="button"
            disabled={isSubmitting}
            onPress={
              initialState.step === 'instance'
                ? saveInstance
                : initialState.step === 'tenant'
                  ? createTenant
                  : createInventory
            }
            style={[styles.primaryButton, isSubmitting ? styles.primaryButtonDisabled : null]}
          >
            {isSubmitting ? (
              <ActivityIndicator color={colors.onAction} />
            ) : (
              <Text style={styles.primaryButtonText}>{buttonLabelForStep(initialState.step)}</Text>
            )}
          </Pressable>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
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
      return 'Create Tenant';
    case 'inventory':
      return 'Create Inventory';
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
    case 'inventory':
      return 'Create inventory';
    case 'complete':
      return 'Continue';
  }
}

function readableError(error: unknown): string {
  return error instanceof Error ? error.message : 'Onboarding failed.';
}

const styles = StyleSheet.create({
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
  title: {
    color: colors.text,
    fontSize: 32,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 38,
    marginBottom: spacing.sm
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
  primaryButtonDisabled: {
    opacity: 0.7
  },
  primaryButtonText: {
    color: colors.onAction,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  }
});
