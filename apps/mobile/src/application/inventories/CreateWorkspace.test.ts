import { expect, it } from 'vitest';
import { CreateWorkspace, type WorkspaceCreationPort } from './CreateWorkspace';

class WorkspaceFake implements WorkspaceCreationPort {
  permitted = true;
  authenticated = true;
  created: string[] = [];
  async canCreateInventory(tenantId: string) { if (!this.authenticated) throw new Error('Sign in'); return this.permitted && tenantId === 'home'; }
  async createHousehold(name: string) { if (!this.authenticated) throw new Error('Sign in'); this.created.push(name); return { id: 'new-home', name, canCreateInventory: true }; }
  async createInventory(tenantId: string, name: string) { if (!this.authenticated) throw new Error('Sign in'); this.created.push(`${tenantId}:${name}`); return { id: 'new-inventory', tenantId, name }; }
}
it('creates named resources without switching the current inventory', async () => {
  const port = new WorkspaceFake(); const changes: string[] = [];
  const command = new CreateWorkspace(port, { created: () => changes.push('created') });
  await expect(command.household(' Home ')).resolves.toMatchObject({ name: 'Home' });
  await expect(command.inventory('home', ' Garage ')).resolves.toMatchObject({ name: 'Garage' });
  expect(port.created).toEqual(['Home', 'home:Garage']);
  expect(changes).toHaveLength(2);
});
it.each(['denied', 'other-household', 'signed-out', 'blank'])('rejects %s before inventory creation', async reason => {
  const port = new WorkspaceFake();
  port.permitted = reason !== 'denied'; port.authenticated = reason !== 'signed-out';
  const command = new CreateWorkspace(port, { created() {} });
  await expect(command.inventory(reason === 'other-household' ? 'other' : 'home', reason === 'blank' ? ' ' : 'Main')).rejects.toThrow();
  expect(port.created).toEqual([]);
});
