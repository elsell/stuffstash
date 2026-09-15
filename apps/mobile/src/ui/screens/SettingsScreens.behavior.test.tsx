import { setScreenFocused, attemptNavigation, dispatchedActions, resetNavigation } from '../../test-support/navigation';
import { AddProviderProfileScreen, ProviderProfileDetailScreen, ProviderProfileListScreen } from './ProviderProfileScreens';
import React from 'react';
import { AppearanceProvider } from '../theme/AppearanceContext';
import { AppearancePreferenceController, type AppearancePreference } from '../../application/settings/AppearancePreference';
import { describe, expect, it } from 'vitest';
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
  ProviderPromptScreen,
  VoiceCapabilityScreen
} from './VoiceSettingsScreens';

import { SettingsQuery } from '../../application/settings/SettingsQuery';
import { HouseholdSettingsScreen } from './ScopedSettingsScreens';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { latestAlert, pressAlertButton } from '../../test-support/react-native';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
async function mount(element: React.ReactElement) {
  const harness = new MobileRenderHarness();
  const client = createMobileQueryClient();
  await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant-home', inventoryId: 'inventory-home' })}><AppFeedbackProvider>{element}</AppFeedbackProvider></MobileServerStateProvider>);
  await settle(harness); await settle(harness);
  return { harness, client };
}
function settingsQuery(permissions: readonly string[] = ['view', 'configure'], getScope?: () => Promise<Awaited<ReturnType<SettingsQuery['getSelectedScope']>>>) {
  return new SettingsQuery({ getCurrentPrincipal: async () => ({ id: 'principal', email: 'john@example.com' }) }, { getDiagnostics: () => ({ apiBaseUrl: 'https://stash.home.test/api', appVersion: 'test', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: getScope ?? (async () => ({ tenant: { id: 'tenant-home', name: 'Home', permissions }, inventory: { id: 'inventory-home', name: 'Household', permissions: ['view', 'share'] } })) });
}
const textButton = (harness: MobileRenderHarness, label: string) => harness.allByType('Pressable').find((node) => node.queryAll((child) => child.type === 'Text' && child.children.includes(label)).length > 0);

describe('mounted Settings behavior', () => {
  it('orders appearance choices and marks the selected radio', async () => {
    const { harness } = await mount(<AppearanceSettingsScreen />);
    try {
      await harness.press(harness.byLabel('Choose appearance'));
      const choices = harness.allByType('Pressable').filter((node) => node.props.accessibilityRole === 'radio');
      expect(choices.map((node) => node.props.accessibilityLabel)).toEqual(['System', 'Light', 'Dark']);
      expect(choices.map((node) => node.props.accessibilityState)).toEqual([{ checked: true }, { checked: false }, { checked: false }]);
    } finally { await harness.unmount(); }
  });
  it('changes appearance in place and restores the saved choice after persistence failure', async () => {
    let saved: AppearancePreference = 'system'; let fail = false;
    const controller = new AppearancePreferenceController({ load: async () => saved, save: async next => { if (fail) throw new Error('Storage unavailable'); saved = next; } });
    const navigation: string[] = [];
    const { harness } = await mount(<AppearanceProvider controller={controller}><SettingsScreen settingsQuery={settingsQuery()} onNavigate={destination => navigation.push(destination)} /></AppearanceProvider>);
    try {
      await harness.press(harness.byLabel('Choose appearance'));
      await harness.press(harness.byLabel('Dark'));
      expect(saved).toBe('dark'); expect(navigation).toEqual([]);
      fail = true;
      await harness.press(harness.byLabel('Choose appearance'));
      await harness.press(harness.byLabel('Light'));
      expect(saved).toBe('dark'); expect(navigation).toEqual([]);
      expect(harness.byText('Appearance not saved')).toBeDefined();
      await harness.press(harness.byLabel('Choose appearance'));
      expect(harness.byLabel('Dark')?.props.accessibilityState.checked).toBe(true);
    } finally { await harness.unmount(); }
  });
  it.each(['signOut', 'server'] as const)('confirms %s and reports a rejected action', async (kind) => {
    let calls = 0;
    const action = async () => { calls++; throw new Error('Action failed safely.'); };
    const { harness } = await mount(kind === 'signOut' ? <AccountSettingsScreen settingsQuery={settingsQuery()} onSignOut={action} /> : <ConnectionSettingsScreen settingsQuery={settingsQuery()} onChangeServer={action} />);
    try {
      await harness.press(harness.byLabel(kind === 'signOut' ? 'Sign out john@example.com' : 'Change Stuff Stash server from stash.home.test'));
      expect(latestAlert()?.message).toContain(kind === 'signOut' ? 'This Stuff Stash server will stay saved on your device.' : 'Your Stuff Stash data won’t be deleted.');
      await harness.run(() => pressAlertButton(kind === 'signOut' ? 'Sign Out' : 'Change Server')); await settle(harness);
      expect(calls).toBe(1);
      expect(harness.allText()).toContain(kind === 'signOut' ? 'Could not sign out' : 'Could not change server');
    } finally { await harness.unmount(); }
  });
  it('denies household configuration for a viewer', async () => {
    const { harness } = await mount(<HouseholdSettingsScreen settingsQuery={settingsQuery(['view'])} onNavigate={() => undefined} />);
    try { expect(harness.allText()).toContain('Settings unavailable'); expect(harness.byLabel('Open Voice setup for Home')).toBeUndefined(); }
    finally { await harness.unmount(); }
  });
  it('retries a failed scope request', async () => {
    let calls = 0;
    const query = settingsQuery([], async () => { if (++calls === 1) throw new Error('Unavailable'); return { tenant: { id: 'tenant-home', name: 'Home', permissions: [] }, inventory: { id: 'inventory-home', name: 'Household', permissions: [] } }; });
    const { harness } = await mount(<SettingsScreen settingsQuery={query} onNavigate={() => undefined} />);
    try { await harness.press(textButton(harness, 'Retry')); await settle(harness); expect(calls).toBe(2); expect(harness.allText()).toContain('Household'); }
    finally { await harness.unmount(); }
  });
});

it('opens provider profiles without requesting voice configuration', async () => {
  const repository = new FakeProviderRepository();
  let configReads = 0;
  repository.getVoiceProviderConfiguration = () => { configReads++; return new Promise(() => undefined); };
  const { harness } = await mount(<ProviderProfileListScreen query={new ProviderProfileSettingsQuery(repository)} onAdd={() => undefined} onOpenProfile={() => undefined} />);
  try { expect(harness.allText()).toContain('Gemini language'); expect(configReads).toBe(0); }
  finally { await harness.unmount(); }
});

describe('mounted voice settings actions', () => {
  it.each([['add_profile', 'Add Profile'], ['replace_credential', 'Add Credential']] as const)('routes %s to its focused action', async (action, label) => {
    const repository = new FakeProviderRepository(slot(action));
    const navigation: string[] = [];
    const { harness } = await mount(<VoiceCapabilityScreen capability="language_inference" manageCommand={new ManageProviderProfileCommand(repository)} testCommand={new TestProviderProfileCommand(repository)} query={new ProviderProfileSettingsQuery(repository)} onAddProfile={() => navigation.push('add')} onEditCredential={(id) => navigation.push(id)} onEditProfile={() => undefined} />);
    try { await harness.press(textButton(harness, label)); expect(navigation).toEqual([action === 'add_profile' ? 'add' : 'profile-language']); }
    finally { await harness.unmount(); }
  });
  it.each([['test_profile', 'Test Connection'], ['enable_profile', 'Enable Service']] as const)('suppresses duplicate %s submissions', async (action, label) => {
    const repository = new FakeProviderRepository(slot(action));
    const pending = deferred<ProviderProfileSummary | ProviderProfileTestResult>();
    repository.pendingAction = pending.promise;
    const { harness } = await mount(<VoiceCapabilityScreen capability="language_inference" manageCommand={new ManageProviderProfileCommand(repository)} testCommand={new TestProviderProfileCommand(repository)} query={new ProviderProfileSettingsQuery(repository)} onAddProfile={() => undefined} onEditCredential={() => undefined} onEditProfile={() => undefined} />);
    try {
      const button = textButton(harness, label)!;
      await harness.run(() => { button.props.onPress(); button.props.onPress(); });
      expect(action === 'test_profile' ? repository.testCalls : repository.lifecycleCalls).toHaveLength(1);
      await harness.run(() => pending.resolve(action === 'test_profile' ? testResult() : profile({ lifecycleState: 'enabled' })));
    } finally { await harness.unmount(); }
  });
  it('keeps credential input secure and out of query caches, and navigates only after saving', async () => {
    const repository = new FakeProviderRepository(slot('replace_credential'));
    const pending = deferred<ProviderProfileSummary>(); repository.pendingCredential = pending.promise;
    let saved = 0;
    const { harness, client } = await mount(<ProviderCredentialScreen manageCommand={new ManageProviderProfileCommand(repository)} onSaved={() => { saved++; }} profileId="profile-language" query={new ProviderProfileSettingsQuery(repository)} />);
    try {
      expect(harness.byLabel('API key')?.props).toMatchObject({ autoCapitalize: 'none', autoCorrect: false, secureTextEntry: true });
      await harness.changeText(harness.byLabel('API key'), '  secret-value  ');
      const button = harness.byLabel('Save Credential')!;
      await harness.run(() => { button.props.onPress(); button.props.onPress(); });
      expect(repository.credentialInputs).toEqual([{ providerProfileId: 'profile-language', purpose: 'api_key', credential: 'secret-value' }]);
      expect(saved).toBe(0);
      expect(JSON.stringify(client.getQueryCache().getAll().map((query) => query.state.data))).not.toContain('secret-value');
      await harness.run(() => pending.resolve(profile({ credentialStatus: 'configured' }))); await settle(harness);
      expect(saved).toBe(1);
    } finally { await harness.unmount(); }
  });
});

class FakeProviderRepository implements ProviderProfileRepository {
  readonly profile = profile({});
  extraProfiles: ProviderProfileSummary[] = [];
  pendingSelection?: Promise<VoiceProviderConfiguration>;
  selectionInputs: UpdateVoiceProviderConfigurationInput[] = [];
  readonly configuration: VoiceProviderConfiguration;
  credentialInputs: ReplaceProviderProfileCredentialInput[] = [];
  lifecycleCalls: Array<{ id: string; action: ProviderProfileLifecycleAction }> = [];
  testCalls: string[] = [];
  pendingAction?: Promise<ProviderProfileSummary | ProviderProfileTestResult>;
  pendingCredential?: Promise<ProviderProfileSummary>;
  pendingPrompt?: Promise<ProviderProfileSummary>;
  pendingCreation?: Promise<ProviderProfileSummary>;

  constructor(voiceSlot = slot('none')) {
    this.configuration = {
      tenantId: 'tenant-home', readiness: voiceSlot.readiness === 'ready' ? 'ready' : 'needs_attention',
      profileIds: { languageInference: voiceSlot.selectedProfileId }, slots: [voiceSlot]
    };
  }
  async listProviderProfiles() { return [this.profile, ...this.extraProfiles]; }
  async getVoiceProviderConfiguration() { return this.configuration; }
  async updateVoiceProviderConfiguration(input: UpdateVoiceProviderConfigurationInput) { this.selectionInputs.push(input); return this.pendingSelection ?? this.configuration; }
  async createProviderProfile(_input: CreateProviderProfileInput) { return this.pendingCreation ?? this.profile; }
  async updateProviderProfile(_input: UpdateProviderProfileInput) { return this.pendingPrompt ?? this.profile; }
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

function profile(overrides: Partial<ProviderProfileSummary>): ProviderProfileSummary {
  return { id: 'profile-language', capability: 'language_inference', providerKind: 'gemini', displayName: 'Gemini language', modelName: 'gemini-2.5-flash-lite', credentialStatus: 'missing', credentialPurpose: 'api_key', lifecycleState: 'disabled', hasPromptTemplate: false, ...overrides };
}
function slot(recommendedAction: VoiceProviderRecommendedAction) {
  const selected = recommendedAction === 'add_profile' ? undefined : profile({});
  const readiness: VoiceProviderSlotReadiness = recommendedAction === 'test_profile' ? 'untested' : recommendedAction === 'enable_profile' ? 'disabled' : recommendedAction === 'replace_credential' ? 'credential_missing' : recommendedAction === 'add_profile' ? 'missing' : 'ready';
  return { capability: 'language_inference', label: 'Language inference', selectedProfileId: selected?.id, selectedProfile: selected, selectionSource: selected ? 'explicit' : 'missing', readiness, issues: [], recommendedAction, duplicateProfiles: [] } as const;
}
function testResult(): ProviderProfileTestResult { return { providerProfileId: 'profile-language', capability: 'language_inference', providerKind: 'gemini', status: 'success', message: 'Succeeded.', testedAt: '2026-07-14T12:00:00Z' }; }

function deferred<T>() { let resolve!: (value: T) => void; let reject!: (error: Error) => void; const promise = new Promise<T>((done, fail) => { resolve = done; reject = fail; }); return { promise, resolve, reject }; }

 it.each(['test', 'enable', 'archive'] as const)('shows only the pending %s profile operation and restores actions after failure', async operation => {
  const repository = new FakeProviderRepository();
  let rejectAction: ((error: Error) => void) | undefined;
  repository.pendingAction = new Promise((_resolve, reject) => { rejectAction = reject; });
  const { harness, client } = await mount(<ProviderProfileDetailScreen profileId="profile-language" query={new ProviderProfileSettingsQuery(repository)} manageCommand={new ManageProviderProfileCommand(repository)} testCommand={new TestProviderProfileCommand(repository)} onEditCredential={() => {}} onEditPrompt={() => {}} />);
  try {
    const label = operation === 'test' ? 'Test connection for Gemini language' : operation === 'enable' ? 'enable Gemini language' : 'Archive Gemini language';
    await harness.press(harness.byLabel(label));
    if (operation === 'archive') await harness.run(() => pressAlertButton('Archive'));
    expect(harness.allText()).toContain(operation === 'test' ? 'Testing…' : operation === 'enable' ? 'Updating…' : 'Archiving…');
    for (const [kind, text] of [['test', 'Testing…'], ['enable', 'Updating…'], ['archive', 'Archiving…']]) {
      if (kind !== operation) expect(harness.allText()).not.toContain(text);
    }
    expect(harness.byLabel('Replace credential for Gemini language')?.props.disabled).toBe(true);
    expect(harness.byLabel('Edit prompt guidance for Gemini language')?.props.disabled).toBe(true);
    await harness.run(() => rejectAction?.(new Error('offline')));
    expect(harness.byLabel(label)?.props.disabled).toBe(false);
    expect(harness.byLabel('Replace credential for Gemini language')?.props.disabled).toBe(false);
    expect(harness.allText()).toContain('Gemini language');
  } finally { await harness.unmount(); client.clear(); }
});

it.each(['credential', 'prompt'] as const)('preserves the submitted %s draft after rejected save', async kind => {
  const repository = new FakeProviderRepository();
  let rejectSave: ((error: Error) => void) | undefined;
  const pending = new Promise<ProviderProfileSummary>((_resolve, reject) => { rejectSave = reject; });
  if (kind === 'credential') repository.pendingCredential = pending; else repository.pendingPrompt = pending;
  const Editor = kind === 'credential' ? ProviderCredentialScreen : ProviderPromptScreen;
  const { harness, client } = await mount(<Editor profileId="profile-language" query={new ProviderProfileSettingsQuery(repository)} manageCommand={new ManageProviderProfileCommand(repository)} onSaved={() => {}} />);
  const label = kind === 'credential' ? 'API key' : 'New prompt guidance';
  try {
    await harness.changeText(harness.byLabel(label), 'original draft');
    await harness.press(harness.byLabel(kind === 'credential' ? 'Save Credential' : 'Save Guidance'));
    expect(harness.byLabel(label)?.props.editable).toBe(false);
    await harness.changeText(harness.byLabel(label), 'later edit');
    await harness.run(() => rejectSave?.(new Error('offline')));
    expect(harness.byLabel(label)?.props.editable).toBe(true);
    expect(harness.byLabel(label)?.props.value).toBe('original draft');
    expect(JSON.stringify(client.getQueryCache().getAll().map(query => query.state.data))).not.toContain('original draft');
  } finally { await harness.unmount(); client.clear(); }
});

it('selects a voice service in place without implying a connection test is running', async () => {
  const repository = new FakeProviderRepository(slot('test_profile'));
  repository.extraProfiles = [profile({ id: 'other', displayName: 'Other language' }), profile({ id: 'archived', displayName: 'Archived language', lifecycleState: 'archived' })];
  let rejectSelection: ((error: Error) => void) | undefined;
  repository.pendingSelection = new Promise((_resolve, reject) => { rejectSelection = reject; });
  const { harness, client } = await mount(<VoiceCapabilityScreen capability="language_inference" query={new ProviderProfileSettingsQuery(repository)} manageCommand={new ManageProviderProfileCommand(repository)} testCommand={new TestProviderProfileCommand(repository)} onAddProfile={() => {}} onEditProfile={() => {}} onEditCredential={() => {}} />);
  try {
    await harness.press(harness.byLabel('Choose voice service'));
    expect(harness.byLabel('Archived language')).toBeUndefined();
    await harness.press(harness.byLabel('Gemini language'));
    expect(repository.selectionInputs).toEqual([]);
    await harness.press(harness.byLabel('Choose voice service'));
    await harness.press(harness.byLabel('Other language'));
    expect(repository.selectionInputs).toHaveLength(1);
    expect(repository.selectionInputs[0]?.languageInferenceProfileId).toBe('other');
    expect(harness.allText()).toContain('Selecting service…');
    expect(harness.allText()).not.toContain('Testing…');
    expect(harness.byLabel('Open provider profile Gemini language')?.props.disabled).toBe(true);
    await harness.run(() => rejectSelection?.(new Error('offline')));
    expect(harness.byLabel('Choose voice service')?.props.disabled).toBe(false);
    expect(harness.allText()).toContain('Gemini language');
  } finally { await harness.unmount(); client.clear(); }
});

it('keeps account and connection recovery reachable when inventory settings fail', async () => {
  const destinations: string[] = [];
  const { harness } = await mount(<SettingsScreen settingsQuery={settingsQuery([], async () => { throw Object.assign(new Error('inventory removed'), { status: 403 }); })} onNavigate={destination => destinations.push(destination)} />);
  try {
    await harness.press(harness.byLabel('Open Account settings'));
    await harness.press(harness.byLabel('Open Stuff Stash server settings'));
    expect(destinations).toEqual(['account', 'connection']);
  } finally { await harness.unmount(); }
});
it('permits confirmed sign out without loading inventory settings', async () => {
  let scopeReads = 0; let signedOut = 0;
  const { harness } = await mount(<AccountSettingsScreen settingsQuery={settingsQuery([], async () => { scopeReads++; throw new Error('offline'); })} onSignOut={async () => { signedOut++; }} />);
  try {
    await harness.press(harness.byLabel('Sign out john@example.com'));
    await harness.run(() => pressAlertButton('Sign Out'));
    expect(signedOut).toBe(1);
    expect(scopeReads).toBe(0);
  } finally { await harness.unmount(); }
});

it.each(['pending', 'failed'] as const)('keeps sign out available while identity is %s', async identity => {
  let signedOut = 0;
  const query = new SettingsQuery({ getCurrentPrincipal: async () => {
    if (identity === 'failed') throw new Error('offline');
    return new Promise(() => undefined);
  } }, { getDiagnostics: () => ({ apiBaseUrl: 'https://stash.home.test/api', appVersion: 'test', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => { throw new Error('Account must not load inventory settings'); } });
  const { harness } = await mount(<AccountSettingsScreen settingsQuery={query} onSignOut={async () => { signedOut++; }} />);
  try {
    await harness.press(harness.byLabel('Sign out Current account'));
    await harness.run(() => pressAlertButton('Sign Out'));
    expect(signedOut).toBe(1);
  } finally { await harness.unmount(); }
});


it.each((['credential', 'prompt', 'create', 'detail'] as const).flatMap(kind => (['success', 'failure'] as const).map(outcome => [kind, outcome] as const)))('does not publish or navigate from a departed %s task after returning: %s', async (kind, outcome) => {
  resetNavigation();
  const repository = new FakeProviderRepository();
  const pending = deferred<ProviderProfileSummary>();
  repository.pendingCredential = pending.promise;
  repository.pendingPrompt = pending.promise;
  repository.pendingAction = pending.promise;
  repository.pendingCreation = pending.promise;
  let navigations = 0;
  const command = new ManageProviderProfileCommand(repository);
  const query = new ProviderProfileSettingsQuery(repository);
  const editorProps = { manageCommand: command, query, profileId: 'profile-language', onSaved: () => { navigations++; } };
  const element = kind === 'credential' ? <ProviderCredentialScreen {...editorProps} />
    : kind === 'prompt' ? <ProviderPromptScreen {...editorProps} />
    : kind === 'create' ? <AddProviderProfileScreen manageCommand={command} onCreated={() => { navigations++; }} />
    : <ProviderProfileDetailScreen manageCommand={command} query={query} profileId="profile-language" testCommand={new TestProviderProfileCommand(repository)} onEditCredential={() => undefined} onEditPrompt={() => undefined} />;
  const { harness, client } = await mount(element);
  try {
    if (kind === 'credential' || kind === 'prompt') await harness.changeText(harness.byLabel(kind === 'credential' ? 'API key' : 'New prompt guidance'), 'private draft');
    const button = kind === 'credential' || kind === 'prompt' ? harness.byLabel(kind === 'credential' ? 'Save Credential' : 'Save Guidance')
      : kind === 'create' ? harness.all().find(node => String(node.props.accessibilityLabel ?? '').startsWith('Create draft '))
      : harness.byLabel('enable Gemini language');
    await harness.press(button);
    await harness.run(() => setScreenFocused(false));
    await harness.run(() => setScreenFocused(true));
    await harness.run(() => outcome === 'success' ? pending.resolve(profile({ credentialStatus: 'configured' })) : pending.reject(new Error('Late private error')));
    await settle(harness);
    expect(navigations).toBe(0);
    if (kind === 'credential' || kind === 'prompt') {
      expect(harness.byLabel(kind === 'credential' ? 'API key' : 'New prompt guidance')?.props.value).toBe(outcome === 'success' ? '' : 'private draft');
      if (outcome === 'success') {
        await harness.run(() => attemptNavigation({ type: 'manual-back' }));
        expect(dispatchedActions()).toEqual([{ type: 'manual-back' }]);
      }
    }
    for (const title of ['Credential saved', 'Prompt guidance saved', 'Draft profile created', 'Profile enabled', 'Credential not saved', 'Prompt not saved', 'Could not create profile', 'Profile action failed', 'Late private error']) expect(harness.byText(title)).toBeUndefined();
  } finally { await harness.unmount(); client.clear(); setScreenFocused(true); }
});

it('rejects an archive confirmation retained from a previous focus session', async () => {
  const repository = new FakeProviderRepository();
  const { harness, client } = await mount(<ProviderProfileDetailScreen manageCommand={new ManageProviderProfileCommand(repository)} query={new ProviderProfileSettingsQuery(repository)} profileId="profile-language" testCommand={new TestProviderProfileCommand(repository)} onEditCredential={() => undefined} onEditPrompt={() => undefined} />);
  try {
    await harness.press(harness.byLabel('Archive Gemini language'));
    const confirm = latestAlert()?.buttons?.find(button => button.text === 'Archive')?.onPress;
    expect(confirm).toBeDefined();
    await harness.run(() => setScreenFocused(false));
    await harness.run(() => setScreenFocused(true));
    await harness.run(() => confirm?.());
    expect(repository.lifecycleCalls).toEqual([]);
  } finally { await harness.unmount(); client.clear(); setScreenFocused(true); }
});

it('keeps a replacement profile credential when the previous form finishes saving', async () => {
  const repository = new FakeProviderRepository();
  repository.extraProfiles = [profile({ id: 'other', displayName: 'Other language' })];
  const pending = deferred<ProviderProfileSummary>();
  repository.pendingCredential = pending.promise;
  const command = new ManageProviderProfileCommand(repository);
  const query = new ProviderProfileSettingsQuery(repository);
  let replace!: (id: string) => void;
  let navigations = 0;
  function Editor() {
    const [id, setId] = React.useState('profile-language');
    replace = setId;
    return <ProviderCredentialScreen manageCommand={command} query={query} profileId={id} onSaved={() => { navigations++; }} />;
  }
  const { harness, client } = await mount(<Editor />);
  try {
    await harness.changeText(harness.byLabel('API key'), 'old secret');
    await harness.press(harness.byLabel('Save Credential'));
    await harness.run(() => replace('other'));
    await harness.changeText(harness.byLabel('API key'), 'replacement secret');
    await harness.run(() => pending.resolve(profile({ credentialStatus: 'configured' })));
    expect(harness.byLabel('API key')?.props.value).toBe('replacement secret');
    expect(harness.byLabel('API key')?.props.editable).toBe(true);
    expect(navigations).toBe(0);
    expect(harness.byText('Credential saved')).toBeUndefined();
  } finally { await harness.unmount(); client.clear(); }
});

it.each((['select', 'test', 'enable'] as const).flatMap(kind => (['success', 'failure'] as const).flatMap(outcome => [true, false].map(departed => [kind, outcome, departed] as const))))('scopes voice-stage %s completion (%s, departed=%s)', async (kind, outcome, departed) => {
  const repository = new FakeProviderRepository(slot(kind === 'enable' ? 'enable_profile' : 'test_profile'));
  repository.extraProfiles = [profile({ id: 'other', displayName: 'Other language' })];
  const action = deferred<ProviderProfileSummary | ProviderProfileTestResult>();
  const selection = deferred<VoiceProviderConfiguration>();
  repository.pendingAction = action.promise;
  repository.pendingSelection = selection.promise;
  const { harness, client } = await mount(<VoiceCapabilityScreen capability="language_inference" manageCommand={new ManageProviderProfileCommand(repository)} query={new ProviderProfileSettingsQuery(repository)} testCommand={new TestProviderProfileCommand(repository)} onAddProfile={() => undefined} onEditCredential={() => undefined} onEditProfile={() => undefined} />);
  try {
    if (kind === 'select') {
      await harness.press(harness.byLabel('Choose voice service'));
      await harness.press(harness.byLabel('Other language'));
      expect(repository.selectionInputs).toHaveLength(1);
    } else {
      await harness.press(textButton(harness, kind === 'test' ? 'Test Connection' : 'Enable Service'));
      expect(kind === 'test' ? repository.testCalls : repository.lifecycleCalls).toHaveLength(1);
    }
    if (departed) {
      await harness.run(() => setScreenFocused(false));
      await harness.run(() => setScreenFocused(true));
    }
    await harness.run(() => {
      if (kind === 'select') {
        if (outcome === 'success') selection.resolve(repository.configuration);
        else selection.reject(new Error('Late stage error'));
      } else if (outcome === 'success') action.resolve(kind === 'test' ? testResult() : profile({ lifecycleState: 'enabled' }));
      else action.reject(new Error('Late stage error'));
    });
    await settle(harness);
    if (departed) {
      for (const title of ['Voice service selected', 'Connection tested', 'Service enabled', 'Could not update voice', 'Late stage error']) expect(harness.byText(title)).toBeUndefined();
    } else {
      const title = outcome === 'failure' ? 'Could not update voice' : kind === 'select' ? 'Voice service selected' : kind === 'test' ? 'Connection tested' : 'Service enabled';
      expect(harness.byText(title)).toBeDefined();
    }
    expect(harness.byLabel('Choose voice service')?.props.disabled).toBe(false);
  } finally { await harness.unmount(); client.clear(); setScreenFocused(true); }
});


it.each(['credential', 'prompt'] as const)('protects the %s replacement draft and validates native Save before submission', async kind => {
  resetNavigation();
  const repository = new FakeProviderRepository();
  const pending = deferred<ProviderProfileSummary>();
  repository.pendingCredential = pending.promise;
  repository.pendingPrompt = pending.promise;
  let saves = 0;
  const props = { manageCommand: new ManageProviderProfileCommand(repository), query: new ProviderProfileSettingsQuery(repository), profileId: 'profile-language', onSaved: () => { saves++; attemptNavigation({ type: 'saved-back' }); } };
  const { harness, client } = await mount(kind === 'credential' ? <ProviderCredentialScreen {...props} /> : <ProviderPromptScreen {...props} />);
  const field = () => harness.byLabel(kind === 'credential' ? 'API key' : 'New prompt guidance');
  const save = () => harness.byLabel(kind === 'credential' ? 'Save Credential' : 'Save Guidance');
  try {
    expect(save()).toBeDefined();
    expect(save()?.props.disabled).toBe(true);
    await harness.changeText(field(), '   ');
    expect(save()?.props.disabled).toBe(true);
    await harness.changeText(field(), 'private replacement');
    expect(save()?.props.disabled).toBe(false);
    await harness.run(() => attemptNavigation({ type: 'back' }));
    expect(latestAlert()?.title).toBe('Discard changes?');
    await harness.run(() => pressAlertButton('Keep Editing'));
    expect(dispatchedActions()).toEqual([]);
    expect(field()?.props.value).toBe('private replacement');
    await harness.press(save());
    await harness.run(() => attemptNavigation({ type: 'during-save' }));
    expect(dispatchedActions()).toEqual([]);
    await harness.run(() => pending.reject(new Error('Try this replacement again')));
    expect(field()?.props.value).toBe('private replacement');
    const failure = harness.byText('Try this replacement again');
    expect(failure).toBeDefined();
    let ancestor = failure?.parent;
    while (ancestor && ancestor.type !== 'ScrollView') ancestor = ancestor.parent;
    expect(ancestor?.type).toBe('ScrollView');
    repository.pendingCredential = undefined; repository.pendingPrompt = undefined;
    await harness.press(save()); await settle(harness);
    expect(saves).toBe(1);
    expect(dispatchedActions()).toEqual([{ type: 'saved-back' }]);
  } finally { await harness.unmount(); client.clear(); resetNavigation(); }
});

it.each(['credential', 'prompt'] as const)('limits %s discard confirmation to the current visit and dispatches it once', async kind => {
  resetNavigation();
  const repository = new FakeProviderRepository();
  const props = { manageCommand: new ManageProviderProfileCommand(repository), query: new ProviderProfileSettingsQuery(repository), profileId: 'profile-language', onSaved: () => undefined };
  const { harness, client } = await mount(kind === 'credential' ? <ProviderCredentialScreen {...props} /> : <ProviderPromptScreen {...props} />);
  try {
    await harness.changeText(harness.byLabel(kind === 'credential' ? 'API key' : 'New prompt guidance'), 'replacement');
    await harness.run(() => attemptNavigation({ type: 'old-back' }));
    const oldDiscard = latestAlert()?.buttons?.find(button => button.text === 'Discard')?.onPress;
    expect(oldDiscard).toBeDefined();
    await harness.run(() => setScreenFocused(false));
    await harness.run(() => setScreenFocused(true));
    await harness.run(() => oldDiscard?.());
    expect(dispatchedActions()).toEqual([]);
    await harness.run(() => attemptNavigation({ type: 'current-back' }));
    const discard = latestAlert()?.buttons?.find(button => button.text === 'Discard')?.onPress;
    expect(discard).toBeDefined();
    await harness.run(() => { discard?.(); discard?.(); });
    expect(dispatchedActions()).toEqual([{ type: 'current-back' }]);
  } finally { await harness.unmount(); client.clear(); resetNavigation(); }
});

it('allows server credentials without a secret and blocks immediate Back while saving', async () => {
  resetNavigation();
  const repository = new FakeProviderRepository();
  repository.extraProfiles = [profile({ id: 'server', credentialPurpose: 'server_adc' })];
  const pending = deferred<ProviderProfileSummary>();
  repository.pendingCredential = pending.promise;
  const { harness, client } = await mount(<ProviderCredentialScreen manageCommand={new ManageProviderProfileCommand(repository)} query={new ProviderProfileSettingsQuery(repository)} profileId="server" onSaved={() => attemptNavigation({ type: 'saved-back' })} />);
  try {
    const save = harness.byLabel('Save Credential');
    expect(save).toBeDefined();
    expect(save?.props.disabled).toBe(false);
    await harness.run(() => { save?.props.onPress(); attemptNavigation({ type: 'immediate-back' }); });
    expect(dispatchedActions()).toEqual([]);
    expect(repository.credentialInputs).toEqual([{ providerProfileId: 'server', purpose: 'server_adc', credential: undefined }]);
    await harness.run(() => pending.resolve(profile({ id: 'server', credentialPurpose: 'server_adc', credentialStatus: 'configured' })));
    expect(dispatchedActions()).toEqual([{ type: 'saved-back' }]);
  } finally { await harness.unmount(); client.clear(); resetNavigation(); }
});

it.each(['signOut', 'server'] as const)('rejects an earlier-visit %s confirmation and keeps current retry', async kind => {
  let calls = 0;
  const action = async () => { calls++; throw new Error('Retry this action'); };
  const { harness } = await mount(kind === 'signOut' ? <AccountSettingsScreen settingsQuery={settingsQuery()} onSignOut={action} /> : <ConnectionSettingsScreen settingsQuery={settingsQuery()} onChangeServer={action} />);
  const label = kind === 'signOut' ? 'Sign out john@example.com' : 'Change Stuff Stash server from stash.home.test';
  const button = kind === 'signOut' ? 'Sign Out' : 'Change Server';
  try {
    await harness.press(harness.byLabel(label));
    const stale = latestAlert()?.buttons.find(value => value.text === button)?.onPress;
    expect(stale).toBeTypeOf('function');
    await harness.run(() => setScreenFocused(false)); await harness.run(() => setScreenFocused(true));
    await harness.run(() => stale?.()); await settle(harness);
    expect(calls).toBe(0);
    await harness.press(harness.byLabel(label));
    const current = latestAlert()?.buttons.find(value => value.text === button)?.onPress;
    await harness.run(() => current?.()); await settle(harness);
    expect(calls).toBe(1);
    await harness.run(() => current?.()); await settle(harness);
    expect(calls).toBe(1);
    await harness.press(harness.byLabel(label)); await harness.run(() => pressAlertButton(button)); await settle(harness);
    expect(calls).toBe(2);
  } finally { await harness.unmount(); setScreenFocused(true); }
});

it.each(['signOut', 'server'] as const)('suppresses a departed %s failure and unlocks retry', async kind => {
  const pending = deferred<void>(); let calls = 0;
  const action = () => { calls++; return pending.promise; };
  const { harness } = await mount(kind === 'signOut' ? <AccountSettingsScreen settingsQuery={settingsQuery()} onSignOut={action} /> : <ConnectionSettingsScreen settingsQuery={settingsQuery()} onChangeServer={action} />);
  const label = kind === 'signOut' ? 'Sign out john@example.com' : 'Change Stuff Stash server from stash.home.test';
  try {
    await harness.press(harness.byLabel(label)); await harness.run(() => pressAlertButton(kind === 'signOut' ? 'Sign Out' : 'Change Server'));
    await harness.run(() => setScreenFocused(false)); await harness.run(() => setScreenFocused(true));
    await harness.run(() => pending.reject(new Error('Departed action failure'))); await settle(harness);
    expect(harness.allText()).not.toContain('Departed action failure');
    expect(harness.byLabel(label)?.props.disabled).toBe(false);
    expect(calls).toBe(1);
  } finally { await harness.unmount(); setScreenFocused(true); }
});

it.each(['signOut', 'server'] as const)('rejects %s confirmation after settings-query replacement', async kind => {
  let calls = 0;
  const action = async () => { calls++; };
  const screen = (query: SettingsQuery) => kind === 'signOut' ? <AccountSettingsScreen settingsQuery={query} onSignOut={action} /> : <ConnectionSettingsScreen settingsQuery={query} onChangeServer={action} />;
  const { harness, client } = await mount(screen(settingsQuery()));
  const label = kind === 'signOut' ? 'Sign out john@example.com' : 'Change Stuff Stash server from stash.home.test';
  const button = kind === 'signOut' ? 'Sign Out' : 'Change Server';
  try {
    await harness.press(harness.byLabel(label));
    const stale = latestAlert()?.buttons.find(value => value.text === button)?.onPress;
    expect(stale).toBeTypeOf('function');
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant-home', inventoryId: 'inventory-home' })}><AppFeedbackProvider>{screen(settingsQuery())}</AppFeedbackProvider></MobileServerStateProvider>);
    await harness.run(() => stale?.());
    expect(calls).toBe(0);
    await harness.press(harness.byLabel(label)); await harness.run(() => pressAlertButton(button));
    expect(calls).toBe(1);
  } finally { await harness.unmount(); }
});
