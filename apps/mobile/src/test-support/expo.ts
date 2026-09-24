import { createElement, useImperativeHandle, type ComponentType, type Ref } from 'react';

type DetachedBoundary = { measureInKeyboardWindow(): Promise<null> };

/** A detached native boundary cannot provide a keyboard-window measurement. */
export function requireNativeView<Props extends object>(name: string): ComponentType<Props> {
  if (name === 'StuffStashPhotoSystemBars') return (props: Props) => createElement(name, props);
  if (name !== 'StuffStashSheetBoundary') throw new Error(`Unsupported native view: ${name}`);
  return function NativeView(props: Props) {
    const { ref, ...attributes } = props as Props & { ref?: Ref<DetachedBoundary> };
    useImperativeHandle(ref, () => ({ measureInKeyboardWindow: async () => null }), []);
    return createElement(name, attributes);
  };
}
