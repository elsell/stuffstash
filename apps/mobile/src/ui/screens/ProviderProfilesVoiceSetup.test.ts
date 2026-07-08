import { describe, expect, it } from 'vitest';
import { formatVoiceProviderReadinessLabel } from './ProviderProfilesVoiceSetupPresentation';

describe('formatVoiceProviderReadinessLabel', () => {
  it.each([
    ['ready', 'Ready'],
    ['missing', 'Missing'],
    ['disabled', 'Disabled'],
    ['archived', 'Archived'],
    ['credential_missing', 'Needs credentials'],
    ['untested', 'Needs test'],
    ['duplicate_candidates', 'Choose profile'],
    ['invalid_selection', 'Fix selection']
  ])('maps %s to a product-owned setup label', (readiness, label) => {
    expect(formatVoiceProviderReadinessLabel(readiness)).toBe(label);
  });

  it('does not render unknown backend readiness values directly', () => {
    expect(formatVoiceProviderReadinessLabel('providerSessionId:abc123')).toBe('Needs attention');
    expect(formatVoiceProviderReadinessLabel('raw_prompt_injected')).toBe('Needs attention');
  });
});
