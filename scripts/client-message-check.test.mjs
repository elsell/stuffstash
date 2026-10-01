import assert from 'node:assert/strict';
import { test } from 'node:test';
import { embeddedDisplayMessages } from './client-message-check.mjs';

test('flags rendered literals but preserves protocol values, user content and catalog calls', () => {
  const source = `const state = 'Saved'; const x = <View><Text>{t('save.done')}</Text><Text>{item.title}</Text><Button accessibilityLabel="Save item" /><Text>{busy ? 'Saving item' : 'Save item'}</Text></View>`;
  assert.deepEqual(embeddedDisplayMessages(source, 'Example.tsx').map(x => x.text), ['Save item', 'Saving item', 'Save item']);
});

test('checks Svelte text, attributes and conditional rendered expressions', () => {
  const source = `<script>let status = 'Saved';</script><button aria-label="Save item">Save item</button><span>{busy ? 'Saving item' : t('save.done')}</span><span>{item.title}</span>`;
  assert.deepEqual(embeddedDisplayMessages(source, 'Example.svelte').map(x => x.text), ['Save item', 'Save item', 'Saving item']);
});

test('checks rendered short-circuit and concatenation copy without interpreting predicates', () => {
  for (const filename of ['Example.tsx', 'Example.svelte']) {
    const tag = filename.endsWith('.tsx') ? 'Text' : 'span';
    const source = `<${tag}>{status === 'Ready' && 'Ready'}</${tag}><${tag}>{'Hello ' + name}</${tag}>`;
    assert.deepEqual(embeddedDisplayMessages(source, filename).map(x => x.text), ['Ready', 'Hello']);
  }
});

test('checks Svelte expression-valued display attributes', () => {
  const source = `<span aria-label={ready ? 'Ready to save' : 'Still loading'}></span>`;
  assert.deepEqual(embeddedDisplayMessages(source, 'Example.svelte').map(x => x.text), ['Ready to save', 'Still loading']);
});
