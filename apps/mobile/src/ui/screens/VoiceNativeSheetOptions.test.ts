import { expect, it } from 'vitest';
import { voiceNativeSheetOptions } from './VoiceNativeSheetOptions';
import { colors } from '../theme/tokens';

it('exposes Conversation header actions on Android without changing iOS detents', () => {
  expect(voiceNativeSheetOptions(colors, 'android')).toMatchObject({ presentation: 'card', headerShown: true, title: 'Conversation' });
  expect(voiceNativeSheetOptions(colors, 'android')).not.toHaveProperty('sheetAllowedDetents');
  expect(voiceNativeSheetOptions(colors, 'ios')).toMatchObject({ presentation: 'formSheet', sheetAllowedDetents: [0.42, 0.88], sheetInitialDetentIndex: 1 });
});
