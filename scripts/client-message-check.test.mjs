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

test('rejects raw error state used in rendered copy while allowing private diagnostics', () => {
  const failure = 'caught instanceof Error ? caught.message : fallback';
  const svelte = `<script>let saveError = ''; async function save() { try { await submit(); } catch (caught) { saveError = (${failure}); } }</script><p role="alert">{saveError}</p>`;
  const jsx = `const message = ${failure}; const view = <Text>{message}</Text>;`;
  for (const [source, file] of [[svelte, 'Recovery.svelte'], [jsx, 'Recovery.tsx']]) {
    const issues = embeddedDisplayMessages(source, file);
    assert.deepEqual(issues.map(x => x.text), ['ordinary Error.message requires catalog recovery']);
    assert.match(source.slice(issues[0].offset), /^caught instanceof Error/);
  }
  const privateDiagnostic = `<script>let note = ''; const diagnostic = ${failure}; classify(diagnostic);</script><p>{note}</p>`;
  assert.deepEqual(embeddedDisplayMessages(privateDiagnostic, 'Recovery.svelte'), []);
  assert.deepEqual(embeddedDisplayMessages(svelte.replace(`(${failure})`, 'safeWorkspaceErrorMessage(caught, fallback)'), 'Recovery.svelte'), []);
});

test('does not confuse private diagnostic bindings with same-named rendered props', () => {
  const source = `function classifyFailure(caught, fallback) {
    const message = caught instanceof Error ? caught.message : fallback;
    classify(message);
  }
  function Notice({message}) { return <Text>{message}</Text>; }`;
  assert.deepEqual(embeddedDisplayMessages(source, 'Recovery.tsx'), []);
  const svelte = `<script>let message = ''; function classifyFailure(caught, fallback) {
    const message = caught instanceof Error ? caught.message : fallback; classify(message);
  }</script><p>{message}</p>`;
  assert.deepEqual(embeddedDisplayMessages(svelte, 'Recovery.svelte'), []);
});

test('does not confuse Svelte each values with diagnostic state', () => {
  const source = `<script>const message = caught instanceof Error ? caught.message : fallback;</script>
    {#each notes as message}<p>{message}</p>{/each}`;
  assert.deepEqual(embeddedDisplayMessages(source, 'Recovery.svelte'), []);
});

test('does not confuse Svelte await and snippet bindings with diagnostic state', () => {
  const prefix = `<script>const message = caught instanceof Error ? caught.message : fallback;</script>`;
  for (const template of [
    '{#await task then message}<p>{message}</p>{/await}',
    '{#await task}<p>{pending}</p>{:catch message}<p>{message}</p>{/await}',
    '{#snippet notice(message)}<p>{message}</p>{/snippet}',
  ]) assert.deepEqual(embeddedDisplayMessages(prefix + template, 'Recovery.svelte'), []);
});

test('leaves ambiguous Svelte bindings to caller review without disabling direct checks', () => {
  const prefix = `<script>const message = caught instanceof Error ? caught.message : fallback;</script>`;
  for (const template of [
    '{#if visible}{@const message = note}<p>{message}</p>{/if}',
    '<Widget let:message><p>{message}</p></Widget>',
  ]) assert.deepEqual(embeddedDisplayMessages(prefix + template, 'Recovery.svelte'), []);
  assert.deepEqual(embeddedDisplayMessages(`<script module>const message = caught instanceof Error ? caught.message : fallback;</script><script>const message = value;</script><p>{message}</p>`, 'Recovery.svelte'), []);
  assert.equal(embeddedDisplayMessages(prefix + '<Widget let:message><p>{caught instanceof Error ? caught.message : fallback}</p></Widget>', 'Recovery.svelte').length, 1);
});

test('checks human labels in inline Svelte each-block arrays', () => {
  const source = `{#each [['soon', 'Expiring soon'], ['expired', 'Expired'], ['all', 'All dates']] as [mode, label]}<a href={mode}>{label}</a>{/each}`;
  assert.deepEqual(embeddedDisplayMessages(source, 'Expiration.svelte').map(issue => issue.text), ['Expiring soon', 'Expired', 'All dates']);
  assert.deepEqual(embeddedDisplayMessages(`{#each [['soon', t('soon')], ['expired', t('expired')]] as [mode, label]}<a href={mode}>{label}</a>{/each}`, 'Expiration.svelte'), []);
});

test('does not inspect non-display attribute arrays as each-block labels', () => {
  assert.deepEqual(embeddedDisplayMessages(`<div class={['rounded-lg shadow-sm']}>{value}</div>`, 'Styled.svelte'), []);
});
