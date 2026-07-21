import { describe, expect, it } from 'vitest';
import { personalWorkspaceNames, provisionPersonalWorkspace } from './personalWorkspaceProvisioning';

describe('personal workspace provisioning', () => {
  it('uses the normalized OIDC display name for the household', () => {
    expect(personalWorkspaceNames({ id: 'principal-one', displayName: '  Alex Rivera  ' })).toEqual({
      tenantName: 'Alex Rivera\u2019s household',
      inventoryName: 'Home'
    });
  });

  it('uses a calm fallback when no safe display name is available', () => {
    expect(personalWorkspaceNames({ id: 'principal-one', email: 'alex@example.test' })).toEqual({
      tenantName: 'My household',
      inventoryName: 'Home'
    });
  });

  it('serializes browser-tab provisioning and delegates the retry recheck to the repository', async () => {
    const calls: string[] = [];
    const repository = {
      provisionPersonalWorkspace: async (names: { tenantName: string; inventoryName: string }) => {
        calls.push(`${names.tenantName}:${names.inventoryName}`);
        return { context: { tenants: [] } };
      }
    } as any;
    const lock = {
      request: async <T>(name: string, callback: () => Promise<T>) => {
        calls.push(name);
        return callback();
      }
    };

    await provisionPersonalWorkspace(repository, { id: 'principal-one', displayName: 'Alex' }, lock);

    expect(calls).toEqual(['stuffstash.personal-workspace-provisioning', 'Alex’s household:Home']);
  });
});
