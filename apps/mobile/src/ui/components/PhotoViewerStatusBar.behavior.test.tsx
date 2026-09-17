import React from 'react';
import { expect, it } from 'vitest';
import { Platform } from '../../test-support/react-native';
import { MobileRenderHarness } from '../../test-support/render';
import { FullScreenPhotoViewer } from './FullScreenPhotoViewer';

it('owns light status content only during a valid iOS photo presentation', async () => {
  const h = new MobileRenderHarness();
  const platform = Platform.OS;
  const photos = [{ id: 'photo', label: 'Photo', uri: 'https://example.invalid/photo' }];
  const render = (index: number | undefined, available = photos) => <FullScreenPhotoViewer
    canRemove={false} currentIndex={index} photos={available}
    onClose={() => {}} onSelectIndex={() => {}} />;
  try {
    Platform.OS = 'ios';
    await h.render(render(undefined)); expect(h.allByType('StatusBar')).toHaveLength(0);
    await h.render(render(0)); expect(h.allByType('StatusBar')[0]?.props.barStyle).toBe('light-content');
    await h.render(render(undefined)); expect(h.allByType('StatusBar')).toHaveLength(0);
    await h.render(render(0)); expect(h.allByType('StatusBar')).toHaveLength(1);
    await h.render(render(0, [])); expect(h.allByType('StatusBar')).toHaveLength(0);
    await h.render(render(3)); expect(h.allByType('StatusBar')).toHaveLength(0);
    Platform.OS = 'android';
    await h.render(render(0)); expect(h.allByType('StatusBar')).toHaveLength(0);
  } finally { await h.unmount(); Platform.OS = platform; }
});
