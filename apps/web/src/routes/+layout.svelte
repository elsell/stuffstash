<script lang="ts">
  import '../styles.css';
  import { Toaster } from '$lib/components/ui/sonner/index.js';

  let { children } = $props();
</script>

<Toaster />
{@render children()}
