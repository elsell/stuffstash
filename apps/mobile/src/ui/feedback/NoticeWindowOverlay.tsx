import type { ReactNode } from 'react';

export type NoticeWindowOverlayProps = { readonly children: (windowTop?: number) => ReactNode };
export function NoticeWindowOverlay({ children }: NoticeWindowOverlayProps) {
  return <>{children()}</>;
}
