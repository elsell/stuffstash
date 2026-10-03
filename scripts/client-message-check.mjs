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
  const seen = new Set();
  const rawErrorBindings = new Map();
  const renderedBindings = new Set();
  const moduleBindings = new Map();
  const bindingNames = new Map();
  const ambiguousNames = new Set();
  const add = (text, offset) => {
    const key = JSON.stringify([offset, text.trim()]);
    if (human(text.replace(/\{[^}]+\}/g, '').trim()) && !seen.has(key)) {
      seen.add(key); found.push({ text: text.trim(), offset });
    }
  };
  function script(text, offset = 0, expression = false, rendered = true, templateBindings = new Map(), moduleScript = false) {
    if (expression) { text = `(${text})`; offset -= 1; }
    const ast = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true, filename.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS);
    const scopes = new Map();
    const scopeFor = node => {
      if (ts.isSourceFile(node)) return moduleBindings;
      if (!scopes.has(node)) scopes.set(node, new Map());
      return scopes.get(node);
    };
    const isScope = node => ts.isSourceFile(node) || ts.isBlock(node) ||
      ts.isFunctionLike(node) || ts.isCatchClause(node) || ts.isForStatement(node) ||
      ts.isForOfStatement(node) || ts.isForInStatement(node);
    function declare(name, owner) {
      if (ts.isIdentifier(name)) {
        if (moduleScript) ambiguousNames.add(name.text);
        const scope = scopeFor(owner);
        if (!scope.has(name.text)) scope.set(name.text, {});
      } else if (ts.isObjectBindingPattern(name) || ts.isArrayBindingPattern(name)) {
        for (const element of name.elements) if (ts.isBindingElement(element)) declare(element.name, owner);
      }
    }
    function declarations(node) {
      if (ts.isVariableDeclaration(node)) {
        let owner = node.parent;
        const lexical = ts.isVariableDeclarationList(owner) && (owner.flags & ts.NodeFlags.BlockScoped);
        while (owner && !(lexical ? isScope(owner) : ts.isFunctionLike(owner) || ts.isSourceFile(owner) || ts.isCatchClause(owner))) owner = owner.parent;
        declare(node.name, owner ?? ast);
      }
      if (ts.isParameter(node)) declare(node.name, node.parent);
      ts.forEachChild(node, declarations);
    }
    declarations(ast);
    function binding(node) {
      for (let owner = node.parent; owner; owner = owner.parent) {
        if (ts.isSourceFile(owner) && templateBindings.has(node.text)) return templateBindings.get(node.text);
        if (isScope(owner) && scopeFor(owner).has(node.text)) return scopeFor(owner).get(node.text);
      }
      if (!moduleBindings.has(node.text)) moduleBindings.set(node.text, {});
      return moduleBindings.get(node.text);
    }
    function rawErrorFallback(node) {
      if (!node || !ts.isConditionalExpression(node)) return false;
      const condition = node.condition;
      const value = node.whenTrue;
      return ts.isBinaryExpression(condition) && condition.operatorToken.kind === ts.SyntaxKind.InstanceOfKeyword &&
        ts.isIdentifier(condition.right) && condition.right.text === 'Error' &&
        ts.isIdentifier(condition.left) && ts.isPropertyAccessExpression(value) && value.name.text === 'message' &&
        ts.isIdentifier(value.expression) && value.expression.text === condition.left.text;
    }
    function rejectRawError(node) {
      while (node && ts.isParenthesizedExpression(node)) node = node.expression;
      if (!rawErrorFallback(node)) return false;
      add('ordinary Error.message requires catalog recovery', offset + node.getStart(ast));
      return true;
    }
    function rememberRawErrorBinding(name, node) {
      while (node && ts.isParenthesizedExpression(node)) node = node.expression;
      if (ts.isIdentifier(name) && rawErrorFallback(node)) {
        const key = binding(name);
        bindingNames.set(key, name.text);
        const offsets = rawErrorBindings.get(key) ?? [];
        offsets.push(offset + node.getStart(ast));
        rawErrorBindings.set(key, offsets);
      }
    }
    function output(node) {
      if (!node || rejectRawError(node)) return;
      if (ts.isIdentifier(node)) renderedBindings.add(binding(node));
      else if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) add(node.text, offset + node.getStart(ast));
      else if (ts.isParenthesizedExpression(node)) output(node.expression);
      else if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'String' && node.arguments.length === 1) output(node.arguments[0]);
      else if (ts.isConditionalExpression(node)) { output(node.whenTrue); output(node.whenFalse); }
      else if (ts.isBinaryExpression(node)) {
        if ([ts.SyntaxKind.QuestionQuestionToken, ts.SyntaxKind.BarBarToken, ts.SyntaxKind.PlusToken].includes(node.operatorToken.kind)) { output(node.left); output(node.right); }
        else if (node.operatorToken.kind === ts.SyntaxKind.AmpersandAmpersandToken) output(node.right);
      }
      else if (ts.isTemplateExpression(node)) {
        add(node.head.text + node.templateSpans.map(span => `{value}${span.literal.text}`).join(''), offset + node.getStart(ast));
        for (const span of node.templateSpans) output(span.expression);
      }
    }
    function properties(node) {
      if (ts.isVariableDeclaration(node)) rememberRawErrorBinding(node.name, node.initializer);
      if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.EqualsToken) rememberRawErrorBinding(node.left, node.right);
      if (ts.isReturnStatement(node)) rejectRawError(node.expression);
      if (ts.isArrowFunction(node) && !ts.isBlock(node.body)) rejectRawError(node.body);
      if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 't') {
        const values = node.arguments[1];
        if (values && ts.isObjectLiteralExpression(values)) {
          for (const property of values.properties) {
            if (ts.isPropertyAssignment(property)) output(property.initializer);
          }
        }
      }
      if (ts.isPropertyAssignment(node)) {
        const name = node.name.getText(ast).replace(/^['"]|['"]$/g, '');
        if (['classes', 'classNames', 'style', 'styles'].includes(name)) return;
        if (displayAttribute(name)) output(node.initializer);
      }
      ts.forEachChild(node, properties);
    }
    function visit(node) {
      if (ts.isJsxText(node)) add(node.getText(ast), offset + node.getStart(ast));
      else if (ts.isJsxAttribute(node)) {
        if (displayAttribute(node.name.getText(ast))) output(ts.isJsxExpression(node.initializer ?? {}) ? node.initializer.expression : node.initializer);
        return;
      } else if (ts.isJsxExpression(node)) output(node.expression);
      ts.forEachChild(node, visit);
    }
    properties(ast);
    if (expression) { const statement = ast.statements[0]; if (rendered && statement && ts.isExpressionStatement(statement)) output(statement.expression); }
    else visit(ast);
  }
  function finish() {
    for (const name of renderedBindings) {
      if (ambiguousNames.has(bindingNames.get(name))) continue;
      for (const offset of rawErrorBindings.get(name) ?? []) add('ordinary Error.message requires catalog recovery', offset);
    }
    return found;
  }
  if (!filename.endsWith('.svelte')) { script(source); return finish(); }
  const ast = parse(source, { modern: true });
  for (const section of [ast.instance, ast.module].filter(Boolean)) {
    script(source.slice(section.content.start, section.content.end), section.content.start, false, true, new Map(), section === ast.module);
  }
  function bindTemplate(parent, patterns) {
    const nested = new Map(parent);
    function bind(pattern) {
      if (!pattern) return;
      if (pattern.type === 'Identifier') nested.set(pattern.name, {});
      else if (pattern.type === 'ObjectPattern') pattern.properties.forEach(p => bind(p.value ?? p.argument));
      else if (pattern.type === 'ArrayPattern') pattern.elements.forEach(bind);
      else if (pattern.type === 'RestElement') bind(pattern.argument);
      else if (pattern.type === 'AssignmentPattern') bind(pattern.left);
    }
    patterns.forEach(bind);
    return nested;
  }
  function visit(node, templateBindings = new Map()) {
    if (!node || typeof node !== 'object') return;
    // Unsupported binding forms are left to caller review rather than linked
    // by name. Direct rendered expressions still receive all existing checks.
    if (node.type === 'ConstTag') {
      for (const declaration of node.declaration.declarations) {
        for (const name of bindTemplate(new Map(), [declaration.id]).keys()) ambiguousNames.add(name);
      }
    }
    if (node.type === 'LetDirective') {
      if (node.expression) {
        for (const name of bindTemplate(new Map(), [node.expression]).keys()) ambiguousNames.add(name);
      } else ambiguousNames.add(node.name);
    }
    if (node.type === 'Attribute') {
      for (const value of Array.isArray(node.value) ? node.value : [node.value]) {
        if (value?.type === 'ExpressionTag') {
          script(source.slice(value.expression.start, value.expression.end), value.expression.start, true, displayAttribute(node.name), templateBindings);
        } else if (displayAttribute(node.name)) visit(value, templateBindings);
      }
      return;
    }
    if (node.type === 'EachBlock') {
      script(source.slice(node.expression.start, node.expression.end), node.expression.start, true, false, templateBindings);
      const nested = bindTemplate(templateBindings, [node.context]);
      if (node.index) nested.set(node.index, {});
      visit(node.body, nested);
      visit(node.fallback, templateBindings);
      return;
    }
    if (node.type === 'AwaitBlock') {
      script(source.slice(node.expression.start, node.expression.end), node.expression.start, true, false, templateBindings);
      visit(node.pending, templateBindings);
      visit(node.then, bindTemplate(templateBindings, [node.value]));
      visit(node.catch, bindTemplate(templateBindings, [node.error]));
      return;
    }
    if (node.type === 'SnippetBlock') {
      visit(node.body, bindTemplate(templateBindings, node.parameters));
      return;
    }
    if (node.type === 'Text') { add(node.data, node.start); return; }
    if (node.type === 'ExpressionTag') { script(source.slice(node.expression.start, node.expression.end), node.expression.start, true, true, templateBindings); return; }
    if (node.type === 'Comment') return;
    for (const [key, value] of Object.entries(node)) {
      if (['loc', 'metadata', 'expression'].includes(key)) continue;
      if (Array.isArray(value)) value.forEach(child => visit(child, templateBindings));
      else if (value && typeof value === 'object') visit(value, templateBindings);
    }
  }
  visit(ast.fragment);
  return finish();
}

async function check(root) {
  let failures = 0;
  async function visit(directory) {
    for (const entry of await readdir(directory, { withFileTypes: true })) {
      const filename = path.join(directory, entry.name);
      if (/test-support|fixtures|generated|\/src\/test\/|\/testing\//.test(filename)) continue;
      if (entry.isDirectory()) { await visit(filename); continue; }
      if (!/\.(tsx?|svelte)$/.test(filename) || /\.(test[-.]|spec\.|bench\.|d\.ts)/.test(filename)) continue;
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
