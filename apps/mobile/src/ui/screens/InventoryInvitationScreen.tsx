import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, ScrollView, StyleSheet, Text, View, useWindowDimensions } from 'react-native';
import { CheckCircle2, MailCheck } from 'lucide-react-native';
import type { AcceptInventoryInvitationCommand } from '../../application/invitations/AcceptInventoryInvitationCommand';
import {
  InventoryInvitationAuthenticationRequiredError,
  InventoryInvitationEmailMismatchError,
  InventoryInvitationInvalidError,
  InventoryInvitationInvalidResponseError,
  type InventoryInvitationPreview,
  type InventoryInvitationReference
} from '../../application/invitations/InventoryInvitationRepository';
import type { PreviewInventoryInvitationQuery } from '../../application/invitations/PreviewInventoryInvitationQuery';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { BrandMark } from '../components/BrandMark';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly preview: InventoryInvitationPreview; readonly reference: InventoryInvitationReference }
  | { readonly status: 'accepting'; readonly preview: InventoryInvitationPreview; readonly reference: InventoryInvitationReference }
  | { readonly status: 'accepted'; readonly inventoryId: string; readonly inventoryName: string }
  | { readonly status: 'opening'; readonly inventoryId: string; readonly inventoryName: string }
  | { readonly status: 'open_error'; readonly inventoryId: string; readonly inventoryName: string }
  | { readonly status: 'error'; readonly title: string; readonly message: string; readonly retryable: boolean; readonly canSwitchAccount?: boolean; readonly retryStartOver?: boolean };

export function InventoryInvitationScreen({
  acceptCommand,
  invalidLink,
  initialized,
  onAccepted,
  onDismiss,
  onSwitchAccount,
  onStartOver,
  previewQuery,
  reference
}: {
  readonly acceptCommand: Pick<AcceptInventoryInvitationCommand, 'execute'>;
  readonly invalidLink: boolean;
  readonly initialized: boolean;
  readonly onAccepted: (inventoryId: string) => Promise<void>;
  readonly onDismiss: () => void;
  readonly onSwitchAccount: () => void;
  readonly onStartOver?: () => Promise<void>;
  readonly previewQuery: Pick<PreviewInventoryInvitationQuery, 'execute'>;
  readonly reference?: InventoryInvitationReference;
}) {
  const colors = useAppearanceAwarePalette();
  const { fontScale, width } = useWindowDimensions();
  const styles = createStyles(colors, fontScale >= 2 || width < 340);
  const [state, setState] = useState<ScreenState>({ status: 'loading' });
  const requestGeneration = useRef(0);
  const [startingOver, setStartingOver] = useState(false);

  const load = async () => {
    if (!reference) return;
    const generation = ++requestGeneration.current;
    setState({ status: 'loading' });
    try {
      const preview = await previewQuery.execute(reference);
      if (generation !== requestGeneration.current) return;
      if (preview.status === 'accepted') {
        setState({ status: 'accepted', inventoryId: preview.inventoryId, inventoryName: preview.inventoryName });
      } else if (preview.status !== 'pending' || preview.isExpired) {
        setState(terminalState(preview));
      } else {
        setState({ status: 'ready', preview, reference });
      }
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      setState(errorState(error));
    }
  };

  useEffect(() => {
    setStartingOver(false);
    if (!initialized) return;
    if (invalidLink || !reference) {
      setState({
        status: 'error',
        title: 'Invitation not available',
        message: 'This invitation link is incomplete or invalid. Ask the sender for a new link.',
        retryable: false
      });
      return;
    }
    void load();
    return () => { requestGeneration.current += 1; };
  }, [initialized, invalidLink, reference]);

  const accept = async () => {
    if (state.status !== 'ready') return;
    const preview = state.preview;
    const acceptedReference = state.reference;
    const generation = ++requestGeneration.current;
    setState({ status: 'accepting', preview, reference: acceptedReference });
    try {
      const acceptance = await acceptCommand.execute(acceptedReference);
      if (generation !== requestGeneration.current) return;
      setState({ status: 'accepted', inventoryId: acceptance.inventoryId, inventoryName: preview.inventoryName });
    } catch (error) {
      if (generation !== requestGeneration.current) return;
      setState(errorState(error));
    }
  };

  return (
    <ScrollView contentContainerStyle={styles.page} style={styles.scroller}>
      <BrandMark />
      <View accessibilityLiveRegion="polite" style={styles.card}>
        {state.status === 'loading' ? (
          <StateMessage icon={<ActivityIndicator color={colors.action} />} title="Checking invitation"
            message="Confirming that this invitation is still available…" styles={styles} />
        ) : state.status === 'ready' || state.status === 'accepting' ? (
          <>
            <MailCheck color={colors.action} size={34} />
            <Text accessibilityRole="header" style={styles.title}>You’re invited</Text>
            <Text style={styles.message}>{state.preview.inventoryName}</Text>
            <View style={styles.details}>
              <View style={styles.detailRow}>
                <Text style={styles.detailLabel}>Access</Text>
                <Text style={styles.detailValue}>{relationshipLabel(state.preview.relationship)}</Text>
              </View>
              <View style={styles.detailRow}>
                <Text style={styles.detailLabel}>Expires</Text>
                <Text style={styles.detailValue}>{expirationLabel(state.preview.expiresAt)}</Text>
              </View>
            </View>
            <View style={styles.commands}>
              <NativeCommandButton label="Join inventory" prominence="primary"
                disabled={state.status === 'accepting' || startingOver} onPress={() => void accept()} />
              {state.status === 'accepting' ? <InvitationProgress label="Joining…" /> : null}
              {onStartOver ? <NativeCommandButton label="Sign out and start over"
                disabled={startingOver || state.status === 'accepting'} onPress={() => void startOver()} /> : null}
              {startingOver ? <InvitationProgress label="Starting over…" /> : null}
              <NativeCommandButton label="Not now" disabled={startingOver} onPress={onDismiss} />
            </View>
          </>
        ) : state.status === 'accepted' || state.status === 'opening' || state.status === 'open_error' ? (
          <>
            <CheckCircle2 color={colors.success} size={36} />
            <Text accessibilityRole="header" style={styles.title}>You’re in</Text>
            <Text style={styles.message}>You now have access to {state.inventoryName}.</Text>
            {state.status === 'open_error' ? <Text style={styles.message}>The inventory could not be opened. Your access was still added.</Text> : null}
            <View style={styles.commands}>
              <NativeCommandButton label="Open inventory" prominence="primary" disabled={state.status === 'opening'}
                onPress={() => void openAcceptedInventory(state.inventoryId, state.inventoryName)} />
              {state.status === 'opening' ? <InvitationProgress label="Opening…" /> : null}
            </View>
          </>
        ) : (
          <>
            <Text accessibilityRole="header" style={styles.title}>{state.title}</Text>
            <Text style={styles.message}>{state.message}</Text>
            <View style={styles.commands}>
              {state.retryable ? <NativeCommandButton label="Try again" prominence="primary" disabled={startingOver}
                onPress={() => void (state.retryStartOver ? startOver() : load())} /> : null}
              {startingOver ? <InvitationProgress label="Starting over…" /> : null}
              {state.canSwitchAccount ? <NativeCommandButton label="Switch account" prominence="primary" onPress={onSwitchAccount} /> : null}
              <NativeCommandButton label="Done" onPress={onDismiss} />
            </View>
          </>
        )}
      </View>
    </ScrollView>
  );

  async function startOver() {
    if (!onStartOver || startingOver) return;
    const generation = ++requestGeneration.current;
    setStartingOver(true);
    try { await onStartOver(); }
    catch { if (generation === requestGeneration.current) setState({ status: 'error', title: 'Could not start over', message: 'Try again to sign out.', retryable: true, retryStartOver: true }); }
    finally { if (generation === requestGeneration.current) setStartingOver(false); }
  }

  async function openAcceptedInventory(inventoryId: string, inventoryName: string): Promise<void> {
    const generation = ++requestGeneration.current;
    setState({ status: 'opening', inventoryId, inventoryName });
    try {
      await onAccepted(inventoryId);
    } catch {
      if (generation === requestGeneration.current) setState({ status: 'open_error', inventoryId, inventoryName });
    }
  }
}

function InvitationProgress({ label }: { readonly label: string }) {
  const colors = useAppearanceAwarePalette();
  return <View accessibilityRole="progressbar" accessibilityLabel={label} accessibilityState={{ busy: true }}>
    <ActivityIndicator color={colors.action} />
    <Text style={{ color: colors.textMuted, textAlign: 'center' }}>{label}</Text>
  </View>;
}

function StateMessage({ icon, message, styles, title }: {
  readonly icon: React.ReactNode;
  readonly message: string;
  readonly styles: ReturnType<typeof createStyles>;
  readonly title: string;
}) {
  return <>{icon}<Text accessibilityRole="header" style={styles.title}>{title}</Text><Text style={styles.message}>{message}</Text></>;
}

function relationshipLabel(relationship: InventoryInvitationPreview['relationship']): string {
  return relationship === 'editor' ? 'Editor' : 'Viewer';
}

function expirationLabel(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

function terminalState(preview: InventoryInvitationPreview): ScreenState {
  if (preview.isExpired || preview.status === 'expired') {
    return { status: 'error', title: 'Invitation expired', message: 'Ask the sender for a new invitation.', retryable: false };
  }
  const label = preview.status === 'revoked' ? 'revoked' : 'cancelled';
  return { status: 'error', title: `Invitation ${label}`, message: 'This invitation can no longer be accepted.', retryable: false };
}

function errorState(error: unknown): ScreenState {
  if (error instanceof InventoryInvitationEmailMismatchError) {
    return { status: 'error', title: 'Different account needed', message: error.message, retryable: false, canSwitchAccount: true };
  }
  if (error instanceof InventoryInvitationAuthenticationRequiredError) {
    return { status: 'error', title: 'Sign in required', message: error.message, retryable: true };
  }
  if (error instanceof InventoryInvitationInvalidResponseError) {
    return { status: 'error', title: 'Could not verify invitation', message: 'The server returned an invalid response.', retryable: false };
  }
  if (error instanceof InventoryInvitationInvalidError) {
    return { status: 'error', title: 'Invitation not available', message: error.message, retryable: false };
  }
  return { status: 'error', title: 'Could not check invitation', message: 'Check your connection and try again.', retryable: true };
}

function createStyles(colors: MobileColorPalette, accessibilityLayout = false) {
  return StyleSheet.create({
    scroller: { backgroundColor: colors.background, flex: 1 },
    page: {
      alignItems: 'center',
      flexGrow: 1,
      justifyContent: accessibilityLayout ? 'flex-start' : 'center',
      paddingHorizontal: accessibilityLayout ? spacing.md : spacing.lg,
      paddingVertical: spacing.lg
    },
    card: {
      alignItems: 'center',
      backgroundColor: colors.surface,
      borderRadius: accessibilityLayout ? radius.md : radius.lg,
      marginTop: accessibilityLayout ? spacing.md : spacing.xl,
      maxWidth: 520,
      padding: accessibilityLayout ? spacing.md : spacing.lg,
      width: '100%'
    },
    title: { color: colors.text, fontSize: accessibilityLayout ? 22 : 28, fontWeight: '800', marginTop: spacing.md, textAlign: 'center' },
    message: { color: colors.textMuted, fontSize: accessibilityLayout ? 16 : 17, lineHeight: accessibilityLayout ? 23 : 24, marginTop: spacing.sm, textAlign: 'center' },
    details: { alignSelf: 'stretch', backgroundColor: colors.surfaceMuted, borderRadius: radius.md, marginTop: spacing.lg, paddingHorizontal: spacing.md },
    detailRow: { alignItems: 'flex-start', alignSelf: 'stretch', minHeight: 48, paddingVertical: spacing.sm },
    detailLabel: { color: colors.textMuted, fontSize: 15 },
    detailValue: { color: colors.text, fontSize: 16, fontWeight: '700' },
    commands: { alignSelf: 'stretch', gap: spacing.sm, marginTop: spacing.lg }
  });
}
