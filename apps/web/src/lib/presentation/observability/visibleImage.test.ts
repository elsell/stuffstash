import {expect, it} from 'vitest';
import {visibleImage} from './visibleImage';
import type {PerformanceContext, PerformanceOutcome, PerformanceObserver} from '$lib/ports/performanceObserver';

function fixture() {
 const events: {context: PerformanceContext; outcomes: PerformanceOutcome[]}[] = [];
 const observer: PerformanceObserver = {start(context) { const event = {context, outcomes: [] as PerformanceOutcome[]}; events.push(event); return outcome => event.outcomes.push(outcome); }};
 const image = document.createElement('img');
 Object.defineProperties(image, {complete: {value: false, configurable: true}, naturalWidth: {value: 0, configurable: true}});
 return {events, observer, image};
}
it('records each rendered source once and cancels replacement and unmount without exporting URLs', () => {
 const {events, observer, image} = fixture();
 const first = {source: 'blob:secret', observer, surface: 'home' as const, variant: 'small' as const};
 const action = visibleImage(image, first);
 action.update({...first});
 action.update({...first, source: 'blob:replacement'});
 image.dispatchEvent(new Event('load')); image.dispatchEvent(new Event('error'));
 action.update({...first, source: 'blob:third'});
 action.destroy(); image.dispatchEvent(new Event('load'));
 expect(events.map(event => event.outcomes)).toEqual([['cancelled'], ['success'], ['cancelled']]);
 expect(events[0].context).toEqual({operation: 'image', surface: 'home', variant: 'small'});
});
it('reports cached success, cached failure and loading error distinctly', () => {
 for (const width of [0, 20]) {
  const {events, observer, image} = fixture();
  Object.defineProperties(image, {complete: {value: true}, naturalWidth: {value: width}});
  const action = visibleImage(image, {source: 'blob:cached', observer, surface: 'detail', variant: 'medium'});
  action.destroy(); expect(events[0].outcomes).toEqual([width ? 'success' : 'failure']);
 }
 const {events, observer, image} = fixture();
 const action = visibleImage(image, {source: 'blob:failed', observer, surface: 'gallery', variant: 'small'});
 image.dispatchEvent(new Event('error')); action.destroy();
 expect(events[0].outcomes).toEqual(['failure']);
});
it('keeps image lifecycle usable when telemetry start or finish throws', () => {
 for (const observer of [
  {start() {throw Error('telemetry start');}},
  {start() {return () => {throw Error('telemetry finish');};}}
 ]) {
  const {image} = fixture();
  const action = visibleImage(image, {source: 'blob:photo', observer, surface: 'upload', variant: 'original'});
  expect(() => image.dispatchEvent(new Event('load'))).not.toThrow();
  expect(() => action.destroy()).not.toThrow();
 }
});
