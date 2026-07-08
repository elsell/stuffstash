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
