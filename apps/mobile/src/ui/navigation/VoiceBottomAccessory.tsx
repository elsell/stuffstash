import { router, usePathname } from 'expo-router';
import { NativeTabs } from 'expo-router/unstable-native-tabs';
import { Check, Mic, Pause, X } from 'lucide-react-native';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { colors, spacing } from '../theme/tokens';
import { useVoiceInteractionState, VoiceInteractionStage } from './VoiceInteractionStateContext';

export function VoiceBottomAccessory() {
  const placement = NativeTabs.BottomAccessory.usePlacement();
  const isInline = placement === 'inline';
  const pathname = usePathname();
  const { reset, setStage, state } = useVoiceInteractionState();
  const contextLabel = describeVoiceContext(pathname);
  const status = describeVoiceStatus(state.stage);

  function advanceVoiceStage(): void {
    if (state.stage === 'ready') {
      setStage('listening');
      return;
    }

    if (state.stage === 'listening') {
      setStage('review');
      return;
    }

    router.push('/voice');
  }

  return (
    <View
      pointerEvents="box-none"
      style={[styles.accessoryArea, isInline ? styles.inlineArea : styles.regularArea]}
    >
      {isInline ? null : (
        <Pressable
          accessibilityRole="button"
          onPress={() => router.push('/voice')}
          style={styles.statusRegion}
        >
          <View style={[styles.statusDot, status.dotStyle]} />
          <View style={styles.statusText}>
            <Text numberOfLines={1} style={styles.statusTitle}>
              {status.title}
            </Text>
            <Text numberOfLines={1} style={styles.statusSubtitle}>
              {status.subtitle} · {contextLabel}
            </Text>
          </View>
        </Pressable>
      )}
      {state.stage === 'review' && !isInline ? (
        <View style={styles.reviewActions}>
          <Pressable
            accessibilityLabel="Cancel voice plan preview"
            accessibilityRole="button"
            onPress={reset}
            style={styles.smallAction}
          >
            <X color={colors.textMuted} size={17} strokeWidth={2.5} />
          </Pressable>
          <Pressable
            accessibilityLabel="Review voice plan"
            accessibilityRole="button"
            onPress={() => router.push('/voice')}
            style={[styles.smallAction, styles.reviewAction]}
          >
            <Check color={colors.onAction} size={17} strokeWidth={2.5} />
          </Pressable>
        </View>
      ) : null}
      <Pressable
        accessibilityLabel={state.stage === 'listening' ? 'Stop listening preview' : 'Start Voice'}
        accessibilityRole="button"
        accessibilityState={{ selected: state.stage === 'listening' }}
        onPress={advanceVoiceStage}
        style={({ pressed }) => [
          styles.button,
          isInline ? styles.inlineButton : styles.regularButton,
          state.stage === 'listening' && styles.listeningButton,
          state.stage === 'review' && styles.reviewButton,
          pressed && styles.buttonPressed
        ]}
      >
        {state.stage === 'listening' ? (
          <Pause color={colors.onAction} size={isInline ? 21 : 22} strokeWidth={2.5} />
        ) : (
          <Mic color={colors.onAction} size={isInline ? 22 : 23} strokeWidth={2.5} />
        )}
      </Pressable>
    </View>
  );
}

function describeVoiceStatus(stage: VoiceInteractionStage): {
  readonly dotStyle: object;
  readonly subtitle: string;
  readonly title: string;
} {
  if (stage === 'listening') {
    return {
      dotStyle: styles.listeningDot,
      subtitle: 'Capturing preview audio',
      title: 'Listening'
    };
  }

  if (stage === 'review') {
    return {
      dotStyle: styles.reviewDot,
      subtitle: 'Plan needs approval',
      title: 'Review voice plan'
    };
  }

  return {
    dotStyle: styles.readyDot,
    subtitle: 'Tap mic to start',
    title: 'Voice ready'
  };
}

function describeVoiceContext(pathname: string): string {
  if (pathname.startsWith('/assets/')) {
    return 'asset context';
  }

  if (pathname.startsWith('/locations/')) {
    return 'location context';
  }

  if (pathname === '/search') {
    return 'search context';
  }

  if (pathname === '/add') {
    return 'add context';
  }

  if (pathname === '/voice') {
    return 'voice workspace';
  }

  return 'current inventory';
}

const styles = StyleSheet.create({
  accessoryArea: {
    alignItems: 'center',
    flex: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'center',
  },
  inlineArea: {
    justifyContent: 'flex-end',
    paddingRight: spacing.sm
  },
  regularArea: {
    paddingHorizontal: spacing.md
  },
  statusRegion: {
    alignItems: 'center',
    flex: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 52,
    minWidth: 0
  },
  statusText: {
    flex: 1,
    minWidth: 0
  },
  statusTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  statusSubtitle: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '700',
    letterSpacing: 0,
    marginTop: 2
  },
  statusDot: {
    borderRadius: 5,
    height: 10,
    width: 10
  },
  readyDot: {
    backgroundColor: colors.accent
  },
  listeningDot: {
    backgroundColor: colors.success
  },
  reviewDot: {
    backgroundColor: colors.brandAmber
  },
  button: {
    alignItems: 'center',
    backgroundColor: colors.action,
    justifyContent: 'center',
    shadowColor: '#000000',
    shadowOffset: { width: 0, height: 3 },
    shadowOpacity: 0.16,
    shadowRadius: 8
  },
  listeningButton: {
    backgroundColor: colors.success
  },
  reviewButton: {
    backgroundColor: colors.brandAmber
  },
  inlineButton: {
    borderRadius: 14,
    height: 44,
    width: 44
  },
  regularButton: {
    borderRadius: 16,
    height: 52,
    width: 52
  },
  buttonPressed: {
    backgroundColor: colors.actionPressed,
    transform: [{ scale: 0.96 }]
  },
  reviewActions: {
    flexDirection: 'row',
    gap: spacing.xs
  },
  smallAction: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderRadius: 14,
    height: 34,
    justifyContent: 'center',
    width: 34
  },
  reviewAction: {
    backgroundColor: colors.action
  }
});
