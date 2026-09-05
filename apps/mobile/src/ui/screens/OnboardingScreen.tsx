import { OnboardingPartialSetupError } from '../../application/onboarding/HouseholdSetup';
import { useEffect, useRef, useState } from 'react';
import {
  AccessibilityInfo, ActivityIndicator, findNodeHandle, KeyboardAvoidingView,
  Platform, Pressable, ScrollView, Text, View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import { OnboardingCommand, OnboardingSupersededError, type OnboardingStartState } from '../../application/onboarding/OnboardingCommand';
import { MobileAuthenticationRequiredError } from '../../application/auth/MobileAuthSession';
import { BrandMark } from '../components/BrandMark';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { initialInventoryName, onboardingError, onboardingStyles } from './OnboardingPresentation';

type OnboardingScreenProps = {
  readonly command: OnboardingCommand;
  readonly initialApiBaseUrl?: string;
  readonly initialState: OnboardingStartState;
  readonly invitationPending?: boolean;
  readonly onStateChange: (state: OnboardingStartState) => void;
  readonly onComplete: (profile: ConnectionProfile) => void;
  readonly onStartOver?: () => void;
};

export function OnboardingScreen({ command, initialApiBaseUrl, initialState, invitationPending = false,
  onStateChange, onComplete, onStartOver }: OnboardingScreenProps) {
  const colors = useAppearanceAwarePalette();
  const styles = onboardingStyles(colors);
  const [apiBaseUrl, setApiBaseUrl] = useState(initialState.profile?.apiBaseUrl ?? initialApiBaseUrl ?? '');
  const [householdName, setHouseholdName] = useState('');
  const [inventoryName, setInventoryName] = useState(initialInventoryName);
  const [helpVisible, setHelpVisible] = useState(false);
  const [error, setError] = useState<string>();
  const [submitting, setSubmitting] = useState(false);
  const pending = useRef(false);
  const generation = useRef(0);
  const heading = useRef<Text>(null);
  const connection = initialState.step === 'instance' || initialState.step === 'signIn';
  const household = initialState.step === 'tenant';
  const title = connection ? 'Connect to Stuff Stash' : household ? 'Set up your household' : 'Create your first inventory';
  const actionLabel = connection ? 'Connect and sign in' : household ? 'Create household' : 'Create inventory';

  useEffect(() => {
    const handle = findNodeHandle(heading.current);
    if (handle) AccessibilityInfo.setAccessibilityFocus(handle);
  }, [title]);
  useEffect(() => () => { generation.current++; }, []);

  async function submit(action: () => Promise<void>) {
    if (pending.current) return;
    pending.current = true;
    const current = generation.current;
    setSubmitting(true);
    setError(undefined);
    try { await action(); }
    catch (failure) {
      if (current !== generation.current || failure instanceof OnboardingSupersededError) return;
      if (failure instanceof OnboardingPartialSetupError) {
        onStateChange(failure.state);
        setError(`${failure.message} ${onboardingError(failure.failure)}`);
        return;
      }
      if (failure instanceof MobileAuthenticationRequiredError && initialState.profile) {
        onStateChange({ step: 'signIn', profile: initialState.profile });
      }
      setError(onboardingError(failure));
    } finally {
      if (current === generation.current) { pending.current = false; setSubmitting(false); }
    }
  }

  async function proceed() {
    const current = generation.current;
    await submit(async () => {
      let next: OnboardingStartState;
      if (connection) next = await command.connectAndSignIn({ apiBaseUrl });
      else {
        const profile = initialState.profile;
        if (!profile) throw new MobileAuthenticationRequiredError();
        next = household
          ? await command.createHousehold({ profile, householdName, inventoryName })
          : await command.completeInventorySetup({ profile, inventoryName });
      }
      if (current !== generation.current) return;
      if (next.step === 'complete' && next.profile) onComplete(next.profile);
      else onStateChange(next);
    });
  }

  async function startOver() {
    await submit(async () => {
      await command.reset();
      setApiBaseUrl(''); setHouseholdName(''); setInventoryName(initialInventoryName);
      setHelpVisible(false);
      onStartOver?.();
      onStateChange({ step: 'instance' });
    });
  }

  function input(label: string, value: string, onChangeText: (value: string) => void, placeholder: string, url = false) {
    return <View style={styles.field}>
      <Text style={styles.label}>{label}</Text>
      <AppTextInput accessibilityLabel={label} value={value} onChangeText={onChangeText}
        placeholder={placeholder} placeholderTextColor={colors.textMuted} autoCorrect={false}
        autoCapitalize={url ? 'none' : 'sentences'} keyboardType={url ? 'url' : 'default'}
        editable={!submitting} returnKeyType="go" onSubmitEditing={() => void proceed()} style={styles.input} />
    </View>;
  }

  return <SafeAreaView style={styles.shell} edges={['top', 'left', 'right', 'bottom']}>
    <KeyboardAvoidingView style={styles.shell} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView contentContainerStyle={styles.content} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
        <View style={styles.brand}><BrandMark showWordmark /></View>
        <Text ref={heading} accessibilityRole="header" style={styles.heading}>{title}</Text>
        {invitationPending ? <View style={styles.notice}><Text style={styles.body}>Your invitation is waiting. Sign in to review it.</Text></View> : null}
        {connection ? <>
          {input('Server address', apiBaseUrl, setApiBaseUrl, 'https://stash.example.com', true)}
          <Pressable accessibilityRole="button" accessibilityLabel="Need help connecting?"
            accessibilityState={{ expanded: helpVisible }} onPress={() => setHelpVisible(value => !value)} style={styles.helpAction}>
            <Text style={styles.helpLink}>Need help connecting?</Text>
          </Pressable>
          {helpVisible ? <View style={styles.help}><Text style={styles.body}>
            Enter your Stuff Stash server’s full address, including a port or path if needed.
            {'\n\n'}You’ll need a running Stuff Stash server to connect. If you’re joining someone else’s inventory, ask them for its server address.
          </Text></View> : null}
        </> : <>
          {household ? input('Household name', householdName, setHouseholdName, 'e.g. Maple Street household') : null}
          {input(household ? 'First inventory' : 'Inventory name', inventoryName, setInventoryName, 'e.g. Home Inventory')}
        </>}
        {error ? <Text accessibilityRole="alert" accessibilityLiveRegion="assertive" style={styles.error}>{error}</Text> : null}
        <View style={styles.footer}>
          {connection ? <Text style={styles.note}>Your browser will open for sign-in, then bring you back here.</Text> : null}
          <Pressable accessibilityRole="button" accessibilityLabel={actionLabel} accessibilityState={{ disabled: submitting, busy: submitting }}
            disabled={submitting} onPress={proceed} style={({ pressed }) => [styles.button, pressed && styles.buttonPressed, submitting && styles.buttonDisabled]}>
            {submitting ? <ActivityIndicator accessibilityLabel="Setup in progress" color={colors.onAction} /> : <Text style={styles.buttonText}>{actionLabel}</Text>}
          </Pressable>
          {!connection ? <Pressable accessibilityRole="button" accessibilityLabel="Sign out and start over" disabled={submitting}
            onPress={startOver} style={({ pressed }) => [styles.button, styles.ghost, pressed && styles.ghostPressed, submitting && styles.buttonDisabled]}>
            <Text style={[styles.buttonText, styles.ghostText]}>Sign out and start over</Text>
          </Pressable> : null}
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  </SafeAreaView>;
}
