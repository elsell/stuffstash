import { useState } from 'react';
import { Stack } from 'expo-router';
import { Text, View } from 'react-native';
import { InventoryArchiveScreen } from '../src/ui/screens/InventoryArchiveScreen';
import type { ArchiveJob, InventoryArchiveWorkspace } from '../src/application/archives/InventoryArchive';

// Real task controls and lifecycle, controlled repository/picker. This fixture
// establishes layout/approval behavior, not server or native file transfer evidence.
export function InventoryArchiveFixture() {
  const [closed, setClosed] = useState(false);
  const [evidence, setEvidence] = useState('Nothing restored');
  const [workspace] = useState<InventoryArchiveWorkspace>(() => {
    let job: ArchiveJob | undefined;
    const history: ArchiveJob[] = Array.from({ length: 12 }, (_, index) => ({ id: `history-${index}`, kind: index % 2 ? 'restore' : 'export', state: 'cancelled', phase: 'cancelled', createdAt: `2026-09-${String(20 - index).padStart(2, '0')}T00:00:00Z`, expiresAt: '2100-01-01T00:00:00Z', photos: true, otherFiles: true }));
    history.push({ id: 'oldest-export', kind: 'export', state: 'ready', phase: 'complete', createdAt: '2026-09-01T00:00:00Z', expiresAt: '2100-01-01T00:00:00Z', photos: true, otherFiles: true });
    return {
      newRequestKey: () => 'archive-audit-request',
      files: {
        pick: async () => ({ name: 'Household.zip', uri: 'file:///controlled-archive.zip', dispose() {} }),
        share: async () => { throw new Error('Not an export fixture'); }
      },
      repository: {
        list: async scope => {
          if (scope.inventoryId) throw new Error('Activity must use household scope');
          return { jobs: job ? [job, ...history] : history };
        },
        get: async () => { if (!job) throw new Error('Missing job'); return job; },
        create: async () => { throw new Error('Not an export fixture'); },
        upload: async () => {
          job = { id: 'restore-job', kind: 'restore', state: 'awaiting_approval', phase: 'validation', createdAt: '2026-10-02T00:00:00Z', expiresAt: '2026-10-03T00:00:00Z', photos: true, otherFiles: true };
          return job;
        },
        preview: async () => ({ inventoryName: 'Camping equipment', assets: 18, tags: 6, customAssetTypes: 2, customFields: 3, photos: 12, otherFiles: 2, omittedAttachments: 0, keyRemappings: [] }),
        approve: async (_tenant, _id, name) => {
          if (!job) throw new Error('Missing validation');
          setEvidence(`Restored: ${name}`);
          job = { ...job, state: 'ready', destinationInventoryId: 'restored-inventory' }; return job;
        },
        cancel: async () => { if (!job) throw new Error('Missing job'); job = { ...job, state: 'cancelled' }; return job; },
        retry: async () => { throw new Error('No failed job'); },
        download: async () => { throw new Error('Not an export fixture'); }
      }
    };
  });
  return <View style={{ flex: 1 }}>
    <Stack.Screen options={{ title: 'Import and export' }} />
    {closed ? <Text accessibilityLabel={evidence}>{evidence}</Text> : <InventoryArchiveScreen workspace={workspace} scope={{ tenantId: 'household', inventoryId: 'inventory' }} onClose={() => setClosed(true)} onOpen={async () => setClosed(true)} />}
  </View>;
}
