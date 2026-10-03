<script lang="ts">
  import {page} from '$app/state';
  import WorkspaceRoot from '../+page.svelte';
  import LabelLanding from '$lib/components/labels/LabelLanding.svelte';
  const labelPath=$derived(/\/l\/v[0-9]+\//.test(page.url.pathname));
</script>
{#if labelPath}<LabelLanding/>{:else}<WorkspaceRoot/>{/if}
