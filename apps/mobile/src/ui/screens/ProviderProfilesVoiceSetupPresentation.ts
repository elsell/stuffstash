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
      return ['Choose a provider profile for this slot.'];
    case 'choose_profile':
      return ['Choose which profile this voice slot should use.'];
    case 'replace_credential':
      return ['Add a credential for the selected profile.'];
    case 'enable_profile':
      return ['Enable the selected provider profile.'];
    case 'test_profile':
      return ['Test the selected profile before using voice.'];
    default:
      return voiceProviderSetupIssueLabelsForReadiness(readiness);
  }
}

function voiceProviderSetupIssueLabelsForReadiness(readiness: string): readonly string[] {
  switch (readiness) {
    case 'ready':
      return [];
    case 'missing':
      return ['Choose a provider profile for this slot.'];
    case 'disabled':
      return ['Enable the selected provider profile.'];
    case 'archived':
      return ['Choose an active provider profile.'];
    case 'credential_missing':
      return ['Add a credential for the selected profile.'];
    case 'untested':
      return ['Test the selected profile before using voice.'];
    case 'duplicate_candidates':
      return ['Choose which ready profile this voice slot should use.'];
    case 'invalid_selection':
      return ['Choose a valid profile for this slot.'];
    default:
      return ['Review this voice provider slot.'];
  }
}
