<script lang="ts">
  import '../styles.css';
  import { onMount } from 'svelte';
  import { localization } from '$lib/presentation/localization';

  onMount(() => {
    document.documentElement.lang = localization.messageLanguage;
    document.documentElement.dir = localization.direction;
  });
  import { Toaster } from '$lib/components/ui/sonner/index.js';

  let { children } = $props();
</script>

<Toaster />
{@render children()}
