import React from 'react';
import { afterEach, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { FullSpectrumTagColorPicker } from './FullSpectrumTagColorPicker';

let screen: MobileRenderHarness;
afterEach(async () => { await screen?.unmount(); });

it('retained color gestures use the current parent and stop after disable or teardown', async () => {
  screen = new MobileRenderHarness();
  const old: string[] = [], current: string[] = [];
  await screen.render(<FullSpectrumTagColorPicker value="#FF0000" onChange={value => old.push(value)} />);
  const move = screen.byLabel('Saturation and brightness')!.props.onPanResponderMove;
  const hue = screen.byLabel('Hue')!.props.onPanResponderMove;
  const adjust = screen.byLabel('Hue')!.props.onAccessibilityAction;
  const increaseHue = screen.byLabel('Increase hue')!.props.onPress;
  await screen.render(<FullSpectrumTagColorPicker value="#00FF00" onChange={value => current.push(value)} />);
  await screen.run(() => move({ nativeEvent: { locationX: 320, locationY: 0 } }));
  expect(old).toEqual([]);
  expect(current).toEqual(['#00FF00']);
  await screen.run(() => increaseHue());
  expect(current).toEqual(['#00FF00', '#00FF15']);
  await screen.render(<FullSpectrumTagColorPicker disabled value="#00FF00" onChange={value => current.push(value)} />);
  await screen.run(() => move({ nativeEvent: { locationX: 100, locationY: 100 } }));
  await screen.run(() => hue({ nativeEvent: { locationX: 100 } }));
  await screen.run(() => adjust({ nativeEvent: { actionName: 'increment' } }));
  await screen.run(() => increaseHue());
  expect(old).toEqual([]);
  expect(current).toEqual(['#00FF00', '#00FF15']);
  await screen.unmount();
  await screen.run(() => move({ nativeEvent: { locationX: 100, locationY: 100 } }));
  await screen.run(() => adjust({ nativeEvent: { actionName: 'increment' } }));
  await screen.run(() => increaseHue());
  expect(old).toEqual([]);
  expect(current).toEqual(['#00FF00', '#00FF15']);
});
