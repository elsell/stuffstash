import { expect, it } from 'vitest';
import { returnToPreviousOrHome } from './returnToPreviousOrHome';

it.each([true, false])('returns to the previous screen or Home with history=%s', hasHistory => {
  const actions: string[] = [];
  returnToPreviousOrHome({ canGoBack: () => hasHistory,
    back: () => { actions.push('back'); }, replace: path => { actions.push(`replace ${path}`); }
  });
  expect(actions).toEqual([hasHistory ? 'back' : 'replace /']);
});
