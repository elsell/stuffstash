import type { ReactNode } from 'react';
import type { NativeSheetActionsProps } from './NativeSheetActions.types';

export type NativeFilterSheetProps = {
  readonly title: string;
  readonly search?: {
    readonly query: string; readonly placeholder: string;
    readonly onChange: (text: string) => void;
    readonly onSubmit: (text: string) => void;
    readonly onClear: () => void;
  };
  readonly children: ReactNode;
  readonly actions: Omit<NativeSheetActionsProps, 'keyboardAvoidance'>;
  readonly footerTestID: string;
};
