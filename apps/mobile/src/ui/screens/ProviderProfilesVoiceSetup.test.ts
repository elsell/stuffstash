import { describe, expect, it, vi } from 'vitest';
import { ProviderProfileSummary, VoiceProviderConfiguration } from '../../application/providerProfiles/ProviderProfileRepository';
import { VoiceSetupPanel } from './ProviderProfilesVoiceSetup';
import {
  formatProviderProfileCredentialStatusLabel,
  formatProviderProfileLifecycleLabel,
  formatProviderProfileTestStatusLabel,
  formatVoiceProviderCapabilityLabel,
  formatVoiceProviderReadinessLabel,
  formatVoiceProviderSelectionSourceLabel,
  voiceProviderSetupIssueLabels
} from './ProviderProfilesVoiceSetupPresentation';

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Pressable: 'Pressable',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  View: 'View'
}));

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

describe('VoiceSetupPanel', () => {
  it('renders safe setup labels instead of raw backend metadata', () => {
    const tree = VoiceSetupPanel({
      configuration: configuration({
        slots: [
          {
            capability: 'speech_to_text',
            label: 'Speech input',
            selectedProfileId: 'profile-speech',
            selectedProfile: profile({
              id: 'profile-speech',
              capability: 'speech_to_text',
              credentialStatus: 'apiKey:secret',
              lastTestedAt: 'providerSessionId:last-test',
              lifecycleState: 'stack_trace_here'
            }),
            selectionSource: 'providerSessionId:abc123',
            readiness: 'raw_prompt_injected',
            issues: ['bearer secret stack_trace_here'],
            recommendedAction: 'none',
            duplicateProfiles: []
          }
        ]
      }),
      profiles: [],
      onEditCredential: () => undefined,
      onSelectSlotProfile: () => undefined,
      onTestProfile: () => undefined
    });

    expect(textContent(tree)).toContain('Needs attention');
    expect(textContent(tree)).toContain('Speech input / Selection unknown');
    expect(textContent(tree)).toContain('Unknown');
    expect(textContent(tree)).not.toContain('raw_prompt_injected');
    expect(textContent(tree)).not.toContain('providerSessionId:abc123');
    expect(textContent(tree)).not.toContain('providerSessionId:last-test');
    expect(textContent(tree)).not.toContain('apiKey:secret');
    expect(textContent(tree)).not.toContain('bearer secret stack_trace_here');
    expect(textContent(tree)).not.toContain('stack_trace_here');
  });
});

function profile(overrides: Partial<ProviderProfileSummary> = {}): ProviderProfileSummary {
  return {
    id: 'profile-language',
    capability: 'language_inference',
    providerKind: 'gemini',
    displayName: 'Gemini language',
    modelName: 'gemini-2.5-flash-lite',
    credentialStatus: 'configured',
    lifecycleState: 'enabled',
    hasPromptTemplate: false,
    ...overrides
  };
}

function configuration(overrides: Partial<VoiceProviderConfiguration> = {}): VoiceProviderConfiguration {
  return {
    tenantId: 'tenant-home',
    readiness: 'needs_attention',
    profileIds: {},
    slots: [],
    ...overrides
  };
}

function textContent(node: unknown): string {
  if (node === undefined || node === null || typeof node === 'boolean') {
    return '';
  }
  if (typeof node === 'string' || typeof node === 'number') {
    return String(node);
  }
  if (Array.isArray(node)) {
    return node.map(textContent).join('');
  }
  if (!isElementNode(node)) {
    return '';
  }
  if (typeof node.type === 'function') {
    return textContent(node.type(node.props));
  }
  return textContent(node.props?.children);
}

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}

type ElementNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};

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
    ['ready', []],
    ['missing', ['Choose a provider profile for this slot.']],
    ['disabled', ['Enable the selected provider profile.']],
    ['archived', ['Choose an active provider profile.']],
    ['credential_missing', ['Add a credential for the selected profile.']],
    ['untested', ['Test the selected profile before using voice.']],
    ['duplicate_candidates', ['Choose which ready profile this voice slot should use.']],
    ['invalid_selection', ['Choose a valid profile for this slot.']],
    ['raw_prompt_injected', ['Review this voice provider slot.']]
  ])('derives safe issue labels from readiness %s', (readiness, labels) => {
    expect(voiceProviderSetupIssueLabels(readiness)).toEqual(labels);
  });

  it('uses safe fallbacks for unknown setup metadata', () => {
    expect(formatVoiceProviderCapabilityLabel('raw_prompt_injected')).toBe('Unknown capability');
    expect(formatVoiceProviderSelectionSourceLabel('providerSessionId:abc123')).toBe('Selection unknown');
    expect(formatProviderProfileCredentialStatusLabel('apiKey:secret')).toBe('Unknown');
    expect(formatProviderProfileLifecycleLabel('stack_trace_here')).toBe('Unknown');
  });
});
