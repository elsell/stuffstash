import {useLayoutEffect, useMemo} from 'react';
import {Image, type ImageProps} from 'react-native';
import type {PerformanceContext, PerformanceOutcome} from '../../application/observability/PerformanceObserver';
import {useImagePerformanceObserver} from './ImagePerformanceContext';

export function MeasuredImage({surface = 'list', variant = 'none', ...props}: ImageProps & {
 surface?: PerformanceContext['surface'];
 variant?: PerformanceContext['variant'];
}) {
 const observer = useImagePerformanceObserver();
 // Identity stays local, including headers: only the bounded context crosses the port.
 const identity = JSON.stringify(props.source);
 const lifecycle = useMemo(() => {
  let active = false;
  let started = false;
  let finished = false;
  let finish: (outcome: PerformanceOutcome) => void = () => {};
  function start() {
   started = true; finished = false;
   try {finish = observer.start({operation: 'image', surface, variant});} catch {finish = () => {};}
  }
  function settle(outcome: PerformanceOutcome) {
   if (!active || finished) return;
   if (!started) start();
   finished = true;
   try {finish(outcome);} catch { /* Telemetry must not affect display. */ }
  }
  return {
   mount() {active = true;},
   dispose() {if (started) settle('cancelled'); active = false;},
   start() {if (!active) return; if (started && !finished) settle('cancelled'); start();},
   settle
  };
 }, [identity, observer, surface, variant]);
 useLayoutEffect(() => {lifecycle.mount(); return () => lifecycle.dispose();}, [lifecycle]);
 return <Image {...props}
  onLoadStart={() => {lifecycle.start(); props.onLoadStart?.();}}
  onLoad={event => {lifecycle.settle('success'); props.onLoad?.(event);}}
  onError={event => {lifecycle.settle('failure'); props.onError?.(event);}}
 />;
}
