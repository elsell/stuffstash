import { Children, createElement, isValidElement, type ReactNode } from 'react';
type Props = { children?: ReactNode; expanded?: boolean; [key: string]: unknown };
const view = (name: string) => (props: Props) => createElement(name, props);
const Items = view('ComposeDropdownItems');
export const DropdownMenu = Object.assign((props: Props) => createElement('ComposeDropdownMenu', {
  ...props, children: Children.toArray(props.children).filter(child => props.expanded || !isValidElement(child) || child.type !== Items)
}), { Trigger: view('ComposeDropdownTrigger'), Items });
export const DropdownMenuItem = Object.assign(view('ComposeDropdownMenuItem'), {
  Text: view('ComposeDropdownText'), TrailingIcon: view('ComposeDropdownTrailingIcon')
});
export const Host = view('ComposeHost');
export const HorizontalDivider = view('ComposeDivider');
export const Icon = view('ComposeIcon');
export const Text = view('Text');
export const TextButton = view('ComposeTextButton');
export const OutlinedButton = view('ComposeOutlinedButton');
