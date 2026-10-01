import { t } from '../../presentation/localization';
import { OnboardingPartialSetupError } from '../../application/onboarding/HouseholdSetup';
import { useEffect, useRef, useState } from 'react';
import {
  AccessibilityInfo, ActivityIndicator, findNodeHandle, KeyboardAvoidingView,
  Platform, Pressable, ScrollView, Text, View, type TextInputProps
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import { OnboardingCommand, OnboardingSupersededError, type OnboardingStartState } from '../../application/onboarding/OnboardingCommand';
import { MobileAuthenticationRequiredError } from '../../application/auth/MobileAuthSession';
import { OnboardingAddressInput } from './OnboardingAddressInput';
import { BrandMark } from '../components/BrandMark';
import { NativeCommandButton } from '../components/NativeCommandButton';
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
  const title = connection ? t('mobile.OnboardingScreen.connectToStuffStash') : household ? t('mobile.OnboardingScreen.setUpYourHousehold') : t('mobile.OnboardingScreen.createYourFirstInventory');
  const actionLabel = connection ? t('mobile.OnboardingScreen.connectAndSignIn') : household ? t('mobile.OnboardingScreen.createHousehold') : t('mobile.OnboardingScreen.createInventory');
  const requiredMessage = connection
    ? (!apiBaseUrl.trim() ? t('mobile.OnboardingScreen.enterAServerAddressToContinue') : undefined)
    : household && !householdName.trim()
      ? t('mobile.OnboardingScreen.enterAHouseholdNameToContinue')
      : !inventoryName.trim() ? t('mobile.OnboardingScreen.enterAnInventoryNameToContinue') : undefined;
  const actionDisabled = submitting || Boolean(requiredMessage);

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
    if (actionDisabled) return;
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
    const current = generation.current;
    await submit(async () => {
      await command.reset();
      if (current !== generation.current) return;
      setApiBaseUrl(''); setHouseholdName(''); setInventoryName(initialInventoryName);
      setHelpVisible(false);
      onStartOver?.();
      onStateChange({ step: 'instance' });
    });
  }

  function input(label: string, value: string, onChangeText: (value: string) => void, placeholder: string, url = false) {
    return <View style={styles.field}>
      <Text style={styles.label}>{label}</Text>
      {url ? <OnboardingAddressInput initialValue={value} onChangeText={onChangeText}
        disabled={submitting} onSubmit={() => void proceed()} /> : <OnboardingTextInput key={label} accessibilityLabel={label} initialValue={value} onChangeText={onChangeText}
        placeholder={placeholder} placeholderTextColor={colors.textMuted} autoCorrect={false}
        autoCapitalize="sentences" keyboardType="default"
        editable={!submitting} returnKeyType="go" onSubmitEditing={() => void proceed()} style={styles.input} />}
    </View>;
  }

  return <SafeAreaView style={styles.shell} edges={['top', 'left', 'right', 'bottom']}>
    <KeyboardAvoidingView style={styles.shell} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView contentContainerStyle={styles.content} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
        <View style={styles.form}>
          <View style={styles.brand}><BrandMark showWordmark /></View>
          <Text ref={heading} accessibilityRole="header" style={styles.heading}>{title}</Text>
          {invitationPending ? <View style={styles.notice}><Text style={styles.body}>{t('mobile.OnboardingScreen.yourInvitationIsWaitingSignInToReviewIt')}</Text></View> : null}
          {connection ? <>
            {input(t('mobile.OnboardingScreen.serverAddress'), apiBaseUrl, setApiBaseUrl, 'https://stash.example.com', true)}
            <Pressable accessibilityRole="button" accessibilityLabel={t('mobile.OnboardingScreen.needHelpConnecting')}
              accessibilityState={{ expanded: helpVisible }} onPress={() => setHelpVisible(value => !value)} style={styles.helpAction}>
              <Text style={styles.helpLink}>{t('mobile.OnboardingScreen.needHelpConnecting')}</Text>
            </Pressable>
            {helpVisible ? <View style={styles.help}><Text style={styles.body}>{t('mobile.OnboardingScreen.enterYourStuffStashServerSFullAddressIncluding')}{'\n\n'}{t('mobile.OnboardingScreen.youLlNeedARunningStuffStashServerTo')}</Text></View> : null}
          </> : <>
            {household ? input(t('mobile.OnboardingScreen.householdName'), householdName, setHouseholdName, t("onboarding.householdExample")) : null}
            {input(household ? t('mobile.OnboardingScreen.firstInventory') : t('mobile.OnboardingScreen.inventoryName'), inventoryName, setInventoryName, t("onboarding.inventoryExample"))}
          </>}
          {error ? <Text accessibilityRole="alert" accessibilityLiveRegion="assertive" style={styles.error}>{error}</Text> : null}
          <View style={styles.footer}>
            {requiredMessage ? <Text style={styles.note}>{requiredMessage}</Text> : null}
            {connection ? <Text style={styles.note}>{t('mobile.OnboardingScreen.yourBrowserWillOpenForSignInThenBring')}</Text> : null}
            {submitting ? <ActivityIndicator accessibilityLabel={t('mobile.OnboardingScreen.setupInProgress')} color={colors.action} /> : null}
            <NativeCommandButton label={actionLabel} prominence="primary" disabled={actionDisabled} onPress={() => void proceed()} />
            {!connection ? <NativeCommandButton label={t('mobile.OnboardingScreen.signOutAndStartOver')} disabled={submitting} onPress={() => void startOver()} /> : null}
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  </SafeAreaView>;
}


function OnboardingTextInput({ initialValue, ...props }: TextInputProps & { readonly initialValue: string }) {
  // Preserve native editing state across validation/submission renders. A new
  // field key seeds a new form step; ordinary keystrokes never replace its text.
  const seed = useRef(initialValue);
  return <AppTextInput {...props} defaultValue={seed.current} />;
}
