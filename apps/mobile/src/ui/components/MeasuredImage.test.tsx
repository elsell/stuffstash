import React from 'react';
import {expect, it} from 'vitest';
import {MobileRenderHarness} from '../../test-support/render';
import {MeasuredImage} from './MeasuredImage';
import {ImagePerformanceProvider} from './ImagePerformanceContext';
import type {PerformanceContext, PerformanceOutcome, PerformanceObserver} from '../../application/observability/PerformanceObserver';

it('measures native image attempts and ignores old credential callbacks after replacement', async () => {
 const events: {context: PerformanceContext; outcomes: PerformanceOutcome[]}[] = [];
 const observer: PerformanceObserver = {start(context) {const event = {context, outcomes: [] as PerformanceOutcome[]}; events.push(event); return outcome => event.outcomes.push(outcome);}};
 const harness = new MobileRenderHarness();
 const render = (authorization: string) => harness.render(<ImagePerformanceProvider value={observer}><MeasuredImage surface="home" variant="small" source={{uri: 'https://private.invalid/photo', headers: {Authorization: authorization}}} /></ImagePerformanceProvider>);
 try {
  await render('first');
  const old = harness.allByType('Image')[0]!.props;
  await harness.run(() => old.onLoadStart());
  await render('second');
  await harness.run(() => old.onLoad());
  const current = harness.allByType('Image')[0]!.props;
  await harness.run(() => {current.onLoadStart(); current.onError(); current.onLoad();});
  await render('third');
  await harness.run(() => harness.allByType('Image')[0]!.props.onLoadStart());
  await harness.unmount();
  expect(events.map(event => event.outcomes)).toEqual([['cancelled'], ['failure'], ['cancelled']]);
  expect(events[0].context).toEqual({operation: 'image', surface: 'home', variant: 'small'});
 } finally {await harness.unmount();}
});

it('preserves product callbacks when the observer throws and handles cached completion', async () => {
 const harness = new MobileRenderHarness(); let loaded = 0;
 try {
  await harness.render(<ImagePerformanceProvider value={{start() {throw Error('telemetry');}}}><MeasuredImage surface="detail" variant="medium" source={{uri: 'file:local'}} onLoad={() => loaded++} /></ImagePerformanceProvider>);
  await harness.run(() => harness.allByType('Image')[0]!.props.onLoad());
  expect(loaded).toBe(1);
 } finally {await harness.unmount();}
});

it('settles cached success once and treats another load-start as a new attempt', async () => {
 const outcomes: string[][] = [];
 const observer: PerformanceObserver = {start() {const values: string[] = []; outcomes.push(values); return value => values.push(value);}};
 const harness = new MobileRenderHarness();
 try {
  await harness.render(<ImagePerformanceProvider value={observer}><MeasuredImage source={{uri: 'file:cached'}} /></ImagePerformanceProvider>);
  const props = harness.allByType('Image')[0]!.props;
  await harness.run(() => {props.onLoad(); props.onLoad(); props.onLoadStart(); props.onLoadStart(); props.onLoad();});
  expect(outcomes).toEqual([['success'], ['cancelled'], ['success']]);
 } finally {await harness.unmount();}
});
