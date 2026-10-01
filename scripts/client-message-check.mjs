import ts from '../apps/mobile/node_modules/typescript/lib/typescript.js';
import { parse } from '../apps/web/node_modules/svelte/src/compiler/index.js';
import { readFile, readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const human = text => /[A-Za-z]{2}/.test(text);
const displayAttribute = name => ['title', 'label', 'message', 'description', 'placeholder', 'accessibilityLabel', 'accessibilityHint', 'aria-label', 'data-cell-label', 'alt', 'footer', 'hint', 'subtitle', 'caption', 'text'].includes(name) || /(?:Label|Title|Placeholder|Description|Legend|Hint)$/.test(name);

/** Only directly rendered copy. State, protocol and variable provenance need source review. */
export function embeddedDisplayMessages(source, filename) {
  const found = [];
  const add = (text, offset) => { if (human(text.replace(/\{[^}]+\}/g, '').trim())) found.push({ text: text.trim(), offset }); };
  function script(text, offset = 0, expression = false) {
    const ast = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true, filename.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS);
    function output(node) {
      if (!node) return;
      if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) add(node.text, offset + node.getStart(ast));
      else if (ts.isParenthesizedExpression(node)) output(node.expression);
      else if (ts.isConditionalExpression(node)) { output(node.whenTrue); output(node.whenFalse); }
      else if (ts.isBinaryExpression(node)) {
        if ([ts.SyntaxKind.QuestionQuestionToken, ts.SyntaxKind.BarBarToken, ts.SyntaxKind.PlusToken].includes(node.operatorToken.kind)) { output(node.left); output(node.right); }
        else if (node.operatorToken.kind === ts.SyntaxKind.AmpersandAmpersandToken) output(node.right);
      }
      else if (ts.isTemplateExpression(node)) add(node.head.text + node.templateSpans.map(span => `{value}${span.literal.text}`).join(''), offset + node.getStart(ast));
    }
    function visit(node) {
      if (ts.isJsxText(node)) add(node.getText(ast), offset + node.getStart(ast));
      else if (ts.isJsxAttribute(node)) {
        if (displayAttribute(node.name.getText(ast))) output(ts.isJsxExpression(node.initializer ?? {}) ? node.initializer.expression : node.initializer);
        return;
      } else if (ts.isJsxExpression(node)) output(node.expression);
      ts.forEachChild(node, visit);
    }
    if (expression) { const statement = ast.statements[0]; if (statement && ts.isExpressionStatement(statement)) output(statement.expression); }
    else visit(ast);
  }
  if (!filename.endsWith('.svelte')) { script(source); return found; }
  const ast = parse(source, { modern: true });
  function visit(node) {
    if (!node || typeof node !== 'object') return;
    if (node.type === 'Attribute') { if (displayAttribute(node.name)) { if (Array.isArray(node.value)) node.value.forEach(visit); else visit(node.value); } return; }
    if (node.type === 'Text') { add(node.data, node.start); return; }
    if (node.type === 'ExpressionTag') { script(source.slice(node.expression.start, node.expression.end), node.expression.start, true); return; }
    if (node.type === 'Comment') return;
    for (const [key, value] of Object.entries(node)) {
      if (['loc', 'metadata', 'expression'].includes(key)) continue;
      if (Array.isArray(value)) value.forEach(visit);
      else if (value && typeof value === 'object') visit(value);
    }
  }
  visit(ast.fragment);
  return found;
}

async function check(root) {
  let failures = 0;
  async function visit(directory) {
    for (const entry of await readdir(directory, { withFileTypes: true })) {
      const filename = path.join(directory, entry.name);
      if (/test-support|fixtures|generated|\/src\/test\//.test(filename)) continue;
      if (entry.isDirectory()) { await visit(filename); continue; }
      if (!/\.(tsx?|svelte)$/.test(filename) || /\.(test[-.]|spec\.|d\.ts)/.test(filename)) continue;
      const source = await readFile(filename, 'utf8');
      for (const issue of embeddedDisplayMessages(source, filename)) {
        process.stderr.write(`${path.relative(root, filename)}:${source.slice(0, issue.offset).split('\n').length}: catalog required for ${JSON.stringify(issue.text)}\n`);
        failures++;
      }
    }
  }
  await visit(path.join(root, 'apps/mobile/src'));
  await visit(path.join(root, 'apps/web/src'));
  return failures;
}
if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
  process.exitCode = await check(root) ? 1 : 0;
}
