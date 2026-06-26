import { useEffect, useRef } from 'react';
import { router, usePathname } from 'expo-router';
import { NativeTabs } from 'expo-router/unstable-native-tabs';
import { Mic, Pause } from 'lucide-react-native';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { colors, spacing } from '../theme/tokens';
import { useVoiceInteractionState } from './VoiceInteractionStateContext';
import { buildVoiceAccessoryPresentation } from './VoiceSessionPresentation';

export function VoiceBottomAccessory() {
  const placement = NativeTabs.BottomAccessory.usePlacement();
  const isInline = placement === 'inline';
  const pathname = usePathname();
  const isOpeningVoiceSheet = useRef(false);
  const { startRealtime, state, stopRealtime } = useVoiceInteractionState();
  const presentation = buildVoiceAccessoryPresentation({
    pathname,
    stage: state.stage,
    status: state.status
  });

  useEffect(() => {
    if (pathname !== '/voice') {
      isOpeningVoiceSheet.current = false;
    }
  }, [pathname]);

  async function handlePrimaryAction(): Promise<void> {
    if (state.status !== 'ready') {
      openVoiceSheet();
      return;
    }

    if (presentation.primaryAction === 'start') {
      openVoiceSheet();
      void startRealtime();
      return;
    }

    if (presentation.primaryAction === 'stop') {
      openVoiceSheet();
      void stopRealtime();
      return;
    }

    openVoiceSheet();
  }

  function openVoiceSheet(): void {
    if (pathname === '/voice' || isOpeningVoiceSheet.current) {
      return;
    }

    isOpeningVoiceSheet.current = true;
    router.navigate('/voice');
  }

  return (
    <View
      pointerEvents="box-none"
      style={[styles.accessoryArea, isInline ? styles.inlineArea : styles.regularArea]}
    >
      {isInline ? null : (
        <Pressable
          accessibilityLabel="Open voice session"
          accessibilityRole="button"
          onPress={openVoiceSheet}
          style={styles.statusRegion}
        >
          <View style={[styles.statusDot, dotStyleForTone(presentation.tone)]} />
          <View style={styles.statusText}>
            <Text numberOfLines={1} style={styles.statusTitle}>
              {presentation.title}
            </Text>
            <Text numberOfLines={1} style={styles.statusSubtitle}>
              {presentation.subtitle}
            </Text>
          </View>
        </Pressable>
      )}

      <Pressable
        accessibilityLabel={presentation.accessibilityLabel}
        accessibilityRole="button"
        accessibilityState={{ selected: state.stage === 'listening' }}
        onPress={() => {
          void handlePrimaryAction();
        }}
        style={({ pressed }) => [
          styles.button,
          isInline ? styles.inlineButton : styles.regularButton,
          presentation.tone === 'active' && styles.activeButton,
          presentation.tone === 'attention' && styles.attentionButton,
          presentation.tone === 'failed' && styles.failedButton,
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

function dotStyleForTone(tone: 'ready' | 'active' | 'attention' | 'failed') {
  switch (tone) {
    case 'active':
      return styles.activeDot;
    case 'attention':
      return styles.attentionDot;
    case 'failed':
      return styles.failedDot;
    case 'ready':
      return styles.readyDot;
  }
}

const styles = StyleSheet.create({
  accessoryArea: {
    alignItems: 'center',
    flex: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'center'
  },
  activeButton: {
    backgroundColor: colors.success
  },
  activeDot: {
    backgroundColor: colors.success
  },
  attentionButton: {
    backgroundColor: colors.brandAmber
  },
  attentionDot: {
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
  buttonPressed: {
    opacity: 0.84,
    transform: [{ scale: 0.98 }]
  },
  failedButton: {
    backgroundColor: colors.danger
  },
  failedDot: {
    backgroundColor: colors.danger
  },
  inlineArea: {
    justifyContent: 'flex-end',
    paddingRight: spacing.sm
  },
  inlineButton: {
    borderRadius: 22,
    height: 44,
    width: 44
  },
  readyDot: {
    backgroundColor: colors.accent
  },
  regularArea: {
    paddingHorizontal: spacing.md
  },
  regularButton: {
    borderRadius: 27,
    height: 54,
    width: 54
  },
  statusDot: {
    borderRadius: 5,
    height: 10,
    width: 10
  },
  statusRegion: {
    alignItems: 'center',
    flex: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 52,
    minWidth: 0
  },
  statusSubtitle: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '700',
    letterSpacing: 0,
    marginTop: 2
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
  }
});
