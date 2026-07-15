<script lang="ts">
  import { Dialog } from 'bits-ui';
</script>

<Dialog.Root open>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content>
      <Dialog.Title>Lifecycle test</Dialog.Title>
      <Dialog.Description>Exercise the pinned primitive's real scroll-lock teardown.</Dialog.Description>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
