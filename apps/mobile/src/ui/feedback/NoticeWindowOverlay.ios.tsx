import { useContext } from 'react';
import { View, useWindowDimensions } from 'react-native';
import { FullWindowOverlay } from 'react-native-screens';
import { NoticeWindowContext } from './NoticeWindowContext';
import type { NoticeWindowOverlayProps } from './NoticeWindowOverlay';

const maximumBannerWidth = 720;

/** Window ownership keeps the banner interactive above native sheets and transitions. */
export function NoticeWindowOverlay({ children }: NoticeWindowOverlayProps) {
  const top = useContext(NoticeWindowContext);
  const { width, height } = useWindowDimensions();
  const bannerWidth = Math.min(width, maximumBannerWidth);
  return <FullWindowOverlay unstable_accessibilityContainerViewIsModal={false}>
    <View pointerEvents="box-none" style={{ position: 'absolute', top: 0,
      left: (width - bannerWidth) / 2, width: bannerWidth, height }}>
      {children(top)}
    </View>
  </FullWindowOverlay>;
}
