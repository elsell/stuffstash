import { t } from '../../presentation/localization';
import { InvitationEmailInput } from './InvitationEmailInput';
import { InventoryInvitationLinkUnavailableError } from '../../application/sharing/InventorySharing';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { usePullRefresh } from '../serverState/usePullRefresh';
import { useInfiniteQuery } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScopeId } from '../navigation/MobileServerStateProvider';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import {
  ActivityIndicator,
  Keyboard,
  Platform,
  Alert,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
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
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { SettingsSection, useSettingsListStyles } from './SettingsList';
import { appKeyboardDismissMode } from '../components/AppTextInput';

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
  const palette = useAppearancePalette();
  const { styles: settingsStyles } = useSettingsListStyles();
  const styles = createStyles(palette);
  const compositionScopeId = useMobileServerStateScopeId();
  const scopeKey = `${compositionScopeId}:${scope.tenantId}:${scope.inventoryId}:${scope.permissions.join(',')}`;
  const [email, setEmail] = useState('');
  const emailScope = useRef(scopeKey);
  const [emailRevision, setEmailRevision] = useState(0);
  const [creationError, setCreationError] = useState<{ title: string; message: string }>();
  const [linkFeedback, setLinkFeedback] = useState<{ title: string; message?: string }>();
  const [cancellationErrors, setCancellationErrors] = useState<Record<string, string>>({});
  const linkOperation = useRef(0);
  const activeLinkOperation = useRef<number | undefined>(undefined);
  const [linkWorking, setLinkWorking] = useState<'copy' | 'share'>();
  const [relationship, setRelationship] = useState<InventoryInvitationRelationship>('viewer');
  const [created, setCreated] = useState<CreatedInventoryInvitation>();
  const [createdScopeKey, setCreatedScopeKey] = useState<string>();
  const [working, setWorking] = useState(false);
  const pendingCancellations = useRef(new Set<string>());
  const [cancellingKeys, setCancellingKeys] = useState<ReadonlySet<string>>(new Set());
  const workingRef = useRef(false);
  const currentScopeKeyRef = useRef(scopeKey);
  currentScopeKeyRef.current = scopeKey;
  const feedbackSession = useRef<object | undefined>(undefined);
  useFocusEffect(useCallback(() => {
    const session = {}; feedbackSession.current = session;
    setCreationError(undefined);
    setLinkFeedback(undefined);
    setCancellationErrors({});
    linkOperation.current += 1;
    activeLinkOperation.current = undefined;
    setLinkWorking(undefined);
    return () => { if (feedbackSession.current === session) feedbackSession.current = undefined; };
  }, [scopeKey]));
  function captureFeedbackOwner(): () => boolean {
    const session = feedbackSession.current;
    const requestedScope = scopeKey;
    return () => session !== undefined && feedbackSession.current === session && currentScopeKeyRef.current === requestedScope;
  }
  const canShare = scope.permissions.includes('share');
  const list = useInfiniteQuery({
    queryKey: mobileQueryKeys.invitations(compositionScopeId, scope.tenantId, scope.inventoryId),
    queryFn: ({ signal, pageParam }) => listQuery.execute(scope, { signal, cursor: pageParam }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last, _pages, _param, params) => last.nextCursor && !params.includes(last.nextCursor) ? last.nextCursor : undefined,
    enabled: canShare,
    subscribed: canShare
  });
  const pullRefresh = usePullRefresh(async () => { await list.refetch({ cancelRefetch: false }); });
  const denied = !canShare || isAccessFailure(list.error);
  const visibleInvitations = denied ? [] : list.data?.pages.flatMap(page => page.items) ?? [];
  const visibleCreated = !denied && createdScopeKey === scopeKey ? created : undefined;
  useEffect(() => {
    setCreated(undefined);
    setCreatedScopeKey(undefined);
    setEmail('');
    emailScope.current = scopeKey;
    setRelationship('viewer');
  }, [scopeKey]);

  async function create(): Promise<void> {
    if (workingRef.current) return;
    Keyboard.dismiss();
    workingRef.current = true;
    setWorking(true);
    setCreationError(undefined);
    setLinkFeedback(undefined);
    linkOperation.current += 1;
    activeLinkOperation.current = undefined;
    setLinkWorking(undefined);
    setCreated(undefined);
    setCreatedScopeKey(undefined);
    const ownsFeedback = captureFeedbackOwner();
    const requestedScopeKey = scopeKey;
    try {
      const invitation = await createCommand.execute(scope, { email, relationship });
      if (currentScopeKeyRef.current !== requestedScopeKey) return;
      setCreated(invitation);
      setCreatedScopeKey(requestedScopeKey);
      setEmail('');
    } catch (error) {
      if (ownsFeedback()) setCreationError(error instanceof InventoryInvitationLinkUnavailableError
        ? { title: t('mobile.InventorySharingScreen.invitationCreatedLinkUnavailable'), message: t('mobile.InventorySharingScreen.ifTheInvitationIsStillPendingBelowCancelIt') }
        : { title: t('mobile.InventorySharingScreen.couldNotCreateInvitation'), message: readableError(error) });
    } finally {
      // A disabled SwiftUI field can restore focus after a menu closes.
      // End this input session on failure too, preserving the React email draft.
      if (currentScopeKeyRef.current === requestedScopeKey) setEmailRevision(value => value + 1);
      workingRef.current = false;
      setWorking(false);
    }
  }

  async function performLinkAction(action: 'copy' | 'share'): Promise<void> {
    if (!visibleCreated || activeLinkOperation.current !== undefined) return;
    const ownsFeedback = captureFeedbackOwner();
    const operation = ++linkOperation.current;
    activeLinkOperation.current = operation;
    setLinkWorking(action);
    setLinkFeedback(undefined);
    const ownsLinkFeedback = () => ownsFeedback() && linkOperation.current === operation;
    try {
      if (action === 'copy') {
        await linkActions.copy(visibleCreated.inviteUrl);
        if (ownsLinkFeedback()) setLinkFeedback({ title: t('mobile.InventorySharingScreen.invitationLinkCopied') });
      } else {
        await linkActions.share({ link: visibleCreated.inviteUrl, inventoryName: scope.inventoryName });
      }
    } catch (error) {
      if (ownsLinkFeedback()) setLinkFeedback({ title: t('mobile.InventorySharingScreen.couldNotInvitation', { action: String(action) }), message: readableError(error) });
    } finally {
      if (activeLinkOperation.current === operation) {
        activeLinkOperation.current = undefined;
        setLinkWorking(undefined);
      }
    }
  }

  async function cancel(invitation: InventoryInvitationSummary): Promise<void> {
    const operationKey = cancellationKey(invitation.id);
    if (pendingCancellations.current.has(operationKey)) return;
    pendingCancellations.current.add(operationKey);
    setCancellingKeys(new Set(pendingCancellations.current));
    setCancellationErrors(current => {
      const next = { ...current }; delete next[invitation.id]; return next;
    });
    const ownsFeedback = captureFeedbackOwner();
    const requestedScopeKey = scopeKey;
    try {
      await cancelCommand.execute(scope, invitation.id);
      if (currentScopeKeyRef.current !== requestedScopeKey) return;

    } catch (error) {
      if (ownsFeedback()) setCancellationErrors(current => ({ ...current, [invitation.id]: readableError(error) }));
    } finally {
      pendingCancellations.current.delete(operationKey);
      setCancellingKeys(new Set(pendingCancellations.current));
    }
  }

  function cancellationKey(id: string): string { return JSON.stringify([scopeKey, id]); }
  function requestCancellation(invitation: InventoryInvitationSummary): void {
    const ownsConfirmation = captureFeedbackOwner();
    if (!ownsConfirmation() || pendingCancellations.current.has(cancellationKey(invitation.id))) return;
    Keyboard.dismiss();
    let confirmed = false;
    confirmCancel(invitation, async value => {
      if (confirmed || !ownsConfirmation()) return;
      confirmed = true;
      await cancel(value);
    });
  }

  if (list.isPending && !denied) {
    return <View style={[settingsStyles.shell, settingsStyles.errorContainer]}><ActivityIndicator color={palette.action} /></View>;
  }
  if (denied || (list.isError && !list.data)) {
    return (
      <ScrollView style={settingsStyles.shell} contentContainerStyle={settingsStyles.errorContainer}>
        <Text accessibilityRole="header" style={settingsStyles.errorTitle}>{denied ? t('mobile.InventorySharingScreen.sharingUnavailable') : t('mobile.InventorySharingScreen.couldNotLoadInvitations')}</Text>
        <Text style={settingsStyles.errorMessage}>{!canShare
          ? t('mobile.InventorySharingScreen.youDonTHavePermissionToManageInvitationsFor', { inventoryName: String(scope.inventoryName) })
          : denied ? t('mobile.InventorySharingScreen.yourAccessToThisInventoryCouldNotBeConfirmed')
          : t('mobile.InventorySharingScreen.yourInvitationsCouldNotBeLoadedTryAgain')}</Text>
        {canShare ? <NativeCommandButton label={denied ? t('mobile.InventorySharingScreen.checkAgain') : t('mobile.InventorySharingScreen.retry')} onPress={() => { void list.refetch({ cancelRefetch: false }); }} /> : null}
      </ScrollView>
    );
  }

  return (
    <ScrollView
      contentContainerStyle={settingsStyles.content}
      keyboardDismissMode={appKeyboardDismissMode()}
      keyboardShouldPersistTaps="handled"
      refreshControl={<RefreshControl refreshing={pullRefresh.refreshing} onRefresh={() => void pullRefresh.refresh()} tintColor={palette.action} />}
      style={settingsStyles.shell}
    >
      <View style={settingsStyles.detailHeader}>
        <Text accessibilityRole="header" style={settingsStyles.detailTitle}>{t('mobile.InventorySharingScreen.share')}{scope.inventoryName}</Text>
        <Text style={settingsStyles.detailSubtitle}>{t('mobile.InventorySharingScreen.inviteSomeoneByEmailAsAViewerOrEditor')}</Text>
      </View>

      <SettingsRefreshNotice visible={list.isRefetchError || list.isFetchNextPageError} onRetry={async () => { await (list.isFetchNextPageError ? list.fetchNextPage({ cancelRefetch: false }) : list.refetch({ cancelRefetch: false })); }} />
      <SettingsSection title={t('mobile.InventorySharingScreen.newInvitation')}>
        <View testID="invitation-creation-form" style={styles.form}>
          {creationError ? <View accessibilityRole="alert" accessibilityLiveRegion="polite">
            <Text style={styles.successTitle}>{creationError.title}</Text>
            <Text style={[settingsStyles.errorMessage, styles.inlineFeedback]}>{creationError.message}</Text>
          </View> : null}
          <Text style={styles.label}>{t('mobile.InventorySharingScreen.email')}</Text>
          <InvitationEmailInput
            key={Platform.OS === 'ios' ? `${scopeKey}:${emailRevision}` : scopeKey}
            editable={!working}
            onChangeText={value => { if (!workingRef.current) setEmail(value); }}
            placeholderTextColor={palette.textMuted}
            style={styles.input}
            email={emailScope.current === scopeKey ? email : ''}
          />
          <SettingsPickerRow label={t('mobile.InventorySharingScreen.access')} accessibilityLabel={t('mobile.InventorySharingScreen.chooseInvitationAccess')} value={relationship}
            options={[{ value: 'viewer', label: t('mobile.InventorySharingScreen.viewer') }, { value: 'editor', label: t('mobile.InventorySharingScreen.editor') }] as const}
            disabled={working} onChange={value => { if (!workingRef.current) setRelationship(value); }} />
          <NativeCommandButton prominence="primary" label={working ? t('mobile.InventorySharingScreen.creating') : t('mobile.InventorySharingScreen.createInvitation')}
            disabled={working || email.trim().length === 0} onPress={() => void create()} />
        </View>
      </SettingsSection>

      {visibleCreated ? (
        <SettingsSection
          footer={t('mobile.InventorySharingScreen.copyOrShareThisLinkBeforeLeavingThisScreen')}
          title={t('mobile.InventorySharingScreen.invitationLink')}
        >
          <View style={styles.oneTimeLink}>
            <Text style={styles.successTitle}>{t('mobile.InventorySharingScreen.invitationReady')}</Text>
            <Text style={styles.linkContext}>
              {visibleCreated.email} · {titleCase(visibleCreated.relationship)}{t('mobile.InventorySharingScreen.expires')}{formatDate(visibleCreated.expiresAt)}
            </Text>
            <Text accessibilityLabel={t('mobile.InventorySharingScreen.completeInvitationLink')} selectable style={styles.linkText}>
              {visibleCreated.inviteUrl}
            </Text>
            <View style={styles.linkActions}>
              <View style={styles.linkCommand}>
                <NativeCommandButton prominence="primary" label={linkWorking === 'share' ? t('mobile.InventorySharingScreen.sharing') : t('mobile.InventorySharingScreen.shareInvitation')} disabled={linkWorking !== undefined} onPress={() => void performLinkAction('share')} />
              </View>
              <View style={styles.linkCommand}>
                <NativeCommandButton label={linkWorking === 'copy' ? t('mobile.InventorySharingScreen.copying') : t('mobile.InventorySharingScreen.copyLink')} disabled={linkWorking !== undefined} onPress={() => void performLinkAction('copy')} />
              </View>
            </View>
            {linkFeedback ? <View accessibilityLiveRegion="polite" accessibilityRole={linkFeedback.message ? 'alert' : undefined}>
              <Text style={styles.successTitle}>{linkFeedback.title}</Text>
              {linkFeedback.message ? <Text style={[settingsStyles.errorMessage, styles.inlineFeedback]}>{linkFeedback.message}</Text> : null}
            </View> : null}
          </View>
        </SettingsSection>
      ) : null}

      <SettingsSection footer={t('mobile.InventorySharingScreen.invitationLinksAreShownOnlyWhenCreatedExistingInvitations')} title={t('mobile.InventorySharingScreen.invitations')}>
        {visibleInvitations.length === 0 ? (
          <View style={styles.empty}><Text style={styles.emptyText}>{t('mobile.InventorySharingScreen.noInvitationsYet')}</Text></View>
        ) : visibleInvitations.map((invitation, index) => (
          <View key={invitation.id}>
            {index > 0 ? <View style={styles.separator} /> : null}
            <View style={styles.invitationRow}>
              <View style={styles.invitationText}>
                <Text style={styles.invitationEmail}>{invitation.email}</Text>
                <Text style={styles.invitationMetadata}>
                  {titleCase(invitation.relationship)} · {statusLabel(invitation)}{t('mobile.InventorySharingScreen.expires')}{formatDate(invitation.expiresAt)}
                </Text>
                {cancellingKeys.has(cancellationKey(invitation.id)) ? <Text accessibilityLiveRegion="polite" style={styles.invitationMetadata}>{t('mobile.InventorySharingScreen.cancelling')}</Text> : null}
                {cancellationErrors[invitation.id] ? <View accessibilityRole="alert" accessibilityLiveRegion="polite">
                  <Text style={styles.successTitle}>{t('mobile.InventorySharingScreen.couldNotCancelInvitation')}</Text>
                  <Text style={[settingsStyles.errorMessage, styles.inlineFeedback]}>{cancellationErrors[invitation.id]}</Text>
                </View> : null}
              </View>
              {invitation.status === 'pending' && !invitation.isExpired ? (
                <NativeActionMenu
                  accessibilityLabel={t('mobile.InventorySharingScreen.invitationActions', { email: invitation.email })}
                  disabled={cancellingKeys.has(cancellationKey(invitation.id))}
                  groups={[{ id: 'invitation', items: [{ id: 'cancel',
                    label: t('mobile.InventorySharingScreen.cancelInvitation'),
                    systemImage: 'xmark.circle', isDestructive: true,
                    onPress: () => requestCancellation(invitation)
                  }] }]} />
              ) : null}
            </View>
          </View>
        ))}
      </SettingsSection>
      {list.hasNextPage ? <View>
        {list.isFetchingNextPage ? <Text accessibilityLiveRegion="polite" style={settingsStyles.errorMessage}>{t('mobile.InventorySharingScreen.loadingOlderInvitations')}</Text> : null}
        <NativeCommandButton label={t('mobile.InventorySharingScreen.loadOlderInvitations')} disabled={list.isFetching} onPress={() => void list.fetchNextPage({ cancelRefetch: false })} />
      </View> : null}
    </ScrollView>
  );
}

function statusLabel(invitation: InventoryInvitationSummary): string {
  if (invitation.isExpired) return t('mobile.InventorySharingScreen.expired');
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
  Alert.alert(t('mobile.InventorySharingScreen.cancelInvitation2'), t('mobile.InventorySharingScreen.willNoLongerBeAbleToUseThisInvitation', { email: String(invitation.email) }), [
    { text: t('mobile.InventorySharingScreen.keepInvitation'), style: 'cancel' },
    { text: t('mobile.InventorySharingScreen.cancelInvitation3'), style: 'destructive', onPress: () => void cancel(invitation) }
  ]);
}

function readableError(error: unknown): string {
  return error instanceof Error ? error.message : t('mobile.InventorySharingScreen.theActionFailedSafelyTryAgain');
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    inlineFeedback: { textAlign: 'auto' },
    form: { gap: spacing.sm, padding: spacing.md },
    label: { color: colors.textMuted, fontSize: 13, fontWeight: '600' },
    input: { backgroundColor: colors.background, borderColor: colors.border, borderRadius: radius.sm, borderWidth: StyleSheet.hairlineWidth, color: colors.text, fontSize: 17, minHeight: 48, paddingHorizontal: spacing.md },
    oneTimeLink: { gap: spacing.sm, padding: spacing.md },
    successTitle: { color: colors.text, fontSize: 17, fontWeight: '700' },
    linkContext: { color: colors.textMuted, fontSize: 14, lineHeight: 20 },
    linkText: { backgroundColor: colors.background, borderRadius: radius.sm, color: colors.text, fontSize: 13, lineHeight: 19, padding: spacing.sm },
    linkActions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm, maxWidth: 560 },
    linkCommand: { width: 160, maxWidth: '100%' },
    empty: { minHeight: 68, justifyContent: 'center', paddingHorizontal: spacing.md },
    emptyText: { color: colors.textMuted, fontSize: 16 },
    separator: { backgroundColor: colors.border, height: StyleSheet.hairlineWidth, marginLeft: spacing.md },
    invitationRow: { alignItems: 'center', flexDirection: 'row', minHeight: 68, paddingHorizontal: spacing.md, paddingVertical: spacing.sm },
    invitationText: { flex: 1, minWidth: 0 },
    invitationEmail: { color: colors.text, fontSize: 16, fontWeight: '600' },
    invitationMetadata: { color: colors.textMuted, fontSize: 13, lineHeight: 18, marginTop: 2 }
  });
}
