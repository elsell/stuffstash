import { useReducedMotionPreference } from '../accessibility/useReducedMotionPreference';
import { createContext, ReactNode, useCallback, useContext, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { AccessibilityInfo, Alert, AlertButton, Animated, Platform, useWindowDimensions, PanResponder, Pressable, StyleSheet, Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import {
  AppNoticeInput,
  AppNoticeTone,
  buildAppNoticePresentation
} from './AppFeedbackPresentation';
import { useNoticeAccessibility } from './useNoticeAccessibility';

export type ShowAppNoticeInput = AppNoticeInput & {
  readonly action?: {
    readonly label: string;
    readonly onPress: () => void;
  };
};

export type ShowAppDialogInput = {
  readonly title: string;
  readonly message: string;
  readonly primaryAction: {
    readonly label: string;
    readonly onPress?: () => void;
  };
  readonly secondaryAction?: {
    readonly label: string;
    readonly onPress?: () => void;
  };
  readonly cancelable?: boolean;
};

export type AppFeedbackContextValue = {
  readonly showNotice: (input: ShowAppNoticeInput) => void;
  readonly showDialog: (input: ShowAppDialogInput) => void;
};

type ActiveNotice = ShowAppNoticeInput & {
  readonly id: number;
  readonly presentation: { entered: boolean };
  readonly owner: { active: boolean };
};

const AppFeedbackContext = createContext<AppFeedbackContextValue | null>(null);
const AppNoticeContext = createContext<{
  readonly notice: ActiveNotice | null;
  readonly placement: 'root' | 'screen';
  readonly dismiss: (id: number) => void;
} | null>(null);


export function AppFeedbackProvider({ children, scopeKey = 'app', noticePlacement = 'root' }: { readonly children: ReactNode; readonly scopeKey?: string; readonly noticePlacement?: 'root' | 'screen' }) {
  const [activeNotice, setActiveNotice] = useState<ActiveNotice | null>(null);
  const noticeSequence = useRef(0);
  const insets = useSafeAreaInsets();
  const noticeOwner = useMemo(() => ({ active: true }), [scopeKey]);
  useLayoutEffect(() => {
    noticeOwner.active = true;
    return () => { noticeOwner.active = false; };
  }, [noticeOwner]);
  useEffect(() => {
    setActiveNotice(current => current?.owner === noticeOwner ? current : null);
  }, [noticeOwner]);

  const showDialog = useCallback((input: ShowAppDialogInput) => {
    const buttons: AlertButton[] = [];
    if (input.secondaryAction) {
      buttons.push({
        text: input.secondaryAction.label,
        style: 'cancel',
        onPress: input.secondaryAction.onPress
      });
    }
    buttons.push({
      text: input.primaryAction.label,
      style: 'default',
      onPress: input.primaryAction.onPress
    });

    Alert.alert(
      input.title,
      input.message,
      buttons,
      { cancelable: input.cancelable ?? false }
    );
  }, []);
  const value = useMemo<AppFeedbackContextValue>(() => ({
    showDialog,
    showNotice: (input) => {
      if (!noticeOwner.active) return;
      setActiveNotice({
        ...input,
        action: input.action ? { ...input.action, onPress: () => {
          if (noticeOwner.active) input.action?.onPress();
        } } : undefined,
        owner: noticeOwner,
        id: ++noticeSequence.current,
        presentation: { entered: false }
      });
    }
  }), [noticeOwner, showDialog]);

  const dismissNotice = useCallback((id: number) => {
    setActiveNotice(current => current?.id === id ? null : current);
  }, []);

  return (
    <AppFeedbackContext.Provider value={value}>
      <AppNoticeContext.Provider value={{ notice: activeNotice?.owner === noticeOwner ? activeNotice : null, placement: noticePlacement, dismiss: dismissNotice }}>
      {children}
      {activeNotice?.owner === noticeOwner ? <NoticeLifetime key={`lifetime-${activeNotice.id}`} notice={activeNotice} onDismiss={dismissNotice} /> : null}
      {noticePlacement === 'root' && activeNotice?.owner === noticeOwner ? (
        <AppNotice
          key={activeNotice.id}
          notice={activeNotice}
          topOffset={insets.top + spacing.sm}
          onDismiss={dismissNotice}
        />
      ) : null}
      </AppNoticeContext.Provider>
    </AppFeedbackContext.Provider>
  );
}

/** A route owns placement, while the provider retains service-scoped action ownership. */
export function AppNoticePresenter({ topOffset }: { readonly topOffset: number }) {
  const state = useContext(AppNoticeContext);
  if (!state || state.placement !== 'screen' || !state.notice) return null;
  return <AppNotice key={state.notice.id} notice={state.notice} topOffset={topOffset} onDismiss={state.dismiss} />;
}

export function useAppFeedback(): AppFeedbackContextValue {
  const feedback = useContext(AppFeedbackContext);
  if (!feedback) {
    throw new Error('App feedback is not available.');
  }
  return feedback;
}

/** Remains mounted through route focus changes, unlike individual native presenters. */
function NoticeLifetime({ notice, onDismiss }: { readonly notice: ActiveNotice; readonly onDismiss: (id: number) => void }) {
  const { screenReader } = useNoticeAccessibility();
  const palette = useAppearancePalette();
  const presentation = buildAppNoticePresentation({ ...notice, actionLabel: notice.action?.label }, palette);
  useEffect(() => {
    if (Platform.OS === 'ios') AccessibilityInfo.announceForAccessibility(presentation.accessibilityLabel);
  }, [presentation.accessibilityLabel]);
  useEffect(() => {
    if (screenReader || presentation.durationMs === null) return;
    const timeout = setTimeout(() => onDismiss(notice.id), presentation.durationMs);
    return () => clearTimeout(timeout);
  }, [notice.id, onDismiss, presentation.durationMs, screenReader]);
  return null;
}

function AppNotice({
  topOffset,
  notice,
  onDismiss
}: {
  readonly topOffset: number;
  readonly notice: ActiveNotice;
  readonly onDismiss: (id: number) => void;
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const reduceMotion = useReducedMotionPreference();
  const animateEntry = useRef(!notice.presentation.entered).current;
  const { fontScale } = useWindowDimensions();
  const opacity = useRef(new Animated.Value(animateEntry ? 0 : 1)).current;
  const translateY = useRef(new Animated.Value(animateEntry ? -120 : 0)).current;
  const isDismissingRef = useRef(false);
  const presentation = buildAppNoticePresentation({
    actionLabel: notice.action?.label,
    message: notice.message,
    title: notice.title,
    tone: notice.tone
  }, palette);

  const dismissWithAnimation = useCallback((afterDismiss?: () => void) => {
    if (isDismissingRef.current) {
      return;
    }

    isDismissingRef.current = true;
    if (reduceMotion) {
      onDismiss(notice.id);
      afterDismiss?.();
      return;
    }
    Animated.parallel([
      Animated.timing(translateY, {
        duration: 170,
        toValue: -120,
        useNativeDriver: true
      }),
      Animated.timing(opacity, {
        duration: 130,
        toValue: 0,
        useNativeDriver: true
      })
    ]).start(() => {
      onDismiss(notice.id);
      afterDismiss?.();
    });
  }, [notice.id, onDismiss, opacity, reduceMotion, translateY]);

  useEffect(() => {
    notice.presentation.entered = true;
    if (reduceMotion || !animateEntry) {
      opacity.stopAnimation();
      translateY.stopAnimation();
      opacity.setValue(1);
      translateY.setValue(0);
      return;
    }
    const animation = Animated.parallel([
      Animated.spring(translateY, {
        speed: 20,
        bounciness: 5,
        toValue: 0,
        useNativeDriver: true
      }),
      Animated.timing(opacity, {
        duration: 160,
        toValue: 1,
        useNativeDriver: true
      })
    ]);
    animation.start();
    return () => animation.stop();
  }, [animateEntry, notice.presentation, opacity, reduceMotion, translateY]);

  const restorePosition = useCallback(() => {
    if (reduceMotion) { translateY.setValue(0); return; }
    Animated.spring(translateY, { speed: 22, bounciness: 4, toValue: 0, useNativeDriver: true }).start();
  }, [reduceMotion, translateY]);

  const panResponder = useMemo(() => PanResponder.create({
    onMoveShouldSetPanResponder: (_event, gestureState) =>
      gestureState.dy < -6 && Math.abs(gestureState.dy) > Math.abs(gestureState.dx),
    onPanResponderMove: (_event, gestureState) => {
      translateY.setValue(Math.min(0, gestureState.dy));
    },
    onPanResponderRelease: (_event, gestureState) => {
      if (gestureState.dy < -28 || gestureState.vy < -0.45) {
        dismissWithAnimation();
        return;
      }

      restorePosition();
    },
    onPanResponderTerminate: restorePosition
  }), [dismissWithAnimation, restorePosition, translateY]);

  return (
    <View
      testID="app-notice-layer"
      accessibilityLiveRegion="polite"
      accessibilityRole="alert"
      pointerEvents="box-none"
      style={[styles.noticeLayer, { top: topOffset }]}
    >
      <Animated.View
        testID="app-notice-container"
        accessibilityLabel={presentation.accessibilityLabel}
        {...panResponder.panHandlers}
        style={[
          styles.notice,
          fontScale >= 1.3 && { flexDirection: 'column', alignItems: 'stretch' },
          {
            backgroundColor: presentation.backgroundColor,
            borderColor: presentation.borderColor,
            opacity,
            transform: [{ translateY }]
          }
        ]}
      >
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={`${presentation.accessibilityLabel}. Dismiss message`}
          hitSlop={spacing.sm}
          onPress={() => dismissWithAnimation()}
          style={[styles.noticeBody, fontScale >= 1.3 && { flex: 0 }]}
        >
          <NoticeToneDot palette={palette} tone={notice.tone} />
          <View style={styles.noticeText}>
            <Text style={[styles.noticeTitle, { color: presentation.textColor }]}>
              {presentation.title}
            </Text>
            {presentation.message ? (
              <Text style={[styles.noticeMessage, { color: presentation.textColor }]}>
                {presentation.message}
              </Text>
            ) : null}
          </View>
        </Pressable>
        {notice.action ? (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={notice.action.label}
            onPress={() => {
              dismissWithAnimation(notice.action?.onPress);
            }}
            style={styles.noticeAction}
          >
            <Text style={styles.noticeActionText}>{notice.action.label}</Text>
          </Pressable>
        ) : null}
      </Animated.View>
    </View>
  );
}

function NoticeToneDot({
  palette,
  tone
}: {
  readonly palette: MobileColorPalette;
  readonly tone: AppNoticeTone;
}) {
  const styles = createStyles(palette);
  const backgroundColor =
    tone === 'success'
      ? palette.success
      : tone === 'warning'
        ? palette.brandAmber
        : tone === 'error'
          ? palette.danger
          : palette.accent;

  return <View style={[styles.noticeDot, { backgroundColor }]} />;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  noticeLayer: {
    left: spacing.md,
    position: 'absolute',
    right: spacing.md,
    zIndex: 1000
  },
  notice: {
    alignItems: 'center',
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 54,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    shadowColor: colors.brandCharcoalDeep,
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.16,
    shadowRadius: 18
  },
  noticeAction: {
    minHeight: 48,
    justifyContent: 'center',
    paddingHorizontal: spacing.xs
  },
  noticeActionText: {
    color: colors.action,
    fontSize: 15,
    fontWeight: '800'
  },
  noticeBody: {
    alignItems: 'center',
    flex: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 48
  },
  noticeDot: {
    borderRadius: 5,
    height: 10,
    width: 10
  },
  noticeMessage: {
    fontSize: 13,
    lineHeight: 17,
    marginTop: 2
  },
  noticeText: {
    flex: 1
  },
  noticeTitle: {
    fontSize: 14,
    fontWeight: '800',
    lineHeight: 18
  }
  });
}
