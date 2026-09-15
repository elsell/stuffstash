import React from 'react';
import { expect, it } from 'vitest';
import { Pressable, Text } from 'react-native';
import { MobileRenderHarness } from '../../test-support/render';
import { VoiceTabContent } from './VoiceTabContent';

it.each([['android', 35, true], ['ios', '18.6', true], ['ios', '26.0', false]] as const)(
  'keeps tab content and exposes a voice command only without native accessory support: %s %s', async (platform, version, fallback) => {
    const h = new MobileRenderHarness();
    let opened = 0;
    try {
      await h.render(<VoiceTabContent platform={platform} version={version}
        accessory={<Pressable accessibilityLabel="Open voice" onPress={() => opened++}><Text>Voice</Text></Pressable>}>
        <Text>Inventory content</Text>
      </VoiceTabContent>);
      expect(h.allText()).toContain('Inventory content');
      const button = h.byLabel('Open voice');
      if (fallback) { expect(button).toBeDefined(); await h.press(button); expect(opened).toBe(1); }
      else { expect(button).toBeUndefined(); expect(opened).toBe(0); }
    } finally { await h.unmount(); }
  }
);
