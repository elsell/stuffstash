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

test('checks CSS-rendered table labels without treating data identities as copy', () => {
  const source = `<td data-cell-label="Source" data-state="Ready">{value}</td>`;
  assert.deepEqual(embeddedDisplayMessages(source, 'Example.svelte').map(x => x.text), ['Source']);
});

test('checks nested option labels without treating protocol and style values as copy', () => {
  for (const filename of ['Example.tsx', 'Example.svelte']) {
    const source = `<Control label={t('policy')} options={[{value:'defaults',label:'Use defaults'},{value:'off',label:t('off')}]} class={busy ? 'Busy' : 'Idle'} />`;
    assert.deepEqual(embeddedDisplayMessages(source, filename).map(x => x.text), ['Use defaults']);
  }
});

test('checks script-defined display properties in Svelte and TypeScript', () => {
  const definition = `const options = [{value:'archived',label:'Archived items'}];`;
  for (const filename of ['Example.tsx', 'Example.svelte']) {
    const source = filename.endsWith('.svelte') ? `<script>${definition}</script><Control options={options} />` : definition;
    assert.deepEqual(embeddedDisplayMessages(source, filename).map(x => x.text), ['Archived items']);
  }
});

test('checks direct Svelte option objects and preserves literal source offsets', () => {
  const source = `<Control config={{label:'Use defaults',value:'defaults',classes:{title:'toast-title'}}} />`;
  const issues = embeddedDisplayMessages(source, 'Example.svelte');
  assert.deepEqual(issues.map(issue => issue.text), ['Use defaults']);
  const [issue] = issues;
  assert.equal(source.slice(issue.offset, issue.offset + 14), "'Use defaults'");
});

test('checks each-block collection display properties and leaves protocol values intact', () => {
  const source = '<script>let local = [];</script>{#each level === "inventory" ? [{ value: "List", label: "List", title: `From ${tenant.name}` }] : [{ value: "Map", label: t("browse.map"), title: t("group", { name: inventory.name }) }] as option}<span>{option.label}</span>{/each}';
  assert.deepEqual(embeddedDisplayMessages(source, 'Example.svelte').map(x => x.text), ['List', 'From {value}']);
});

test('checks literal interpolation values without flagging keys or user text', () => {
  for (const filename of ['Example.tsx', 'Example.svelte']) {
    const expression = "t('current', { value: String(name || 'Top level'), selected: active ? 'Changes' : 'All events', literal: 'Photos', translated: t('photos'), title: item.title, count: String(12) })";
    const source = filename.endsWith('.svelte') ? '<span>{' + expression + '}</span>' : 'const label = ' + expression;
    assert.deepEqual(embeddedDisplayMessages(source, filename).map(x => x.text), ['Top level', 'Changes', 'All events', 'Photos']);
  }
});

test('reports a translated interpolation literal once at its source offset', () => {
  const source = "const label = t('notice', { title: 'Saved item' });";
  const issues = embeddedDisplayMessages(source, 'Example.ts');
  assert.deepEqual(issues.map(x => x.text), ['Saved item']);
  assert.equal(source.slice(issues[0].offset, issues[0].offset + 12), "'Saved item'");
});

test('checks literals hidden inside interpolation template spans', () => {
  const source = "const label = t('current', { value: `${active ? 'Changes' : 'All events'}` });";
  assert.deepEqual(embeddedDisplayMessages(source, 'Example.ts').map(x => x.text), ['Changes', 'All events']);
});

test('rejects ordinary error text in rendered recovery and returned helpers', () => {
  for (const source of [
    'function recovery(error, fallback) { return error instanceof Error ? error.message : fallback; }',
    'function recovery(error, fallback) { return (error instanceof Error ? error.message : fallback); }',
    'const recovery = (error, fallback) => (error instanceof Error ? error.message : fallback);',
    '<Text>{error instanceof Error ? error.message : fallback}</Text>',
    'const notice = { message: error instanceof Error ? error.message : fallback };'
  ]) assert.equal(embeddedDisplayMessages(source, 'Recovery.tsx').length, 1);
  assert.deepEqual(embeddedDisplayMessages('const diagnostic = error instanceof Error ? error.message : fallback; classify(diagnostic);', 'Recovery.tsx'), []);
  assert.deepEqual(embeddedDisplayMessages('function recovery(error, fallback) { return catalogRecoveryMessage(error, fallback); }', 'Recovery.tsx'), []);
});
