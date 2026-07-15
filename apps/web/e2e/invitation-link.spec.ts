import { expect, test } from '@playwright/test';

const token = 'A'.repeat(43);
const invitationPath = `/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one#token=${token}`;

test('clickable invitation previews before explicit acceptance and enters the scoped inventory', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'One browser project is sufficient for the end-to-end contract.');
  await page.addInitScript(() => {
    sessionStorage.setItem('stuffstash.oidc.session', JSON.stringify({ idToken: 'invitee-token', expiresAt: Date.now() + 60_000 }));
  });
  let acceptCalls = 0;
  const requestURLs: string[] = [];
  await page.route('http://127.0.0.1:18080/**', async (route) => {
    const request = route.request();
    requestURLs.push(request.url());
    if (request.url().endsWith('/preview')) {
      await route.fulfill({ json: { data: {
        inventoryId: 'inventory-one', inventoryName: 'Workshop tools', relationship: 'viewer',
        status: 'pending', isExpired: false, expiresAt: '2026-07-21T12:00:00Z'
      }, meta: {} } });
      return;
    }
    if (request.url().endsWith('/accept')) {
      acceptCalls += 1;
      await route.fulfill({ json: { data: {
        invitation: {
          id: 'invite-one', tenantId: 'tenant-one', inventoryId: 'inventory-one', email: 'invitee@example.test',
          relationship: 'viewer', status: 'accepted', isExpired: false, expiresAt: '2026-07-21T12:00:00Z',
          inviterPrincipalId: 'inviter-one', acceptedPrincipalId: 'invitee-one'
        },
        grant: { tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'invitee-one', relationship: 'viewer' }
      }, meta: {} } });
      return;
    }
    await route.abort();
  });

  await page.goto(invitationPath);
  await expect(page).toHaveURL(/\/invitations\/accept\?tenant=tenant-one&inventory=inventory-one&invitation=invite-one$/);
  await expect(page.getByRole('heading', { name: 'Join Workshop tools' })).toBeVisible();
  expect(acceptCalls).toBe(0);
  await page.getByRole('button', { name: 'Accept invitation' }).click();
  await expect(page.getByRole('heading', { name: 'You joined Workshop tools' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Open inventory' })).toHaveAttribute(
    'href', '/tenants/tenant-one/inventories/inventory-one'
  );
  expect(acceptCalls).toBe(1);
  expect(requestURLs.every((url) => !url.includes(token))).toBe(true);
});

test('signed-out invitation survives two browser OIDC identities and an email mismatch', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'One browser project is sufficient for the identity handoff contract.');
  const providerRequests: string[] = [];
  const tokenRequests: string[] = [];
  let identityNumber = 0;

  await page.route('http://127.0.0.1:5556/dex/auth**', async (route) => {
    const providerURL = new URL(route.request().url());
    providerRequests.push(providerURL.toString());
    identityNumber += 1;
    const state = providerURL.searchParams.get('state');
    expect(state).toBeTruthy();
    await route.fulfill({
      status: 302,
      headers: { location: `http://127.0.0.1:5197/callback?code=identity-${identityNumber}&state=${encodeURIComponent(state ?? '')}` }
    });
  });
  await page.route('http://127.0.0.1:5556/dex/token', async (route) => {
    tokenRequests.push(route.request().postData() ?? '');
    await route.fulfill({
      headers: { 'access-control-allow-origin': '*' },
      json: { id_token: identityNumber === 1 ? 'first-identity-token' : 'invited-identity-token', expires_in: 60 }
    });
  });
  await page.route('http://127.0.0.1:18080/**', async (route) => {
    const request = route.request();
    const authorization = request.headers().authorization;
    if (request.url().endsWith('/preview') && authorization === 'Bearer first-identity-token') {
      await route.fulfill({
        status: 403,
        json: { error: { code: 'invitation_email_mismatch', message: 'Invitation does not match this account.', details: [] }, meta: {} }
      });
      return;
    }
    if (request.url().endsWith('/preview') && authorization === 'Bearer invited-identity-token') {
      await route.fulfill({ json: { data: {
        inventoryId: 'inventory-one', inventoryName: 'Workshop tools', relationship: 'editor',
        status: 'pending', isExpired: false, expiresAt: '2026-07-21T12:00:00Z'
      }, meta: {} } });
      return;
    }
    if (request.url().endsWith('/accept') && authorization === 'Bearer invited-identity-token') {
      await route.fulfill({ json: { data: acceptedResponse('editor'), meta: {} } });
      return;
    }
    await route.abort();
  });

  await page.goto(invitationPath);
  await expect(page).toHaveURL(/\/invitations\/accept\?tenant=tenant-one&inventory=inventory-one&invitation=invite-one$/);
  await page.getByRole('button', { name: 'Continue to sign in' }).click();
  await expect(page.getByRole('heading', { name: 'This invitation is for another account' })).toBeVisible();
  await page.getByRole('button', { name: 'Switch account' }).click();
  await expect(page.getByRole('heading', { name: 'Join Workshop tools' })).toBeVisible();
  await expect(page.getByText('Can edit')).toBeVisible();
  await page.getByRole('button', { name: 'Accept invitation' }).click();
  await expect(page.getByRole('heading', { name: 'You joined Workshop tools' })).toBeVisible();

  expect(identityNumber).toBe(2);
  expect(providerRequests.every((value) => !value.includes(token) && !value.includes('returnTo'))).toBe(true);
  expect(tokenRequests.every((value) => !value.includes(token))).toBe(true);
});

test.describe('invitation terminal states', () => {
  for (const state of [
    { status: 'pending', isExpired: true, heading: 'This invitation expired' },
    { status: 'revoked', isExpired: false, heading: 'This invitation was revoked' },
    { status: 'cancelled', isExpired: false, heading: 'This invitation was cancelled' },
    { status: 'accepted', isExpired: false, heading: 'You already joined Workshop tools' }
  ] as const) {
    test(`shows ${state.status}${state.isExpired ? ' expired' : ''} without accepting`, async ({ page }, testInfo) => {
      test.skip(testInfo.project.name !== 'desktop-chromium', 'One browser project is sufficient for terminal API state mapping.');
      await page.addInitScript(() => {
        sessionStorage.setItem('stuffstash.oidc.session', JSON.stringify({ idToken: 'invitee-token', expiresAt: Date.now() + 60_000 }));
      });
      let acceptCalls = 0;
      await page.route('http://127.0.0.1:18080/**', async (route) => {
        if (route.request().url().endsWith('/accept')) acceptCalls += 1;
        await route.fulfill({ json: { data: {
          inventoryId: 'inventory-one', inventoryName: 'Workshop tools', relationship: 'viewer',
          status: state.status, isExpired: state.isExpired, expiresAt: '2026-07-21T12:00:00Z'
        }, meta: {} } });
      });

      await page.goto(invitationPath);
      await expect(page.getByRole('heading', { name: state.heading })).toBeVisible();
      expect(acceptCalls).toBe(0);
      await expect(page.getByRole('button', { name: 'Accept invitation' })).toHaveCount(0);
    });
  }
});

function acceptedResponse(relationship: 'viewer' | 'editor') {
  return {
    invitation: {
      id: 'invite-one', tenantId: 'tenant-one', inventoryId: 'inventory-one', email: 'invitee@example.test',
      relationship, status: 'accepted', isExpired: false, expiresAt: '2026-07-21T12:00:00Z',
      inviterPrincipalId: 'inviter-one', acceptedPrincipalId: 'invitee-one'
    },
    grant: { tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'invitee-one', relationship }
  };
}
