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
