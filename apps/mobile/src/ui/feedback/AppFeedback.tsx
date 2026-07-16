import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { Alert, AlertButton, Animated, PanResponder, Pressable, StyleSheet, Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import {
  AppNoticeInput,
  AppNoticeTone,
  buildAppNoticePresentation
} from './AppFeedbackPresentation';

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
};

const AppFeedbackContext = createContext<AppFeedbackContextValue | null>(null);

export function AppFeedbackProvider({ children }: { readonly children: ReactNode }) {
  const [activeNotice, setActiveNotice] = useState<ActiveNotice | null>(null);
  const insets = useSafeAreaInsets();

  const value = useMemo<AppFeedbackContextValue>(() => ({
    showDialog: (input) => {
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
    },
    showNotice: (input) => {
      setActiveNotice((current) => ({
        ...input,
        id: (current?.id ?? 0) + 1
      }));
    }
  }), []);

  const dismissNotice = useCallback(() => {
    setActiveNotice(null);
  }, []);

  return (
    <AppFeedbackContext.Provider value={value}>
      {children}
      {activeNotice ? (
        <AppNotice
          notice={activeNotice}
          topOffset={insets.top + spacing.sm}
          onDismiss={dismissNotice}
        />
      ) : null}
    </AppFeedbackContext.Provider>
  );
}

export function useAppFeedback(): AppFeedbackContextValue {
  const feedback = useContext(AppFeedbackContext);
  if (!feedback) {
    throw new Error('App feedback is not available.');
  }
  return feedback;
}

function AppNotice({
  topOffset,
  notice,
  onDismiss
}: {
  readonly topOffset: number;
  readonly notice: ActiveNotice;
  readonly onDismiss: () => void;
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const opacity = useRef(new Animated.Value(0)).current;
  const translateY = useRef(new Animated.Value(-120)).current;
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
      onDismiss();
      afterDismiss?.();
    });
  }, [onDismiss, opacity, translateY]);

  useEffect(() => {
    Animated.parallel([
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
    ]).start();

    const timeout = setTimeout(() => {
      dismissWithAnimation();
    }, presentation.durationMs);

    return () => {
      clearTimeout(timeout);
    };
  }, [dismissWithAnimation, opacity, presentation.durationMs, translateY]);

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

      Animated.spring(translateY, {
        speed: 22,
        bounciness: 4,
        toValue: 0,
        useNativeDriver: true
      }).start();
    },
    onPanResponderTerminate: () => {
      Animated.spring(translateY, {
        speed: 22,
        bounciness: 4,
        toValue: 0,
        useNativeDriver: true
      }).start();
    }
  }), [dismissWithAnimation, translateY]);

  return (
    <View
      accessibilityLiveRegion="polite"
      accessibilityRole="alert"
      pointerEvents="box-none"
      style={[styles.noticeLayer, { top: topOffset }]}
    >
      <Animated.View
        accessibilityLabel={presentation.accessibilityLabel}
        {...panResponder.panHandlers}
        style={[
          styles.notice,
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
          accessibilityLabel="Dismiss message"
          hitSlop={spacing.sm}
          onPress={() => dismissWithAnimation()}
          style={styles.noticeBody}
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
    minHeight: 38,
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
    minHeight: 38
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
