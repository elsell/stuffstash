import { describe, expect, it } from 'vitest';
import { voiceResponseMarkdown } from './VoiceResponseMarkdown';
describe('native response formatting', () => {
 it('formats the baby-clothes answer with separate list rows and bold names', () => {
  const result = voiceResponseMarkdown('Your baby clothes are in the Garage:\n\n* **0–3M Clothes**: In Bin 58.\n* **6–9M Clothes**: In Bin 106.');
  expect(result.map(block => block.prefix)).toEqual(['', '• ', '• ']);
  expect(result[1].spans).toEqual([{text:'0–3M Clothes', bold:true}, {text:': In Bin 58.'}]);
 });
 it('supports headings, ordered lists, emphasis and code without opening markdown URLs', () => {
  const blocks = voiceResponseMarkdown('# Found\n1. *Garage* and `Bin 58`\n[Drill](https://untrusted.example)');
  expect(blocks[0].heading).toBe(true);
  expect(blocks[1].prefix).toBe('1. ');
  expect(blocks[1].spans).toContainEqual({text:'Garage', italic:true});
  expect(blocks[1].spans).toContainEqual({text:'Bin 58', code:true});
  expect(blocks[2].spans).toEqual([{text:'Drill'}]);
 });
 it('leaves unmatched punctuation and asset title underscores intact', () => {
  expect(voiceResponseMarkdown('Bin_A and 3 * 4')[0].spans).toEqual([{text:'Bin_A and 3 * 4'}]);
 });
});
