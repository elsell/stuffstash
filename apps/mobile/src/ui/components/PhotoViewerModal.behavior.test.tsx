import React, { useEffect, useState, useImperativeHandle, forwardRef } from 'react';
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import ts from 'typescript';
import { expect, it } from 'vitest';
import useImageIndexChange from 'react-native-image-viewing/dist/hooks/useImageIndexChange';
import { MobileRenderHarness } from '../../test-support/render';

// Execute the installed viewer, with in-memory native presentation/list ports.
// Native image events are explicit inputs; no fake implements paging decisions.
it('keeps one presentation across pages and gives scrolling ownership to the selected image', async () => {
  let presentations = 0;
  let dismissals = 0;
  let scrollEnabled = true;
  const selections: number[] = [];
  function Modal({ children }: { children: React.ReactNode }) {
    useEffect(() => { presentations++; return () => { dismissals++; }; }, []);
    return <>{children}</>;
  }
  const List = forwardRef(function List(props: any, ref) {
    useImperativeHandle(ref, () => ({
      scrollToIndex: ({ index }: { index: number }) => selections.push(index),
      setNativeProps: (values: { scrollEnabled: boolean }) => { scrollEnabled = values.scrollEnabled; }
    }));
    return React.createElement('NativePager', props,
      props.data.map((item: unknown, index: number) => React.createElement(React.Fragment, { key: props.keyExtractor(item, index) }, props.renderItem({ item, index }))));
  });
  const native = { Modal, VirtualizedList: List, View: 'View', Animated: { View: 'View' },
    Dimensions: { get: () => ({ width: 400, height: 800 }) }, StyleSheet: { create: (value: unknown) => value } };
  const require = createRequire(import.meta.url);
  const code = ts.transpileModule(readFileSync(require.resolve('react-native-image-viewing/dist/ImageViewing.js'), 'utf8'), {
    fileName: 'ImageViewing.jsx', compilerOptions: { allowJs: true, esModuleInterop: true, jsx: ts.JsxEmit.React, module: ts.ModuleKind.CommonJS }
  }).outputText;
  const dependencies: Record<string, unknown> = {
    react: React, 'react-native': native,
    './components/ImageItem/ImageItem': 'NativeImage',
    './components/ImageDefaultHeader': () => null,
    './components/StatusBarManager': () => null,
    './hooks/useImageIndexChange': useImageIndexChange,
    './hooks/useAnimatedComponents': () => [[], [], () => {}],
    './hooks/useRequestClose': (close: () => void) => [1, close]
  };
  const module = { exports: {} as { default?: React.ComponentType<any> } };
  new Function('require', 'module', 'exports', code)((name: string) => {
    if (!(name in dependencies)) throw new Error(`Unhandled native boundary: ${name}`);
    return dependencies[name];
  }, module, module.exports);
  const Viewer = module.exports.default!;
  const h = new MobileRenderHarness();
  const images = [{ uri: 'first' }, { uri: 'second' }];
  const render = (index: number, visible = true) => <Viewer images={images} imageIndex={index} visible={visible} onRequestClose={() => {}} />;
  try {
    await h.render(render(0));
    const first = h.allByType('NativeImage')[0];
    await h.run(() => first.props.onZoom(true));
    expect(scrollEnabled).toBe(false);
    await h.render(render(1));
    expect(presentations).toBe(1);
    expect(dismissals).toBe(0);
    expect(selections).toEqual([1]);
    expect(scrollEnabled).toBe(true);
    // A delayed event from the offscreen page cannot disable the active page.
    await h.run(() => first.props.onZoom(true));
    expect(scrollEnabled).toBe(true);
    await h.render(render(0));
    expect(scrollEnabled).toBe(false);
    await h.run(() => h.allByType('NativeImage')[0].props.onZoom(false));
    await h.run(() => h.byType('NativePager')!.props.onMomentumScrollEnd({ nativeEvent: { contentOffset: { x: 400 } } }));
    await h.render(render(1));
    expect(selections).toEqual([1, 0]);
    expect(presentations).toBe(1);
    await h.render(render(1, false));
    expect(dismissals).toBe(1);
    await h.render(render(1));
    expect(presentations).toBe(2);
    let close!: () => void;
    function ControlledViewer() {
      const [index, setIndex] = useState<number | undefined>(0);
      close = () => setIndex(undefined);
      return <Viewer images={images} imageIndex={index ?? 0} visible={index !== undefined}
        onRequestClose={close} onImageIndexChange={setIndex} />;
    }
    await h.render(<ControlledViewer />);
    const beforeClose = presentations;
    await h.run(() => close());
    expect(h.allByType('NativePager')).toHaveLength(0);
    expect(presentations).toBe(beforeClose);
  } finally { await h.unmount(); }
});
