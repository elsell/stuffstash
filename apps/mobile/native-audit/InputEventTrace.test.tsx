import React from 'react';
import { TextInput } from 'react-native';
import { afterEach, expect, it } from 'vitest';
import { MobileRenderHarness } from '../src/test-support/render';
import { useInputEventTrace } from './InputEventTrace';

let harness: MobileRenderHarness | undefined;
afterEach(async () => { await harness?.unmount(); harness = undefined; });

function setupProbe(onRender: () => void) {
  return function Probe() {
    onRender();
    const trace = useInputEventTrace('');
    return <><TextInput accessibilityLabel="Traced input" onKeyPress={trace.onKeyPress}
      onChange={trace.onChange} onSelectionChange={trace.onSelectionChange} />{trace.controls}</>;
  };
}

async function snapshot() {
  await harness!.press(harness!.all().find(node => node.props.title === 'Capture input events'));
  return JSON.parse(harness!.byTestId('audit-input-event-trace')!.props.children);
}

it('records key, change and selection ordering without rendering until capture', async () => {
  let renders = 0;
  const Probe = setupProbe(() => { renders += 1; });
  harness = new MobileRenderHarness();
  await harness.render(<Probe />);
  const before = renders;
  const input = harness.byLabel('Traced input')!;
  expect(input.props.onKeyPress).toBeTypeOf('function');
  await harness.run(() => {
    input.props.onKeyPress({ nativeEvent: { key: 'N' } });
    input.props.onChange({ nativeEvent: { text: 'N', eventCount: 1 } });
    input.props.onSelectionChange({ nativeEvent: { selection: { start: 1, end: 1 } } });
  });
  expect(renders).toBe(before);
  expect(harness.byTestId('audit-input-event-trace')).toBeUndefined();
  expect(await snapshot()).toEqual({ value: '', dropped: 0, events: [
    { kind: 'commit', text: '' }, { kind: 'key', key: 'N' },
    { kind: 'change', text: 'N', count: 1 }, { kind: 'selection', start: 1, end: 1 }
  ] });
});

it('bounds retained key events and reports discarded entries', async () => {
  const Probe = setupProbe(() => {});
  harness = new MobileRenderHarness();
  await harness.render(<Probe />);
  const input = harness.byLabel('Traced input')!;
  expect(input.props.onKeyPress).toBeTypeOf('function');
  await harness.run(() => {
    for (let index = 0; index < 300; index += 1) input.props.onKeyPress({ nativeEvent: { key: String(index) } });
  });
  const result = await snapshot();
  expect(result.dropped).toBe(45);
  expect(result.events).toHaveLength(256);
  expect(result.events[0]).toEqual({ kind: 'key', key: '44' });
  expect(result.events.at(-1)).toEqual({ kind: 'key', key: '299' });
});
