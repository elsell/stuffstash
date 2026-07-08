export function formatVoiceProviderReadinessLabel(readiness: string): string {
  switch (readiness) {
    case 'ready':
      return 'Ready';
    case 'missing':
      return 'Missing';
    case 'disabled':
      return 'Disabled';
    case 'archived':
      return 'Archived';
    case 'credential_missing':
      return 'Needs credentials';
    case 'untested':
      return 'Needs test';
    case 'duplicate_candidates':
      return 'Choose profile';
    case 'invalid_selection':
      return 'Fix selection';
    default:
      return 'Needs attention';
  }
}

export function formatVoiceProviderCapabilityLabel(capability: string): string {
  switch (capability) {
    case 'speech_to_text':
      return 'Speech input';
    case 'language_inference':
      return 'Agent brain';
    case 'text_to_speech':
      return 'Spoken output';
    default:
      return 'Unknown capability';
  }
}

export function formatVoiceProviderSelectionSourceLabel(selectionSource: string): string {
  switch (selectionSource) {
    case 'explicit':
      return 'Selected';
    case 'implicit':
      return 'Auto-selected';
    case 'missing':
      return 'Missing';
    default:
      return 'Selection unknown';
  }
}

export function formatProviderProfileCredentialStatusLabel(credentialStatus: string): string {
  switch (credentialStatus) {
    case 'configured':
      return 'Configured';
    case 'missing':
      return 'Missing';
    default:
      return 'Unknown';
  }
}

export function formatProviderProfileLifecycleLabel(lifecycleState: string): string {
  switch (lifecycleState) {
    case 'enabled':
      return 'Enabled';
    case 'disabled':
      return 'Disabled';
    case 'archived':
      return 'Archived';
    default:
      return 'Unknown';
  }
}

export function formatProviderProfileTestStatusLabel(lastTestedAt?: string): string {
  return lastTestedAt ? 'Tested' : 'Needs test';
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
