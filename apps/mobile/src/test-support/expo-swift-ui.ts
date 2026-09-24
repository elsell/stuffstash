import { createElement, type ReactNode } from 'react';
export const Host='SwiftUIHost';
export const Button='SwiftUIButton';
export const VStack='SwiftUIVStack';
export const HStack='SwiftUIHStack';
export const Spacer='SwiftUISpacer';
export const Text='Text';
export const Picker='SwiftUIPicker';
export const LabeledContent='SwiftUILabeledContent';
export const TextField='SwiftUITextField';
export const ColorPicker='SwiftUIColorPicker';
export const Menu='SwiftUIMenu';
export function Section({ header, footer, children, ...props }: { header?: ReactNode; footer?: ReactNode; children?: ReactNode; title?: string }) {
  return createElement('SwiftUISection', props, header, children, footer);
}
export const Image='SwiftUIImage';

export const List='SwiftUIList';
export const RNHostView='SwiftUIRNHostView';
