import { createElement, type ComponentType } from 'react';

type ImageViewingProps = {
  readonly visible: boolean;
  readonly imageIndex: number;
  readonly FooterComponent?: ComponentType<{ imageIndex: number }>;
  readonly [key: string]: unknown;
};

// Present the library-owned footer only while the native viewer is visible.
export default function ImageViewing(props: ImageViewingProps) {
  return createElement('ImageViewing', props, props.visible && props.FooterComponent
    ? createElement(props.FooterComponent, { imageIndex: props.imageIndex }) : null);
}
