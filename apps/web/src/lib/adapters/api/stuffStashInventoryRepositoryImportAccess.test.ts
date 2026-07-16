import { describe, expect, it } from 'vitest';
import { StuffStashInventoryRepository } from './stuffStashInventoryRepository';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
import { config, fakeFetch } from './stuffStashInventoryRepository.test-helpers';

describe('StuffStashInventoryRepository imports, access, and audit', () => {
  it('normalizes live Homebox import requests before sending them through the generated client', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await repository.previewImportJob('tenant-home', 'inventory-household', {
      sourceType: 'legacy_homebox',
      baseUrl: 'stuff.jsksell.com',
      username: ' codex@jsksell.com ',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    });

    const previewRequest = requests.find((request) => request.method === 'POST' && request.url.includes('/imports/jobs/preview'));
    expect(previewRequest).toBeTruthy();
    expect(await previewRequest?.json()).toMatchObject({
      sourceType: 'legacy_homebox',
      baseUrl: 'https://stuff.jsksell.com',
      username: 'codex@jsksell.com'
    });
  });

  it('preserves an explicitly entered http Homebox import URL', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const job = await repository.previewImportJob('tenant-home', 'inventory-household', {
      sourceType: 'legacy_homebox',
      baseUrl: 'http://homebox.local:3100',
      username: 'codex@jsksell.com',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: true,
      allowInsecureTLS: false
    });

    const previewRequest = requests.find((request) => request.method === 'POST' && request.url.includes('/imports/jobs/preview'));
    expect(previewRequest).toBeTruthy();
    expect(await previewRequest?.json()).toMatchObject({
      sourceType: 'legacy_homebox',
      baseUrl: 'http://homebox.local:3100'
    });
    expect(job.source).toMatchObject({
      allowPrivateNetwork: true,
      allowInsecureTLS: false
    });
  });

  it('records safe import job observability events at the API adapter boundary', async () => {
    const { fetch } = fakeFetch();
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);
    const source = {
      sourceType: 'legacy_homebox' as const,
      baseUrl: 'stuff.jsksell.com',
      username: 'codex@jsksell.com',
      password: 'super-secret-password',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    };

    await repository.listImportJobs('tenant-home', 'inventory-household');
    const previewed = await repository.previewImportJob('tenant-home', 'inventory-household', source);
    await repository.startImportJob('tenant-home', 'inventory-household', previewed.id, source);
    await repository.cancelImportJob('tenant-home', 'inventory-household', previewed.id, 'discard_partial_progress');
    await repository.removeImportJobFromHistory('tenant-home', 'inventory-household', previewed.id);

    expect(observer.events).toEqual([
      { eventName: 'workspace.import_jobs_load_started', attributes: {} },
      { eventName: 'workspace.import_jobs_loaded', attributes: { jobCount: 1 } },
      { eventName: 'workspace.import_job_preview_started', attributes: { sourceType: 'legacy_homebox' } },
      {
        eventName: 'workspace.import_job_preview_completed',
        attributes: { sourceType: 'legacy_homebox', assetCount: 0, warningCount: 0, errorCount: 0 }
      },
      { eventName: 'workspace.import_job_start_started', attributes: { sourceType: 'legacy_homebox' } },
      { eventName: 'workspace.import_job_started', attributes: { sourceType: 'legacy_homebox', jobId: 'import-job-one' } },
      {
        eventName: 'workspace.import_job_cancel_started',
        attributes: { mode: 'discard_partial_progress', jobId: 'import-job-one' }
      },
      {
        eventName: 'workspace.import_job_cancel_requested',
        attributes: { mode: 'discard_partial_progress', jobId: 'import-job-one' }
      },
      { eventName: 'workspace.import_job_history_remove_started', attributes: { jobId: 'import-job-one' } },
      { eventName: 'workspace.import_job_history_removed', attributes: { jobId: 'import-job-one' } }
    ]);
    for (const event of observer.events) {
      expect(JSON.stringify(event.attributes)).not.toContain('super-secret-password');
      expect(JSON.stringify(event.attributes)).not.toContain('codex@jsksell.com');
      expect(JSON.stringify(event.attributes)).not.toContain('stuff.jsksell.com');
    }
  });

  it('sends safe request correlation IDs for import mutations', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const source = {
      sourceType: 'legacy_homebox' as const,
      baseUrl: 'stuff.jsksell.com',
      username: 'codex@jsksell.com',
      password: 'super-secret-password',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    };

    const previewed = await repository.previewImportJob('tenant-home', 'inventory-household', source);
    await repository.startImportJob('tenant-home', 'inventory-household', previewed.id, source);
    await repository.cancelImportJob('tenant-home', 'inventory-household', previewed.id, 'discard_partial_progress');
    await repository.removeImportJobFromHistory('tenant-home', 'inventory-household', previewed.id);

    const mutationRequests = requests.filter((request) => request.method === 'POST' || request.method === 'DELETE');
    expect(mutationRequests.map((request) => new URL(request.url).pathname)).toEqual([
      '/tenants/tenant-home/inventories/inventory-household/imports/jobs/preview',
      '/tenants/tenant-home/inventories/inventory-household/imports/jobs/import-job-one/start',
      '/tenants/tenant-home/inventories/inventory-household/imports/jobs/import-job-one/cancel',
      '/tenants/tenant-home/inventories/inventory-household/imports/jobs/import-job-one'
    ]);
    for (const request of mutationRequests) {
      const requestId = request.headers.get('X-Request-ID') ?? '';
      expect(requestId).toMatch(/^web-import-(preview|start|cancel|remove)-/);
      expect(requestId.length).toBeLessThanOrEqual(128);
      expect(requestId).not.toContain('super-secret-password');
      expect(requestId).not.toContain('codex@jsksell.com');
      expect(requestId).not.toContain('stuff.jsksell.com');
    }
  });

  it('does not include raw CSV source content in import observability events', async () => {
    const { fetch } = fakeFetch();
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);
    const source = {
      sourceType: 'legacy_homebox_csv' as const,
      fileName: 'homebox-secret-export.csv',
      contentBase64: 'U0VDUkVUX0NTVl9DT05URU5U'
    };

    const previewed = await repository.previewImportJob('tenant-home', 'inventory-household', source);
    await repository.startImportJob('tenant-home', 'inventory-household', previewed.id, source);

    expect(observer.events).toEqual([
      { eventName: 'workspace.import_job_preview_started', attributes: { sourceType: 'legacy_homebox_csv' } },
      {
        eventName: 'workspace.import_job_preview_completed',
        attributes: { sourceType: 'legacy_homebox_csv', assetCount: 0, warningCount: 0, errorCount: 0 }
      },
      { eventName: 'workspace.import_job_start_started', attributes: { sourceType: 'legacy_homebox_csv' } },
      { eventName: 'workspace.import_job_started', attributes: { sourceType: 'legacy_homebox_csv', jobId: 'import-job-one' } }
    ]);
    for (const event of observer.events) {
      const attributes = JSON.stringify(event.attributes);
      expect(attributes).not.toContain('U0VDUkVUX0NTVl9DT05URU5U');
      expect(attributes).not.toContain('SECRET_CSV_CONTENT');
      expect(attributes).not.toContain('homebox-secret-export.csv');
    }
  });

  it('records safe import job failure observability events at the API adapter boundary', async () => {
    const { fetch } = fakeFetch({ failedImportOperations: ['list', 'preview', 'start', 'cancel', 'remove'] });
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);
    const source = {
      sourceType: 'legacy_homebox' as const,
      baseUrl: 'stuff.jsksell.com',
      username: 'codex@jsksell.com',
      password: 'super-secret-password',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    };

    await expect(repository.listImportJobs('tenant-home', 'inventory-household')).rejects.toThrow();
    await expect(repository.previewImportJob('tenant-home', 'inventory-household', source)).rejects.toThrow();
    await expect(repository.startImportJob('tenant-home', 'inventory-household', 'import-job-one', source)).rejects.toThrow();
    await expect(
      repository.cancelImportJob('tenant-home', 'inventory-household', 'import-job-one', 'keep_partial_progress')
    ).rejects.toThrow();
    await expect(repository.removeImportJobFromHistory('tenant-home', 'inventory-household', 'import-job-one')).rejects.toThrow();

    expect(observer.events.map((event) => event.eventName)).toEqual([
      'workspace.import_jobs_load_started',
      'workspace.import_jobs_load_failed',
      'workspace.import_job_preview_started',
      'workspace.import_job_preview_failed',
      'workspace.import_job_start_started',
      'workspace.import_job_start_failed',
      'workspace.import_job_cancel_started',
      'workspace.import_job_cancel_failed',
      'workspace.import_job_history_remove_started',
      'workspace.import_job_history_remove_failed'
    ]);
    for (const event of observer.events) {
      expect(JSON.stringify(event.attributes)).not.toContain('super-secret-password');
      expect(JSON.stringify(event.attributes)).not.toContain('codex@jsksell.com');
      expect(JSON.stringify(event.attributes)).not.toContain('stuff.jsksell.com');
      expect(JSON.stringify(event.attributes)).not.toContain('provider-stacktrace');
    }
  });

  it('manages access grants through generated client-backed repository methods', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(repository.listInventoryAccessGrants('tenant-home', 'inventory-household')).resolves.toMatchObject({
      items: [
        {
          tenantId: 'tenant-home',
          inventoryId: 'inventory-household',
          principalId: 'principal-two',
          relationship: 'viewer'
        }
      ],
      pagination: { limit: 50, nextCursor: null, hasMore: false }
    });
    await expect(
      repository.grantInventoryAccess('tenant-home', 'inventory-household', 'principal-three', 'editor')
    ).resolves.toMatchObject({ principalId: 'principal-three', relationship: 'editor' });
    await expect(
      repository.revokeInventoryAccess('tenant-home', 'inventory-household', 'principal-two', 'viewer')
    ).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/access-grants?limit=50',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/access-grants',
      'DELETE http://api.local/tenants/tenant-home/inventories/inventory-household/access-grants/principal-two/viewer'
    ]);
    expect(await requests[1]?.json()).toEqual({ principalId: 'principal-three', relationship: 'editor' });
  });

  it('manages access invitations through generated client-backed repository methods', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch, 'https://stash.example.test');

    await expect(repository.listInventoryAccessInvitations('tenant-home', 'inventory-household', 'pending')).resolves.toMatchObject({
      items: [expect.objectContaining({ id: 'invite-one', email: 'friend@example.test', relationship: 'viewer' })],
      pagination: { limit: 50, nextCursor: null, hasMore: false }
    });
    await expect(
      repository.createInventoryAccessInvitation('tenant-home', 'inventory-household', 'editor@example.test', 'editor')
    ).resolves.toMatchObject({
      invitation: { email: 'editor@example.test', relationship: 'editor' },
      inviteUrl: 'https://stash.example.test/invitations/accept?tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
    });
    await expect(
      repository.updateInventoryAccessInvitationExpiration(
        'tenant-home',
        'inventory-household',
        'invite-one',
        '2026-07-01T00:00:00Z'
      )
    ).resolves.toMatchObject({ id: 'invite-one', expiresAt: '2026-07-01T00:00:00Z' });
    await expect(repository.cancelInventoryAccessInvitation('tenant-home', 'inventory-household', 'invite-one')).resolves.toBeUndefined();
    await expect(repository.deleteInventoryAccessInvitation('tenant-home', 'inventory-household', 'invite-one')).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations?limit=50&status=pending',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations',
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one/expiration',
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one/cancel',
      'DELETE http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one'
    ]);
    expect(await requests[1]?.json()).toEqual({ email: 'editor@example.test', relationship: 'editor' });
    expect(await requests[2]?.json()).toEqual({ expiresAt: '2026-07-01T00:00:00Z' });
  });

  it.each([
    ['missing URL', { inviteUrl: undefined }],
    ['remote HTTP URL', { inviteUrl: 'http://stash.example.test/invitations/accept?tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA' }],
    ['untrusted HTTPS origin', { inviteUrl: 'https://phish.example.test/invitations/accept?tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA' }],
    ['wrong URL scope', { inviteUrl: 'https://stash.example.test/invitations/accept?tenant=tenant-other&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA' }],
    ['duplicate URL field', { inviteUrl: 'https://stash.example.test/invitations/accept?tenant=tenant-home&tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA' }],
    ['short token', { inviteUrl: 'https://stash.example.test/invitations/accept?tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=short' }],
    ['response scope', { tenantId: 'tenant-other' }]
  ])('rejects a created invitation with %s', async (_label, createdInvitationOverride) => {
    const { fetch } = fakeFetch({ createdInvitationOverride });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch, 'https://stash.example.test');
    await expect(
      repository.createInventoryAccessInvitation('tenant-home', 'inventory-household', 'editor@example.test', 'editor')
    ).rejects.toThrow('invalid invitation link');
  });

  it('rejects a matching private HTTP origin when the explicit switch is off', async () => {
    const origin = 'http://192.168.1.117:5173';
    const inviteUrl = `${origin}/invitations/accept?tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`;
    const { fetch } = fakeFetch({ createdInvitationOverride: { inviteUrl } });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch, origin);
    await expect(
      repository.createInventoryAccessInvitation('tenant-home', 'inventory-household', 'editor@example.test', 'editor')
    ).rejects.toThrow('invalid invitation link');
  });

  it.each([
    'http://localhost:5173',
    'http://127.0.0.1:5173',
    'http://[::1]:5173',
    'http://10.0.0.7:5173',
    'http://172.16.0.1:5173',
    'http://172.31.255.254:5173',
    'http://192.168.1.117:5173'
  ])('accepts an explicitly trusted local invitation origin: %s', async (origin) => {
    const inviteUrl = `${origin}/invitations/accept?tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`;
    const { fetch } = fakeFetch({ createdInvitationOverride: { inviteUrl } });
    const repository = new StuffStashInventoryRepository(
      { ...config, invitationAllowInsecureLocalHTTP: true },
      () => 'id-token',
      new InMemoryWorkspaceObserver(),
      fetch,
      origin
    );
    await expect(
      repository.createInventoryAccessInvitation('tenant-home', 'inventory-household', 'editor@example.test', 'editor')
    ).resolves.toMatchObject({ inviteUrl });
  });

  it.each([
    'http://8.8.8.8:5173',
    'http://172.32.0.1:5173',
    'http://stash.example.test'
  ])('rejects public HTTP even when the local switch is enabled: %s', async (origin) => {
    const inviteUrl = `${origin}/invitations/accept?tenant=tenant-home&inventory=inventory-household&invitation=invite-created#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`;
    const { fetch } = fakeFetch({ createdInvitationOverride: { inviteUrl } });
    const repository = new StuffStashInventoryRepository(
      { ...config, invitationAllowInsecureLocalHTTP: true },
      () => 'id-token',
      new InMemoryWorkspaceObserver(),
      fetch,
      origin
    );
    await expect(
      repository.createInventoryAccessInvitation('tenant-home', 'inventory-household', 'editor@example.test', 'editor')
    ).rejects.toThrow('invalid invitation link');
  });

  it('lists tenant and inventory audit records through generated client paths', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(repository.listTenantAuditRecords('tenant-home')).resolves.toMatchObject({
      items: [{ id: 'audit-one', action: 'asset.created', inventoryId: 'inventory-household' }],
      pagination: { limit: 50, nextCursor: null, hasMore: false }
    });
    await expect(repository.listInventoryAuditRecords('tenant-home', 'inventory-household', 'next-page')).resolves.toMatchObject({
      items: [{ id: 'audit-one', action: 'asset.created', inventoryId: 'inventory-household' }]
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/audit-records?limit=50',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/audit-records?limit=50&cursor=next-page'
    ]);
  });
});
