import type { ReactNode } from 'react';
import type { NativeSheetActionsProps } from './NativeSheetActions.types';

export type NativeFilterSheetProps = {
  readonly title: string;
  readonly children: ReactNode;
  readonly actions: Omit<NativeSheetActionsProps, 'keyboardAvoidance'>;
  readonly footerTestID: string;
};
