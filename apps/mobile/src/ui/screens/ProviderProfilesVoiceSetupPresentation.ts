import { t } from '../../presentation/localization';
export function formatVoiceProviderReadinessLabel(readiness: string): string {
  switch (readiness) {
    case 'ready':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.ready');
    case 'missing':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.missing');
    case 'disabled':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.disabled');
    case 'archived':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.archived');
    case 'credential_missing':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.needsCredentials');
    case 'untested':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.needsTest');
    case 'duplicate_candidates':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.chooseProfile');
    case 'invalid_selection':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.fixSelection');
    default:
      return t('mobile.ProviderProfilesVoiceSetupPresentation.needsAttention');
  }
}

export function formatVoiceProviderCapabilityLabel(capability: string): string {
  switch (capability) {
    case 'speech_to_text':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.speechInput');
    case 'language_inference':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.agentBrain');
    case 'text_to_speech':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.spokenOutput');
    default:
      return t('mobile.ProviderProfilesVoiceSetupPresentation.unknownCapability');
  }
}

export function formatVoiceProviderSelectionSourceLabel(selectionSource: string): string {
  switch (selectionSource) {
    case 'explicit':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.selected');
    case 'implicit':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.autoSelected');
    case 'missing':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.missing');
    default:
      return t('mobile.ProviderProfilesVoiceSetupPresentation.selectionUnknown');
  }
}

export function formatProviderProfileCredentialStatusLabel(credentialStatus: string): string {
  switch (credentialStatus) {
    case 'configured':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.configured');
    case 'missing':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.missing');
    default:
      return t('mobile.ProviderProfilesVoiceSetupPresentation.unknown');
  }
}

export function formatProviderProfileLifecycleLabel(lifecycleState: string): string {
  switch (lifecycleState) {
    case 'enabled':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.enabled');
    case 'disabled':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.disabled');
    case 'archived':
      return t('mobile.ProviderProfilesVoiceSetupPresentation.archived');
    default:
      return t('mobile.ProviderProfilesVoiceSetupPresentation.unknown');
  }
}

export function formatProviderProfileTestStatusLabel(lastTestedAt?: string): string {
  return lastTestedAt ? t('mobile.ProviderProfilesVoiceSetupPresentation.tested') : t('mobile.ProviderProfilesVoiceSetupPresentation.needsTest');
}

export function voiceProviderSetupIssueLabels(readiness: string, recommendedAction: string): readonly string[] {
  switch (recommendedAction) {
    case 'none':
      return readiness === 'ready' ? [] : voiceProviderSetupIssueLabelsForReadiness(readiness);
    case 'add_profile':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.chooseAProviderProfileForThisSlot')];
    case 'choose_profile':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.chooseWhichProfileThisVoiceSlotShouldUse')];
    case 'replace_credential':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.addACredentialForTheSelectedProfile')];
    case 'enable_profile':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.enableTheSelectedProviderProfile')];
    case 'test_profile':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.testTheSelectedProfileBeforeUsingVoice')];
    default:
      return voiceProviderSetupIssueLabelsForReadiness(readiness);
  }
}

function voiceProviderSetupIssueLabelsForReadiness(readiness: string): readonly string[] {
  switch (readiness) {
    case 'ready':
      return [];
    case 'missing':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.chooseAProviderProfileForThisSlot')];
    case 'disabled':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.enableTheSelectedProviderProfile')];
    case 'archived':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.chooseAnActiveProviderProfile')];
    case 'credential_missing':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.addACredentialForTheSelectedProfile')];
    case 'untested':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.testTheSelectedProfileBeforeUsingVoice')];
    case 'duplicate_candidates':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.chooseWhichReadyProfileThisVoiceSlotShouldUse')];
    case 'invalid_selection':
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.chooseAValidProfileForThisSlot')];
    default:
      return [t('mobile.ProviderProfilesVoiceSetupPresentation.reviewThisVoiceProviderSlot')];
  }
}
