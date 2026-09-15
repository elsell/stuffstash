import type { ReactNode } from 'react';
import { useHeaderHeight } from '@react-navigation/elements';
import { useIsFocused } from '@react-navigation/native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { spacing } from '../theme/tokens';
import { AppNoticePresenter } from './AppFeedback';

type NoticeScreenLayoutProps = {
  readonly children: ReactNode;
  readonly route: { readonly name: string };
  readonly options: { readonly headerTransparent?: boolean; readonly headerShown?: boolean };
};

/** Preserve the native screen's direct scroll child; container routes delegate to leaves. */
export function AppNoticeScreenLayout({ children, route, options }: NoticeScreenLayoutProps) {
  return <>{children}{route.name === '(tabs)' ? null : <FocusedNotice options={options} />}</>;
}

function FocusedNotice({ options }: Pick<NoticeScreenLayoutProps, 'options'>) {
  const focused = useIsFocused();
  const headerHeight = useHeaderHeight();
  const insets = useSafeAreaInsets();
  const top = options.headerShown === false ? insets.top : options.headerTransparent ? headerHeight : 0;
  return focused ? <AppNoticePresenter topOffset={top + spacing.sm} /> : null;
}
