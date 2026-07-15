import { describe, expect, it } from 'vitest';
import { InvitationFailure } from '$lib/domain/invitation';
import { StuffStashInventoryInvitationRepository } from './stuffStashInventoryInvitationRepository';

const material = {
  tenantId: 'tenant-one',
  inventoryId: 'inventory-one',
  invitationId: 'invite-one',
  token: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
};

describe('StuffStashInventoryInvitationRepository', () => {
  it('previews and accepts through generated client paths', async () => {
    const requests: Request[] = [];
    const repository = new StuffStashInventoryInvitationRepository('https://api.example.test', () => 'id-token', async (input, init) => {
      const request = new Request(input, init);
      requests.push(request);
      if (request.url.endsWith('/preview')) {
        return Response.json({
          data: {
            inventoryId: 'inventory-one',
            inventoryName: 'Workshop tools',
            relationship: 'viewer',
            status: 'pending',
            isExpired: false,
            expiresAt: '2026-07-21T12:00:00Z'
          },
          meta: {}
        });
      }
      return Response.json({
        data: {
          invitation: {
            id: 'invite-one', tenantId: 'tenant-one', inventoryId: 'inventory-one', email: 'person@example.test',
            relationship: 'viewer', status: 'accepted', isExpired: false, expiresAt: '2026-07-21T12:00:00Z',
            inviterPrincipalId: 'owner', acceptedPrincipalId: 'invitee'
          },
          grant: { tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'invitee', relationship: 'viewer' }
        },
        meta: {}
      });
    });

    await expect(repository.preview(material)).resolves.toMatchObject({ inventoryName: 'Workshop tools', status: 'pending' });
    await expect(repository.accept(material)).resolves.toEqual({ tenantId: 'tenant-one', inventoryId: 'inventory-one', status: 'accepted' });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST https://api.example.test/tenants/tenant-one/inventories/inventory-one/access-invitations/invite-one/preview',
      'POST https://api.example.test/tenants/tenant-one/inventories/inventory-one/access-invitations/invite-one/accept'
    ]);
    expect(await requests[0]?.json()).toEqual({ acceptanceToken: material.token });
    expect(await requests[1]?.json()).toEqual({ acceptanceToken: material.token });
  });

  it.each([
    { status: 401, code: 'authentication_required', kind: 'authentication_required' },
    { status: 403, code: 'invitation_email_mismatch', kind: 'email_mismatch' },
    { status: 404, code: 'invitation_invalid', kind: 'invalid' },
    { status: 500, code: 'internal_error', kind: 'unavailable' }
  ] as const)('maps $code without exposing transport messages', async ({ status, code, kind }) => {
    const repository = new StuffStashInventoryInvitationRepository('https://api.example.test', () => 'id-token', async () => Response.json({
      error: { code, message: `unsafe ${material.token}`, details: [] }, meta: {}
    }, { status }));

    const error = await repository.preview(material).catch((caught) => caught);
    expect(error).toBeInstanceOf(InvitationFailure);
    expect((error as InvitationFailure).kind).toBe(kind);
    expect((error as Error).message).not.toContain(material.token);
  });

  it('rejects preview data for a different inventory', async () => {
    const repository = new StuffStashInventoryInvitationRepository('https://api.example.test', () => 'id-token', async () => Response.json({
      data: { inventoryId: 'inventory-other', inventoryName: 'Other', relationship: 'viewer', status: 'pending', isExpired: false, expiresAt: '2026-07-21T12:00:00Z' },
      meta: {}
    }));
    await expect(repository.preview(material)).rejects.toMatchObject({ kind: 'invalid' });
  });

  it.each([
    ['invitation id', { invitation: { id: 'invite-other' } }],
    ['invitation scope', { invitation: { tenantId: 'tenant-other' } }],
    ['grant scope', { grant: { inventoryId: 'inventory-other' } }],
    ['relationship', { grant: { relationship: 'editor' } }],
    ['status', { invitation: { status: 'pending' } }]
  ] as const)('rejects an accepted response with mismatched %s', async (_label, override) => {
    const invitation = {
      id: 'invite-one', tenantId: 'tenant-one', inventoryId: 'inventory-one', email: 'person@example.test',
      relationship: 'viewer', status: 'accepted', isExpired: false, expiresAt: '2026-07-21T12:00:00Z',
      inviterPrincipalId: 'owner', acceptedPrincipalId: 'invitee', ...('invitation' in override ? override.invitation : {})
    };
    const grant = {
      tenantId: 'tenant-one', inventoryId: 'inventory-one', principalId: 'invitee', relationship: 'viewer', ...('grant' in override ? override.grant : {})
    };
    const repository = new StuffStashInventoryInvitationRepository('https://api.example.test', () => 'id-token', async () => Response.json({ data: { invitation, grant }, meta: {} }));
    await expect(repository.accept(material)).rejects.toMatchObject({ kind: 'invalid' });
  });
});
