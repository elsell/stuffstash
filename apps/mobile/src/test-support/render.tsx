import React, { act, type ReactElement } from 'react';
import { createRoot, type Root, type TestInstance } from 'test-renderer';

(globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

export class MobileRenderHarness {
  private readonly root: Root = createRoot({ textComponentTypes: ['Text'] });

  async render(element: ReactElement) { await act(async () => { this.root.render(element); }); }
  async settle() { await act(async () => { await Promise.resolve(); await Promise.resolve(); await Promise.resolve(); }); }
  async unmount() { await act(async () => this.root.unmount()); }
  all() { return this.root.container.queryAll(() => true); }
  byLabel(label: string) { return this.all().find((node) => node.props.accessibilityLabel === label); }
  byTestId(testId: string) { return this.all().find((node) => node.props.testID === testId); }
  byType(type: string) { return this.all().find((node) => node.type === type); }
  allByType(type: string) { return this.all().filter((node) => node.type === type); }
  byText(text: string) { return this.all().find((node) => node.type === 'Text' && node.children.includes(text)); }
  allText() { return this.allByType('Text').flatMap((node) => node.children.filter((child): child is string => typeof child === 'string')); }
  async press(node: TestInstance | undefined) { if (!node) throw new Error('Missing press target'); await act(async () => { await node.props.onPress?.(); }); }
  async changeText(node: TestInstance | undefined, value: string) { if (!node) throw new Error('Missing text input'); await act(async () => { node.props.onChangeText?.(value); }); }
  async change(node: TestInstance | undefined, value: string) { if (!node) throw new Error('Missing change target'); await act(async () => { (node.props.onChange ?? node.props.onValueChange)?.(value); }); }
  async accessibilityAction(node: TestInstance | undefined, actionName: string) { if (!node) throw new Error('Missing accessibility target'); await act(async () => { node.props.onAccessibilityAction?.({ nativeEvent: { actionName } }); }); }
  async run(action: () => unknown) { await act(async () => { await action(); }); }
}
