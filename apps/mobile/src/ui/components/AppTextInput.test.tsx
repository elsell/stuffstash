import React, { createRef } from 'react';
import { Platform, type TextInput } from 'react-native';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNativeTestState } from '../../test-support/react-native';
import {
  keyboardControllerDismissCount,
  resetKeyboardControllerTestState
} from '../../test-support/react-native-keyboard-controller';
import { AppKeyboardAccessory as IOSAppKeyboardAccessory } from './AppKeyboardAccessory.ios';
import { AppKeyboardAccessory as DefaultAppKeyboardAccessory } from './AppKeyboardAccessory';
import { AppKeyboardProvider as IOSAppKeyboardProvider } from './AppKeyboardProvider.ios';
import { AppKeyboardProvider as DefaultAppKeyboardProvider } from './AppKeyboardProvider';
import { appKeyboardDismissMode, AppTextInput } from './AppTextInput';
// @ts-expect-error Vite provides raw source imports to structural integration tests.
import rootLayoutSource from '../../app/_layout.tsx?raw';

let harness: MobileRenderHarness | undefined;
const originalPlatform = Platform.OS;
const mutablePlatform = Platform as unknown as { OS: string };

beforeEach(() => {
  mutablePlatform.OS = 'ios';
  resetNativeTestState();
  resetKeyboardControllerTestState();
});

afterEach(async () => {
  mutablePlatform.OS = originalPlatform;
  await harness?.unmount();
  harness = undefined;
});

describe('AppTextInput', () => {
  it('forwards single-line and multiline input props and refs through the shared adapter', async () => {
    const ref = createRef<TextInput>();
    const changes: string[] = [];
    harness = new MobileRenderHarness();
    await harness.render(
      <AppTextInput
        accessibilityLabel="Item name"
        autoCapitalize="words"
        multiline
        onChangeText={(value) => changes.push(value)}
        ref={ref}
        value="Camera"
      />
    );

    const input = harness.byLabel('Item name');
    expect(input?.type).toBe('TextInput');
    expect(input?.props).toMatchObject({
      autoCapitalize: 'words',
      multiline: true,
      value: 'Camera'
    });
    expect(input?.props.inputAccessoryViewID).toBeUndefined();
    await harness.changeText(input, 'Camera bag');
    ref.current?.focus();
    expect(changes).toEqual(['Camera bag']);
  });

  it('preserves disabled input semantics on every platform', async () => {
    harness = new MobileRenderHarness();
    await harness.render(<AppTextInput accessibilityLabel="Read only" editable={false} value="Fixed" />);
    expect(harness.byLabel('Read only')?.props.editable).toBe(false);

    mutablePlatform.OS = 'android';
    await harness.render(<AppTextInput accessibilityLabel="Android field" value="Editable" />);
    expect(harness.byLabel('Android field')?.props.value).toBe('Editable');
  });

  it('provides native scroll-dismiss modes for both platforms', () => {
    expect(appKeyboardDismissMode()).toBe('interactive');
    mutablePlatform.OS = 'android';
    expect(appKeyboardDismissMode()).toBe('on-drag');
  });

  it('does not create an accessory instance for each input', async () => {
    harness = new MobileRenderHarness();
    await harness.render(
      <>
        <AppTextInput accessibilityLabel="First field" />
        <AppTextInput accessibilityLabel="Second field" />
      </>
    );
    expect(harness.all().filter((node) => node.type === 'InputAccessoryView')).toHaveLength(0);
  });
});

describe('AppKeyboardAccessory', () => {
  it('mounts one project-owned provider and one keyboard extension at the application root', () => {
    expect(rootLayoutSource.match(/<AppKeyboardProvider>/g)).toHaveLength(1);
    expect(rootLayoutSource.match(/<AppKeyboardAccessory\s*\/>/g)).toHaveLength(1);
  });

  it('disables provider preloading so app launch never flashes the keyboard', async () => {
    harness = new MobileRenderHarness();
    await harness.render(<IOSAppKeyboardProvider><AppTextInput accessibilityLabel="Name" /></IOSAppKeyboardProvider>);
    expect(harness.byType('KeyboardProvider')?.props.preload).toBe(false);
  });

  it('renders one accessible iOS keyboard-down action with a 44-point target', async () => {
    let changes = 0;
    let submissions = 0;
    harness = new MobileRenderHarness();
    await harness.render(
      <>
        <AppTextInput
          accessibilityLabel="Description"
          multiline
          onChangeText={() => { changes += 1; }}
          onSubmitEditing={() => { submissions += 1; }}
          value="Still editing"
        />
        <IOSAppKeyboardAccessory />
      </>
    );

    expect(harness.all().filter((node) => node.type === 'KeyboardExtender')).toHaveLength(1);
    const bar = harness.all().find((node) => node.props.style?.some?.((style: unknown) => (
      typeof style === 'object' && style !== null && 'width' in style && style.width === '100%'
    )));
    expect(bar?.props.style).toEqual(expect.arrayContaining([
      expect.objectContaining({ height: 44, width: '100%' }),
      expect.objectContaining({
        paddingRight: 20
      })
    ]));
    const action = harness.byLabel('Dismiss keyboard');
    expect(action?.props.accessibilityRole).toBe('button');
    expect(action?.props.style).toEqual(expect.objectContaining({ height: 44, width: 44 }));
    expect(harness.byType('ChevronDownIcon')?.props).toMatchObject({
      color: 'platform:link',
      size: 24,
      strokeWidth: 2.2
    });

    await harness.press(action);
    expect(keyboardControllerDismissCount()).toBe(1);
    expect(changes).toBe(0);
    expect(submissions).toBe(0);
    expect(harness.byLabel('Description')?.props.value).toBe('Still editing');
  });

  it('keeps the default platform accessory empty and provider transparent', async () => {
    harness = new MobileRenderHarness();
    await harness.render(
      <DefaultAppKeyboardProvider>
        <AppTextInput accessibilityLabel="System keyboard field" />
        <DefaultAppKeyboardAccessory />
      </DefaultAppKeyboardProvider>
    );
    expect(harness.all().filter((node) => node.type === 'KeyboardExtender')).toHaveLength(0);
    expect(harness.all().filter((node) => node.type === 'KeyboardProvider')).toHaveLength(0);
    expect(harness.byLabel('Dismiss keyboard')).toBeUndefined();
    expect(harness.byLabel('System keyboard field')).toBeDefined();
  });
});
