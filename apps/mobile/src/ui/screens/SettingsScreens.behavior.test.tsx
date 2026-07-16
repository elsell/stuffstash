import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ManageProviderProfileCommand } from '../../application/providerProfiles/ManageProviderProfileCommand';
import type {
  CreateProviderProfileInput,
  ProviderProfileLifecycleAction,
  ProviderProfileRepository,
  ProviderProfileSummary,
  ProviderProfileTestResult,
  ReplaceProviderProfileCredentialInput,
  UpdateProviderProfileInput,
  UpdateVoiceProviderConfigurationInput,
  VoiceProviderConfiguration,
  VoiceProviderRecommendedAction,
  VoiceProviderSlotReadiness
} from '../../application/providerProfiles/ProviderProfileRepository';
import { ProviderProfileSettingsQuery } from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import { TestProviderProfileCommand } from '../../application/providerProfiles/TestProviderProfileCommand';
import type { SettingsViewModel } from '../../application/settings/SettingsQuery';
import {
  AccountSettingsScreen,
  AppearanceSettingsScreen,
  ConnectionSettingsScreen
} from './SettingsDetailScreens';
import { SettingsScreen } from './SettingsScreen';
import {
  ProviderCredentialScreen,
  VoiceCapabilityScreen
} from './VoiceSettingsScreens';

const hooks = vi.hoisted(() => ({ index: 0, values: [] as unknown[] }));
const alerts = vi.hoisted(() => vi.fn());
const feedback = vi.hoisted(() => ({ showNotice: vi.fn(), showDialog: vi.fn() }));
const appearance = vi.hoisted(() => ({
  preference: 'system' as 'system' | 'light' | 'dark',
  setPreference: vi.fn()
}));

vi.mock('react', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react')>()),
  useCallback: <T,>(callback: T) => callback,
  useEffect: vi.fn(),
  useMemo: <T,>(factory: () => T) => factory(),
  useRef: <T,>(initial: T) => {
    const index = hooks.index++;
    if (index >= hooks.values.length) hooks.values[index] = { current: initial };
    return hooks.values[index] as { current: T };
  },
  useState: <T,>(initial: T) => {
    const index = hooks.index++;
    if (index >= hooks.values.length) hooks.values[index] = initial;
    const setValue = (next: T | ((current: T) => T)) => {
      const current = hooks.values[index] as T;
      hooks.values[index] = typeof next === 'function'
        ? (next as (current: T) => T)(current)
        : next;
    };
    return [hooks.values[index] as T, setValue] as const;
  }
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Alert: { alert: alerts },
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: { create: <T,>(styles: T) => styles },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View',
  useWindowDimensions: () => ({ fontScale: 1, height: 844, width: 390 })
}));

vi.mock('lucide-react-native', () => ({
  Activity: 'ActivityIcon',
  AudioLines: 'AudioLinesIcon',
  Check: 'CheckIcon',
  ChevronRight: 'ChevronRightIcon',
  Info: 'InfoIcon',
  Server: 'ServerIcon',
  Share2: 'ShareIcon',
  SunMedium: 'SunIcon',
  UserRound: 'UserIcon'
}));

vi.mock('../feedback/AppFeedback', () => ({ useAppFeedback: () => feedback }));
vi.mock('../theme/AppearanceContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../theme/AppearanceContext')>()),
  useAppearance: () => ({
    isHydrated: true,
    palette,
    preference: appearance.preference,
    resolvedColorScheme: 'light',
    setPreference: appearance.setPreference
  }),
  useAppearancePalette: () => palette
}));

describe('Settings detail behavior', () => {
  beforeEach(resetHarness);

  it('orders appearance choices System, Light, then Dark and marks the selection', () => {
    appearance.preference = 'system';
    hooks.values = [];

    const tree = render(AppearanceSettingsScreen, {});
    const choices = findAllByType(tree, 'Pressable').filter((node) => node.props?.accessibilityRole === 'radio');

    expect(choices.map((node) => node.props?.accessibilityLabel)).toEqual([
      'System appearance',
      'Light appearance',
      'Dark appearance'
    ]);
    expect(choices.map((node) => node.props?.accessibilityState)).toEqual([
      { checked: true },
      { checked: false },
      { checked: false }
    ]);
  });

  it('confirms sign out with exact saved-server language and reports a rejected action', async () => {
    const onSignOut = vi.fn().mockRejectedValue(new Error('Secure session could not be cleared.'));
    hooks.values = [false, { current: false }, readySettingsState()];
    const tree = render(AccountSettingsScreen, { onSignOut, settingsQuery: settingsQuery() });

    press(findByLabel(tree, 'Sign out john@example.com'));
    expect(alerts).toHaveBeenCalledWith(
      'Sign out?',
      'You’ll need to sign in again as john@example.com. This Stuff Stash server will stay saved on your device.',
      [
        { text: 'Cancel', style: 'cancel' },
        expect.objectContaining({ text: 'Sign Out' })
      ]
    );

    await pressAlertAction('Sign Out');
    expect(onSignOut).toHaveBeenCalledOnce();
    expect(feedback.showNotice).toHaveBeenCalledWith({
      tone: 'error',
      title: 'Could not sign out',
      message: 'Secure session could not be cleared.'
    });
  });

  it('confirms server change with exact non-deletion language and reports a rejected action', async () => {
    const onChangeServer = vi.fn().mockRejectedValue(new Error('Profile reset failed safely.'));
    hooks.values = [false, { current: false }, readySettingsState()];
    const tree = render(ConnectionSettingsScreen, { onChangeServer, settingsQuery: settingsQuery() });

    press(findByLabel(tree, 'Change Stuff Stash server from stash.home.test'));
    expect(alerts).toHaveBeenCalledWith(
      'Change Stuff Stash server?',
      'You’ll be signed out of stash.home.test, and this device will forget its saved server and household selection. Your Stuff Stash data won’t be deleted.',
      [
        { text: 'Cancel', style: 'cancel' },
        expect.objectContaining({ text: 'Change Server' })
      ]
    );

    await pressAlertAction('Change Server');
    expect(onChangeServer).toHaveBeenCalledOnce();
    expect(feedback.showNotice).toHaveBeenCalledWith({
      tone: 'error',
      title: 'Could not change server',
      message: 'Profile reset failed safely.'
    });
  });
});

describe('Settings root behavior', () => {
  beforeEach(resetHarness);

  it('does not render tenant Voice Setup for a principal without configure permission', () => {
    hooks.values = [readySettingsState(['view']), undefined];
    const tree = render(SettingsScreen, {
      onNavigate: vi.fn(),
      providerProfileSettingsQuery: { execute: vi.fn() } as never,
      settingsQuery: settingsQuery(['view'])
    });

    expect(findText(tree, 'Voice Setup')).toBeUndefined();
    expect(findText(tree, 'Tenant Administration')).toBeUndefined();
  });

  it('offers Retry after a root load failure and re-executes the settings query', async () => {
    const execute = vi.fn().mockResolvedValue(settings());
    hooks.values = [{ status: 'error', message: 'Settings are unavailable.' }, undefined];
    const tree = render(SettingsScreen, {
      onNavigate: vi.fn(),
      providerProfileSettingsQuery: { execute: vi.fn() } as never,
      settingsQuery: { execute } as never
    });

    press(findTextButton(tree, 'Retry'));
    await flushPromises();
    expect(execute).toHaveBeenCalledOnce();
  });
});

describe('Voice capability actions', () => {
  beforeEach(resetHarness);

  it.each([
    ['add_profile', 'Add Profile', 'add', 'Add provider profile for Understand'],
    ['replace_credential', 'Add Credential', 'credential', 'Add credential for Gemini language in Understand']
  ] as const)('routes %s directly through the expected action', (recommendedAction, label, expected, accessibilityLabel) => {
    const repository = new FakeProviderRepository(slot(recommendedAction));
    const onAddProfile = vi.fn();
    const onEditCredential = vi.fn();
    hooks.values = [readyProviderState(repository), false, { current: false }];
    const tree = renderVoiceCapability(repository, { onAddProfile, onEditCredential });

    expect(findByLabel(tree, accessibilityLabel)).toBeDefined();
    press(findTextButton(tree, label));
    expect(expected === 'add' ? onAddProfile : onEditCredential).toHaveBeenCalledOnce();
    if (expected === 'credential') expect(onEditCredential).toHaveBeenCalledWith('profile-language');
  });

  it.each([
    ['test_profile', 'Test Connection', 'test', 'Test Gemini language for Understand'],
    ['enable_profile', 'Enable Service', 'enable', 'Enable Gemini language for Understand']
  ] as const)('runs %s once, disables the action while busy, and ignores a second submission', async (recommendedAction, label, expected, accessibilityLabel) => {
    const pending = deferred<ProviderProfileSummary | ProviderProfileTestResult>();
    const repository = new FakeProviderRepository(slot(recommendedAction));
    repository.pendingAction = pending.promise;
    hooks.values = [readyProviderState(repository), false, { current: false }];

    let tree = renderVoiceCapability(repository);
    expect(findByLabel(tree, accessibilityLabel)).toBeDefined();
    const action = findTextButton(tree, label);
    press(action);
    press(action);
    tree = renderVoiceCapability(repository);
    const busyButton = findTextButton(tree, expected === 'test' ? 'Testing…' : 'Enabling…');

    expect(busyButton?.props?.disabled).toBe(true);
    if (!busyButton?.props?.disabled) press(busyButton);
    expect(expected === 'test' ? repository.testCalls : repository.lifecycleCalls).toHaveLength(1);

    pending.resolve(expected === 'test' ? testResult() : profile({ lifecycleState: 'enabled' }));
    await flushPromises();
  });
});

describe('Provider credential behavior', () => {
  beforeEach(resetHarness);

  it('uses a secure non-correcting input and saves before navigation', async () => {
    const repository = new FakeProviderRepository(slot('replace_credential'));
    const pending = deferred<ProviderProfileSummary>();
    repository.pendingCredential = pending.promise;
    const onSaved = vi.fn();
    hooks.values = [readyProviderState(repository), '', false, { current: false }];

    let tree = render(ProviderCredentialScreen, {
      manageCommand: new ManageProviderProfileCommand(repository),
      onCancel: vi.fn(),
      onSaved,
      profileId: 'profile-language',
      query: new ProviderProfileSettingsQuery(repository)
    });
    const input = findByLabel(tree, 'API key');
    expect(input?.props).toMatchObject({
      autoCapitalize: 'none',
      autoCorrect: false,
      secureTextEntry: true
    });

    (input?.props?.onChangeText as ((value: string) => void) | undefined)?.('  secret-value  ');
    tree = render(ProviderCredentialScreen, {
      manageCommand: new ManageProviderProfileCommand(repository),
      onCancel: vi.fn(),
      onSaved,
      profileId: 'profile-language',
      query: new ProviderProfileSettingsQuery(repository)
    });
    const save = findTextButton(tree, 'Save Credential');
    press(save);
    press(save);

    expect(repository.credentialInputs).toEqual([{
      providerProfileId: 'profile-language',
      purpose: 'api_key',
      credential: 'secret-value'
    }]);
    expect(onSaved).not.toHaveBeenCalled();
    expect(feedback.showNotice).not.toHaveBeenCalledWith(expect.objectContaining({ title: 'Credential saved' }));

    pending.resolve(profile({ credentialStatus: 'configured' }));
    await flushPromises();
    expect(feedback.showNotice).toHaveBeenCalledWith({
      tone: 'success',
      title: 'Credential saved',
      message: 'Gemini language is ready to test.'
    });
    expect(onSaved).toHaveBeenCalledOnce();
  });
});

class FakeProviderRepository implements ProviderProfileRepository {
  readonly profile = profile({});
  readonly configuration: VoiceProviderConfiguration;
  credentialInputs: ReplaceProviderProfileCredentialInput[] = [];
  lifecycleCalls: Array<{ id: string; action: ProviderProfileLifecycleAction }> = [];
  testCalls: string[] = [];
  pendingAction?: Promise<ProviderProfileSummary | ProviderProfileTestResult>;
  pendingCredential?: Promise<ProviderProfileSummary>;

  constructor(voiceSlot = slot('none')) {
    this.configuration = {
      tenantId: 'tenant-home', readiness: voiceSlot.readiness === 'ready' ? 'ready' : 'needs_attention',
      profileIds: { languageInference: voiceSlot.selectedProfileId }, slots: [voiceSlot]
    };
  }
  async listProviderProfiles() { return [this.profile]; }
  async getVoiceProviderConfiguration() { return this.configuration; }
  async updateVoiceProviderConfiguration(_input: UpdateVoiceProviderConfigurationInput) { return this.configuration; }
  async createProviderProfile(_input: CreateProviderProfileInput) { return this.profile; }
  async updateProviderProfile(_input: UpdateProviderProfileInput) { return this.profile; }
  async replaceProviderProfileCredential(input: ReplaceProviderProfileCredentialInput) {
    this.credentialInputs.push(input);
    return this.pendingCredential ?? this.profile;
  }
  async changeProviderProfileLifecycle(id: string, action: ProviderProfileLifecycleAction) {
    this.lifecycleCalls.push({ id, action });
    return (this.pendingAction ?? this.profile) as Promise<ProviderProfileSummary>;
  }
  async testProviderProfile(id: string) {
    this.testCalls.push(id);
    return (this.pendingAction ?? testResult()) as Promise<ProviderProfileTestResult>;
  }
}

function renderVoiceCapability(repository: FakeProviderRepository, overrides: Partial<Parameters<typeof VoiceCapabilityScreen>[0]> = {}) {
  return render(VoiceCapabilityScreen, {
    capability: 'language_inference',
    manageCommand: new ManageProviderProfileCommand(repository),
    onAddProfile: vi.fn(),
    onEditCredential: vi.fn(),
    onEditProfile: vi.fn(),
    query: new ProviderProfileSettingsQuery(repository),
    testCommand: new TestProviderProfileCommand(repository),
    ...overrides
  });
}

function settings(permissions: readonly string[] = ['view', 'configure']): SettingsViewModel {
  return {
    principal: { id: 'principal-subject', primaryLabel: 'john@example.com' },
    selectedTenant: { id: 'tenant-home', name: 'Home', permissions },
    selectedInventory: { id: 'inventory-home', name: 'Household', permissions: ['view', 'share'] },
    serverUrl: 'https://stash.home.test/api', appVersion: '0.0.0', authenticationMode: 'oidc-sso'
  };
}
function readySettingsState(permissions?: readonly string[]) { return { status: 'ready', settings: settings(permissions) } as const; }
function settingsQuery(permissions?: readonly string[]) { return { execute: async () => settings(permissions) } as never; }
function readyProviderState(repository: FakeProviderRepository) {
  return { status: 'ready', viewModel: { profiles: [repository.profile], configuration: repository.configuration, missingCapabilities: [] } } as const;
}
function profile(overrides: Partial<ProviderProfileSummary>): ProviderProfileSummary {
  return { id: 'profile-language', capability: 'language_inference', providerKind: 'gemini', displayName: 'Gemini language', modelName: 'gemini-2.5-flash-lite', credentialStatus: 'missing', credentialPurpose: 'api_key', lifecycleState: 'disabled', hasPromptTemplate: false, ...overrides };
}
function slot(recommendedAction: VoiceProviderRecommendedAction) {
  const selected = recommendedAction === 'add_profile' ? undefined : profile({});
  const readiness: VoiceProviderSlotReadiness = recommendedAction === 'test_profile' ? 'untested' : recommendedAction === 'enable_profile' ? 'disabled' : recommendedAction === 'replace_credential' ? 'credential_missing' : recommendedAction === 'add_profile' ? 'missing' : 'ready';
  return { capability: 'language_inference', label: 'Language inference', selectedProfileId: selected?.id, selectedProfile: selected, selectionSource: selected ? 'explicit' : 'missing', readiness, issues: [], recommendedAction, duplicateProfiles: [] } as const;
}
function testResult(): ProviderProfileTestResult { return { providerProfileId: 'profile-language', capability: 'language_inference', providerKind: 'gemini', status: 'success', message: 'Succeeded.', testedAt: '2026-07-14T12:00:00Z' }; }

function resetHarness() { hooks.index = 0; hooks.values = []; alerts.mockReset(); feedback.showNotice.mockReset(); feedback.showDialog.mockReset(); appearance.preference = 'system'; appearance.setPreference.mockReset(); }
function render<Props>(component: (props: Props) => unknown, props: Props): unknown {
  hooks.index = 0;
  return materialize(component(props));
}
function findByLabel(tree: unknown, label: string) { return findAll(tree).find((node) => node.props?.accessibilityLabel === label); }
function findText(tree: unknown, text: string) { return findAllByType(tree, 'Text').find((node) => node.props?.children === text); }
function findTextButton(tree: unknown, text: string) { const textNode = findText(tree, text); return textNode ? findAll(tree).find((node) => node.type === 'Pressable' && contains(node, textNode)) : undefined; }
function press(node: ElementNode | undefined) { expect(node, 'expected an interactive element').toBeDefined(); (node?.props?.onPress as (() => void) | undefined)?.(); }
async function pressAlertAction(label: string) { const buttons = alerts.mock.calls.at(-1)?.[2] as Array<{ text: string; onPress?: () => void }>; buttons.find((button) => button.text === label)?.onPress?.(); await flushPromises(); }
async function flushPromises() { await Promise.resolve(); await Promise.resolve(); await Promise.resolve(); }
function contains(root: unknown, target: unknown): boolean { if (root === target) return true; return findAll(root).some((node) => node === target); }
function findAllByType(tree: unknown, type: unknown) { return findAll(tree).filter((node) => node.type === type); }
function findAll(tree: unknown): ElementNode[] { if (Array.isArray(tree)) return tree.flatMap(findAll); if (!isElement(tree)) return []; return [tree, ...findAll(tree.props?.children)]; }
function materialize(tree: unknown): unknown {
  if (Array.isArray(tree)) return tree.map(materialize);
  if (!isElement(tree)) return tree;
  if (typeof tree.type === 'function') return materialize(tree.type(tree.props));
  return { ...tree, props: { ...tree.props, children: materialize(tree.props?.children) } };
}
function isElement(value: unknown): value is ElementNode { return Boolean(value && typeof value === 'object' && 'props' in value); }
type ElementNode = { readonly type?: unknown; readonly props?: Record<string, unknown> & { readonly children?: unknown } };
function deferred<T>() { let resolve!: (value: T) => void; const promise = new Promise<T>((done) => { resolve = done; }); return { promise, resolve }; }

const palette = {
  accent: '#6B90AA', action: '#0066CC', background: '#F7FAFB', border: '#C5D0D7',
  brandAmber: '#F5AB4B', brandCharcoal: '#243038', brandCharcoalDeep: '#172027',
  brandDustyBlueSoft: '#E8F0F5', controlBorder: '#B8C4CC', danger: '#B42318',
  onAction: '#FFFFFF', selected: '#DCECF8', success: '#18794E', surface: '#FFFFFF',
  surfaceMuted: '#EEF3F6', text: '#243038', textMuted: '#52616B', warning: '#8A4F00',
  warningSurface: '#FFF3DF'
} as const;
