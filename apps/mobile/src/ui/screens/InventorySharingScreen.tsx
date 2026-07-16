import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react';
import {
  ActivityIndicator,
  Alert,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { Check, Copy, Send, X } from 'lucide-react-native';
import type {
  CancelInventoryInvitationCommand,
  CreatedInventoryInvitation,
  CreateInventoryInvitationCommand,
  InvitationLinkActions,
  InventoryInvitationRelationship,
  InventoryInvitationSummary,
  InventorySharingScope,
  ListInventoryInvitationsQuery
} from '../../application/sharing/InventorySharing';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { SettingsSection, useSettingsListStyles } from './SettingsList';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';

export function InventorySharingScreen({
  cancelCommand,
  createCommand,
  linkActions,
  listQuery,
  scope
}: {
  readonly cancelCommand: CancelInventoryInvitationCommand;
  readonly createCommand: CreateInventoryInvitationCommand;
  readonly linkActions: InvitationLinkActions;
  readonly listQuery: ListInventoryInvitationsQuery;
  readonly scope: InventorySharingScope;
}) {
  const feedback = useAppFeedback();
  const palette = useAppearancePalette();
  const { layout, styles: settingsStyles } = useSettingsListStyles();
  const styles = createStyles(palette);
  const scopeKey = `${scope.tenantId}:${scope.inventoryId}:${scope.permissions.join(',')}`;
  const [email, setEmail] = useState('');
  const [relationship, setRelationship] = useState<InventoryInvitationRelationship>('viewer');
  const [invitations, setInvitations] = useState<readonly InventoryInvitationSummary[]>([]);
  const [invitationsScopeKey, setInvitationsScopeKey] = useState(scopeKey);
  const [created, setCreated] = useState<CreatedInventoryInvitation>();
  const [createdScopeKey, setCreatedScopeKey] = useState<string>();
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading');
  const [working, setWorking] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [cancellingId, setCancellingId] = useState<string>();
  const workingRef = useRef(false);
  const requestGenerationRef = useRef(0);
  const currentScopeKeyRef = useRef(scopeKey);
  currentScopeKeyRef.current = scopeKey;
  const visibleInvitations = invitationsScopeKey === scopeKey ? invitations : [];
  const visibleCreated = createdScopeKey === scopeKey ? created : undefined;

  const load = useCallback(async (refresh = false) => {
    const generation = ++requestGenerationRef.current;
    if (refresh) setRefreshing(true); else setStatus('loading');
    try {
      const loaded = await listQuery.execute(scope);
      if (generation !== requestGenerationRef.current) return;
      setInvitations(loaded);
      setInvitationsScopeKey(scopeKey);
      setStatus('ready');
    } catch (error) {
      if (generation !== requestGenerationRef.current) return;
      if (refresh) {
        feedback.showNotice({ tone: 'error', title: 'Could not refresh invitations', message: readableError(error) });
      } else {
        setStatus('error');
      }
    } finally {
      if (generation === requestGenerationRef.current) setRefreshing(false);
    }
  }, [feedback, listQuery, scopeKey]);

  useEffect(() => {
    setCreated(undefined);
    setCreatedScopeKey(undefined);
    setInvitations([]);
    setInvitationsScopeKey(scopeKey);
    void load();
    return () => { requestGenerationRef.current += 1; };
  }, [load, scopeKey]);

  async function create(): Promise<void> {
    if (workingRef.current) return;
    workingRef.current = true;
    setWorking(true);
    const requestedScopeKey = scopeKey;
    try {
      const invitation = await createCommand.execute(scope, { email, relationship });
      if (currentScopeKeyRef.current !== requestedScopeKey) return;
      setCreated(invitation);
      setCreatedScopeKey(requestedScopeKey);
      setInvitations((current) => [withoutLink(invitation), ...current]);
      setInvitationsScopeKey(requestedScopeKey);
      setEmail('');
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: 'Could not create invitation', message: readableError(error) });
    } finally {
      workingRef.current = false;
      setWorking(false);
    }
  }

  async function performLinkAction(action: 'copy' | 'share'): Promise<void> {
    if (!visibleCreated) return;
    try {
      if (action === 'copy') {
        await linkActions.copy(visibleCreated.inviteUrl);
        feedback.showNotice({ tone: 'success', title: 'Invitation link copied' });
      } else {
        await linkActions.share({ link: visibleCreated.inviteUrl, inventoryName: scope.inventoryName });
      }
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: `Could not ${action} invitation`, message: readableError(error) });
    }
  }

  async function cancel(invitation: InventoryInvitationSummary): Promise<void> {
    setCancellingId(invitation.id);
    const requestedScopeKey = scopeKey;
    try {
      await cancelCommand.execute(scope, invitation.id);
      if (currentScopeKeyRef.current !== requestedScopeKey) return;
      setInvitations((current) => current.map((item) => item.id === invitation.id
        ? { ...item, status: 'cancelled' }
        : item));
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: 'Could not cancel invitation', message: readableError(error) });
    } finally {
      setCancellingId(undefined);
    }
  }

  if (status === 'loading') {
    return <View style={[settingsStyles.shell, settingsStyles.errorContainer]}><ActivityIndicator color={palette.action} /></View>;
  }
  if (status === 'error') {
    return (
      <View style={[settingsStyles.shell, settingsStyles.errorContainer]}>
        <Text accessibilityRole="header" style={settingsStyles.errorTitle}>Could not load invitations</Text>
        <Text style={settingsStyles.errorMessage}>Your invitation settings are still safe. Try again.</Text>
        <Pressable accessibilityRole="button" onPress={() => void load()} style={settingsStyles.retryButton}>
          <Text style={settingsStyles.retryText}>Retry</Text>
        </Pressable>
      </View>
    );
  }

  return (
    <ScrollView
      contentContainerStyle={settingsStyles.content}
      keyboardDismissMode={appKeyboardDismissMode()}
      keyboardShouldPersistTaps="handled"
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => void load(true)} tintColor={palette.action} />}
      style={settingsStyles.shell}
    >
      <View style={settingsStyles.detailHeader}>
        <Text accessibilityRole="header" style={settingsStyles.detailTitle}>Share {scope.inventoryName}</Text>
        <Text style={settingsStyles.detailSubtitle}>Invite someone by email as a viewer or editor.</Text>
      </View>

      <SettingsSection title="New Invitation">
        <View style={styles.form}>
          <Text style={styles.label}>Email</Text>
          <AppTextInput
            autoCapitalize="none"
            autoComplete="email"
            accessibilityLabel="Invitee email"
            keyboardType="email-address"
            onChangeText={setEmail}
            placeholder="friend@example.com"
            placeholderTextColor={palette.textMuted}
            style={styles.input}
            value={email}
          />
          <Text style={styles.label}>Access</Text>
          <View
            accessibilityRole="radiogroup"
            style={[styles.roleGroup, layout.stacksChoiceRows && styles.roleGroupStacked]}
          >
            {(['viewer', 'editor'] as const).map((role) => {
              const selected = relationship === role;
              return (
                <Pressable
                  accessibilityRole="radio"
                  accessibilityState={{ checked: selected }}
                  key={role}
                  onPress={() => setRelationship(role)}
                  style={[styles.roleButton, selected && styles.roleButtonSelected]}
                >
                  {selected ? <Check color={palette.action} size={17} /> : null}
                  <Text style={[styles.roleText, selected && styles.roleTextSelected]}>{titleCase(role)}</Text>
                </Pressable>
              );
            })}
          </View>
          <Pressable
            accessibilityRole="button"
            accessibilityState={{ busy: working, disabled: working || email.trim().length === 0 }}
            disabled={working || email.trim().length === 0}
            onPress={() => void create()}
            style={[styles.primaryButton, (working || email.trim().length === 0) && styles.disabled]}
          >
            {working ? <ActivityIndicator color={palette.onAction} /> : <Send color={palette.onAction} size={18} />}
            <Text style={styles.primaryButtonText}>{working ? 'Creating…' : 'Create Invitation'}</Text>
          </Pressable>
        </View>
      </SettingsSection>

      {visibleCreated ? (
        <SettingsSection
          footer="This complete link cannot be recovered after you leave this screen. Copy or share it now."
          title="Invitation Link"
        >
          <View style={styles.oneTimeLink}>
            <Text style={styles.successTitle}>Invitation ready</Text>
            <Text style={styles.linkContext}>
              {visibleCreated.email} · {titleCase(visibleCreated.relationship)} · Expires {formatDate(visibleCreated.expiresAt)}
            </Text>
            <Text accessibilityLabel="Complete invitation link" selectable style={styles.linkText}>
              {visibleCreated.inviteUrl}
            </Text>
            <View style={styles.linkActions}>
              <LinkButton icon={<Copy color={palette.action} size={18} />} label="Copy link" onPress={() => void performLinkAction('copy')} />
              <LinkButton icon={<Send color={palette.action} size={18} />} label="Share invitation" onPress={() => void performLinkAction('share')} />
            </View>
          </View>
        </SettingsSection>
      ) : null}

      <SettingsSection footer="Invitation links are shown only when created. Existing invitations never reveal their links again." title="Invitations">
        {visibleInvitations.length === 0 ? (
          <View style={styles.empty}><Text style={styles.emptyText}>No invitations yet.</Text></View>
        ) : visibleInvitations.map((invitation, index) => (
          <View key={invitation.id}>
            {index > 0 ? <View style={styles.separator} /> : null}
            <View style={styles.invitationRow}>
              <View style={styles.invitationText}>
                <Text style={styles.invitationEmail}>{invitation.email}</Text>
                <Text style={styles.invitationMetadata}>
                  {titleCase(invitation.relationship)} · {statusLabel(invitation)} · Expires {formatDate(invitation.expiresAt)}
                </Text>
              </View>
              {invitation.status === 'pending' && !invitation.isExpired ? (
                <Pressable
                  accessibilityLabel={`Cancel invitation for ${invitation.email}`}
                  accessibilityRole="button"
                  disabled={cancellingId === invitation.id}
                  onPress={() => confirmCancel(invitation, cancel)}
                  style={styles.cancelButton}
                >
                  {cancellingId === invitation.id
                    ? <ActivityIndicator color={palette.danger} />
                    : <X color={palette.danger} size={19} />}
                </Pressable>
              ) : null}
            </View>
          </View>
        ))}
      </SettingsSection>
    </ScrollView>
  );
}

function LinkButton({ icon, label, onPress }: { readonly icon: ReactNode; readonly label: string; readonly onPress: () => void }) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <Pressable accessibilityRole="button" onPress={onPress} style={({ pressed }) => [styles.linkButton, pressed && styles.pressed]}>
      {icon}<Text style={styles.linkButtonText}>{label}</Text>
    </Pressable>
  );
}

function withoutLink(invitation: CreatedInventoryInvitation): InventoryInvitationSummary {
  const { inviteUrl: _inviteUrl, ...safe } = invitation;
  return safe;
}

function statusLabel(invitation: InventoryInvitationSummary): string {
  if (invitation.isExpired) return 'Expired';
  return titleCase(invitation.status);
}

function titleCase(value: string): string {
  return `${value.charAt(0).toUpperCase()}${value.slice(1)}`;
}

function formatDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString(undefined, { dateStyle: 'medium' });
}

function confirmCancel(invitation: InventoryInvitationSummary, cancel: (value: InventoryInvitationSummary) => Promise<void>): void {
  Alert.alert('Cancel invitation?', `${invitation.email} will no longer be able to use this invitation link.`, [
    { text: 'Keep Invitation', style: 'cancel' },
    { text: 'Cancel Invitation', style: 'destructive', onPress: () => void cancel(invitation) }
  ]);
}

function readableError(error: unknown): string {
  return error instanceof Error ? error.message : 'The action failed safely. Try again.';
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    form: { gap: spacing.sm, padding: spacing.md },
    label: { color: colors.textMuted, fontSize: 13, fontWeight: '600' },
    input: { backgroundColor: colors.background, borderColor: colors.border, borderRadius: radius.sm, borderWidth: StyleSheet.hairlineWidth, color: colors.text, fontSize: 17, minHeight: 48, paddingHorizontal: spacing.md },
    roleGroup: { flexDirection: 'row', gap: spacing.sm },
    roleGroupStacked: { flexDirection: 'column' },
    roleButton: { alignItems: 'center', borderColor: colors.border, borderRadius: radius.sm, borderWidth: StyleSheet.hairlineWidth, flex: 1, flexDirection: 'row', gap: spacing.xs, justifyContent: 'center', minHeight: 44 },
    roleButtonSelected: { backgroundColor: colors.selected, borderColor: colors.action },
    roleText: { color: colors.textMuted, fontSize: 16, fontWeight: '600' },
    roleTextSelected: { color: colors.action },
    primaryButton: { alignItems: 'center', backgroundColor: colors.action, borderRadius: radius.md, flexDirection: 'row', gap: spacing.sm, justifyContent: 'center', marginTop: spacing.xs, minHeight: 48, paddingHorizontal: spacing.md },
    primaryButtonText: { color: colors.onAction, fontSize: 17, fontWeight: '700' },
    disabled: { opacity: 0.5 },
    oneTimeLink: { gap: spacing.sm, padding: spacing.md },
    successTitle: { color: colors.text, fontSize: 17, fontWeight: '700' },
    linkContext: { color: colors.textMuted, fontSize: 14, lineHeight: 20 },
    linkText: { backgroundColor: colors.background, borderRadius: radius.sm, color: colors.text, fontSize: 13, lineHeight: 19, padding: spacing.sm },
    linkActions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
    linkButton: { alignItems: 'center', borderColor: colors.border, borderRadius: radius.sm, borderWidth: StyleSheet.hairlineWidth, flexDirection: 'row', gap: spacing.xs, justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.md },
    linkButtonText: { color: colors.action, fontSize: 16, fontWeight: '600' },
    pressed: { backgroundColor: colors.selected },
    empty: { minHeight: 68, justifyContent: 'center', paddingHorizontal: spacing.md },
    emptyText: { color: colors.textMuted, fontSize: 16 },
    separator: { backgroundColor: colors.border, height: StyleSheet.hairlineWidth, marginLeft: spacing.md },
    invitationRow: { alignItems: 'center', flexDirection: 'row', minHeight: 68, paddingHorizontal: spacing.md, paddingVertical: spacing.sm },
    invitationText: { flex: 1, minWidth: 0 },
    invitationEmail: { color: colors.text, fontSize: 16, fontWeight: '600' },
    invitationMetadata: { color: colors.textMuted, fontSize: 13, lineHeight: 18, marginTop: 2 },
    cancelButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 44 }
  });
}
