import { afterEach, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setPathname } from '../../test-support/expo-router';
import { dispatchedActions, resetNavigation } from '../../test-support/navigation';
import { labelLinkSourceFake } from '../../test-support/expo-linking';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { LabelLinkProvider } from './LabelLinkContext';
import { PendingLabelNavigation } from './PendingLabelNavigation';

const link = 'stuffstash://labels/v1/01ARZ3NDEKTSV4RRFFQ69G5FAV/01ARZ3NDEKTSV4RRFFQ69G5FAW';
afterEach(() => { resetNavigation(); setPathname('/'); labelLinkSourceFake.reset(); });
const view = <LabelLinkProvider><AppFeedbackProvider><PendingLabelNavigation /></AppFeedbackProvider></LabelLinkProvider>;

it.each(['/', '/search'])('opens a valid incoming label once from %s', async path => {
  const h = new MobileRenderHarness(); setPathname(path);
  try {
    await h.render(view);
    await h.run(() => labelLinkSourceFake.receive(link));
    expect(dispatchedActions()).toEqual([{ type: 'push', href: '/scan-label' }]);
    await h.run(() => setPathname('/scan-label'));
    await h.run(() => setPathname(path));
    expect(dispatchedActions()).toHaveLength(1);
    expect(h.byText('Your label is ready to open.')).toBeUndefined();
  } finally { await h.unmount(); }
});

it.each(['/add', '/assets/asset/edit', '/settings/inventory/fields/new', '/search/other'])('protects %s with one explicit navigation offer per incoming revision', async path => {
  const h = new MobileRenderHarness(); setPathname(path);
  try {
    await h.render(view);
    await h.run(() => labelLinkSourceFake.receive(link));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byText('Your label is ready to open.')).toBeDefined();
    await h.press(h.all().find(node => String(node.props.accessibilityLabel).endsWith('. Dismiss message')));
    await h.run(() => setPathname('/'));
    expect(h.byText('Your label is ready to open.')).toBeUndefined();
    expect(dispatchedActions()).toEqual([]);
    await h.run(() => setPathname(path));
    await h.run(() => labelLinkSourceFake.receive(link));
    await h.press(h.byLabel('Open label'));
    expect(dispatchedActions()).toEqual([{ type: 'push', href: '/scan-label' }]);
  } finally { await h.unmount(); }
});

it('reports malformed incoming links honestly and does not automatically navigate', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(view);
    await h.run(() => labelLinkSourceFake.receive('stuffstash://labels/v9/invalid/invalid'));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byText('Your label is ready to open.')).toBeUndefined();
    expect(h.byText('This is not a supported Stuff Stash label.')).toBeDefined();
  } finally { await h.unmount(); }
});

it('leaves incoming references on the existing scanner without stacking another route', async () => {
  const h = new MobileRenderHarness(); setPathname('/scan-label');
  try {
    await h.render(view);
    await h.run(() => labelLinkSourceFake.receive(link));
    await h.run(() => setPathname('/'));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byText('Your label is ready to open.')).toBeUndefined();
  } finally { await h.unmount(); }
});
