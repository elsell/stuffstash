<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import WorkspaceTaskSheet from './WorkspaceTaskSheet.svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';

  let { busy = false, closeHref }: { busy?: boolean; closeHref?: string } = $props();
</script>

<WorkspaceTaskSheet open title={t('web.WorkspaceTaskSheetHarness.editAsset')} description={t('web.WorkspaceTaskSheetHarness.changeTheAssetDetails')} {busy} {closeHref}>
  <Label for="task-name">{t('web.WorkspaceTaskSheetHarness.name')}</Label>
  <Input id="task-name" />
  {#snippet footer()}<Button.Root>{t('web.WorkspaceTaskSheetHarness.save')}</Button.Root>{/snippet}
</WorkspaceTaskSheet>
