import { describe, expect, it } from 'vitest';
import {
  parseCreatedInventoryInvitationLink,
  parseInventoryInvitationLink
} from './InvitationLinkParser';

const token = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';
const appLink =
  `stuffstash://invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one#token=${token}`;
const webLink =
  `https://stash.example.test/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one#token=${token}`;

describe('parseInventoryInvitationLink', () => {
  it('parses the app scheme and configured HTTPS origin without retaining the source URL', () => {
    expect(parseInventoryInvitationLink(appLink, 'https://stash.example.test')).toEqual({
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      invitationId: 'invite-one',
      acceptanceToken: token
    });
    expect(parseInventoryInvitationLink(webLink, 'https://stash.example.test')).toEqual({
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      invitationId: 'invite-one',
      acceptanceToken: token
    });
  });

  it('accepts an exact private LAN HTTP origin only with explicit local-development opt-in', () => {
    const lanLink = webLink.replace('https://stash.example.test', 'http://192.168.1.117:5173');
    expect(parseInventoryInvitationLink(lanLink, 'http://192.168.1.117:5173', true)).toMatchObject({
      invitationId: 'invite-one'
    });
    expect(() => parseInventoryInvitationLink(lanLink, 'http://192.168.1.117:5173')).toThrow(
      'This invitation link is invalid.'
    );
    expect(() => parseInventoryInvitationLink(
      lanLink.replace('192.168.1.117', '8.8.8.8'),
      'http://8.8.8.8:5173',
      true
    )).toThrow('This invitation link is invalid.');
  });

  it.each([
    ['unconfigured HTTPS origin', webLink, 'https://other.example.test'],
    ['insecure public link', webLink.replace('https:', 'http:'), 'https://stash.example.test'],
    ['wrong app route', appLink.replace('/accept?', '/other?'), 'https://stash.example.test'],
    ['wrong HTTPS route', webLink.replace('/invitations/accept?', '/other?'), 'https://stash.example.test'],
    ['username in public link', webLink.replace('https://', 'https://user@'), 'https://stash.example.test'],
    ['missing tenant', appLink.replace('tenant=tenant-one&', ''), 'https://stash.example.test'],
    ['missing inventory', appLink.replace('inventory=inventory-one&', ''), 'https://stash.example.test'],
    ['missing invitation', appLink.replace('invitation=invite-one', ''), 'https://stash.example.test'],
    ['missing token', appLink.replace(`#token=${token}`, ''), 'https://stash.example.test'],
    ['duplicate tenant', appLink.replace('tenant=tenant-one', 'tenant=tenant-one&tenant=tenant-two'), 'https://stash.example.test'],
    ['duplicate token', `${appLink}&token=other-token`, 'https://stash.example.test'],
    ['token in query', appLink.replace('&invitation=', `&token=${token}&invitation=`), 'https://stash.example.test'],
    ['unknown query field', appLink.replace('&invitation=', '&campaign=summer&invitation='), 'https://stash.example.test'],
    ['unknown fragment field', `${appLink}&campaign=summer`, 'https://stash.example.test'],
    ['whitespace in identifier', appLink.replace('tenant-one', 'tenant%20one'), 'https://stash.example.test'],
    ['control character in token', appLink.replace(token, 'raw-token%0A'), 'https://stash.example.test'],
    ['short token', appLink.replace(token, 'raw-token_123'), 'https://stash.example.test'],
    ['literal newline normalized by URL parser', `${appLink}\n`, 'https://stash.example.test'],
    ['literal tab normalized by URL parser', `\t${appLink}`, 'https://stash.example.test'],
    ['surrounding spaces', ` ${appLink} `, 'https://stash.example.test'],
    ['oversized identifier', appLink.replace('tenant-one', 't'.repeat(201)), 'https://stash.example.test'],
    ['oversized token', appLink.replace(token, 't'.repeat(44)), 'https://stash.example.test'],
    ['oversized URL', `${appLink}${'x'.repeat(4097)}`, 'https://stash.example.test']
  ])('fails closed for %s', (_label, link, configuredOrigin) => {
    expect(() => parseInventoryInvitationLink(link, configuredOrigin)).toThrow(
      'This invitation link is invalid.'
    );
  });

  it('does not put invitation material in parser error messages', () => {
    let message = '';
    try {
      parseInventoryInvitationLink(appLink.replace('/accept?', '/wrong?'), 'https://stash.example.test');
    } catch (error) {
      message = error instanceof Error ? error.message : String(error);
    }

    expect(message).toBe('This invitation link is invalid.');
    expect(message).not.toContain(token);
    expect(message).not.toContain('tenant-one');
    expect(message).not.toContain(appLink);
  });
});

describe('parseCreatedInventoryInvitationLink', () => {
  it('accepts canonical HTTPS and explicitly enabled unconfigured loopback HTTP creation responses', () => {
    expect(parseCreatedInventoryInvitationLink(webLink, 'https://stash.example.test')).toMatchObject({
      tenantId: 'tenant-one', inventoryId: 'inventory-one', invitationId: 'invite-one'
    });
    expect(parseCreatedInventoryInvitationLink(webLink.replace(
      'https://stash.example.test',
      'http://127.0.0.1:8081'
    ), undefined, true)).toMatchObject({ invitationId: 'invite-one' });
    expect(() => parseCreatedInventoryInvitationLink(webLink.replace(
      'https://stash.example.test',
      'http://127.0.0.1:8081'
    ), 'https://stash.example.test', true)).toThrow('This invitation link is invalid.');
  });

  it('accepts an exact private LAN HTTP creation response only with explicit opt-in', () => {
    const lanLink = webLink.replace('https://stash.example.test', 'http://192.168.1.117:5173');
    expect(parseCreatedInventoryInvitationLink(
      lanLink,
      'http://192.168.1.117:5173',
      true
    )).toMatchObject({ invitationId: 'invite-one' });
    expect(() => parseCreatedInventoryInvitationLink(
      lanLink,
      'http://192.168.1.117:5173'
    )).toThrow('This invitation link is invalid.');
  });

  it.each([
    ['non-loopback HTTP', webLink.replace('https:', 'http:')],
    ['custom scheme', appLink],
    ['unknown query field', webLink.replace('&inventory=', '&campaign=x&inventory=')],
    ['duplicate token', `${webLink}&token=${token}`],
    ['short token', webLink.replace(token, 'short')]
  ])('rejects %s', (_label, link) => {
    expect(() => parseCreatedInventoryInvitationLink(link, 'https://stash.example.test')).toThrow('This invitation link is invalid.');
  });

  it('rejects an arbitrary HTTPS origin and fails closed when no trusted origin is configured', () => {
    expect(() => parseCreatedInventoryInvitationLink(webLink, 'https://other.example.test')).toThrow(
      'This invitation link is invalid.'
    );
    expect(() => parseCreatedInventoryInvitationLink(webLink)).toThrow('This invitation link is invalid.');
  });
});
