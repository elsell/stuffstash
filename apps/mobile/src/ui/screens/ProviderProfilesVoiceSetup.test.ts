import { describe, expect, it } from 'vitest';
import {
  formatProviderProfileCredentialStatusLabel,
  formatProviderProfileLifecycleLabel,
  formatProviderProfileTestStatusLabel,
  formatVoiceProviderCapabilityLabel,
  formatVoiceProviderReadinessLabel,
  formatVoiceProviderSelectionSourceLabel,
  voiceProviderSetupIssueLabels
} from './ProviderProfilesVoiceSetupPresentation';

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

describe('voice provider setup labels', () => {
  it.each([
    ['speech_to_text', 'Speech input'],
    ['language_inference', 'Agent brain'],
    ['text_to_speech', 'Spoken output']
  ])('maps capability %s to a product-owned setup label', (capability, label) => {
    expect(formatVoiceProviderCapabilityLabel(capability)).toBe(label);
  });

  it.each([
    ['explicit', 'Selected'],
    ['implicit', 'Auto-selected'],
    ['missing', 'Missing']
  ])('maps selection source %s to a product-owned setup label', (selectionSource, label) => {
    expect(formatVoiceProviderSelectionSourceLabel(selectionSource)).toBe(label);
  });

  it.each([
    ['configured', 'Configured'],
    ['missing', 'Missing']
  ])('maps credential status %s to a product-owned setup label', (credentialStatus, label) => {
    expect(formatProviderProfileCredentialStatusLabel(credentialStatus)).toBe(label);
  });

  it.each([
    ['enabled', 'Enabled'],
    ['disabled', 'Disabled'],
    ['archived', 'Archived']
  ])('maps lifecycle state %s to a product-owned setup label', (lifecycleState, label) => {
    expect(formatProviderProfileLifecycleLabel(lifecycleState)).toBe(label);
  });

  it('maps test timestamps to bounded status labels without rendering timestamp text', () => {
    expect(formatProviderProfileTestStatusLabel('providerSessionId:last-test')).toBe('Tested');
    expect(formatProviderProfileTestStatusLabel(undefined)).toBe('Needs test');
  });

  it.each([
    ['ready', 'none', []],
    ['missing', 'add_profile', ['Choose a provider profile for this slot.']],
    ['invalid_selection', 'choose_profile', ['Choose which profile this voice slot should use.']],
    ['credential_missing', 'replace_credential', ['Add a credential for the selected profile.']],
    ['disabled', 'enable_profile', ['Enable the selected provider profile.']],
    ['untested', 'test_profile', ['Test the selected profile before using voice.']]
  ])('derives safe issue labels from recommended action %s', (readiness, recommendedAction, labels) => {
    expect(voiceProviderSetupIssueLabels(readiness, recommendedAction)).toEqual(labels);
  });

  it.each([
    ['missing', ['Choose a provider profile for this slot.']],
    ['disabled', ['Enable the selected provider profile.']],
    ['archived', ['Choose an active provider profile.']],
    ['credential_missing', ['Add a credential for the selected profile.']],
    ['untested', ['Test the selected profile before using voice.']],
    ['duplicate_candidates', ['Choose which ready profile this voice slot should use.']],
    ['invalid_selection', ['Choose a valid profile for this slot.']],
    ['raw_prompt_injected', ['Review this voice provider slot.']]
  ])('falls back to safe readiness labels for unknown action keys on %s', (readiness, labels) => {
    expect(voiceProviderSetupIssueLabels(readiness, 'providerSessionId:action')).toEqual(labels);
  });

  it('uses safe fallbacks for unknown setup metadata', () => {
    expect(formatVoiceProviderCapabilityLabel('raw_prompt_injected')).toBe('Unknown capability');
    expect(formatVoiceProviderSelectionSourceLabel('providerSessionId:abc123')).toBe('Selection unknown');
    expect(formatProviderProfileCredentialStatusLabel('apiKey:secret')).toBe('Unknown');
    expect(formatProviderProfileLifecycleLabel('stack_trace_here')).toBe('Unknown');
  });
});
