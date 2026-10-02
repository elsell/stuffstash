#!/usr/bin/env node
// Triage inventory, not an assertion that every candidate is user-facing copy.
import { readFile, readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import ts from '../apps/mobile/node_modules/typescript/lib/typescript.js';
import { parse } from '../apps/web/node_modules/svelte/src/compiler/index.js';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const files = [];
const displayAttributes = new Set(['title', 'label', 'aria-label', 'placeholder', 'alt', 'description', 'message', 'accessibilityLabel', 'accessibilityHint']);
const human = (text) => /[A-Za-z]{2}/.test(text) && (/\s/.test(text) || /^[A-Z][a-z]+$/.test(text));

function scriptCandidates(source, filename) {
  const ast = ts.createSourceFile(filename, source, ts.ScriptTarget.Latest, true,
    filename.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS);
  let count = 0;
  function visit(node) {
    if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node) || ts.isLiteralTypeNode(node)) return;
    if (ts.isCallExpression(node) && node.expression.getText(ast) === 't') return;
    if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node) || ts.isTemplateHead(node) || ts.isTemplateMiddle(node) || ts.isTemplateTail(node)) {
      if (human(node.text)) count++;
      return;
    }
    if (ts.isJsxText(node)) { if (human(node.getText(ast).trim())) count++; return; }
    ts.forEachChild(node, visit);
  }
  visit(ast);
  return count;
}

function componentCandidates(source, filename) {
  const ast = parse(source, { modern: true });
  let count = 0;
  for (const script of [ast.instance, ast.module].filter(Boolean)) {
    const start = source.indexOf('>', script.start) + 1;
    count += scriptCandidates(source.slice(start, source.indexOf('</script>', start)), `${filename}.ts`);
  }
  function visit(node) {
    if (!node || typeof node !== 'object') return;
    if (node.type === 'Attribute') {
      for (const value of Array.isArray(node.value) ? node.value : [node.value]) {
        if (value?.type === 'ExpressionTag' || displayAttributes.has(node.name)) visit(value);
      }
      return;
    }
    if (node.type === 'EachBlock') {
      count += scriptCandidates(source.slice(node.expression.start, node.expression.end), `${filename}.ts`);
    }
    if (node.type === 'Text') { if (human(node.data.trim())) count++; return; }
    if (node.type === 'ExpressionTag') {
      count += scriptCandidates(source.slice(node.expression.start, node.expression.end), `${filename}.ts`);
      return;
    }
    if (node.type === 'Comment') return;
    for (const [key, value] of Object.entries(node)) {
      if (['loc', 'metadata', 'expression'].includes(key)) continue;
      if (Array.isArray(value)) value.forEach(visit);
      else if (value && typeof value === 'object') visit(value);
    }
  }
  visit(ast.fragment);
  return count;
}

async function visit(directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const filename = path.join(directory, entry.name);
    if (/test-support|fixtures|generated/.test(filename)) continue;
    if (entry.isDirectory()) { await visit(filename); continue; }
    if (!/\.(ts|tsx|svelte)$/.test(filename) || /\.(test[-.]|spec\.|d\.ts)/.test(filename)) continue;
    const source = await readFile(filename, 'utf8');
    const candidates = filename.endsWith('.svelte') ? componentCandidates(source, filename) : scriptCandidates(source, filename);
    if (candidates) files.push({ file: path.relative(root, filename), candidates });
  }
}
await visit(path.join(root, 'apps/mobile/src'));
await visit(path.join(root, 'apps/web/src'));
files.sort((a, b) => b.candidates - a.candidates || a.file.localeCompare(b.file));
process.stdout.write(JSON.stringify({
  purpose: 'Remaining literal-message candidates for manual classification. Includes technical errors and constants; not a test-cell count or proof of a product defect.',
  reviewedClassifications: JSON.parse(await readFile(path.join(root, 'scripts/client-message-classifications.json'), 'utf8')),
  files,
}, null, 2) + '\n');
