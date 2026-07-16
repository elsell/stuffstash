<script lang="ts">
  import PopoverPortal from './popover/popover-portal.svelte';
  import SheetPortal from './sheet/sheet-portal.svelte';
</script>

<PopoverPortal disabled>
  <span data-testid="popover-portal-child">Popover content</span>
</PopoverPortal>

<SheetPortal disabled>
  <span data-testid="sheet-portal-child">Sheet content</span>
</SheetPortal>
