<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import AddAssetTray from './AddAssetTray.svelte';

  let open = $state(false);
</script>

<Button.Root id="open-add" onclick={() => { open = true; }}>Open add</Button.Root>

<AddAssetTray
  {open}
  closeHref="/tenants/tenant-home/inventories/inventory-household"
  parentTargets={[]}
  mediaPolicy={{ supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 }}
  customAssetTypes={[]}
  customFieldDefinitions={[]}
  saving={false}
  onClose={() => { open = false; }}
  onSave={async () => ({ saved: true })}
/>
