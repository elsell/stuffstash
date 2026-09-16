import { expect, it } from 'vitest';
import { returnFromFilterScreen } from './returnFromFilterScreen';

it.each([true, false])('leaves unapplied filters with history=%s', hasHistory => {
  const actions: string[] = [];
  returnFromFilterScreen({ canGoBack: () => hasHistory,
    back: () => { actions.push('back'); }, replace: path => { actions.push(`replace ${path}`); }
  });
  expect(actions).toEqual([hasHistory ? 'back' : 'replace /']);
});
